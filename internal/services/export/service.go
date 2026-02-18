package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
)

// Format represents export format
type Format string

const (
	FormatCSV    Format = "csv"
	FormatJSON  Format = "json"
	FormatExcel Format = "excel"
)

// ProductExporter exports products
type ProductExporter struct{}

// ExportProducts exports products to specified format
func (e *ProductExporter) Export(products []models.Product, format Format) ([]byte, error) {
	switch format {
	case FormatCSV:
		return e.exportCSV(products)
	case FormatJSON:
		return e.exportJSON(products)
	default:
		return e.exportCSV(products)
	}
}

func (e *ProductExporter) exportCSV(products []models.Product) ([]byte, error) {
	var builder strings.Builder
	writer := csv.NewWriter(&builder)

	// Header
	header := []string{"ID", "Name", "Category", "Unit", "Cost Price", "Selling Price", "Stock", "Low Stock Threshold", "Barcode"}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	// Data
	for _, p := range products {
		row := []string{
			fmt.Sprintf("%d", p.ID),
			p.Name,
			p.Category,
			p.Unit,
			fmt.Sprintf("%.2f", p.CostPrice),
			fmt.Sprintf("%.2f", p.SellingPrice),
			fmt.Sprintf("%d", p.CurrentStock),
			fmt.Sprintf("%d", p.LowStockThreshold),
			p.Barcode,
		}
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return []byte(builder.String()), nil
}

func (e *ProductExporter) exportJSON(products []models.Product) ([]byte, error) {
	return json.MarshalIndent(products, "", "  ")
}

// SalesExporter exports sales
type SalesExporter struct{}

// ExportSales exports sales to specified format
func (e *SalesExporter) Export(sales []models.Sale, format Format) ([]byte, error) {
	switch format {
	case FormatCSV:
		return e.exportCSV(sales)
	case FormatJSON:
		return e.exportJSON(sales)
	default:
		return e.exportCSV(sales)
	}
}

func (e *SalesExporter) exportCSV(sales []models.Sale) ([]byte, error) {
	var builder strings.Builder
	writer := csv.NewWriter(&builder)

	// Header
	header := []string{"ID", "Date", "Product", "Quantity", "Unit Price", "Total", "Cost", "Profit", "Payment Method", "Receipt"}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	// Data
	for _, s := range sales {
		row := []string{
			fmt.Sprintf("%d", s.ID),
			s.CreatedAt.Format("2006-01-02 15:04"),
			s.Product.Name,
			fmt.Sprintf("%d", s.Quantity),
			fmt.Sprintf("%.2f", s.UnitPrice),
			fmt.Sprintf("%.2f", s.TotalAmount),
			fmt.Sprintf("%.2f", s.CostAmount),
			fmt.Sprintf("%.2f", s.Profit),
			string(s.PaymentMethod),
			s.MpesaReceipt,
		}
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return []byte(builder.String()), nil
}

func (e *SalesExporter) exportJSON(sales []models.Sale) ([]byte, error) {
	type SaleJSON struct {
		ID             uint    `json:"id"`
		Date           string  `json:"date"`
		ProductName    string  `json:"product_name"`
		Quantity       int     `json:"quantity"`
		UnitPrice      float64 `json:"unit_price"`
		TotalAmount    float64 `json:"total_amount"`
		CostAmount     float64 `json:"cost_amount"`
		Profit         float64 `json:"profit"`
		PaymentMethod  string  `json:"payment_method"`
		MpesaReceipt  string  `json:"mpesa_receipt"`
	}

	result := make([]SaleJSON, len(sales))
	for i, s := range sales {
		result[i] = SaleJSON{
			ID:            s.ID,
			Date:          s.CreatedAt.Format(time.RFC3339),
			ProductName:   s.Product.Name,
			Quantity:      s.Quantity,
			UnitPrice:     s.UnitPrice,
			TotalAmount:   s.TotalAmount,
			CostAmount:    s.CostAmount,
			Profit:       s.Profit,
			PaymentMethod: string(s.PaymentMethod),
			MpesaReceipt:  s.MpesaReceipt,
		}
	}

	return json.MarshalIndent(result, "", "  ")
}

// ReportExporter exports reports
type ReportExporter struct{}

// DailyReportData represents daily report data
type DailyReportData struct {
	Date          string  `json:"date"`
	TotalSales    float64 `json:"total_sales"`
	TotalProfit   float64 `json:"total_profit"`
	TransactionCount int  `json:"transaction_count"`
	AverageSale   float64 `json:"average_sale"`
	TopProducts   []ProductSale `json:"top_products"`
}

// ProductSale represents product sale in report
type ProductSale struct {
	Name       string `json:"name"`
	Quantity   int    `json:"quantity"`
	Revenue    float64 `json:"revenue"`
}

// ExportDaily exports daily report
func (e *ReportExporter) ExportDaily(report DailyReportData, format Format) ([]byte, error) {
	switch format {
	case FormatCSV:
		return e.exportDailyCSV(report)
	case FormatJSON:
		return json.MarshalIndent(report, "", "  ")
	default:
		return e.exportDailyCSV(report)
	}
}

func (e *ReportExporter) exportDailyCSV(report DailyReportData) ([]byte, error) {
	var builder strings.Builder
	writer := csv.NewWriter(&builder)

	// Header
	if err := writer.Write([]string{"DAILY REPORT"}); err != nil {
		return nil, err
	}
	if err := writer.Write([]string{"Date", report.Date}); err != nil {
		return nil, err
	}
	if err := writer.Write([]string{}); err != nil {
		return nil, err
	}

	// Summary
	if err := writer.Write([]string{"Metric", "Value"}); err != nil {
		return nil, err
	}
	if err := writer.Write([]string{"Total Sales", fmt.Sprintf("KSh %.2f", report.TotalSales)}); err != nil {
		return nil, err
	}
	if err := writer.Write([]string{"Total Profit", fmt.Sprintf("KSh %.2f", report.TotalProfit)}); err != nil {
		return nil, err
	}
	if err := writer.Write([]string{"Transactions", fmt.Sprintf("%d", report.TransactionCount)}); err != nil {
		return nil, err
	}
	if err := writer.Write([]string{"Average Sale", fmt.Sprintf("KSh %.2f", report.AverageSale)}); err != nil {
		return nil, err
	}

	if err := writer.Write([]string{}); err != nil {
		return nil, err
	}

	// Top Products
	if err := writer.Write([]string{"Top Products"}); err != nil {
		return nil, err
	}
	if err := writer.Write([]string{"Product", "Quantity", "Revenue"}); err != nil {
		return nil, err
	}

	for _, p := range report.TopProducts {
		row := []string{p.Name, fmt.Sprintf("%d", p.Quantity), fmt.Sprintf("KSh %.2f", p.Revenue)}
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return []byte(builder.String()), nil
}

// InventoryValue calculates total inventory value
func (e *ProductExporter) InventoryValue(products []models.Product) float64 {
	var total float64
	for _, p := range products {
		total += p.SellingPrice * float64(p.CurrentStock)
	}
	return total
}

// ProfitSummary calculates profit summary
func (e *SalesExporter) ProfitSummary(sales []models.Sale) map[string]interface{} {
	var totalRevenue, totalCost, totalProfit float64
	paymentMethods := make(map[string]float64)

	for _, s := range sales {
		totalRevenue += s.TotalAmount
		totalCost += s.CostAmount
		totalProfit += s.Profit
		paymentMethods[string(s.PaymentMethod)] += s.TotalAmount
	}

	return map[string]interface{}{
		"total_revenue":    totalRevenue,
		"total_cost":      totalCost,
		"total_profit":    totalProfit,
		"transaction_count": len(sales),
		"average_sale":    totalRevenue / float64(len(sales)),
		"by_payment_method": paymentMethods,
	}
}
