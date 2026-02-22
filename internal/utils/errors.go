package utils

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

type UserError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func (e *UserError) Error() string {
	return e.Message
}

func NewUserError(code, message string) *UserError {
	return &UserError{Code: code, Message: message}
}

func NewFieldError(code, message, field string) *UserError {
	return &UserError{Code: code, Message: message, Field: field}
}

var (
	ErrInvalidCredentials  = NewUserError("INVALID_CREDENTIALS", "Invalid phone/email or password. Please check your credentials and try again.")
	ErrAccountNotFound     = NewUserError("ACCOUNT_NOT_FOUND", "Account not found. Please check your phone number or register a new account.")
	ErrShopNotFound        = NewUserError("SHOP_NOT_FOUND", "Shop not found. The shop may have been deleted or you may not have access.")
	ErrProductNotFound     = NewUserError("PRODUCT_NOT_FOUND", "Product not found. It may have been deleted or the ID is incorrect.")
	ErrInsufficientStock   = NewUserError("INSUFFICIENT_STOCK", "Not enough stock. The requested quantity exceeds available inventory.")
	ErrInvalidQuantity     = NewUserError("INVALID_QUANTITY", "Invalid quantity. Please enter a valid number greater than zero.")
	ErrInvalidPrice        = NewUserError("INVALID_PRICE", "Invalid price. Please enter a valid amount greater than zero.")
	ErrUnauthorized        = NewUserError("UNAUTHORIZED", "You are not authorized to perform this action. Please log in and try again.")
	ErrPaymentFailed       = NewUserError("PAYMENT_FAILED", "Payment failed. Please try again or use a different payment method.")
	ErrNetworkError        = NewUserError("NETWORK_ERROR", "Network error. Please check your connection and try again.")
	ErrRateLimited         = NewUserError("RATE_LIMITED", "Too many requests. Please wait a moment before trying again.")
	ErrFeatureNotAvailable = NewUserError("FEATURE_NOT_AVAILABLE", "This feature is not available on your current plan. Upgrade to unlock!")
)

func HandleDBError(err error) *UserError {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrProductNotFound
	}
	if strings.Contains(err.Error(), "duplicate key") {
		return NewUserError("DUPLICATE_ENTRY", "This record already exists. Please use a different value.")
	}
	if strings.Contains(err.Error(), "foreign key constraint") {
		return NewUserError("REFERENCE_ERROR", "This record is linked to other data and cannot be deleted.")
	}
	return NewUserError("DATABASE_ERROR", "Something went wrong. Please try again later.")
}

func HandleValidationError(field, message string) *UserError {
	return NewFieldError("VALIDATION_ERROR", message, field)
}
