package export

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
)

type Format string

const (
	FormatCSV   Format = "csv"
	FormatJSON  Format = "json"
	FormatExcel Format = "excel"
	FormatPDF   Format = "pdf"
)

type ProductExporter struct{}

func (e *ProductExporter) Export(products []models.Product, format Format) ([]byte, error) {
	switch format {
	case FormatCSV:
		return e.exportCSV(products)
	case FormatJSON:
		return e.exportJSON(products)
	case FormatExcel:
		return e.exportExcel(products)
	case FormatPDF:
		return e.exportPDF(products)
	default:
		return e.exportCSV(products)
	}
}

func (e *ProductExporter) exportCSV(products []models.Product) ([]byte, error) {
	var builder strings.Builder
	writer := csv.NewWriter(&builder)

	header := []string{"ID", "Name", "Category", "Unit", "Cost Price", "Selling Price", "Stock", "Low Stock Threshold", "Barcode"}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

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

func (e *ProductExporter) exportPDF(products []models.Product) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, "Product Inventory Report")
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 10)
	headers := []string{"ID", "Name", "Category", "Unit", "Cost", "Price", "Stock", "Threshold"}
	colWidths := []float64{15, 45, 30, 20, 25, 25, 20, 25}

	for i, h := range headers {
		pdf.Cell(colWidths[i], 8, h)
	}
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 9)
	for _, p := range products {
		pdf.CellFormat(colWidths[0], 7, fmt.Sprintf("%d", p.ID), "0", 0, "", false, 0, "")
		pdf.CellFormat(colWidths[1], 7, p.Name, "0", 0, "", false, 0, "")
		pdf.CellFormat(colWidths[2], 7, p.Category, "0", 0, "", false, 0, "")
		pdf.CellFormat(colWidths[3], 7, p.Unit, "0", 0, "", false, 0, "")
		pdf.CellFormat(colWidths[4], 7, fmt.Sprintf("%.2f", p.CostPrice), "0", 0, "", false, 0, "")
		pdf.CellFormat(colWidths[5], 7, fmt.Sprintf("%.2f", p.SellingPrice), "0", 0, "", false, 0, "")
		pdf.CellFormat(colWidths[6], 7, fmt.Sprintf("%d", p.CurrentStock), "0", 0, "", false, 0, "")
		pdf.CellFormat(colWidths[7], 7, fmt.Sprintf("%d", p.LowStockThreshold), "0", 0, "", false, 0, "")
		pdf.Ln(-1)
	}

	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(190, 5, fmt.Sprintf("Generated: %s", time.Now().Format("2006-01-02 15:04:05")))

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (e *ProductExporter) exportExcel(products []models.Product) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	f.SetCellValue("Sheet1", "A1", "ID")
	f.SetCellValue("Sheet1", "B1", "Name")
	f.SetCellValue("Sheet1", "C1", "Category")
	f.SetCellValue("Sheet1", "D1", "Unit")
	f.SetCellValue("Sheet1", "E1", "Cost Price")
	f.SetCellValue("Sheet1", "F1", "Selling Price")
	f.SetCellValue("Sheet1", "G1", "Stock")
	f.SetCellValue("Sheet1", "H1", "Low Stock Threshold")
	f.SetCellValue("Sheet1", "I1", "Barcode")

	headers := []string{"A1", "B1", "C1", "D1", "E1", "F1", "G1", "H1", "I1"}
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#00A650"}, Pattern: 1},
	})
	for _, h := range headers {
		f.SetCellStyle("Sheet1", h, h, style)
	}

	for i, p := range products {
		row := i + 2
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", row), p.ID)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", row), p.Name)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", row), p.Category)
		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", row), p.Unit)
		f.SetCellValue("Sheet1", fmt.Sprintf("E%d", row), p.CostPrice)
		f.SetCellValue("Sheet1", fmt.Sprintf("F%d", row), p.SellingPrice)
		f.SetCellValue("Sheet1", fmt.Sprintf("G%d", row), p.CurrentStock)
		f.SetCellValue("Sheet1", fmt.Sprintf("H%d", row), p.LowStockThreshold)
		f.SetCellValue("Sheet1", fmt.Sprintf("I%d", row), p.Barcode)
	}

	f.SetColWidth("Sheet1", "A", "A", 8)
	f.SetColWidth("Sheet1", "B", "B", 25)
	f.SetColWidth("Sheet1", "C", "C", 15)
	f.SetColWidth("Sheet1", "D", "D", 10)
	f.SetColWidth("Sheet1", "E", "E", 12)
	f.SetColWidth("Sheet1", "F", "F", 15)
	f.SetColWidth("Sheet1", "G", "G", 10)
	f.SetColWidth("Sheet1", "H", "H", 18)
	f.SetColWidth("Sheet1", "I", "I", 15)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (e *ProductExporter) InventoryValue(products []models.Product) float64 {
	var total float64
	for _, p := range products {
		total += p.SellingPrice * float64(p.CurrentStock)
	}
	return total
}

