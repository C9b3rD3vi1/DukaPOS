package validation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator provides request validation
type Validator struct {
	validate *validator.Validate
}

// New creates a new validator
func New() *Validator {
	v := validator.New()
	
	// Register custom validators
	v.RegisterValidation("phone", validatePhone)
	v.RegisterValidation("password", validatePassword)
	v.RegisterValidation("money", validateMoney)
	
	return &Validator{validate: v}
}

// Validate validates a struct
func (v *Validator) Validate(i interface{}) ValidationResult {
	err := v.validate.Struct(i)
	if err == nil {
		return ValidationResult{Valid: true}
	}

	errors := make([]FieldError, 0)
	for _, err := range err.(validator.ValidationErrors) {
		errors = append(errors, FieldError{
			Field:   err.Field(),
			Tag:     err.Tag(),
			Value:   err.Value(),
			Message: formatError(err),
		})
	}

	return ValidationResult{
		Valid:  false,
		Errors: errors,
	}
}

// ValidationResult represents validation result
type ValidationResult struct {
	Valid bool
	Error string
	Errors []FieldError
}

// FieldError represents a field validation error
type FieldError struct {
	Field   string
	Tag     string
	Value   interface{}
	Message string
}

func formatError(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", err.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email", err.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", err.Field(), err.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", err.Field(), err.Param())
	case "phone":
		return fmt.Sprintf("%s must be a valid phone number", err.Field())
	case "password":
		return "Password must be at least 6 characters"
	case "money":
		return fmt.Sprintf("%s must be a valid amount", err.Field())
	default:
		return fmt.Sprintf("Invalid %s", err.Field())
	}
}

// Custom validators
func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	// Kenyan phone format
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	
	// Check for valid formats
	// +254XXXXXXXXX, 254XXXXXXXXX, 07XXXXXXXX, 01XXXXXXXX
	regex := regexp.MustCompile(`^(\+?254|0)[7-9][0-9]{8}$`)
	return regex.MatchString(phone)
}

func validatePassword(fl validator.FieldLevel) bool {
	return len(fl.Field().String()) >= 6
}

func validateMoney(fl validator.FieldLevel) bool {
	amount := fl.Field().Float()
	return amount >= 0
}

// =========================================
// Request DTOs with validation tags
// =========================================

// RegisterRequest represents registration request
type RegisterRequest struct {
	Phone    string `json:"phone" validate:"required,phone"`
	Email    string `json:"email" validate:"omitempty,email"`
	Password string `json:"password" validate:"required,min=6"`
	Name     string `json:"name" validate:"required"`
}

// LoginRequest represents login request
type LoginRequest struct {
	Phone    string `json:"phone" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// ProductRequest represents product request
type ProductRequest struct {
	Name            string  `json:"name" validate:"required"`
	Category        string  `json:"category"`
	Unit            string  `json:"unit"`
	CostPrice       float64 `json:"cost_price" validate:"omitempty,money"`
	SellingPrice    float64 `json:"selling_price" validate:"required,money"`
	CurrentStock   int     `json:"current_stock"`
	LowStockThreshold int   `json:"low_stock_threshold"`
	Barcode        string  `json:"barcode"`
}

// SaleRequest represents sale request
type SaleRequest struct {
	ProductID   uint    `json:"product_id" validate:"required"`
	Quantity    int     `json:"quantity" validate:"required,min=1"`
	UnitPrice   float64 `json:"unit_price" validate:"required,money"`
	PaymentMethod string `json:"payment_method" validate:"required"`
}

// STKPushRequest represents M-Pesa STK push request
type STKPushRequest struct {
	Phone        string `json:"phone" validate:"required,phone"`
	Amount       int    `json:"amount" validate:"required,min=1"`
	AccountRef   string `json:"account_ref"`
	Description  string `json:"description"`
}

// WebhookRequest represents webhook request
type WebhookRequest struct {
	Name   string `json:"name" validate:"required"`
	URL    string `json:"url" validate:"required,url"`
	Events string `json:"events" validate:"required"`
}

// StaffRequest represents staff request
type StaffRequest struct {
	Name  string `json:"name" validate:"required"`
	Phone string `json:"phone" validate:"required,phone"`
	Role  string `json:"role" validate:"required"`
}

// APIKeyRequest represents API key request
type APIKeyRequest struct {
	Name        string `json:"name" validate:"required"`
	Permissions string `json:"permissions"`
	RateLimit   int    `json:"rate_limit"`
	ExpiresIn   int    `json:"expires_in"` // days
}

// CustomerRequest represents customer request
type CustomerRequest struct {
	Name  string `json:"name" validate:"required"`
	Phone string `json:"phone" validate:"required,phone"`
	Email string `json:"email" validate:"omitempty,email"`
}

// LoyaltyPointsRequest represents loyalty points request
type LoyaltyPointsRequest struct {
	Points int    `json:"points" validate:"required,min=1"`
	Type   string `json:"type" validate:"required,oneof=add deduct"`
}
