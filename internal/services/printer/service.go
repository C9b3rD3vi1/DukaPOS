package printer

import (
	"fmt"
	"net"
	"strings"
	"time"
)

// Receipt represents a receipt
type Receipt struct {
	ID            string        `json:"id"`
	ShopName      string        `json:"shop_name"`
	ShopPhone     string        `json:"shop_phone"`
	ShopAddress   string        `json:"shop_address"`
	Items         []ReceiptItem `json:"items"`
	Subtotal      float64       `json:"subtotal"`
	Discount      float64       `json:"discount"`
	Tax           float64       `json:"tax"`
	Total         float64       `json:"total"`
	PaymentMethod string        `json:"payment_method"`
	CashGiven     float64       `json:"cash_given"`
	Change        float64       `json:"change"`
	Cashier       string        `json:"cashier"`
	CustomerName  string        `json:"customer_name"`
	CustomerPhone string        `json:"customer_phone"`
	LoyaltyPoints int           `json:"loyalty_points"`
	PrintedAt     time.Time     `json:"printed_at"`
}

// ReceiptItem represents an item on receipt
type ReceiptItem struct {
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Total     float64 `json:"total"`
}

// PrinterConfig represents printer configuration
type PrinterConfig struct {
	Type    string `json:"type"` // thermal, cloud, pdf
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Width   int    `json:"width"` // characters per line (typically 32 or 48)
	CharSet string `json:"char_set"`
	APIKey  string `json:"api_key"`
}

// Service handles receipt generation and printing
type Service struct {
	config *PrinterConfig
}

// New creates a new printer service
func New(config *PrinterConfig) *Service {
	if config == nil {
		config = &PrinterConfig{
			Type:  "thermal",
			Width: 32,
		}
	}
	return &Service{config: config}
}

// GenerateReceipt creates a receipt from sale data
func (s *Service) GenerateReceipt(saleID uint, shopName, shopPhone string, items []ReceiptItem, paymentMethod string, cashGiven float64) *Receipt {
	subtotal := 0.0
	for _, item := range items {
		subtotal += item.Total
	}

	receipt := &Receipt{
		ID:            fmt.Sprintf("RCP-%d-%d", saleID, time.Now().Unix()),
		ShopName:      shopName,
		ShopPhone:     shopPhone,
		Items:         items,
		Subtotal:      subtotal,
		Discount:      0,
		Tax:           0,
		Total:         subtotal,
		PaymentMethod: paymentMethod,
		CashGiven:     cashGiven,
		Change:        cashGiven - subtotal,
		PrintedAt:     time.Now(),
	}

	return receipt
}

// FormatText generates plain text receipt
func (s *Service) FormatText(receipt *Receipt) string {
	width := s.config.Width
	var sb strings.Builder

	// Header
	sb.WriteString(s.center("🏪 "+receipt.ShopName, width))
	sb.WriteString("\n")
	sb.WriteString(s.center(receipt.ShopPhone, width))
	sb.WriteString("\n")
	if receipt.ShopAddress != "" {
		sb.WriteString(s.center(receipt.ShopAddress, width))
		sb.WriteString("\n")
	}
	sb.WriteString(strings.Repeat("-", width))
	sb.WriteString("\n")

	// Receipt info
	sb.WriteString(fmt.Sprintf("Receipt: %s\n", receipt.ID))
	sb.WriteString(fmt.Sprintf("Date: %s\n", receipt.PrintedAt.Format("02/01/2006 15:04")))
	if receipt.Cashier != "" {
		sb.WriteString(fmt.Sprintf("Cashier: %s\n", receipt.Cashier))
	}
	sb.WriteString(strings.Repeat("-", width))
	sb.WriteString("\n")

	// Items
	for _, item := range receipt.Items {
		name := item.Name
		if len(name) > 15 {
			name = name[:15]
		}
		qty := fmt.Sprintf("%d x", item.Quantity)
		price := fmt.Sprintf("KSh %.0f", item.UnitPrice)
		total := fmt.Sprintf("KSh %.0f", item.Total)

		padding := strings.Repeat(" ", width-len(name)-len(qty)-len(price)-len(total)-2)
		line := fmt.Sprintf("%s %s\n%s%s", name, qty, padding, price+total)
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	sb.WriteString(strings.Repeat("-", width))
	sb.WriteString("\n")

	// Totals
	sb.WriteString(s.formatLine("Subtotal:", fmt.Sprintf("KSh %.0f", receipt.Subtotal), width))
	if receipt.Discount > 0 {
		sb.WriteString(s.formatLine("Discount:", fmt.Sprintf("-KSh %.0f", receipt.Discount), width))
	}
	if receipt.Tax > 0 {
		sb.WriteString(s.formatLine("Tax:", fmt.Sprintf("KSh %.0f", receipt.Tax), width))
	}
	sb.WriteString(strings.Repeat("=", width))
	sb.WriteString("\n")
	sb.WriteString(s.formatLine("TOTAL:", fmt.Sprintf("KSh %.0f", receipt.Total), width))
	sb.WriteString("\n")

	// Payment info
	sb.WriteString(strings.Repeat("-", width))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Payment: %s\n", receipt.PaymentMethod))

	if receipt.PaymentMethod == "cash" && receipt.CashGiven > 0 {
		sb.WriteString(s.formatLine("Cash:", fmt.Sprintf("KSh %.0f", receipt.CashGiven), width))
		sb.WriteString(s.formatLine("Change:", fmt.Sprintf("KSh %.0f", receipt.Change), width))
	}

	// Loyalty points
	if receipt.LoyaltyPoints > 0 {
		sb.WriteString(strings.Repeat("-", width))
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("🎁 You earned %d loyalty points!\n", receipt.LoyaltyPoints))
	}

	// Footer
	sb.WriteString(strings.Repeat("-", width))
	sb.WriteString("\n")
	sb.WriteString(s.center("Thank you for shopping", width))
	sb.WriteString("\n")
	sb.WriteString(s.center("with us!", width))
	sb.WriteString("\n\n")
	sb.WriteString(s.center("Please come again", width))
	sb.WriteString("\n")
	sb.WriteString("\n\n\n") // Paper feed

	return sb.String()
}

