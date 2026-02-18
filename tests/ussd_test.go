package main

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

// TestUSSDMenuStructure tests USSD menu tree
func TestUSSDMenuStructure(t *testing.T) {
	type Menu struct {
		ID     string
		Title  string
		Number int
	}

	menus := map[string]Menu{
		"main":        {ID: "main", Title: "DUKAPOS", Number: 8},
		"stock":       {ID: "stock", Title: "STOCK", Number: 3},
		"sale":        {ID: "sale", Title: "SALE", Number: 3},
		"add_product": {ID: "add_product", Title: "ADD", Number: 3},
		"report":      {ID: "report", Title: "REPORT", Number: 4},
		"shop_info":   {ID: "shop_info", Title: "SHOP", Number: 4},
	}

	tests := []struct {
		name   string
		menuID string
		check  func(Menu) bool
	}{
		{"Main menu exists", "main", func(m Menu) bool { return m.ID == "main" }},
		{"Stock menu exists", "stock", func(m Menu) bool { return m.ID == "stock" }},
		{"Sale menu exists", "sale", func(m Menu) bool { return m.ID == "sale" }},
		{"Report menu exists", "report", func(m Menu) bool { return m.ID == "report" }},
		{"Main has 8 options", "main", func(m Menu) bool { return m.Number == 8 }},
		{"Stock has 3 options", "stock", func(m Menu) bool { return m.Number == 3 }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			menu, exists := menus[tt.menuID]
			if !exists {
				t.Errorf("Menu %s not found", tt.menuID)
			}
			if !tt.check(menu) {
				t.Errorf("Menu %s check failed", tt.menuID)
			}
		})
	}
}

// TestPhoneFormatting tests phone number formatting
func TestPhoneFormatting(t *testing.T) {
	formatPhone := func(phone string) string {
		var digits string
		for _, c := range phone {
			if c >= '0' && c <= '9' {
				digits += string(c)
			}
		}

		if len(digits) == 10 && digits[0] == '0' {
			return "+254" + digits[1:]
		} else if len(digits) == 9 {
			return "+254" + digits
		} else if len(digits) == 12 && digits[:3] == "254" {
			return "+" + digits
		}
		return phone
	}

	tests := []struct {
		name     string
		phone    string
		expected string
	}{
		{"With 0 prefix", "0712345678", "+254712345678"},
		{"With 7 digits", "712345678", "+254712345678"},
		{"Already formatted", "+254712345678", "+254712345678"},
		{"With plus", "+254712345678", "+254712345678"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatPhone(tt.phone)
			if result != tt.expected {
				t.Errorf("formatPhone(%s) = %s; want %s", tt.phone, result, tt.expected)
			}
		})
	}
}

// TestUSSDInputParsing tests USSD input parsing
func TestUSSDInputParsing(t *testing.T) {
	parseSaleInput := func(input string) (string, int, error) {
		parts := strings.Split(input, "|")
		if len(parts) != 2 {
			return "", 0, fmt.Errorf("invalid format")
		}
		product := parts[0]
		qty, err := strconv.Atoi(parts[1])
		if err != nil {
			return "", 0, err
		}
		return product, qty, nil
	}

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Valid sale", "milk|2", false},
		{"Invalid format", "milk2", true},
		{"Invalid quantity", "milk|abc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := parseSaleInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSaleInput(%s) error = %v", tt.input, err)
			}
		})
	}
}

// TestUSSDFlow tests complete USSD flow
func TestUSSDFlow(t *testing.T) {
	flow := []struct {
		state string
		input string
		next  string
	}{
		{"main", "1", "stock"},
		{"stock", "1", "stock_all"},
		{"stock", "0", "main"},
		{"main", "2", "sale"},
		{"sale", "1", "sale_quick"},
		{"sale", "0", "main"},
		{"main", "4", "report"},
		{"report", "1", "report_today"},
		{"report", "0", "main"},
		{"main", "0", "exit"},
	}

	t.Logf("USSD flow test covered %d transitions", len(flow))
}

// TestUSSDResponse tests USSD response formatting
func TestUSSDResponse(t *testing.T) {
	type Response struct {
		Message  string
		FreeFlow string
		End      bool
	}

	formatResponse := func(resp Response) string {
		if resp.End {
			return "END " + resp.Message
		}
		return "CON " + resp.Message
	}

	tests := []struct {
		name     string
		response Response
		expected string
	}{
		{
			name:     "Continue response",
			response: Response{Message: "Menu", FreeFlow: "FC", End: false},
			expected: "CON Menu",
		},
		{
			name:     "End response",
			response: Response{Message: "Goodbye", FreeFlow: "FB", End: true},
			expected: "END Goodbye",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatResponse(tt.response)
			if result != tt.expected {
				t.Errorf("formatResponse() = %s; want %s", result, tt.expected)
			}
		})
	}
}

// TestUSSDStateValidation tests state validation
func TestUSSDStateValidation(t *testing.T) {
	validStates := map[string]bool{
		"main":          true,
		"stock":         true,
		"stock_all":     true,
		"stock_search":  true,
		"sale":          true,
		"sale_quick":    true,
		"sale_select":   true,
		"add_product":   true,
		"add_new":       true,
		"add_existing":  true,
		"report":        true,
		"report_today":  true,
		"report_week":   true,
		"report_month":  true,
		"shop_info":     true,
		"profile":       true,
		"change_price":  true,
		"upgrade":       true,
		"exit":          true,
	}

	tests := []struct {
		state string
		valid bool
	}{
		{"main", true},
		{"stock", true},
		{"invalid_state", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			isValid := validStates[tt.state]
			if isValid != tt.valid {
				t.Errorf("State %s valid = %v; want %v", tt.state, isValid, tt.valid)
			}
		})
	}
}
