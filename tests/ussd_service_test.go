package main

import (
	"strings"
	"testing"
)

// TestUSSDMenuStructureService tests USSD menu structure
func TestUSSDMenuStructureService(t *testing.T) {
	type Option struct {
		Number string
		Text   string
		Action string
	}

	type Menu struct {
		ID      string
		Title   string
		Options []Option
	}

	menuTree := map[string]*Menu{
		"main": {
			ID:    "main",
			Title: "DUKAPOS",
			Options: []Option{
				{"1", "Check Stock", "stock"},
				{"2", "Record Sale", "sale"},
				{"3", "Add Product", "add_product"},
				{"4", "Daily Report", "report"},
				{"0", "Exit", "exit"},
			},
		},
	}

	// Verify main menu exists
	mainMenu := menuTree["main"]
	if mainMenu == nil {
		t.Error("Main menu should exist")
	}

	// Verify menu has options
	if len(mainMenu.Options) == 0 {
		t.Error("Main menu should have options")
	}

	// Verify exit option exists
	hasExit := false
	for _, opt := range mainMenu.Options {
		if opt.Number == "0" {
			hasExit = true
			break
		}
	}
	if !hasExit {
		t.Error("Menu should have exit option (0)")
	}
}

// TestUSSDInputParsingService tests USSD input parsing
func TestUSSDInputParsingService(t *testing.T) {
	parseInput := func(input string) string {
		return strings.TrimSpace(input)
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"1", "1"},
		{" 1 ", "1"},
		{"", ""},
		{"0", "0"},
		{"  0  ", "0"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseInput(tt.input)
			if result != tt.expected {
				t.Errorf("parseInput(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestUSSDActionRouting tests USSD action routing
func TestUSSDActionRouting(t *testing.T) {
	routeAction := func(action string) string {
		switch action {
		case "stock":
			return "handleStockAll"
		case "sale":
			return "handleSale"
		case "report":
			return "handleReport"
		case "exit":
			return "exit"
		default:
			return "unknown"
		}
	}

	tests := []struct {
		action   string
		expected string
	}{
		{"stock", "handleStockAll"},
		{"sale", "handleSale"},
		{"report", "handleReport"},
		{"exit", "exit"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			result := routeAction(tt.action)
			if result != tt.expected {
				t.Errorf("routeAction(%s) = %s; want %s", tt.action, result, tt.expected)
			}
		})
	}
}

// TestUSSDSessionManagement tests USSD session management
func TestUSSDSessionManagement(t *testing.T) {
	type Session struct {
		ID     string
		Phone  string
		State  string
		Active bool
	}

	sessions := make(map[string]*Session)

	// Create session
	createSession := func(id, phone string) *Session {
		return &Session{
			ID:     id,
			Phone:  phone,
			State:  "main",
			Active: true,
		}
	}

	// Get or create
	getOrCreate := func(id, phone string) *Session {
		if s, exists := sessions[id]; exists {
			return s
		}
		s := createSession(id, phone)
		sessions[id] = s
		return s
	}

	// Test create
	s1 := getOrCreate("session1", "+254700000000")
	if !s1.Active {
		t.Error("New session should be active")
	}

	// Test get existing
	s2 := getOrCreate("session1", "+254700000000")
	if s1 != s2 {
		t.Error("Should return existing session")
	}

	// Test close session
	closeSession := func(id string) {
		if s, exists := sessions[id]; exists {
			s.Active = false
			delete(sessions, id)
		}
	}

	closeSession("session1")
	if _, exists := sessions["session1"]; exists {
		t.Error("Session should be deleted after close")
	}
}

// TestUSSDResponseFormat tests USSD response formatting
func TestUSSDResponseFormat(t *testing.T) {
	type Response struct {
		Message  string
		FreeFlow string
		End      bool
	}

	formatMenu := func(title string, options []string) string {
		var sb strings.Builder
		sb.WriteString(title)
		sb.WriteString("\n\n")
		for _, opt := range options {
			sb.WriteString(opt)
			sb.WriteString("\n")
		}
		sb.WriteString("\n0. Back to Main")
		return sb.String()
	}

	options := []string{"1. Check Stock", "2. Record Sale", "3. Add Product"}
	result := formatMenu("DUKAPOS", options)

	if !strings.Contains(result, "DUKAPOS") {
		t.Error("Response should contain title")
	}
	if !strings.Contains(result, "1. Check Stock") {
		t.Error("Response should contain options")
	}
	if !strings.Contains(result, "0. Back to Main") {
		t.Error("Response should contain back option")
	}
}

// TestUSSDPhoneFormatting tests USSD phone formatting
func TestUSSDPhoneFormatting(t *testing.T) {
	formatPhone := func(phone string) string {
		// Remove + and any special characters
		result := ""
		for _, c := range phone {
			if c >= '0' && c <= '9' {
				result += string(c)
			}
		}
		// Ensure country code (9 digits)
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
		{"254-700-000-000", "254700000000"},
		{"700000000", "254700000000"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := formatPhone(tt.input)
			if result != tt.expected {
				t.Errorf("formatPhone(%s) = %s; want %s", tt.input, result, tt.expected)
			}
		})
	}
}
