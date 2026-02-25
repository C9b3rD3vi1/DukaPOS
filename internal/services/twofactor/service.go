package twofactor

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	mathrand "math/rand"
	"sync"
	"time"
)

type Service struct {
	otpService interface{}
}

type TwoFactorConfig struct {
	Issuer          string
	Algorithm       string
	Digits          int
	Period          int
	RequireForLogin bool
	RequireForStaff bool
}

type TwoFactorSecret struct {
	AccountName string
	Issuer      string
	Secret      string
	QRCodeURL   string
	CreatedAt   time.Time
}

func NewService() *Service {
	return &Service{}
}

func GenerateSecret(accountName, issuer string) (*TwoFactorSecret, error) {
	secretBytes := make([]byte, 20)
	if _, err := rand.Read(secretBytes); err != nil {
		return nil, fmt.Errorf("failed to generate secret: %w", err)
	}

	secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secretBytes)

	accountName = sanitizeAccountName(accountName)
	issuer = sanitizeIssuer(issuer)

	otpAuthURL := fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA256&digits=6&period=30",
		issuer, accountName, secret, issuer)

	return &TwoFactorSecret{
		AccountName: accountName,
		Issuer:      issuer,
		Secret:      secret,
		QRCodeURL:   otpAuthURL,
		CreatedAt:   time.Now(),
	}, nil
}

func (s *Service) EnableTwoFactor(accountID uint, accountName, phone string) (*TwoFactorSecret, error) {
	secret, err := GenerateSecret(accountName, "DukaPOS")
	if err != nil {
		return nil, err
	}

	tempCode := GenerateBackupCodes(8)
	_ = tempCode

	return secret, nil
}

func (s *Service) VerifyToken(secret, token string) bool {
	secret = sanitizeSecret(secret)

	upperToken := toUpper(token)

	for offset := -1; offset <= 1; offset++ {
		expectedToken := generateTOTP(secret, time.Now().Add(time.Duration(offset)*30*time.Second))
		if constTimeEqual(expectedToken, upperToken) {
			return true
		}
	}

	return false
}

func GenerateBackupCodes(count int) []string {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		codes[i] = generateBackupCode()
	}
	return codes
}

func HashBackupCodes(codes []string) []string {
	hashed := make([]string, len(codes))
	for i, code := range codes {
		hash := sha256.Sum256([]byte(code))
		hashed[i] = base32.StdEncoding.EncodeToString(hash[:])
	}
	return hashed
}

func VerifyBackupCode(hashedCodes []string, code string) bool {
	code = cleanBackupCode(code)
	hash := sha256.Sum256([]byte(code))
	hashStr := base32.StdEncoding.EncodeToString(hash[:])

	for _, hashed := range hashedCodes {
		if constTimeEqual(hashed, hashStr) {
			return true
		}
	}
	return false
}

func generateTOTP(secret string, timestamp time.Time) string {
	counter := uint64(timestamp.Unix() / 30)
	return generateHOTP(secret, counter)
}

func generateHOTP(secret string, counter uint64) string {
	secret = sanitizeSecret(secret)

	counterBytes := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		counterBytes[i] = byte(counter & 0xff)
		counter >>= 8
	}

	secretBytes, _ := base32.StdEncoding.DecodeString(secret)

	hash := sha256.Sum256(append(counterBytes, secretBytes...))

	offset := hash[len(hash)-1] & 0x0f
	truncated := (uint32(hash[offset]) & 0x7f) << 24
	truncated |= (uint32(hash[offset+1]) & 0xff) << 16
	truncated |= (uint32(hash[offset+2]) & 0xff) << 8
	truncated |= (uint32(hash[offset+3]) & 0xff)

	otp := truncated % 1000000
	return fmt.Sprintf("%06d", otp)
}

func generateBackupCode() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	mathrand.Seed(time.Now().UnixNano())

	code := make([]byte, 8)
	for i := 0; i < 8; i++ {
		code[i] = chars[mathrand.Intn(len(chars))]
	}

	return string(code)
}

func sanitizeAccountName(name string) string {
	if name == "" {
		return "user"
	}
	return name
}

func sanitizeIssuer(issuer string) string {
	if issuer == "" {
		return "DukaPOS"
	}
	return issuer
}

func sanitizeSecret(secret string) string {
	secret = toUpper(secret)
	secret = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte(secret))
	return secret
}

func cleanBackupCode(code string) string {
	var result []byte
	for _, c := range code {
		if c >= 'A' && c <= 'Z' || c >= '2' && c <= '7' {
			result = append(result, byte(c))
		}
	}
	return string(result)
}

func toUpper(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			c -= 32
		}
		result[i] = c
	}
	return string(result)
}

func constTimeEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}

	result := 0
	for i := 0; i < len(a); i++ {
		result |= int(a[i]) ^ int(b[i])
	}

	return result == 0
}

type BackupCodeStore struct {
	codes    map[uint][]string
	attempts map[uint]int
	mu       sync.RWMutex
}

func NewBackupCodeStore() *BackupCodeStore {
	return &BackupCodeStore{
		codes:    make(map[uint][]string),
		attempts: make(map[uint]int),
	}
}

func (s *BackupCodeStore) SetCodes(userID uint, codes []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.codes[userID] = codes
}

func (s *BackupCodeStore) VerifyCode(userID uint, code string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	codes, ok := s.codes[userID]
	if !ok {
		return false
	}

	attempts := s.attempts[userID]
	if attempts >= 10 {
		return false
	}

	cleanCode := cleanBackupCode(code)
	for i, hashed := range codes {
		if constTimeEqual(hashed, cleanCode) {
			codes = append(codes[:i], codes[i+1:]...)
			s.codes[userID] = codes
			return true
		}
	}

	s.attempts[userID] = attempts + 1
	return false
}

func (s *BackupCodeStore) RemainingCodes(userID uint) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.codes[userID])
}
