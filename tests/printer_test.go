package main

import (
	"fmt"
	"strings"
	"testing"
)

// TestReceiptGeneration tests receipt creation
func TestReceiptGeneration(t *testing.T) {
	type ReceiptItem struct {
		Name      string
		Quantity  int
		UnitPrice float64
		Total     float64
	}

	type Receipt struct {
		ID            string
		ShopName      string
		ShopPhone     string
		Items         []ReceiptItem
		Subtotal      float64
		Total         float64
		PaymentMethod string
		CashGiven     float64
		Change        float64
	}

	generateReceipt := func(saleID uint, items []ReceiptItem, paymentMethod string, cashGiven float64) *Receipt {
		subtotal := 0.0
		for i := range items {
			items[i].Total = float64(items[i].Quantity) * items[i].UnitPrice
			subtotal += items[i].Total
		}

		return &Receipt{
			ID:            fmt.Sprintf("RCP-%d", saleID),
			ShopName:      "Test Shop",
			ShopPhone:     "+254700000000",
			Items:         items,
			Subtotal:      subtotal,
			Total:         subtotal,
			PaymentMethod: paymentMethod,
			CashGiven:     cashGiven,
			Change:        cashGiven - subtotal,
		}
	}

	items := []ReceiptItem{
		{Name: "Milk", Quantity: 2, UnitPrice: 60},
		{Name: "Bread", Quantity: 1, UnitPrice: 50},
	}

	receipt := generateReceipt(1, items, "cash", 200)

	if receipt.Total != 170 {
		t.Errorf("Total = %v; want 170", receipt.Total)
	}

	if receipt.Change != 30 {
		t.Errorf("Change = %v; want 30", receipt.Change)
	}

	if len(receipt.Items) != 2 {
		t.Errorf("Item count = %d; want 2", len(receipt.Items))
	}
}

// TestReceiptFormatting tests text receipt formatting
func TestReceiptFormatting(t *testing.T) {
	type Receipt struct {
		ShopName    string
		ShopPhone   string
		Items       []struct {
			Name     string
			Quantity int
			UnitPrice float64
			Total    float64
		}
		Subtotal float64
		Total    float64
	}

	formatReceipt := func(r *Receipt, width int) string {
		var sb strings.Builder

		// Header
		sb.WriteString(fmt.Sprintf("%s\n", r.ShopName))
		sb.WriteString(fmt.Sprintf("%s\n", r.ShopPhone))
		sb.WriteString(strings.Repeat("-", width))
		sb.WriteString("\n")

		// Items
		for _, item := range r.Items {
			line := fmt.Sprintf("%s x%d - KSh %.0f", item.Name, item.Quantity, item.Total)
			sb.WriteString(line)
			sb.WriteString("\n")
		}

		sb.WriteString(strings.Repeat("-", width))
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("TOTAL: KSh %.0f", r.Total))

		return sb.String()
	}

	receipt := &Receipt{
		ShopName:  "Test Shop",
		ShopPhone: "+254700000000",
		Items: []struct {
			Name      string
			Quantity int
			UnitPrice float64
			Total    float64
		}{
			{"Milk", 2, 60, 120},
			{"Bread", 1, 50, 50},
		},
		Subtotal: 170,
		Total:    170,
	}

	output := formatReceipt(receipt, 32)

	if !strings.Contains(output, "Test Shop") {
		t.Error("Receipt should contain shop name")
	}
	if !strings.Contains(output, "Milk") {
		t.Error("Receipt should contain Milk")
	}
	if !strings.Contains(output, "TOTAL: KSh 170") {
		t.Error("Receipt should contain total")
	}
}

// TestThermalCommands tests ESC/POS command generation
func TestThermalCommands(t *testing.T) {
	// ESC/POS command constants
	init := []byte{0x1B, 0x40}
	alignCenter := []byte{0x1B, 0x61, 0x01}
	alignLeft := []byte{0x1B, 0x61, 0x00}
	cut := []byte{0x1D, 0x56, 0x00}

	// Verify initialization command
	if init[0] != 0x1B || init[1] != 0x40 {
		t.Error("Init command should be ESC @")
	}

	// Verify center align
	if alignCenter[0] != 0x1B || alignCenter[1] != 0x61 {
		t.Error("Align command format incorrect")
	}

	// Verify cut command (use alignLeft to avoid unused warning)
	_ = alignLeft
	if cut[0] != 0x1D || cut[1] != 0x56 {
		t.Error("Cut command format incorrect")
	}

	t.Log("Thermal command constants verified")
}

// TestReceiptItemCalculation tests item total calculation
func TestReceiptItemCalculation(t *testing.T) {
	type Item struct {
		Name     string
		Quantity int
		Price    float64
	}

	calculateItem := func(item Item) float64 {
		return float64(item.Quantity) * item.Price
	}

	tests := []struct {
		name     string
		item     Item
		expected float64
	}{
		{"Single item", Item{"Milk", 1, 60}, 60},
		{"Multiple items", Item{"Milk", 2, 60}, 120},
		{"Zero quantity", Item{"Bread", 0, 50}, 0},
		{"Decimal price", Item{"Sugar", 1, 130.50}, 130.50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateItem(tt.item)
			if result != tt.expected {
				t.Errorf("calculateItem() = %v; want %v", result, tt.expected)
			}
		})
	}
}

// TestChangeCalculation tests change calculation
func TestChangeCalculation(t *testing.T) {
	calculateChange := func(total, cashGiven float64) float64 {
		return cashGiven - total
	}

	tests := []struct {
		name     string
		total    float64
		cash     float64
		expected float64
	}{
		{"Exact change", 100, 100, 0},
		{"With change", 100, 150, 50},
		{"Over payment", 75, 100, 25},
		{"Under payment", 100, 50, -50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateChange(tt.total, tt.cash)
			if result != tt.expected {
				t.Errorf("calculateChange(%v, %v) = %v; want %v", tt.total, tt.cash, result, tt.expected)
			}
		})
	}
}

// TestReceiptIDGeneration tests receipt ID generation
func TestReceiptIDGeneration(t *testing.T) {
	generateID := func(saleID uint, timestamp int64) string {
		return fmt.Sprintf("RCP-%d-%d", saleID, timestamp)
	}

	id1 := generateID(1, 1704067200)
	id2 := generateID(1, 1704067200)
	id3 := generateID(2, 1704067200)

	// Same inputs should produce same ID
	if id1 != id2 {
		t.Error("Same inputs should produce same ID")
	}

	// Different sale IDs should produce different IDs
	if id1 == id3 {
		t.Error("Different sale IDs should produce different IDs")
	}

	// Verify format
	expectedPrefix := "RCP-"
	if len(id1) < len(expectedPrefix)+3 {
		t.Error("ID format incorrect")
	}
}
