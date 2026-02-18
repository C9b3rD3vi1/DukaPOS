package otp

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/C9b3rD3vi1/DukaPOS/internal/config"
	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"gorm.io/gorm"
)

type OTPService struct {
	db             *gorm.DB
	config         *config.Config
	whatsappSender func(phone, message string) error
	smsSender      func(phone, message string) error
	rateLimiter    *RateLimiter
	otpStore       *OTPStore
}

type OTPStore struct {
	mu       sync.RWMutex
	otps     map[string]*OTPEntry
	verified map[string]bool
}

type OTPEntry struct {
	Phone       string
	Code        string
	Purpose     string
	Attempts    int
	MaxAttempts int
	CreatedAt   time.Time
	ExpiresAt   time.Time
	Verified    bool
}

type RateLimiter struct {
	mu       sync.RWMutex
	attempts map[string]*RateLimitEntry
}

type RateLimitEntry struct {
	Count        int
	FirstAttempt time.Time
	BlockedUntil time.Time
}

type OTPRequest struct {
	Phone   string `json:"phone"`
	Purpose string `json:"purpose"` // login, register, password_reset, phone_verify
}

type OTPVerifyRequest struct {
	Phone   string `json:"phone"`
	Code    string `json:"code"`
	Purpose string `json:"purpose"`
}

