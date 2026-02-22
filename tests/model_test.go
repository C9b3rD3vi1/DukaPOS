package main

import (
	"strings"
	"testing"
)

// TestProductValidation tests product model validation
func TestProductValidation(t *testing.T) {
	type Product struct {
		Name         string
		SellingPrice float64
		CostPrice    float64
		CurrentStock int
	}

	validateProduct := func(p Product) (bool, string) {
		if p.Name == "" {
			return false, "product name is required"
		}
		if p.SellingPrice <= 0 {
			return false, "selling price must be greater than 0"
		}
		if p.CostPrice < 0 {
			return false, "cost price cannot be negative"
		}
		if p.SellingPrice < p.CostPrice {
			return false, "selling price cannot be less than cost price"
		}
		if p.CurrentStock < 0 {
			return false, "stock cannot be negative"
		}
		return true, ""
	}

	tests := []struct {
		name    string
		product Product
		valid   bool
	}{
		{
			name:    "valid product",
			product: Product{Name: "Bread", SellingPrice: 50, CostPrice: 30, CurrentStock: 10},
			valid:   true,
		},
		{
			name:    "empty name",
			product: Product{Name: "", SellingPrice: 50, CostPrice: 30, CurrentStock: 10},
			valid:   false,
		},
		{
			name:    "zero price",
			product: Product{Name: "Bread", SellingPrice: 0, CostPrice: 30, CurrentStock: 10},
			valid:   false,
		},
		{
			name:    "negative cost",
			product: Product{Name: "Bread", SellingPrice: 50, CostPrice: -10, CurrentStock: 10},
			valid:   false,
		},
		{
			name:    "selling less than cost",
			product: Product{Name: "Bread", SellingPrice: 30, CostPrice: 50, CurrentStock: 10},
			valid:   false,
		},
		{
			name:    "negative stock",
			product: Product{Name: "Bread", SellingPrice: 50, CostPrice: 30, CurrentStock: -5},
			valid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, _ := validateProduct(tt.product)
			if valid != tt.valid {
				t.Errorf("validateProduct(%s) = %v; want %v", tt.name, valid, tt.valid)
			}
		})
	}
}

// TestSaleValidation tests sale model validation
func TestSaleValidation(t *testing.T) {
	type Sale struct {
		Quantity      int
		UnitPrice     float64
		TotalAmount   float64
		PaymentMethod string
	}

	validateSale := func(s Sale) (bool, string) {
		if s.Quantity <= 0 {
			return false, "quantity must be greater than 0"
		}
		if s.UnitPrice <= 0 {
			return false, "unit price must be greater than 0"
		}
		if s.TotalAmount != s.UnitPrice*float64(s.Quantity) {
			return false, "total amount does not match"
		}
		validMethods := map[string]bool{"cash": true, "mpesa": true, "card": true, "bank": true}
		if !validMethods[s.PaymentMethod] {
			return false, "invalid payment method"
		}
		return true, ""
	}

	tests := []struct {
		name  string
		sale  Sale
		valid bool
	}{
		{
			name:  "valid sale",
			sale:  Sale{Quantity: 2, UnitPrice: 50, TotalAmount: 100, PaymentMethod: "cash"},
			valid: true,
		},
		{
			name:  "zero quantity",
			sale:  Sale{Quantity: 0, UnitPrice: 50, TotalAmount: 0, PaymentMethod: "cash"},
			valid: false,
		},
		{
			name:  "invalid payment",
			sale:  Sale{Quantity: 2, UnitPrice: 50, TotalAmount: 100, PaymentMethod: "crypto"},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, _ := validateSale(tt.sale)
			if valid != tt.valid {
				t.Errorf("validateSale(%s) = %v; want %v", tt.name, valid, tt.valid)
			}
		})
	}
}

