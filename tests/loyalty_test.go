package main

import (
	"testing"
)

// TestLoyaltyPointsCalculation tests points earned calculation
func TestLoyaltyPointsCalculation(t *testing.T) {
	calculatePoints := func(amount float64) int {
		points := int(amount * 1.0) // 1 point per KSh
		
		// Bonus for larger purchases
		if amount >= 1000 {
			points = int(float64(points) * 1.25)
		} else if amount >= 500 {
			points = int(float64(points) * 1.10)
		}
		
		return points
	}

	tests := []struct {
		name     string
		amount   float64
		expected int
	}{
		{"small purchase", 100, 100},
		{"medium purchase", 500, 550},   // 10% bonus = 500 + 50 = 550
		{"large purchase", 1000, 1250},   // 25% bonus = 1000 + 250 = 1250
		{"very large", 2000, 2500},        // 25% bonus = 2000 + 500 = 2500
		{"zero", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculatePoints(tt.amount)
			if result != tt.expected {
				t.Errorf("calculatePoints(%v) = %d; want %d", tt.amount, result, tt.expected)
			}
		})
	}
}

// TestRedeemedValue tests points redemption value
func TestRedeemedValue(t *testing.T) {
	calculateValue := func(points int) float64 {
		return float64(points) * 0.10 // 10 points = KSh 1
	}

	tests := []struct {
		points   int
		expected float64
	}{
		{0, 0},
		{10, 1.0},
		{100, 10.0},
		{500, 50.0},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.points)), func(t *testing.T) {
			result := calculateValue(tt.points)
			if result != tt.expected {
				t.Errorf("calculateValue(%d) = %v; want %v", tt.points, result, tt.expected)
			}
		})
	}
}

// TestTierDetermination tests customer tier determination
func TestTierDetermination(t *testing.T) {
	determineTier := func(totalSpent float64) string {
		switch {
		case totalSpent >= 100000:
			return "platinum"
		case totalSpent >= 50000:
			return "gold"
		case totalSpent >= 20000:
			return "silver"
		default:
			return "bronze"
		}
	}

	tests := []struct {
		totalSpent float64
		expected   string
	}{
		{0, "bronze"},
		{10000, "bronze"},
		{20000, "silver"},
		{25000, "silver"},
		{50000, "gold"},
		{75000, "gold"},
		{100000, "platinum"},
		{150000, "platinum"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := determineTier(tt.totalSpent)
			if result != tt.expected {
				t.Errorf("determineTier(%v) = %s; want %s", tt.totalSpent, result, tt.expected)
			}
		})
	}
}

// TestTierMultiplier tests tier-based points multiplier
func TestTierMultiplier(t *testing.T) {
	getMultiplier := func(tier string) float64 {
		switch tier {
		case "platinum":
			return 2.0
		case "gold":
			return 1.5
		case "silver":
			return 1.25
		default:
			return 1.0
		}
	}

	tests := []struct {
		tier     string
		expected float64
	}{
		{"bronze", 1.0},
		{"silver", 1.25},
		{"gold", 1.5},
		{"platinum", 2.0},
		{"unknown", 1.0}, // defaults to bronze
	}

	for _, tt := range tests {
		t.Run(tt.tier, func(t *testing.T) {
			result := getMultiplier(tt.tier)
			if result != tt.expected {
				t.Errorf("getMultiplier(%s) = %v; want %v", tt.tier, result, tt.expected)
			}
		})
	}
}
