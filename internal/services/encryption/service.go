package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"sync"

	"golang.org/x/crypto/pbkdf2"
)

type EncryptionService struct {
	key       []byte
	mu        sync.RWMutex
	algorithm string
}

var (
	ErrInvalidKey       = errors.New("invalid encryption key")
	ErrEncryptionFailed = errors.New("encryption failed")
	ErrDecryptionFailed = errors.New("decryption failed")
	ErrInvalidData      = errors.New("invalid encrypted data")
)

const (
	KeySize    = 32
	SaltSize   = 32
	NonceSize  = 12
	Iterations = 100000
)

func NewEncryptionService(password string, salt []byte) (*EncryptionService, error) {
	if len(password) < 8 {
		return nil, ErrInvalidKey
	}

	if salt == nil {
		salt = make([]byte, SaltSize)
		if _, err := io.ReadFull(rand.Reader, salt); err != nil {
			return nil, fmt.Errorf("failed to generate salt: %w", err)
		}
	}

	key := pbkdf2.Key([]byte(password), salt, Iterations, KeySize, sha256.New)

	return &EncryptionService{
		key:       key,
		algorithm: "AES-256-GCM",
	}, nil
}

func NewEncryptionServiceWithKey(key []byte) (*EncryptionService, error) {
	if len(key) < 32 {
		return nil, ErrInvalidKey
	}

	return &EncryptionService{
		key:       key[:32],
		algorithm: "AES-256-GCM",
	}, nil
}

func (s *EncryptionService) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}

	nonce := make([]byte, NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (s *EncryptionService) Decrypt(encrypted string) (string, error) {
	if encrypted == "" {
		return "", nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("%w: invalid base64", ErrDecryptionFailed)
	}

	if len(ciphertext) < NonceSize {
		return "", ErrInvalidData
	}

	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	nonce := ciphertext[:NonceSize]
	ciphertext = ciphertext[NonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	return string(plaintext), nil
}

func (s *EncryptionService) Hash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (s *EncryptionService) HashWithSalt(data, salt string) string {
	hash := sha256.Sum256([]byte(data + salt))
	return hex.EncodeToString(hash[:])
}

func (s *EncryptionService) EncryptStruct(data interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			str, ok := value.(string)
			if ok {
				encrypted, err := s.Encrypt(str)
				if err != nil {
					result[key] = value
				} else {
					result[key] = encrypted
				}
			} else {
				result[key] = value
			}
		}
	case map[string]string:
		for key, value := range v {
			encrypted, err := s.Encrypt(value)
			if err != nil {
				result[key] = value
			} else {
				result[key] = encrypted
			}
		}
	default:
		return nil, fmt.Errorf("unsupported type")
	}

	return result, nil
}

func (s *EncryptionService) DecryptStruct(data map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for key, value := range data {
		str, ok := value.(string)
		if ok && len(str) > 24 {
			decrypted, err := s.Decrypt(str)
			if err != nil {
				result[key] = value
			} else {
				result[key] = decrypted
			}
		} else {
			result[key] = value
		}
	}

	return result, nil
}

type FieldEncryptor struct {
	encryption *EncryptionService
	fields     map[string]bool
}

func NewFieldEncryptor(encryption *EncryptionService, fields []string) *FieldEncryptor {
	fieldMap := make(map[string]bool)
	for _, f := range fields {
		fieldMap[f] = true
	}

	return &FieldEncryptor{
		encryption: encryption,
		fields:     fieldMap,
	}
}

func (f *FieldEncryptor) EncryptFields(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		if f.fields[key] {
			if str, ok := value.(string); ok {
				encrypted, err := f.encryption.Encrypt(str)
				if err != nil {
					result[key] = value
				} else {
					result[key] = encrypted
				}
			} else {
				result[key] = value
			}
		} else {
			result[key] = value
		}
	}

	return result
}

func (f *FieldEncryptor) DecryptFields(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		if f.fields[key] {
			if str, ok := value.(string); ok {
				decrypted, err := f.encryption.Decrypt(str)
				if err != nil {
					result[key] = value
				} else {
					result[key] = decrypted
				}
			} else {
				result[key] = value
			}
		} else {
			result[key] = value
		}
	}

	return result
}

func GenerateSecureToken(length int) string {
	if length < 16 {
		length = 32
	}

	bytes := make([]byte, length)
	io.ReadFull(rand.Reader, bytes)
	return hex.EncodeToString(bytes)[:length]
}

func GenerateAPIKey() string {
	return GenerateSecureToken(32)
}
