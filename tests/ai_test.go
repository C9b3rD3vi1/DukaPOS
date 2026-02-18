package main

import (
	"math"
	"testing"
)

// TestPredictRestock tests restock prediction logic
func TestPredictRestock(t *testing.T) {
	type SalesData struct {
		Date     string
		Quantity int
	}

	type Product struct {
		ID           int
		Name         string
		CurrentStock int
		SalesData    []SalesData
	}

	// Simulate prediction calculation
	predictRestock := func(product Product) map[string]interface{} {
		if len(product.SalesData) < 7 {
			return map[string]interface{}{
				"confidence": 0,
				"trend":       "insufficient_data",
			}
		}

		// Calculate average
		total := 0
		for _, s := range product.SalesData {
			total += s.Quantity
		}
		avgDaily := float64(total) / float64(len(product.SalesData))

		// Calculate days until stockout
		daysUntilStockout := 0
		if avgDaily > 0 {
			daysUntilStockout = int(math.Ceil(float64(product.CurrentStock) / avgDaily))
		}

		// Recommended order (7 days worth)
		recommendedOrder := int(math.Ceil(avgDaily * 7))

		return map[string]interface{}{
			"product_name":         product.Name,
			"current_stock":        product.CurrentStock,
			"days_until_stockout": daysUntilStockout,
			"recommended_order":    recommendedOrder,
			"confidence":           0.8,
			"trend":               "stable",
		}
	}

	// Test with sufficient data
	product1 := Product{
		ID:           1,
		Name:         "Milk",
		CurrentStock: 20,
		SalesData:    []SalesData{
			{"2024-01-01", 5}, {"2024-01-02", 5}, {"2024-01-03", 5},
			{"2024-01-04", 5}, {"2024-01-05", 5}, {"2024-01-06", 5},
			{"2024-01-07", 5},
		},
	}

	result1 := predictRestock(product1)
	if result1["days_until_stockout"] != 4 {
		t.Errorf("days_until_stockout = %v; want 4", result1["days_until_stockout"])
	}

	// Test with insufficient data
	product2 := Product{
		ID:           2,
		Name:         "Bread",
		CurrentStock: 10,
		SalesData:    []SalesData{
			{"2024-01-01", 5},
		},
	}

	result2 := predictRestock(product2)
	if result2["confidence"] != 0 {
		t.Errorf("confidence = %v; want 0 for insufficient data", result2["confidence"])
	}
}

// TestCalculateAverageDailySales tests sales average calculation
func TestCalculateAverageDailySales(t *testing.T) {
	type SalesData struct {
		Quantity int
	}

	calculateAvg := func(data []SalesData) float64 {
		if len(data) == 0 {
			return 0
		}
		total := 0.0
		for _, d := range data {
			total += float64(d.Quantity)
		}
		return total / float64(len(data))
	}

	tests := []struct {
		name     string
		data     []SalesData
		expected float64
	}{
		{"normal", []SalesData{{5}, {10}, {15}}, 10.0},
		{"empty", []SalesData{}, 0},
		{"single", []SalesData{{10}}, 10.0},
		{"all zeros", []SalesData{{0}, {0}, {0}}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateAvg(tt.data)
			if result != tt.expected {
				t.Errorf("calculateAvg() = %v; want %v", result, tt.expected)
			}
		})
	}
}

// TestSeasonalityMultiplier tests seasonal adjustment
func TestSeasonalityMultiplier(t *testing.T) {
	type Month int

	getMonthFactor := func(month Month) float64 {
		factors := map[Month]float64{
			1:  0.9,  // January
			4:  1.1,  // April
			11: 1.2,  // November
			12: 1.3,  // December
		}
		if f, ok := factors[month]; ok {
			return f
		}
		return 1.0
	}

	tests := []struct {
		month    Month
		expected float64
	}{
		{1, 0.9},
		{4, 1.1},
		{11, 1.2},
		{12, 1.3},
		{6, 1.0},  // Unknown month defaults to 1.0
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.month)), func(t *testing.T) {
			result := getMonthFactor(tt.month)
			if result != tt.expected {
				t.Errorf("getMonthFactor(%d) = %v; want %v", tt.month, result, tt.expected)
			}
		})
	}
}
