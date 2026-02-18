package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	
	"strings"
	"sync"
	"time"

	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"github.com/C9b3rD3vi1/DukaPOS/internal/repository"
)

// Errors
var (
	ErrInvalidKey  = errors.New("invalid API key")
	ErrKeyExpired = errors.New("API key expired")
	ErrKeyInactive = errors.New("API key is inactive")
	ErrRateLimited = errors.New("rate limit exceeded")
)

// RateLimit stores rate limit state per API key
type RateLimit struct {
	Count     int
	ResetTime time.Time
}

var (
	rateLimits   = make(map[uint]*RateLimit)
	rateLimitsMu sync.Mutex
)

// Service handles API key operations
type Service struct {
	apiKeyRepo *repository.APIKeyRepository
	keyPrefix  string
}

// New creates a new API key service
func New(apiKeyRepo *repository.APIKeyRepository) *Service {
	return &Service{
		apiKeyRepo: apiKeyRepo,
		keyPrefix:  "dkp_",
	}
}

// CreateKey creates a new API key
func (s *Service) CreateKey(shopID uint, name string, permissions string, rateLimit int) (*models.APIKey, error) {
	key := s.generateKey()
	secret, err := s.generateSecret()
	if err != nil {
		return nil, err
	}

	hashedSecret := s.hashSecret(secret)

	apiKey := &models.APIKey{
		ShopID:      shopID,
		Name:        name,
		Key:         key,
		SecretHash:  hashedSecret,
		Permissions: permissions,
		RateLimit:   rateLimit,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.apiKeyRepo.Create(apiKey); err != nil {
		return nil, err
	}

	// Return with plain secret only once
	apiKey.Key = key
	apiKey.SecretHash = secret // Return plain secret to user

	return apiKey, nil
}

// ValidateKey validates an API key
func (s *Service) ValidateKey(key string) (*models.APIKey, error) {
	apiKey, err := s.apiKeyRepo.GetByKey(key)
	if err != nil {
		return nil, ErrInvalidKey
	}

	if !apiKey.IsActive {
		return nil, ErrKeyInactive
	}

	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return nil, ErrKeyExpired
	}

	return apiKey, nil
}

// CheckRateLimit checks if the request is within rate limit
func (s *Service) CheckRateLimit(key *models.APIKey) bool {
	rateLimitsMu.Lock()
	defer rateLimitsMu.Unlock()

	now := time.Now()
	rl, exists := rateLimits[key.ID]

	if !exists || now.After(rl.ResetTime) {
		rateLimits[key.ID] = &RateLimit{
			Count:     1,
			ResetTime: now.Add(time.Minute),
		}
		return true
	}

	if rl.Count >= key.RateLimit {
		return false
	}

	rl.Count++
	return true
}

// UpdateLastUsed updates the last used timestamp
func (s *Service) UpdateLastUsed(keyID uint) {
	_ = s.apiKeyRepo.UpdateLastUsed(keyID)
}

// ListByShop lists all API keys for a shop
func (s *Service) ListByShop(shopID uint) ([]models.APIKey, error) {
	return s.apiKeyRepo.GetByShopID(shopID)
}

// RevokeKey revokes an API key
func (s *Service) RevokeKey(id uint) error {
	key, err := s.apiKeyRepo.GetByID(id)
	if err != nil {
		return err
	}

	key.IsActive = false
	return s.apiKeyRepo.Update(key)
}

// GetKey gets an API key by ID
func (s *Service) GetKey(id uint) (*models.APIKey, error) {
	return s.apiKeyRepo.GetByID(id)
}

// Generate the key portion
func (s *Service) generateKey() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return s.keyPrefix + hex.EncodeToString(bytes)[:24]
}

// Generate secret portion
func (s *Service) generateSecret() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Hash secret for storage
func (s *Service) hashSecret(secret string) string {
	hash := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(hash[:])
}

// VerifySecret verifies the secret
func (s *Service) VerifySecret(key *models.APIKey, secret string) bool {
	return key.SecretHash == s.hashSecret(secret)
}

// HasPermission checks if API key has required permission
func (s *Service) HasPermission(key *models.APIKey, permission string) bool {
	perms := strings.Split(key.Permissions, ",")
	for _, p := range perms {
		if strings.TrimSpace(p) == permission || p == "*" {
			return true
		}
	}
	return false
}

// GetAvailablePermissions returns list of available permissions
func (s *Service) GetAvailablePermissions() []string {
	return []string{
		"products:read",
		"products:write",
		"sales:read",
		"sales:write",
		"reports:read",
		"payments:read",
		"payments:write",
		"customers:read",
		"customers:write",
		"staff:read",
		"staff:write",
	}
}

// Documentation returns API documentation
func (s *Service) Documentation() map[string]interface{} {
	return map[string]interface{}{
		"title":          "DukaPOS REST API",
		"version":       "1.0.0",
		"base_url":      "/api/v1",
		"authentication": "X-API-Key header or api_key query param",
		"rate_limits": map[string]int{
			"default":   60,
			"products":   100,
			"sales":     60,
			"reports":   30,
			"payments":  20,
		},
	}
}

// CleanExpiredKeys removes expired keys from rate limit cache
func (s *Service) CleanExpiredKeys() {
	rateLimitsMu.Lock()
	defer rateLimitsMu.Unlock()

	now := time.Now()
	for id, rl := range rateLimits {
		if now.After(rl.ResetTime) {
			delete(rateLimits, id)
		}
	}
}

// StartRateLimitCleaner starts a background cleaner for rate limits
func (s *Service) StartRateLimitCleaner(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		for range ticker.C {
			s.CleanExpiredKeys()
		}
	}()
}

// GetRateLimitStatus gets current rate limit status for a key
func (s *Service) GetRateLimitStatus(key *models.APIKey) map[string]interface{} {
	rateLimitsMu.Lock()
	defer rateLimitsMu.Unlock()

	rl, exists := rateLimits[key.ID]
	if !exists {
		return map[string]interface{}{
			"remaining": key.RateLimit,
			"limit":     key.RateLimit,
			"reset":     time.Now().Add(time.Minute).Unix(),
		}
	}

	remaining := key.RateLimit - rl.Count
	if remaining < 0 {
		remaining = 0
	}

	return map[string]interface{}{
		"remaining": remaining,
		"limit":     key.RateLimit,
		"reset":     rl.ResetTime.Unix(),
		"used":      rl.Count,
	}
}
