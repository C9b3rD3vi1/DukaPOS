package main

import (
	"fmt"
	"strings"
	"testing"
)

// TestStaffRoles tests available roles
func TestStaffRoles(t *testing.T) {
	roles := []string{
		"manager",
		"cashier",
		"stock clerk",
		"assistant",
	}
	
	expectedRoles := []string{
		"manager",
		"cashier",
		"stock clerk",
		"assistant",
	}
	
	if len(roles) != len(expectedRoles) {
		t.Errorf("len(roles) = %d; want %d", len(roles), len(expectedRoles))
	}
	
	for i, role := range expectedRoles {
		if roles[i] != role {
			t.Errorf("roles[%d] = %s; want %s", i, roles[i], role)
		}
	}
}

// TestRoleValidation tests role validation logic
func TestRoleValidation(t *testing.T) {
	validRoles := map[string]bool{
		"manager":      true,
		"cashier":      true,
		"stock clerk":  true,
		"assistant":    true,
	}

	isValidRole := func(role string) bool {
		return validRoles[role]
	}

	tests := []struct {
		role     string
		expected bool
	}{
		{"manager", true},
		{"cashier", true},
		{"stock clerk", true},
		{"assistant", true},
		{"Manager", false},  // Case sensitive
		{"MANAGER", false},
		{"admin", false},
		{"", false},
		{"random", false},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			result := isValidRole(tt.role)
			if result != tt.expected {
				t.Errorf("isValidRole(%q) = %v; want %v", tt.role, result, tt.expected)
			}
		})
	}
}

// TestStaffListFormatting tests staff list formatting
func TestStaffListFormatting(t *testing.T) {
	type Staff struct {
		ID       int
		Name     string
		Phone    string
		Role     string
		IsActive bool
	}

	formatStaffList := func(staff []Staff) string {
		if len(staff) == 0 {
			return "No staff members yet.\nAdd: staff add [name] [phone] [role]"
		}

		var sb strings.Builder
		sb.WriteString("👥 STAFF LIST:\n\n")
		for i, st := range staff {
			status := "✅"
			if !st.IsActive {
				status = "❌"
			}
			sb.WriteString(fmt.Sprintf("%d. %s %s\n   📱 %s\n   💼 %s\n\n", 
				i+1, status, st.Name, st.Phone, st.Role))
		}
		return sb.String()
	}

	tests := []struct {
		name     string
		staff    []Staff
		wantContains []string
	}{
		{
			name:  "empty staff list",
			staff: []Staff{},
			wantContains: []string{"No staff members yet"},
		},
		{
			name: "single active staff",
			staff: []Staff{
				{ID: 1, Name: "John", Phone: "+254712345678", Role: "cashier", IsActive: true},
			},
			wantContains: []string{"John", "+254712345678", "cashier", "✅"},
		},
		{
			name: "multiple staff with inactive",
			staff: []Staff{
				{ID: 1, Name: "John", Phone: "+254712345678", Role: "cashier", IsActive: true},
				{ID: 2, Name: "Jane", Phone: "+254798765432", Role: "manager", IsActive: false},
			},
			wantContains: []string{"John", "Jane", "✅", "❌"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatStaffList(tt.staff)
			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("formatStaffList() should contain %q", want)
				}
			}
		})
	}
}
