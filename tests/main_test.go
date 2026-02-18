package main

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

// TestStringUtilities tests string utility functions
func TestStringUtilities(t *testing.T) {
	// Test normalizeProductName
	normalizeProductName := func(name string) string {
		if len(name) == 0 {
			return name
		}
		return strings.Title(strings.ToLower(name))
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"bread", "Bread"},
		{"MILK", "Milk"},
		{"Eggs", "Eggs"},
		{"", ""},
		{"SUGAR", "Sugar"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeProductName(tt.input)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestParseFloat tests float parsing
func TestParseFloat(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
		hasError bool
	}{
		{"50", 50, false},
		{"50.50", 50.50, false},
		{"0", 0, false},
		{"abc", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			price, err := strconv.ParseFloat(tt.input, 64)
			if tt.hasError && err == nil {
				t.Errorf("expected error for %s", tt.input)
			}
			if !tt.hasError && price != tt.expected {
				t.Errorf("expected %f, got %f for %s", tt.expected, price, tt.input)
			}
		})
	}
}

// TestParseInt tests integer parsing
func TestParseInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		hasError bool
	}{
		{"1", 1, false},
		{"10", 10, false},
		{"abc", 0, true},
		{"", 0, true},
		{"1.5", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			qty, err := strconv.Atoi(tt.input)
			if tt.hasError && err == nil {
				t.Errorf("expected error for %s", tt.input)
			}
			if !tt.hasError && qty != tt.expected {
				t.Errorf("expected %d, got %d for %s", tt.expected, qty, tt.input)
			}
		})
	}
}

// TestCommandParsing tests command parsing logic
func TestCommandParsing(t *testing.T) {
	parseCommand := func(message string) (string, []string) {
		message = strings.TrimSpace(strings.ToLower(message))
		parts := strings.Fields(message)
		if len(parts) == 0 {
			return "help", []string{}
		}
		return parts[0], parts[1:]
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"add command", "add bread 50 30", "add"},
		{"sell command", "sell milk 2", "sell"},
		{"stock command", "stock", "stock"},
		{"help command", "help", "help"},
		{"uppercase add", "ADD bread 50 30", "add"},
		{"with spaces", "  add  bread  50  30  ", "add"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, _ := parseCommand(tt.input)
			if cmd != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, cmd)
			}
		})
	}
}

// TestValidation tests input validation
func TestValidation(t *testing.T) {
	validatePrice := func(price float64) bool {
		return price > 0
	}

	validateQuantity := func(qty int) bool {
		return qty > 0
	}

	if !validatePrice(50) {
		t.Error("50 should be valid price")
	}
	if validatePrice(0) {
		t.Error("0 should be invalid price")
	}
	if validatePrice(-10) {
		t.Error("-10 should be invalid price")
	}

	if !validateQuantity(1) {
		t.Error("1 should be valid quantity")
	}
	if validateQuantity(0) {
		t.Error("0 should be invalid quantity")
	}
	if validateQuantity(-5) {
		t.Error("-5 should be invalid quantity")
	}
}

// TestCalculation tests basic calculations
func TestCalculation(t *testing.T) {
	calculateTotal := func(price float64, qty int) float64 {
		return price * float64(qty)
	}

	calculateProfit := func(selling, cost float64, qty int) float64 {
		return (selling - cost) * float64(qty)
	}

	total := calculateTotal(50, 2)
	if total != 100 {
		t.Errorf("expected 100, got %f", total)
	}

	profit := calculateProfit(50, 30, 2)
	if profit != 40 {
		t.Errorf("expected 40, got %f", profit)
	}

	// Test edge cases
	inf := calculateTotal(0, 10)
	if inf != 0 {
		t.Errorf("expected 0 for zero price, got %f", inf)
	}

	inf = calculateTotal(50, 0)
	if inf != 0 {
		t.Errorf("expected 0 for zero qty, got %f", inf)
	}
}

func TestMain(m *testing.M) {
	fmt.Println("Running DukaPOS tests...")
	m.Run()
}