// TestShopValidation tests shop model validation
func TestShopValidation(t *testing.T) {
	type Shop struct {
		Name  string
		Phone string
		Email string
		Plan  string
	}

	validateShop := func(s Shop) (bool, string) {
		if s.Name == "" {
			return false, "shop name is required"
		}
		if s.Phone == "" {
			return false, "phone number is required"
		}
		validPlans := map[string]bool{"free": true, "pro": true, "business": true}
		if !validPlans[s.Plan] {
			return false, "invalid plan"
		}
		return true, ""
	}

	tests := []struct {
		name  string
		shop  Shop
		valid bool
	}{
		{"valid shop", Shop{Name: "My Shop", Phone: "+254700000000", Plan: "free"}, true},
		{"empty name", Shop{Name: "", Phone: "+254700000000", Plan: "free"}, false},
		{"empty phone", Shop{Name: "My Shop", Phone: "", Plan: "free"}, false},
		{"invalid plan", Shop{Name: "My Shop", Phone: "+254700000000", Plan: "invalid"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, _ := validateShop(tt.shop)
			if valid != tt.valid {
				t.Errorf("validateShop(%s) = %v; want %v", tt.name, valid, tt.valid)
			}
		})
	}
}

// TestLoyaltyPoints tests loyalty points calculation
func TestLoyaltyPoints(t *testing.T) {
	type Customer struct {
		Tier          string
		TotalSpent    float64
		LoyaltyPoints int
	}

	calculatePoints := func(c *Customer) int {
		rate := 1.0
		switch c.Tier {
		case "silver":
			rate = 1.25
		case "gold":
			rate = 1.5
		case "platinum":
			rate = 2.0
		}
		return int(c.TotalSpent * rate)
	}

	tests := []struct {
		name     string
		tier     string
		spent    float64
		expected int
	}{
		{"bronze", "bronze", 1000, 1000},
		{"silver", "silver", 1000, 1250},
		{"gold", "gold", 1000, 1500},
		{"platinum", "platinum", 1000, 2000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Customer{Tier: tt.tier, TotalSpent: tt.spent}
			points := calculatePoints(c)
			if points != tt.expected {
				t.Errorf("calculatePoints() = %d; want %d", points, tt.expected)
			}
		})
	}
}

// TestStockCalculation tests inventory stock calculations
func TestStockCalculation(t *testing.T) {
	type StockOp struct {
		CurrentStock int
		Quantity     int
		Operation    string // "add", "sell", "remove"
	}

	calculateStock := func(op StockOp) int {
		switch op.Operation {
		case "add":
			return op.CurrentStock + op.Quantity
		case "sell":
			return op.CurrentStock - op.Quantity
		case "remove":
			return op.CurrentStock - op.Quantity
		default:
			return op.CurrentStock
		}
	}

	tests := []struct {
		name     string
		current  int
		quantity int
		op       string
		expected int
	}{
		{"add stock", 10, 5, "add", 15},
		{"sell stock", 10, 3, "sell", 7},
		{"remove stock", 10, 2, "remove", 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := StockOp{CurrentStock: tt.current, Quantity: tt.quantity, Operation: tt.op}
			result := calculateStock(op)
			if result != tt.expected {
				t.Errorf("calculateStock() = %d; want %d", result, tt.expected)
			}
		})
	}
}

// TestProfitCalculation tests profit calculation
func TestProfitCalculation(t *testing.T) {
	calculateProfit := func(sellingPrice, costPrice float64, quantity int) float64 {
		return (sellingPrice - costPrice) * float64(quantity)
	}

	profit := calculateProfit(50, 30, 2)
	if profit != 40 {
		t.Errorf("Profit = %.0f; want 40", profit)
	}

	// Edge case: zero quantity
	profit = calculateProfit(50, 30, 0)
	if profit != 0 {
		t.Errorf("Profit with zero qty = %.0f; want 0", profit)
	}
}

// TestPaymentMethodValidation tests payment method validation
func TestPaymentMethodValidation(t *testing.T) {
	validMethods := map[string]bool{
		"cash":   true,
		"mpesa":  true,
		"card":   true,
		"bank":   true,
		"credit": true,
	}

	isValidMethod := func(method string) bool {
		return validMethods[strings.ToLower(method)]
	}

	tests := []struct {
		method   string
		expected bool
	}{
		{"cash", true},
		{"CASH", true},
		{"mpesa", true},
		{"MPESA", true},
		{"crypto", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			result := isValidMethod(tt.method)
			if result != tt.expected {
				t.Errorf("isValidMethod(%s) = %v; want %v", tt.method, result, tt.expected)
			}
		})
	}
}
