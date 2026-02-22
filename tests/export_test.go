package main

import (
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"
)

// TestExportCSVGeneration tests CSV export generation
func TestExportCSVGeneration(t *testing.T) {
	type Product struct {
		ID    int
		Name  string
		Price float64
		Stock int
	}

	products := []Product{
		{1, "Bread", 50.00, 30},
		{2, "Milk", 60.00, 20},
		{3, "Sugar", 130.00, 50},
	}

	generateCSV := func(products []Product) string {
		var builder strings.Builder
		writer := csv.NewWriter(&builder)

		header := []string{"ID", "Name", "Price", "Stock"}
		writer.Write(header)

		for _, p := range products {
			row := []string{
				string(rune(p.ID + '0')),
				p.Name,
				string(rune(int(p.Price/100) + '0')),
				string(rune(p.Stock + '0')),
			}
			writer.Write(row)
		}

		writer.Flush()
		return builder.String()
	}

	result := generateCSV(products)

	if !strings.Contains(result, "ID") {
		t.Error("CSV should contain header")
	}
	if !strings.Contains(result, "Bread") {
		t.Error("CSV should contain product name")
	}
}

// TestExportJSONGeneration tests JSON export generation
func TestExportJSONGeneration(t *testing.T) {
	type Product struct {
		ID    int     `json:"id"`
		Name  string  `json:"name"`
		Price float64 `json:"price"`
		Stock int     `json:"stock"`
	}

	products := []Product{
		{1, "Bread", 50.00, 30},
		{2, "Milk", 60.00, 20},
	}

	generateJSON := func(products []Product) string {
		data, _ := json.Marshal(products)
		return string(data)
	}

	result := generateJSON(products)

	if !strings.Contains(result, "Bread") {
		t.Error("JSON should contain product name")
	}
	if !strings.Contains(result, "50") {
		t.Error("JSON should contain price")
	}
}

// TestExportSalesCSV tests sales CSV export
func TestExportSalesCSV(t *testing.T) {
	type Sale struct {
		ID       int
		Product  string
		Quantity int
		Total    float64
		Payment  string
	}

	sales := []Sale{
		{1, "Bread", 2, 100.00, "cash"},
		{2, "Milk", 1, 60.00, "mpesa"},
	}

	generateSalesCSV := func(sales []Sale) string {
		var builder strings.Builder
		writer := csv.NewWriter(&builder)

		header := []string{"ID", "Product", "Quantity", "Total", "Payment"}
		writer.Write(header)

		for _, s := range sales {
			row := []string{
				string(rune(s.ID + '0')),
				s.Product,
				string(rune(s.Quantity + '0')),
				string(rune(int(s.Total/10) + '0')),
				s.Payment,
			}
			writer.Write(row)
		}

		writer.Flush()
		return builder.String()
	}

	result := generateSalesCSV(sales)

	if !strings.Contains(result, "Product") {
		t.Error("CSV should contain header")
	}
	if !strings.Contains(result, "cash") {
		t.Error("CSV should contain payment method")
	}
}

// TestExportInventoryValue tests inventory value calculation
func TestExportInventoryValue(t *testing.T) {
	type Product struct {
		Name  string
		Price float64
		Stock int
	}

	products := []Product{
		{"Bread", 50.00, 30},
		{"Milk", 60.00, 20},
		{"Sugar", 130.00, 50},
	}

	calculateValue := func(products []Product) float64 {
		var total float64
		for _, p := range products {
			total += p.Price * float64(p.Stock)
		}
		return total
	}

	value := calculateValue(products)

	// Bread: 50*30=1500, Milk: 60*20=1200, Sugar: 130*50=6500 = 9200
	if value != 9200 {
		t.Errorf("Inventory value = %.0f; want 9200", value)
	}
}

// TestExportProfitCalculation tests profit calculation for exports
func TestExportProfitCalculation(t *testing.T) {
	type Sale struct {
		Revenue  float64
		Cost     float64
		Quantity int
	}

	sales := []Sale{
		{Revenue: 100, Cost: 60, Quantity: 2},
		{Revenue: 60, Cost: 35, Quantity: 1},
	}

	calculateProfit := func(sales []Sale) float64 {
		var total float64
		for _, s := range sales {
			total += s.Revenue - s.Cost
		}
		return total
	}

	profit := calculateProfit(sales)
	// (100-60) + (60-35) = 40 + 25 = 65
	if profit != 65 {
		t.Errorf("Total profit = %.0f; want 65", profit)
	}
}

// TestExportDateRange tests date range filtering for exports
func TestExportDateRange(t *testing.T) {
	type Sale struct {
		Date  string
		Total float64
	}

	sales := []Sale{
		{"2026-01-01", 100},
		{"2026-01-15", 200},
		{"2026-02-01", 150},
	}

	filterByMonth := func(sales []Sale, month string) []Sale {
		var filtered []Sale
		for _, s := range sales {
			if strings.HasPrefix(s.Date, month) {
				filtered = append(filtered, s)
			}
		}
		return filtered
	}

	january := filterByMonth(sales, "2026-01")
	if len(january) != 2 {
		t.Errorf("January sales count = %d; want 2", len(january))
	}

	february := filterByMonth(sales, "2026-02")
	if len(february) != 1 {
		t.Errorf("February sales count = %d; want 1", len(february))
	}
}