// FormatThermal generates ESC/POS commands for thermal printer
func (s *Service) FormatThermal(receipt *Receipt) []byte {
	var sb strings.Builder

	// ESC/POS Commands
	initialize := []byte{0x1B, 0x40}        // Initialize printer
	alignCenter := []byte{0x1B, 0x61, 0x01} // Center align
	alignLeft := []byte{0x1B, 0x61, 0x00}   // Left align
	boldOn := []byte{0x1B, 0x45, 0x01}      // Bold on
	boldOff := []byte{0x1B, 0x45, 0x00}     // Bold off
	doubleOn := []byte{0x1B, 0x21, 0x10}    // Double height/width
	doubleOff := []byte{0x1B, 0x21, 0x00}   // Normal size
	cut := []byte{0x1D, 0x56, 0x00}         // Cut paper

	sb.Write(initialize)
	sb.Write(alignCenter)
	sb.Write(boldOn)
	sb.Write(doubleOn)
	sb.WriteString(receipt.ShopName)
	sb.WriteString("\n")
	sb.Write(doubleOff)
	sb.Write(boldOff)

	sb.WriteString(receipt.ShopPhone)
	sb.WriteString("\n")
	if receipt.ShopAddress != "" {
		sb.WriteString(receipt.ShopAddress)
		sb.WriteString("\n")
	}

	sb.Write(alignLeft)
	sb.WriteString("--------------------------------")

	// Items
	for _, item := range receipt.Items {
		name := item.Name
		if len(name) > 16 {
			name = name[:16]
		}
		qty := fmt.Sprintf("%d x", item.Quantity)
		price := fmt.Sprintf("%.0f", item.UnitPrice)
		total := fmt.Sprintf("%.0f", item.Total)

		line := fmt.Sprintf("%-16s %s\n%-32s%s", name, qty, price, total)
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	sb.WriteString("--------------------------------")
	sb.WriteString("\n")

	// Totals
	subtotal := fmt.Sprintf("Subtotal: KSh %.0f", receipt.Subtotal)
	sb.WriteString(subtotal)
	sb.WriteString("\n")

	if receipt.Discount > 0 {
		discount := fmt.Sprintf("Discount: -KSh %.0f", receipt.Discount)
		sb.WriteString(discount)
		sb.WriteString("\n")
	}

	sb.WriteString("================================")
	sb.WriteString("\n")

	total := fmt.Sprintf("TOTAL: KSh %.0f", receipt.Total)
	sb.Write(boldOn)
	sb.WriteString(total)
	sb.Write(boldOff)
	sb.WriteString("\n")

	// Payment
	sb.WriteString("--------------------------------")
	sb.WriteString("\n")

	if receipt.PaymentMethod == "cash" && receipt.CashGiven > 0 {
		cash := fmt.Sprintf("Cash: KSh %.0f", receipt.CashGiven)
		sb.WriteString(cash)
		sb.WriteString("\n")
		change := fmt.Sprintf("Change: KSh %.0f", receipt.Change)
		sb.WriteString(change)
		sb.WriteString("\n")
	}

	// Footer
	sb.WriteString("================================")
	sb.WriteString("\n")
	sb.Write(alignCenter)
	sb.WriteString("Thank you for shopping with us!")
	sb.WriteString("\n")
	sb.WriteString("Please come again")
	sb.WriteString("\n\n\n")

	sb.Write(cut)

	return []byte(sb.String())
}

// FormatPDF generates HTML for PDF receipt
func (s *Service) FormatHTML(receipt *Receipt) string {
	itemsHTML := ""
	for _, item := range receipt.Items {
		itemsHTML += fmt.Sprintf(`
		<tr>
			<td>%s</td>
			<td>%d</td>
			<td>KSh %.0f</td>
			<td>KSh %.0f</td>
		</tr>`, item.Name, item.Quantity, item.UnitPrice, item.Total)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Receipt - %s</title>
    <style>
        body { font-family: 'Courier New', monospace; width: 300px; margin: 0 auto; padding: 20px; }
        .header { text-align: center; margin-bottom: 20px; }
        .shop-name { font-size: 18px; font-weight: bold; }
        .divider { border-bottom: 1px dashed #000; margin: 10px 0; }
        .item { display: flex; justify-content: space-between; }
        .total { font-weight: bold; font-size: 18px; }
        .footer { text-align: center; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="header">
        <div class="shop-name">🏪 %s</div>
        <div>%s</div>
        <div>%s</div>
    </div>
    <div class="divider"></div>
    <div>Receipt: %s</div>
    <div>Date: %s</div>
    <div class="divider"></div>
    <table width="100%%">
        <tr><th>Item</th><th>Qty</th><th>Price</th><th>Total</th></tr>
        %s
    </table>
    <div class="divider"></div>
    <div>Subtotal: KSh %.0f</div>
    %s
    <div class="total">TOTAL: KSh %.0f</div>
    <div class="divider"></div>
    <div>Payment: %s</div>
    %s
    <div class="divider"></div>
    <div class="footer">
        <p>Thank you for shopping with us!</p>
        <p>Please come again</p>
    </div>
</body>
</html>`,
		receipt.ID,
		receipt.ShopName, receipt.ShopPhone, receipt.ShopAddress,
		receipt.ID, receipt.PrintedAt.Format("02/01/2006 15:04"),
		itemsHTML,
		receipt.Subtotal,
		formatDiscount(receipt.Discount),
		receipt.Total,
		receipt.PaymentMethod,
		formatCash(receipt.CashGiven, receipt.Change),
	)
}

// center centers text within width
func (s *Service) center(text string, width int) string {
	if len(text) >= width {
		return text[:width]
	}
	padding := (width - len(text)) / 2
	return strings.Repeat(" ", padding) + text
}

// formatLine formats a line with label on left and value on right
func (s *Service) formatLine(label, value string, width int) string {
	padding := width - len(label) - len(value)
	if padding < 1 {
		padding = 1
	}
	return label + strings.Repeat(" ", padding) + value + "\n"
}

func formatDiscount(discount float64) string {
	if discount <= 0 {
		return ""
	}
	return fmt.Sprintf("<div>Discount: -KSh %.0f</div>", discount)
}

func formatCash(cash, change float64) string {
	if cash <= 0 {
		return ""
	}
	return fmt.Sprintf(`
    <div>Cash: KSh %.0f</div>
    <div>Change: KSh %.0f</div>`, cash, change)
}

// Print sends receipt to printer (placeholder - implement actual printing)
func (s *Service) Print(receipt *Receipt) error {
	switch s.config.Type {
	case "thermal":
		return s.printThermal(receipt)
	case "cloud":
		return s.printCloud(receipt)
	case "pdf":
		return s.printPDF(receipt)
	default:
		return fmt.Errorf("unsupported printer type: %s", s.config.Type)
	}
}

func (s *Service) printThermal(receipt *Receipt) error {
	if s.config.Host == "" {
		return fmt.Errorf("printer host not configured")
	}

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	if s.config.Port == 0 {
		addr = fmt.Sprintf("%s:9100", s.config.Host)
	}

	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to printer: %w", err)
	}
	defer conn.Close()

	data := s.FormatThermal(receipt)
	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to send to printer: %w", err)
	}

	return nil
}

func (s *Service) printCloud(receipt *Receipt) error {
	// In production, this would call cloud print API
	return nil
}

func (s *Service) printPDF(receipt *Receipt) error {
	// In production, this would generate PDF
	_ = s.FormatHTML(receipt)
	return nil
}

// DailyReport generates daily summary receipt
func (s *Service) DailyReport(shopName string, totalSales float64, transactionCount int, topProducts []string) string {
	width := s.config.Width
	var sb strings.Builder

	sb.WriteString(s.center("📊 DAILY REPORT", width))
	sb.WriteString("\n")
	sb.WriteString(s.center(shopName, width))
	sb.WriteString("\n")
	sb.WriteString(s.center(time.Now().Format("02/01/2006"), width))
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("-", width))
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("Total Sales: KSh %.0f\n", totalSales))
	sb.WriteString(fmt.Sprintf("Transactions: %d\n", transactionCount))
	if transactionCount > 0 {
		avg := totalSales / float64(transactionCount)
		sb.WriteString(fmt.Sprintf("Average: KSh %.0f\n", avg))
	}

	sb.WriteString(strings.Repeat("-", width))
	sb.WriteString("\n")
	sb.WriteString("Top Products:\n")
	for i, product := range topProducts {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, product))
	}

	sb.WriteString(strings.Repeat("=", width))
	sb.WriteString("\n")
	sb.WriteString(s.center("End of Report", width))

	return sb.String()
}