type SalesExporter struct{}

func (e *SalesExporter) Export(sales []models.Sale, format Format) ([]byte, error) {
	switch format {
	case FormatCSV:
		return e.exportCSV(sales)
	case FormatJSON:
		return e.exportJSON(sales)
	case FormatExcel:
		return e.exportExcel(sales)
	case FormatPDF:
		return e.exportSalesPDF(sales)
	default:
		return e.exportCSV(sales)
	}
}

func (e *SalesExporter) exportCSV(sales []models.Sale) ([]byte, error) {
	var builder strings.Builder
	writer := csv.NewWriter(&builder)

	header := []string{"ID", "Date", "Product", "Quantity", "Unit Price", "Total", "Cost", "Profit", "Payment Method", "Receipt"}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	for _, s := range sales {
		productName := ""
		if s.Product.Name != "" {
			productName = s.Product.Name
		}
		row := []string{
			fmt.Sprintf("%d", s.ID),
			s.CreatedAt.Format("2006-01-02 15:04"),
			productName,
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
		ID            uint    `json:"id"`
		Date          string  `json:"date"`
		ProductName   string  `json:"product_name"`
		Quantity      int     `json:"quantity"`
		UnitPrice     float64 `json:"unit_price"`
		TotalAmount   float64 `json:"total_amount"`
		CostAmount    float64 `json:"cost_amount"`
		Profit        float64 `json:"profit"`
		PaymentMethod string  `json:"payment_method"`
		MpesaReceipt  string  `json:"mpesa_receipt"`
	}

	result := make([]SaleJSON, len(sales))
	for i, s := range sales {
		productName := ""
		if s.Product.Name != "" {
			productName = s.Product.Name
		}
		result[i] = SaleJSON{
			ID:            s.ID,
			Date:          s.CreatedAt.Format(time.RFC3339),
			ProductName:   productName,
			Quantity:      s.Quantity,
			UnitPrice:     s.UnitPrice,
			TotalAmount:   s.TotalAmount,
			CostAmount:    s.CostAmount,
			Profit:        s.Profit,
			PaymentMethod: string(s.PaymentMethod),
			MpesaReceipt:  s.MpesaReceipt,
		}
	}

	return json.MarshalIndent(result, "", "  ")
}

func (e *SalesExporter) exportExcel(sales []models.Sale) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	f.SetCellValue("Sheet1", "A1", "ID")
	f.SetCellValue("Sheet1", "B1", "Date")
	f.SetCellValue("Sheet1", "C1", "Product")
	f.SetCellValue("Sheet1", "D1", "Quantity")
	f.SetCellValue("Sheet1", "E1", "Unit Price")
	f.SetCellValue("Sheet1", "F1", "Total")
	f.SetCellValue("Sheet1", "G1", "Cost")
	f.SetCellValue("Sheet1", "H1", "Profit")
	f.SetCellValue("Sheet1", "I1", "Payment Method")
	f.SetCellValue("Sheet1", "J1", "Receipt")

	headers := []string{"A1", "B1", "C1", "D1", "E1", "F1", "G1", "H1", "I1", "J1"}
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#00A650"}, Pattern: 1},
	})
	for _, h := range headers {
		f.SetCellStyle("Sheet1", h, h, style)
	}

	for i, s := range sales {
		row := i + 2
		productName := ""
		if s.Product.Name != "" {
			productName = s.Product.Name
		}
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", row), s.ID)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", row), s.CreatedAt.Format("2006-01-02 15:04"))
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", row), productName)
		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", row), s.Quantity)
		f.SetCellValue("Sheet1", fmt.Sprintf("E%d", row), s.UnitPrice)
		f.SetCellValue("Sheet1", fmt.Sprintf("F%d", row), s.TotalAmount)
		f.SetCellValue("Sheet1", fmt.Sprintf("G%d", row), s.CostAmount)
		f.SetCellValue("Sheet1", fmt.Sprintf("H%d", row), s.Profit)
		f.SetCellValue("Sheet1", fmt.Sprintf("I%d", row), string(s.PaymentMethod))
		f.SetCellValue("Sheet1", fmt.Sprintf("J%d", row), s.MpesaReceipt)
	}

	f.SetColWidth("Sheet1", "A", "A", 8)
	f.SetColWidth("Sheet1", "B", "B", 18)
	f.SetColWidth("Sheet1", "C", "C", 20)
	f.SetColWidth("Sheet1", "D", "D", 10)
	f.SetColWidth("Sheet1", "E", "E", 12)
	f.SetColWidth("Sheet1", "F", "F", 12)
	f.SetColWidth("Sheet1", "G", "G", 12)
	f.SetColWidth("Sheet1", "H", "H", 12)
	f.SetColWidth("Sheet1", "I", "I", 15)
	f.SetColWidth("Sheet1", "J", "J", 20)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (e *SalesExporter) exportSalesPDF(sales []models.Sale) ([]byte, error) {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(280, 10, "Sales Report")
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 8)
	headers := []string{"ID", "Date", "Product", "Qty", "Unit Price", "Total", "Cost", "Profit", "Payment"}
	colWidths := []float64{15, 35, 45, 15, 25, 25, 20, 20, 25}

	for i, h := range headers {
		pdf.Cell(colWidths[i], 6, h)
	}
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 7)
	for _, s := range sales {
		productName := s.Product.Name
		if len(productName) > 20 {
			productName = productName[:20]
		}
		pdf.CellFormat(colWidths[0], 5, fmt.Sprintf("%d", s.ID), "0", 0, "", false, 0, "")
		pdf.CellFormat(colWidths[1], 5, s.CreatedAt.Format("2006-01-02 15:04"), "0", 0, "", false, 0, "")
		pdf.CellFormat(colWidths[2], 5, productName, "0", 0, "", false, 0, "")
		pdf.CellFormat(colWidths[3], 5, fmt.Sprintf("%d", s.Quantity), "0", 0, "", false, 0, "")
		pdf.CellFormat(colWidths[4], 5, fmt.Sprintf("%.2f", s.UnitPrice), "0", 0, "", false, 0, "")
		pdf.CellFormat(colWidths[5], 5, fmt.Sprintf("%.2f", s.TotalAmount), "0", 0, "", false, 0, "")
		pdf.CellFormat(colWidths[6], 5, fmt.Sprintf("%.2f", s.CostAmount), "0", 0, "", false, 0, "")
		pdf.CellFormat(colWidths[7], 5, fmt.Sprintf("%.2f", s.Profit), "0", 0, "", false, 0, "")
		pdf.CellFormat(colWidths[8], 5, string(s.PaymentMethod), "0", 0, "", false, 0, "")
		pdf.Ln(-1)
	}

	pdf.Ln(8)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(280, 5, fmt.Sprintf("Generated: %s | Total Transactions: %d", time.Now().Format("2006-01-02 15:04:05"), len(sales)))

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

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
		"total_revenue":     totalRevenue,
		"total_cost":        totalCost,
		"total_profit":      totalProfit,
		"transaction_count": len(sales),
		"average_sale":      totalRevenue / float64(len(sales)),
		"by_payment_method": paymentMethods,
	}
}

