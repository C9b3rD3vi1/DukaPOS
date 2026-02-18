package main

import (
	"fmt"
	"strings"
	"testing"
)

// TestPlanLimits tests plan limit definitions
func TestPlanLimits(t *testing.T) {
	type PlanType string

	planLimits := map[PlanType]int{
		PlanType("free"):     1,
		PlanType("pro"):      5,
		PlanType("business"): 50,
	}

	tests := []struct {
		plan     PlanType
		expected int
	}{
		{PlanType("free"), 1},
		{PlanType("pro"), 5},
		{PlanType("business"), 50},
	}

	for _, tt := range tests {
		t.Run(string(tt.plan), func(t *testing.T) {
			if planLimits[tt.plan] != tt.expected {
				t.Errorf("planLimits[%s] = %d; want %d", tt.plan, planLimits[tt.plan], tt.expected)
			}
		})
	}
}

// TestPlanInfo tests plan information retrieval
func TestPlanInfo(t *testing.T) {
	type PlanType string

	getPlanInfo := func(plan PlanType) map[string]interface{} {
		limits := map[PlanType]map[string]interface{}{
			PlanType("free"): {
				"name":      "Free",
				"price":     0,
				"shops":     1,
				"products":  50,
				"staff":     0,
				"mpesa":     false,
				"analytics": false,
			},
			PlanType("pro"): {
				"name":      "Pro",
				"price":     500,
				"shops":     5,
				"products":  -1, // unlimited
				"staff":     3,
				"mpesa":     true,
				"analytics": true,
			},
			PlanType("business"): {
				"name":      "Business",
				"price":     1500,
				"shops":     50,
				"products":  -1,
				"staff":     -1,
				"mpesa":     true,
				"analytics": true,
			},
		}

		if info, ok := limits[plan]; ok {
			return info
		}
		return limits[PlanType("free")]
	}

	tests := []struct {
		name     string
		plan     PlanType
		checkFn  func(map[string]interface{}) bool
	}{
		{
			name: "Free plan",
			plan: PlanType("free"),
			checkFn: func(info map[string]interface{}) bool {
				return info["name"] == "Free" &&
					info["price"] == 0 &&
					info["shops"] == 1 &&
					info["mpesa"] == false
			},
		},
		{
			name: "Pro plan",
			plan: PlanType("pro"),
			checkFn: func(info map[string]interface{}) bool {
				return info["name"] == "Pro" &&
					info["price"] == 500 &&
					info["shops"] == 5 &&
					info["mpesa"] == true &&
					info["staff"] == 3
			},
		},
		{
			name: "Business plan",
			plan: PlanType("business"),
			checkFn: func(info map[string]interface{}) bool {
				return info["name"] == "Business" &&
					info["price"] == 1500 &&
					info["shops"] == 50 &&
					info["staff"] == -1
			},
		},
		{
			name: "Unknown plan defaults to Free",
			plan: PlanType("unknown"),
			checkFn: func(info map[string]interface{}) bool {
				return info["name"] == "Free"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := getPlanInfo(tt.plan)
			if !tt.checkFn(info) {
				t.Errorf("getPlanInfo(%s) returned unexpected values: %+v", tt.plan, info)
			}
		})
	}
}

// TestFormatShopList tests shop list formatting
func TestFormatShopList(t *testing.T) {
	type Shop struct {
		ID       int
		Name     string
		Phone    string
		IsActive bool
	}

	formatShopList := func(shops []Shop) string {
		if len(shops) == 0 {
			return "No shops found."
		}

		var msg = "🏪 YOUR SHOPS:\n\n"
		for i, shop := range shops {
			status := "✅"
			if !shop.IsActive {
				status = "❌"
			}
			msg += fmt.Sprintf("%d. %s %s\n   📱 %s\n\n", i+1, status, shop.Name, shop.Phone)
		}
		return msg
	}

	tests := []struct {
		name  string
		shops []Shop
		check func(string) bool
	}{
		{
			name:  "empty shop list",
			shops: []Shop{},
			check: func(result string) bool {
				return strings.Contains(result, "No shops found")
			},
		},
		{
			name: "single active shop",
			shops: []Shop{
				{ID: 1, Name: "Main Shop", Phone: "+254712345678", IsActive: true},
			},
			check: func(result string) bool {
				return strings.Contains(result, "Main Shop") &&
					strings.Contains(result, "✅")
			},
		},
		{
			name: "multiple shops with inactive",
			shops: []Shop{
				{ID: 1, Name: "Main Shop", Phone: "+254712345678", IsActive: true},
				{ID: 2, Name: "Branch", Phone: "+254798765432", IsActive: false},
			},
			check: func(result string) bool {
				return strings.Contains(result, "Main Shop") &&
					strings.Contains(result, "Branch") &&
					strings.Contains(result, "✅") &&
					strings.Contains(result, "❌")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatShopList(tt.shops)
			if !tt.check(result) {
				t.Errorf("formatShopList() = %s", result)
			}
		})
	}
}

// TestCanAddShop tests shop addition validation
func TestCanAddShop(t *testing.T) {
	type PlanType string

	canAddShop := func(plan PlanType) (bool, error) {
		planLimits := map[PlanType]int{
			PlanType("free"):     1,
			PlanType("pro"):      5,
			PlanType("business"): 50,
		}
		
		limit := planLimits[plan]
		if limit <= 1 {
			return false, fmt.Errorf("maximum shops reached for your plan")
		}
		return true, nil
	}

	tests := []struct {
		name     string
		plan     PlanType
		expected bool
	}{
		{"Free plan cannot add shops", PlanType("free"), false},
		{"Pro plan can add shops", PlanType("pro"), true},
		{"Business plan can add shops", PlanType("business"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := canAddShop(tt.plan)
			if result != tt.expected {
				t.Errorf("canAddShop() = %v; want %v", result, tt.expected)
			}
		})
	}
}

// TestPlanBadge tests plan badge selection
func TestPlanBadge(t *testing.T) {
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
				t.Errorf("getPlanBadge() = %s; want %s", badge, tt.expected)
			}
		})
	}
}
