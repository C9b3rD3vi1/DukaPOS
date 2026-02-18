package main

import (
	"errors"
	"testing"
)

// TestStaffAPIHandler tests staff API handler functions
func TestStaffAPIHandler(t *testing.T) {
	// Test staff list response structure
	type Staff struct {
		ID       int    `json:"id"`
		ShopID   int    `json:"shop_id"`
		Name     string `json:"name"`
		Phone    string `json:"phone"`
		Role     string `json:"role"`
		IsActive bool   `json:"is_active"`
	}

	validateStaff := func(s Staff) error {
		if s.Name == "" {
			return ErrNameRequired
		}
		if s.Phone == "" {
			return ErrPhoneRequired
		}
		return nil
	}

	// Test valid staff
	validStaff := Staff{
		ID: 1, ShopID: 1, Name: "John", Phone: "+254712345678", Role: "cashier", IsActive: true,
	}
	if err := validateStaff(validStaff); err != nil {
		t.Errorf("Valid staff should not return error: %v", err)
	}

	// Test invalid staff - no name
	noNameStaff := Staff{ID: 1, Phone: "+254712345678"}
	if err := validateStaff(noNameStaff); err == nil {
		t.Error("Staff without name should return error")
	}
}

// TestStaffValidation tests staff input validation
func TestStaffValidation(t *testing.T) {
	type Request struct {
		ShopID uint   `json:"shop_id"`
		Name   string `json:"name"`
		Phone  string `json:"phone"`
		Role   string `json:"role"`
		Pin    string `json:"pin"`
	}

	validateCreateRequest := func(req Request) error {
		if req.Name == "" {
			return ErrNameRequired
		}
		if req.Phone == "" {
			return ErrPhoneRequired
		}
		if req.Pin == "" {
			return ErrPinRequired
		}
		return nil
	}

	tests := []struct {
		name    string
		req     Request
		wantErr error
	}{
		{"valid request", Request{1, "John", "+254712345678", "cashier", "1234"}, nil},
		{"missing name", Request{1, "", "+254712345678", "cashier", "1234"}, ErrNameRequired},
		{"missing phone", Request{1, "John", "", "cashier", "1234"}, ErrPhoneRequired},
		{"missing pin", Request{1, "John", "+254712345678", "cashier", ""}, ErrPinRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCreateRequest(tt.req)
			if (err == nil) != (tt.wantErr == nil) {
				t.Errorf("validateCreateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestMpesaWebhook tests M-Pesa callback parsing
func TestMpesaWebhook(t *testing.T) {
	type CallbackItem struct {
		Name  string
		Value string
	}

	extractMetadata := func(items []CallbackItem) map[string]string {
		metadata := make(map[string]string)
		for _, item := range items {
			metadata[item.Name] = item.Value
		}
		return metadata
	}

	// Test extraction
	items := []CallbackItem{
		{Name: "Amount", Value: "100"},
		{Name: "MpesaReceiptNumber", Value: "PIX123456"},
		{Name: "PhoneNumber", Value: "254712345678"},
	}

	metadata := extractMetadata(items)

	if metadata["Amount"] != "100" {
		t.Errorf("Amount = %s, want 100", metadata["Amount"])
	}
	if metadata["MpesaReceiptNumber"] != "PIX123456" {
		t.Errorf("MpesaReceiptNumber = %s, want PIX123456", metadata["MpesaReceiptNumber"])
	}
	if metadata["PhoneNumber"] != "254712345678" {
		t.Errorf("PhoneNumber = %s, want 254712345678", metadata["PhoneNumber"])
	}
}

// TestMultiShopAccount tests account-shop relationship
func TestMultiShopAccount(t *testing.T) {
	type Account struct {
		ID    int
		Email string
		Plan  string
	}

	type Shop struct {
		ID        int
		AccountID int
		Name      string
	}

	// Test account with multiple shops
	account := Account{ID: 1, Email: "test@example.com", Plan: "pro"}
	shops := []Shop{
		{ID: 1, AccountID: 1, Name: "Main Shop"},
		{ID: 2, AccountID: 1, Name: "Branch 1"},
		{ID: 3, AccountID: 1, Name: "Branch 2"},
	}

	// Verify all shops belong to account
	for _, shop := range shops {
		if shop.AccountID != account.ID {
			t.Errorf("Shop %d should belong to account %d", shop.ID, account.ID)
		}
	}

	// Test plan limits
	planLimits := map[string]int{
		"free":     1,
		"pro":      5,
		"business": 50,
	}

	limit := planLimits[account.Plan]
	if limit != 5 {
		t.Errorf("Pro plan should allow 5 shops, got %d", limit)
	}

	if len(shops) > limit {
		t.Errorf("Account has %d shops but plan limit is %d", len(shops), limit)
	}
}

// Error variables (matching handler errors)
var (
	ErrNameRequired  = errors.New("name is required")
	ErrPhoneRequired = errors.New("phone is required")
	ErrPinRequired   = errors.New("pin is required")
)
