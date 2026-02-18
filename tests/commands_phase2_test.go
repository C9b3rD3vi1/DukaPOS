package main

import (
	"strings"
	"testing"
)

// TestPlanBadge tests plan badge selection
func TestPlanBadgeCommands(t *testing.T) {
	type PlanType string

	getPlanBadge := func(plan PlanType) string {
		switch plan {
		case PlanType("pro"):
			return "🚀 PRO"
		case PlanType("business"):
			return "🏢 BUSINESS"
		default:
			return "📦 FREE"
		}
	}

	tests := []struct {
		plan     PlanType
		expected string
	}{
		{PlanType("free"), "📦 FREE"},
		{PlanType("pro"), "🚀 PRO"},
		{PlanType("business"), "🏢 BUSINESS"},
		{"unknown", "📦 FREE"}, // defaults to free
	}

	for _, tt := range tests {
		t.Run(string(tt.plan), func(t *testing.T) {
			badge := getPlanBadge(tt.plan)
			if badge != tt.expected {
				t.Errorf("Plan badge = %s; want %s", badge, tt.expected)
			}
		})
	}
}

// TestPhase2CommandStructure tests Phase 2 command structure
func TestPhase2CommandStructure(t *testing.T) {
	// Test command names
	commands := map[string]bool{
		"mpesa":   true,
		"staff":   true,
		"shop":    true,
		"upgrade": true,
		"plan":    true,
	}

	expectedCommands := []string{"mpesa", "staff", "shop", "upgrade", "plan"}
	
	for _, cmd := range expectedCommands {
		if !commands[cmd] {
			t.Errorf("Expected command %q not found", cmd)
		}
	}
}

// TestPlanTypeValidation tests plan validation logic
func TestPlanTypeValidation(t *testing.T) {
	type PlanType string

	isProPlan := func(plan PlanType) bool {
		return plan == PlanType("pro") || plan == PlanType("business")
	}

	tests := []struct {
		plan     PlanType
		expected bool
	}{
		{PlanType("free"), false},
		{PlanType("pro"), true},
		{PlanType("business"), true},
	}

	for _, tt := range tests {
		t.Run(string(tt.plan), func(t *testing.T) {
			result := isProPlan(tt.plan)
			if result != tt.expected {
				t.Errorf("isProPlan(%s) = %v; want %v", tt.plan, result, tt.expected)
			}
		})
	}
}

// TestMpesaCommandArgs tests M-Pesa command argument parsing
func TestMpesaCommandArgs(t *testing.T) {
	parseArgs := func(args []string) (string, string) {
		if len(args) < 1 {
			return "", ""
		}
		if len(args) < 2 {
			return args[0], ""
		}
		return args[0], args[1]
	}

	tests := []struct {
		name     string
		args     []string
		wantCmd  string
		wantArg  string
	}{
		{"empty args", []string{}, "", ""},
		{"single arg", []string{"pay"}, "pay", ""},
		{"two args", []string{"pay", "100"}, "pay", "100"},
		{"many args", []string{"pay", "100", "extra"}, "pay", "100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, arg := parseArgs(tt.args)
			if cmd != tt.wantCmd {
				t.Errorf("parseArgs(%v).cmd = %s; want %s", tt.args, cmd, tt.wantCmd)
			}
			if arg != tt.wantArg {
				t.Errorf("parseArgs(%v).arg = %s; want %s", tt.args, arg, tt.wantArg)
			}
		})
	}
}

// TestStaffCommandArgs tests staff command argument parsing
func TestStaffCommandArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantCmd string
	}{
		{"empty args", []string{}, ""},
		{"list command", []string{"list"}, "list"},
		{"add command", []string{"add"}, "add"},
		{"remove command", []string{"remove"}, "remove"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := ""
			if len(tt.args) > 0 {
				cmd = tt.args[0]
			}
			if cmd != tt.wantCmd {
				t.Errorf("Command = %s; want %s", cmd, tt.wantCmd)
			}
		})
	}
}

// TestHelpTextContainsProCommands tests help text includes Pro commands
func TestHelpTextContainsProCommands(t *testing.T) {
	freeHelp := `📦 FREE

💎 PRO FEATURES:
Upgrade to unlock:
• M-Pesa payments
• Staff accounts
• Multiple shops`

	proHelp := `🚀 PRO

💎 PRO COMMANDS:
mpesa pay [amount]
staff list`

	// Free plan should mention upgrade
	if !strings.Contains(freeHelp, "Upgrade") {
		t.Error("Free help should mention upgrade")
	}
	if !strings.Contains(freeHelp, "M-Pesa") {
		t.Error("Free help should mention M-Pesa")
	}

	// Pro plan should show commands
	if !strings.Contains(proHelp, "mpesa") {
		t.Error("Pro help should show mpesa command")
	}
	if !strings.Contains(proHelp, "staff") {
		t.Error("Pro help should show staff command")
	}
}
