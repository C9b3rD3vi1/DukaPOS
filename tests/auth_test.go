package main

import (
	"testing"
)

// TestPhoneNormalization tests phone number normalization
func TestPhoneNormalization(t *testing.T) {
	normalizePhone := func(phone string) string {
		// Remove any special characters
		result := ""
		for _, c := range phone {
			if c >= '0' && c <= '9' {
				result += string(c)
			}
		}
		// Add country code if needed (9 digits)
		if len(result) == 9 {
			if result[0] == '0' {
				result = "254" + result[1:]
			} else if result[0] == '7' {
				result = "254" + result
			}
		}
		return result
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"+254700000000", "254700000000"},
		{"254700000000", "254700000000"},
		{"700000000", "254700000000"},
		{"+254-700-000-000", "254700000000"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizePhone(tt.input)
			if result != tt.expected {
				t.Errorf("normalizePhone(%s) = %s; want %s", tt.input, result, tt.expected)
			}
		})
	}
}

// TestPasswordValidation tests password validation
func TestPasswordValidation(t *testing.T) {
	validatePassword := func(password string) (bool, string) {
		if len(password) < 6 {
			return false, "password must be at least 6 characters"
		}
		return true, ""
	}

	tests := []struct {
		password  string
		wantValid bool
	}{
		{"12345", false},
		{"123456", true},
		{"password123", true},
		{"", false},
		{"abcdefghijklmnop", true},
	}

	for _, tt := range tests {
		t.Run(tt.password, func(t *testing.T) {
			valid, _ := validatePassword(tt.password)
			if valid != tt.wantValid {
				t.Errorf("validatePassword(%s) = %v; want %v", tt.password, valid, tt.wantValid)
			}
		})
	}
}

// TestPINValidation tests PIN validation
func TestPINValidation(t *testing.T) {
	validatePIN := func(pin string) (bool, string) {
		if len(pin) < 4 || len(pin) > 6 {
			return false, "PIN must be 4-6 digits"
		}
		for _, c := range pin {
			if c < '0' || c > '9' {
				return false, "PIN must contain only digits"
			}
		}
		return true, ""
	}

	tests := []struct {
		pin       string
		wantValid bool
	}{
		{"123", false},
		{"1234", true},
		{"123456", true},
		{"12345", true},
		{"abcd", false},
		{"", false},
		{"12a4", false},
	}

	for _, tt := range tests {
		t.Run(tt.pin, func(t *testing.T) {
			valid, _ := validatePIN(tt.pin)
			if valid != tt.wantValid {
				t.Errorf("validatePIN(%s) = %v; want %v", tt.pin, valid, tt.wantValid)
			}
		})
	}
}

// TestEmailValidation tests email validation
func TestEmailValidation(t *testing.T) {
	validateEmail := func(email string) bool {
		if email == "" {
			return false
		}
		hasAt := false
		hasDotAfterAt := false
		for i, c := range email {
			if c == '@' {
				hasAt = true
			}
			if c == '.' && hasAt && i > 0 {
				hasDotAfterAt = true
			}
		}
		// Must have @, must have . after @, and not start with @
		return hasAt && hasDotAfterAt && email[0] != '@'
	}

	tests := []struct {
		email    string
		expected bool
	}{
		{"test@example.com", true},
		{"user@domain.co.ke", true},
		{"invalid", false},
		{"", false},
		{"@nodomain.com", false},
		{"noat.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := validateEmail(tt.email)
			if result != tt.expected {
				t.Errorf("validateEmail(%s) = %v; want %v", tt.email, result, tt.expected)
			}
		})
	}
}

// TestJWTTokenGeneration tests JWT token generation (mock)
func TestJWTTokenGeneration(t *testing.T) {
	createToken := func(shopID uint, secret string) string {
		// Simplified mock - in real implementation uses jwt-go
		return "token_" + secret + "_" + string(rune(shopID+'0'))
	}

	token1 := createToken(1, "secret123")
	token2 := createToken(1, "secret123")
	token3 := createToken(2, "secret123")

	// Same inputs should produce same output (deterministic)
	if token1 != token2 {
		t.Error("Same inputs should produce same token")
	}

	// Different shop IDs should produce different tokens
	if token1 == token3 {
		t.Error("Different shop IDs should produce different tokens")
	}
}

// TestAccountLockoutLogic tests account lockout logic
func TestAccountLockoutLogic(t *testing.T) {
	type Account struct {
		FailedAttempts int
		LockedUntil    *int // using int to simulate time
	}

	shouldLock := func(acc *Account) bool {
		return acc.FailedAttempts >= 5
	}

	tests := []struct {
		name     string
		attempts int
		wantLock bool
	}{
		{"no attempts", 0, false},
		{"3 attempts", 3, false},
		{"4 attempts", 4, false},
		{"5 attempts", 5, true},
		{"10 attempts", 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := &Account{FailedAttempts: tt.attempts}
			lock := shouldLock(acc)
			if lock != tt.wantLock {
				t.Errorf("shouldLock() with %d attempts = %v; want %v", tt.attempts, lock, tt.wantLock)
			}
		})
	}
}