type ReportExporter struct{}

type DailyReportData struct {
	Date             string        `json:"date"`
	TotalSales       float64       `json:"total_sales"`
	TotalProfit      float64       `json:"total_profit"`
	TransactionCount int           `json:"transaction_count"`
	AverageSale      float64       `json:"average_sale"`
	TopProducts      []ProductSale `json:"top_products"`
}

type ProductSale struct {
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Revenue  float64 `json:"revenue"`
}

func (e *ReportExporter) ExportDaily(report DailyReportData, format Format) ([]byte, error) {
	switch format {
	case FormatCSV:
		return e.exportDailyCSV(report)
	case FormatJSON:
		return json.MarshalIndent(report, "", "  ")
	case FormatExcel:
		return e.exportDailyExcel(report)
	case FormatPDF:
		return e.exportDailyPDF(report)
	default:
		return e.exportDailyCSV(report)
	}
}

func (e *ReportExporter) exportDailyCSV(report DailyReportData) ([]byte, error) {
	var builder strings.Builder
	writer := csv.NewWriter(&builder)

	if err := writer.Write([]string{"DAILY REPORT"}); err != nil {
		return nil, err
	}
	if err := writer.Write([]string{"Date", report.Date}); err != nil {
		return nil, err
	}
	if err := writer.Write([]string{}); err != nil {
		return nil, err
	}

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

func (e *ReportExporter) exportDailyExcel(report DailyReportData) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	f.SetCellValue("Sheet1", "A1", "DAILY REPORT")
	f.SetCellValue("Sheet1", "A2", "Date")
	f.SetCellValue("Sheet1", "B2", report.Date)

	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 14},
	})
	f.SetCellStyle("Sheet1", "A1", "A1", titleStyle)

	f.SetCellValue("Sheet1", "A4", "Metric")
	f.SetCellValue("Sheet1", "B4", "Value")
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#00A650"}, Pattern: 1},
	})
	f.SetCellStyle("Sheet1", "A4", "B4", headerStyle)

	f.SetCellValue("Sheet1", "A5", "Total Sales")
	f.SetCellValue("Sheet1", "B5", fmt.Sprintf("KSh %.2f", report.TotalSales))
	f.SetCellValue("Sheet1", "A6", "Total Profit")
	f.SetCellValue("Sheet1", "B6", fmt.Sprintf("KSh %.2f", report.TotalProfit))
	f.SetCellValue("Sheet1", "A7", "Transactions")
	f.SetCellValue("Sheet1", "B7", report.TransactionCount)
	f.SetCellValue("Sheet1", "A8", "Average Sale")
	f.SetCellValue("Sheet1", "B8", fmt.Sprintf("KSh %.2f", report.AverageSale))

	f.SetCellValue("Sheet1", "A10", "Top Products")
	productStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	f.SetCellStyle("Sheet1", "A10", "A10", productStyle)

	f.SetCellValue("Sheet1", "A11", "Product")
	f.SetCellValue("Sheet1", "B11", "Quantity")
	f.SetCellValue("Sheet1", "C11", "Revenue")
	f.SetCellStyle("Sheet1", "A11", "C11", headerStyle)

	row := 12
	for _, p := range report.TopProducts {
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", row), p.Name)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", row), p.Quantity)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", row), fmt.Sprintf("KSh %.2f", p.Revenue))
		row++
	}

	f.SetColWidth("Sheet1", "A", "A", 25)
	f.SetColWidth("Sheet1", "B", "B", 15)
	f.SetColWidth("Sheet1", "C", "C", 15)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (e *ReportExporter) exportDailyPDF(report DailyReportData) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 18)
	pdf.Cell(190, 15, "Daily Sales Report")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(190, 8, fmt.Sprintf("Date: %s", report.Date))
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(190, 10, "Summary")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 12)
	summaryData := []struct {
		label, value string
	}{
		{"Total Sales", fmt.Sprintf("KSh %.2f", report.TotalSales)},
		{"Total Profit", fmt.Sprintf("KSh %.2f", report.TotalProfit)},
		{"Transactions", fmt.Sprintf("%d", report.TransactionCount)},
		{"Average Sale", fmt.Sprintf("KSh %.2f", report.AverageSale)},
	}

	for _, s := range summaryData {
		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(60, 8, s.label)
		pdf.SetFont("Arial", "", 11)
		pdf.Cell(130, 8, s.value)
		pdf.Ln(-1)
	}

	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(190, 10, "Top Products")
	pdf.Ln(8)

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(100, 8, "Product")
	pdf.Cell(45, 8, "Quantity")
	pdf.Cell(45, 8, "Revenue")
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 10)
	for _, p := range report.TopProducts {
		pdf.Cell(100, 7, p.Name)
		pdf.Cell(45, 7, fmt.Sprintf("%d", p.Quantity))
		pdf.Cell(45, 7, fmt.Sprintf("KSh %.2f", p.Revenue))
		pdf.Ln(-1)
	}

	pdf.Ln(15)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(190, 5, fmt.Sprintf("Generated by DukaPOS on %s", time.Now().Format("2006-01-02 15:04:05")))

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