type OTPResponse struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	Attempts  int       `json:"attempts_remaining,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

const (
	PurposeLogin         = "login"
	PurposeRegister      = "register"
	PurposePasswordReset = "password_reset"
	PurposePhoneVerify   = "phone_verify"
	PurposePayment       = "payment"

	OTPExpiryMinutes = 5
	MaxAttempts      = 3
	RateLimitMinutes = 15
	MaxRateLimit     = 5
)

func NewOTPService(db *gorm.DB, cfg *config.Config) *OTPService {
	return &OTPService{
		db:     db,
		config: cfg,
		rateLimiter: &RateLimiter{
			attempts: make(map[string]*RateLimitEntry),
		},
		otpStore: &OTPStore{
			otps:     make(map[string]*OTPEntry),
			verified: make(map[string]bool),
		},
	}
}

func (s *OTPService) SetWhatsAppSender(sender func(phone, message string) error) {
	s.whatsappSender = sender
}

func (s *OTPService) SetSMSSender(sender func(phone, message string) error) {
	s.smsSender = sender
}

func (s *OTPService) GenerateOTP(ctx context.Context, req *OTPRequest) (*OTPResponse, error) {
	phone := normalizePhone(req.Phone)
	if phone == "" {
		return &OTPResponse{Success: false, Message: "Invalid phone number"}, nil
	}

	purpose := req.Purpose
	if purpose == "" {
		purpose = PurposeLogin
	}

	storeKey := fmt.Sprintf("%s:%s", purpose, phone)

	if !s.rateLimiter.Allow(storeKey) {
		return &OTPResponse{
			Success: false,
			Message: "Too many OTP requests. Please try again later.",
		}, nil
	}

	otpCode := generateOTPCode(6)

	entry := &OTPEntry{
		Phone:       phone,
		Code:        otpCode,
		Purpose:     purpose,
		Attempts:    0,
		MaxAttempts: MaxAttempts,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(OTPExpiryMinutes * time.Minute),
		Verified:    false,
	}

	s.otpStore.mu.Lock()
	s.otpStore.otps[storeKey] = entry
	s.otpStore.mu.Unlock()

	message := s.formatOTPMessage(otpCode, purpose)

	if s.whatsappSender != nil {
		go func() {
			_ = s.whatsappSender(phone, message)
		}()
	} else if s.smsSender != nil {
		go func() {
			_ = s.smsSender(phone, message)
		}()
	}

	return &OTPResponse{
		Success:   true,
		Message:   fmt.Sprintf("OTP sent to %s", maskPhone(phone)),
		ExpiresAt: entry.ExpiresAt,
	}, nil
}

func (s *OTPService) VerifyOTP(ctx context.Context, req *OTPVerifyRequest) (*OTPResponse, error) {
	phone := normalizePhone(req.Phone)
	if phone == "" {
		return &OTPResponse{Success: false, Message: "Invalid phone number"}, nil
	}

	purpose := req.Purpose
	if purpose == "" {
		purpose = PurposeLogin
	}

	storeKey := fmt.Sprintf("%s:%s", purpose, phone)

	s.otpStore.mu.RLock()
	entry, exists := s.otpStore.otps[storeKey]
	s.otpStore.mu.RUnlock()

	if !exists {
		return &OTPResponse{Success: false, Message: "No OTP found. Please request a new code."}, nil
	}

	if time.Now().After(entry.ExpiresAt) {
		s.otpStore.mu.Lock()
		delete(s.otpStore.otps, storeKey)
		s.otpStore.mu.Unlock()
		return &OTPResponse{Success: false, Message: "OTP expired. Please request a new code."}, nil
	}

	entry.Attempts++

	if entry.Attempts >= entry.MaxAttempts {
		s.otpStore.mu.Lock()
		delete(s.otpStore.otps, storeKey)
		s.otpStore.mu.Unlock()
		return &OTPResponse{Success: false, Message: "Too many attempts. Please request a new OTP."}, nil
	}

	if subtle.ConstantTimeCompare([]byte(entry.Code), []byte(req.Code)) != 1 {
		remaining := entry.MaxAttempts - entry.Attempts
		return &OTPResponse{
			Success:  false,
			Message:  fmt.Sprintf("Invalid OTP. %d attempts remaining.", remaining),
			Attempts: remaining,
		}, nil
	}

	entry.Verified = true

	s.otpStore.mu.Lock()
	s.otpStore.verified[phone] = true
	s.otpStore.mu.Unlock()

	s.otpStore.mu.Lock()
	delete(s.otpStore.otps, storeKey)
	s.otpStore.mu.Unlock()

	return &OTPResponse{
		Success: true,
		Message: "Phone verified successfully!",
	}, nil
}

func (s *OTPService) IsVerified(phone string) bool {
	phone = normalizePhone(phone)
	s.otpStore.mu.RLock()
	defer s.otpStore.mu.RUnlock()
	return s.otpStore.verified[phone]
}

func (s *OTPService) ClearVerification(phone string) {
	phone = normalizePhone(phone)
	s.otpStore.mu.Lock()
	defer s.otpStore.mu.Unlock()
	delete(s.otpStore.verified, phone)
}

func (s *OTPService) VerifyShopPhone(shopID uint, phone, code string) error {
	storeKey := fmt.Sprintf("%s:shop_%d", PurposePhoneVerify, shopID)

	s.otpStore.mu.RLock()
	entry, exists := s.otpStore.otps[storeKey]
	s.otpStore.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no OTP found for this shop")
	}

	if time.Now().After(entry.ExpiresAt) {
		return fmt.Errorf("OTP has expired")
	}

	if entry.Code != code {
		return fmt.Errorf("invalid OTP code")
	}

	shop := &models.Shop{}
	if err := s.db.Model(shop).Where("id = ?", shopID).Update("phone_verified", true).Error; err != nil {
		return err
	}

	s.otpStore.mu.Lock()
	delete(s.otpStore.otps, storeKey)
	s.otpStore.mu.Unlock()

	return nil
}

func (s *OTPService) RequestShopVerification(ctx context.Context, shopID uint) (*OTPResponse, error) {
	shop := &models.Shop{}
	if err := s.db.First(shop, shopID).Error; err != nil {
		return nil, fmt.Errorf("shop not found")
	}

	if shop.Phone == "" {
		return &OTPResponse{Success: false, Message: "Shop has no phone number"}, nil
	}

	otpCode := generateOTPCode(6)
	storeKey := fmt.Sprintf("%s:shop_%d", PurposePhoneVerify, shopID)

	entry := &OTPEntry{
		Phone:       shop.Phone,
		Code:        otpCode,
		Purpose:     PurposePhoneVerify,
		Attempts:    0,
		MaxAttempts: MaxAttempts,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(OTPExpiryMinutes * time.Minute),
	}

	s.otpStore.mu.Lock()
	s.otpStore.otps[storeKey] = entry
	s.otpStore.mu.Unlock()

	message := fmt.Sprintf("Your DukaPOS verification code is: %s\n\nThis code expires in 5 minutes.", otpCode)

	if s.whatsappSender != nil {
		go func() { _ = s.whatsappSender(shop.Phone, message) }()
	} else if s.smsSender != nil {
		go func() { _ = s.smsSender(shop.Phone, message) }()
	}

	return &OTPResponse{
		Success:   true,
		Message:   fmt.Sprintf("Verification code sent to %s", maskPhone(shop.Phone)),
		ExpiresAt: entry.ExpiresAt,
	}, nil
}

func (s *RateLimiter) Allow(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	entry, exists := s.attempts[key]

	if !exists {
		s.attempts[key] = &RateLimitEntry{
			Count:        1,
			FirstAttempt: now,
		}
		return true
	}

	if now.After(entry.BlockedUntil) {
		if now.Sub(entry.FirstAttempt) > RateLimitMinutes*time.Minute {
			s.attempts[key] = &RateLimitEntry{
				Count:        1,
				FirstAttempt: now,
			}
			return true
		}

		if entry.Count >= MaxRateLimit {
			entry.BlockedUntil = now.Add(RateLimitMinutes * time.Minute)
			return false
		}

		entry.Count++
		return true
	}

	return false
}

func normalizePhone(phone string) string {
	phone = fmt.Sprintf("+%s", phone)
	phone = phone[1:]
	phone = removeSpecialChars(phone)

	if len(phone) == 12 && phone[:3] == "254" {
		return phone
	}
	if len(phone) == 10 && phone[0] == '0' {
		return "254" + phone[1:]
	}
	if len(phone) == 9 {
		return "254" + phone
	}

	return ""
}

func removeSpecialChars(s string) string {
	result := ""
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result += string(c)
		}
	}
	return result
}

func generateOTPCode(length int) string {
	const digits = "0123456789"
	code := ""
	for i := 0; i < length; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		code += string(digits[n.Int64()])
	}
	return code
}

func maskPhone(phone string) string {
	if len(phone) < 4 {
		return "****"
	}
	return "****" + phone[len(phone)-4:]
}

func (s *OTPService) formatOTPMessage(code, purpose string) string {
	switch purpose {
	case PurposeLogin:
		return fmt.Sprintf("Your DukaPOS login code is: %s\n\nThis code expires in 5 minutes.\n\nNever share this code with anyone.", code)
	case PurposeRegister:
		return fmt.Sprintf("Welcome to DukaPOS!\n\nYour verification code is: %s\n\nThis code expires in 5 minutes.", code)
	case PurposePasswordReset:
		return fmt.Sprintf("Your DukaPOS password reset code is: %s\n\nThis code expires in 5 minutes.\n\nIf you didn't request this, please ignore.", code)
	case PurposePhoneVerify:
		return fmt.Sprintf("Your DukaPOS phone verification code is: %s\n\nThis code expires in 5 minutes.", code)
	case PurposePayment:
		return fmt.Sprintf("Your DukaPOS payment verification code is: %s\n\nThis code expires in 5 minutes.", code)
	default:
		return fmt.Sprintf("Your DukaPOS verification code is: %s\n\nThis code expires in 5 minutes.", code)
	}
}
