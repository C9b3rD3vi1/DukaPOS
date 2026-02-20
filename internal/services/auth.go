package services

import (
	"errors"
	"time"

	"github.com/C9b3rD3vi1/DukaPOS/internal/config"
	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"github.com/C9b3rD3vi1/DukaPOS/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrShopExists         = errors.New("shop already exists with this phone/email")
	ErrTokenExpired       = errors.New("token has expired")
)

// AuthService handles authentication
type AuthService struct {
	shopRepo    *repository.ShopRepository
	accountRepo *repository.AccountRepository
	cfg         *config.Config
}

// NewAuthService creates a new auth service
func NewAuthService(shopRepo *repository.ShopRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		shopRepo: shopRepo,
		cfg:      cfg,
	}
}

// SetAccountRepo sets the account repository
func (s *AuthService) SetAccountRepo(accountRepo *repository.AccountRepository) {
	s.accountRepo = accountRepo
}

// Register creates a new shop account
func (s *AuthService) Register(shop *models.Shop, password string) error {
	// Check if phone already exists
	existing, err := s.shopRepo.GetByPhone(shop.Phone)
	if err == nil && existing != nil {
		return ErrShopExists
	}

	// Check if email already exists
	if shop.Email != "" {
		existing, err = s.shopRepo.GetByEmail(shop.Email)
		if err == nil && existing != nil {
			return ErrShopExists
		}
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	shop.PasswordHash = string(hashedPassword)
	shop.IsActive = true
	shop.Plan = models.PlanFree

	// Create shop
	return s.shopRepo.Create(shop)
}

// Login authenticates a shop and returns a token
func (s *AuthService) Login(phoneOrEmail, password string) (*models.Shop, string, *models.Account, error) {
	var shop *models.Shop
	var err error

	// Try shop by phone first
	shop, err = s.shopRepo.GetByPhone(phoneOrEmail)
	if err != nil {
		// Try shop by email
		shop, err = s.shopRepo.GetByEmail(phoneOrEmail)
		if err != nil && s.accountRepo != nil {
			// Try account by email (for admin login)
			account, accountErr := s.accountRepo.GetByEmail(phoneOrEmail)
			if accountErr == nil && account != nil {
				// Verify password
				if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(password)); err != nil {
					return nil, "", nil, ErrInvalidCredentials
				}

				// Find or create shop for this account
				shop, _ = s.shopRepo.GetByAccountID(account.ID)
				if shop == nil {
					// Create a default shop for admin
					shop = &models.Shop{
						AccountID:    account.ID,
						Name:         account.Name,
						Phone:        account.Phone,
						OwnerName:    account.Name,
						Plan:         account.Plan,
						IsActive:     account.IsActive,
						Email:        account.Email,
						PasswordHash: account.PasswordHash,
					}
					s.shopRepo.Create(shop)
					shop, _ = s.shopRepo.GetByAccountID(account.ID)
				}

				// Generate token
				token, tokenErr := s.generateToken(shop)
				if tokenErr != nil {
					return nil, "", nil, tokenErr
				}

				return shop, token, account, nil
			}
			return nil, "", nil, ErrInvalidCredentials
		}
		return nil, "", nil, ErrInvalidCredentials
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(shop.PasswordHash), []byte(password)); err != nil {
		return nil, "", nil, ErrInvalidCredentials
	}

	// Get account
	var account *models.Account
	if s.accountRepo != nil {
		account, _ = s.accountRepo.GetByID(shop.AccountID)
	}

	// Generate JWT token
	token, err := s.generateToken(shop)
	if err != nil {
		return nil, "", nil, err
	}

	return shop, token, account, nil
}

// GetAccountByID retrieves an account by ID
func (s *AuthService) GetAccountByID(id uint) (*models.Account, error) {
	if s.accountRepo == nil {
		return nil, nil
	}
	return s.accountRepo.GetByID(id)
}

// ValidateToken validates a JWT token and returns the shop
func (s *AuthService) ValidateToken(tokenString string) (*models.Shop, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		shopID := uint(claims["shop_id"].(float64))
		shop, err := s.shopRepo.GetByID(shopID)
		if err != nil {
			return nil, err
		}
		return shop, nil
	}

	return nil, ErrTokenExpired
}

// ChangePassword changes a shop's password
func (s *AuthService) ChangePassword(shopID uint, oldPassword, newPassword string) error {
	shop, err := s.shopRepo.GetByID(shopID)
	if err != nil {
		return err
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(shop.PasswordHash), []byte(oldPassword)); err != nil {
		return ErrInvalidCredentials
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	shop.PasswordHash = string(hashedPassword)
	return s.shopRepo.Update(shop)
}

// ResetPassword resets a shop's password (admin only in production)
func (s *AuthService) ResetPassword(shopID uint, newPassword string) error {
	shop, err := s.shopRepo.GetByID(shopID)
	if err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	shop.PasswordHash = string(hashedPassword)
	return s.shopRepo.Update(shop)
}

func (s *AuthService) generateToken(shop *models.Shop) (string, error) {
	claims := jwt.MapClaims{
		"shop_id": shop.ID,
		"phone":   shop.Phone,
		"plan":    shop.Plan,
		"exp":     time.Now().Add(s.cfg.GetJWTDuration()).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

// AuthMiddleware provides JWT authentication middleware
type AuthMiddleware struct {
	authService *AuthService
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(authService *AuthService) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

// ExtractShopID extracts shop ID from token (call after middleware validation)
func (m *AuthMiddleware) ExtractShopID(c interface{}) (uint, error) {
	// This would typically extract from Fiber context
	// Implementation depends on how the middleware is set up
	return 0, nil
}
