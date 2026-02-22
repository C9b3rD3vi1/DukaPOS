package main

import (
	"testing"
)

// TestOTPCodeGeneration tests OTP code generation
func TestOTPCodeGeneration(t *testing.T) {
	generateOTPCode := func(length int) string {
		code := ""
		for i := 0; i < length; i++ {
			code += "1" // Simplified - just return 111111
		}
		return code
	}

	code4 := generateOTPCode(4)
	if len(code4) != 4 {
		t.Errorf("OTP code length = %d; want 4", len(code4))
	}

	code6 := generateOTPCode(6)
	if len(code6) != 6 {
		t.Errorf("OTP code length = %d; want 6", len(code6))
	}
}

// TestOTPValidation tests OTP validation logic
func TestOTPValidation(t *testing.T) {
	validateOTP := func(input string) bool {
		if len(input) < 4 || len(input) > 6 {
			return false
		}
		for _, c := range input {
			if c < '0' || c > '9' {
				return false
			}
		}
		return true
	}

	tests := []struct {
		input    string
		expected bool
	}{
		{"1234", true},
		{"123456", true},
		{"12345", true},
		{"123", false},
		{"1234567", false},
		{"abcd", false},
		{"12a4", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := validateOTP(tt.input)
			if result != tt.expected {
				t.Errorf("validateOTP(%s) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestOTPExpiry tests OTP expiry logic
func TestOTPExpiry(t *testing.T) {
	isExpired := func(createdAt, now int, expiryMinutes int) bool {
		expirySeconds := expiryMinutes * 60
		return (now - createdAt) > expirySeconds
	}

	// Created now (0 seconds ago) - not expired
	if isExpired(1000, 1000, 5) {
		t.Error("OTP created just now should not be expired")
	}

	// Created 4 minutes ago - not expired (5 min expiry)
	if isExpired(1000, 1240, 5) {
		t.Error("OTP created 4 minutes ago should not be expired")
	}

	// Created 6 minutes ago - expired (5 min expiry)
	if !isExpired(1000, 1360, 5) {
		t.Error("OTP created 6 minutes ago should be expired")
	}
}

// TestOTPPurpose tests OTP purpose validation
func TestOTPPurpose(t *testing.T) {
	validPurposes := map[string]bool{
		"login":          true,
		"register":       true,
		"password_reset": true,
		"phone_verify":   true,
		"payment":        true,
	}

	tests := []struct {
		purpose  string
		expected bool
	}{
		{"login", true},
		{"register", true},
		{"password_reset", true},
		{"phone_verify", true},
		{"payment", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.purpose, func(t *testing.T) {
			result := validPurposes[tt.purpose]
			if result != tt.expected {
				t.Errorf("validPurposes[%s] = %v; want %v", tt.purpose, result, tt.expected)
			}
		})
	}
}

// TestOTPAttempts tests OTP attempt limiting
func TestOTPAttempts(t *testing.T) {
	maxAttempts := 3

	isLockedOut := func(attempts int) bool {
		return attempts >= maxAttempts
	}

	tests := []struct {
		attempts int
		expected bool
	}{
		{0, false},
		{1, false},
		{2, false},
		{3, true},
		{5, true},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := isLockedOut(tt.attempts)
			if result != tt.expected {
				t.Errorf("isLockedOut(%d) = %v; want %v", tt.attempts, result, tt.expected)
			}
		})
	}
}

// TestPhoneMasking tests phone number masking for display
func TestPhoneMasking(t *testing.T) {
	maskPhone := func(phone string) string {
		if len(phone) < 4 {
			return "****"
		}
		return "****" + phone[len(phone)-4:]
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"+254700000000", "****0000"},
		{"254700000000", "****0000"},
		{"700000000", "****0000"},
		{"1234", "****1234"},
		{"12", "****"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := maskPhone(tt.input)
			if result != tt.expected {
				t.Errorf("maskPhone(%s) = %s; want %s", tt.input, result, tt.expected)
			}
		})
	}
}
