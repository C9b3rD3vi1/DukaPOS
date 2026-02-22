package handler

import (
	"fmt"
	"strconv"
	"time"

	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"github.com/C9b3rD3vi1/DukaPOS/internal/repository"
	"github.com/C9b3rD3vi1/DukaPOS/internal/services/export"
	"github.com/gofiber/fiber/v2"
)

type ExportHandler struct {
	productRepo *repository.ProductRepository
	saleRepo    *repository.SaleRepository
	summaryRepo *repository.DailySummaryRepository
}

func NewExportHandler(
	productRepo *repository.ProductRepository,
	saleRepo *repository.SaleRepository,
	summaryRepo *repository.DailySummaryRepository,
) *ExportHandler {
	return &ExportHandler{
		productRepo: productRepo,
		saleRepo:    saleRepo,
		summaryRepo: summaryRepo,
	}
}

func (h *ExportHandler) RegisterRoutes(protected fiber.Router) {
	exportRoutes := protected.Group("/export")
	exportRoutes.Get("/products", h.ExportProducts)
	exportRoutes.Get("/sales", h.ExportSales)
	exportRoutes.Get("/report", h.ExportReport)
	exportRoutes.Get("/inventory", h.ExportInventory)
}

type ExportQuery struct {
	Format string `query:"format"`
	From   string `query:"from"`
	To     string `query:"to"`
}

func (h *ExportHandler) ExportProducts(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)
	query := new(ExportQuery)
	if err := c.QueryParser(query); err != nil {
		query.Format = "csv"
	}

	format := export.FormatCSV
	if query.Format == "json" {
		format = export.FormatJSON
	}

	products, err := h.productRepo.GetByShopID(shopID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch products",
		})
	}

	exporter := &export.ProductExporter{}
	data, err := exporter.Export(products, format)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to export products",
		})
	}

	filename := fmt.Sprintf("products_%s.%s", time.Now().Format("20060102"), query.Format)
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	if query.Format == "json" {
		c.Set("Content-Type", "application/json")
	} else {
		c.Set("Content-Type", "text/csv")
	}

	return c.Send(data)
}

func (h *ExportHandler) ExportSales(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)
	query := new(ExportQuery)
	if err := c.QueryParser(query); err != nil {
		query.Format = "csv"
	}

	format := export.FormatCSV
	if query.Format == "json" {
		format = export.FormatJSON
	}

	var sales []models.Sale
	var err error

	if query.From != "" && query.To != "" {
		from, _ := time.Parse("2006-01-02", query.From)
		to, _ := time.Parse("2006-01-02", query.To)
		to = to.Add(24 * time.Hour)
		sales, err = h.saleRepo.GetByDateRange(shopID, from, to)
	} else {
		sales, err = h.saleRepo.GetTodaySales(shopID)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch sales",
		})
	}

	exporter := &export.SalesExporter{}
	data, err := exporter.Export(sales, format)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to export sales",
		})
	}

	filename := fmt.Sprintf("sales_%s.%s", time.Now().Format("20060102"), query.Format)
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	if query.Format == "json" {
		c.Set("Content-Type", "application/json")
	} else {
		c.Set("Content-Type", "text/csv")
	}

	return c.Send(data)
}

func (h *ExportHandler) ExportReport(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)
	query := new(ExportQuery)
	if err := c.QueryParser(query); err != nil {
		query.Format = "csv"
	}

	format := export.FormatCSV
	if query.Format == "json" {
		format = export.FormatJSON
	}

	reportDate := time.Now()
	if query.From != "" {
		reportDate, _ = time.Parse("2006-01-02", query.From)
	}

	summaries, err := h.summaryRepo.GetByDateRange(shopID, reportDate.AddDate(0, 0, -30), reportDate)
	if err != nil {
		summaries = nil
	}

	sales, err := h.saleRepo.GetByDateRange(shopID, reportDate.AddDate(0, 0, -30), reportDate)
	if err != nil {
		sales = nil
	}

	var totalSales, totalProfit float64
	for _, s := range summaries {
		totalSales += s.TotalSales
		totalProfit += s.TotalProfit
	}

	avgSale := 0.0
	if len(sales) > 0 {
		avgSale = totalSales / float64(len(sales))
	}

	report := export.DailyReportData{
		Date:             reportDate.Format("2006-01-02"),
		TotalSales:       totalSales,
		TotalProfit:      totalProfit,
		TransactionCount: len(sales),
		AverageSale:      avgSale,
		TopProducts:      []export.ProductSale{},
	}

	productSales := make(map[string]export.ProductSale)
	for _, s := range sales {
		if ps, ok := productSales[s.Product.Name]; ok {
			ps.Quantity += s.Quantity
			ps.Revenue += s.TotalAmount
			productSales[s.Product.Name] = ps
		} else {
			productSales[s.Product.Name] = export.ProductSale{
				Name:     s.Product.Name,
				Quantity: s.Quantity,
				Revenue:  s.TotalAmount,
			}
		}
	}

	for _, ps := range productSales {
		report.TopProducts = append(report.TopProducts, ps)
	}

	exporter := &export.ReportExporter{}
	data, err := exporter.ExportDaily(report, format)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to export report",
		})
	}

	filename := fmt.Sprintf("report_%s.%s", reportDate.Format("20060102"), query.Format)
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	if query.Format == "json" {
		c.Set("Content-Type", "application/json")
	} else {
		c.Set("Content-Type", "text/csv")
	}

	return c.Send(data)
}

func (h *ExportHandler) ExportInventory(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)
	query := new(ExportQuery)
	if err := c.QueryParser(query); err != nil {
		query.Format = "csv"
	}

	products, err := h.productRepo.GetByShopID(shopID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch products",
		})
	}

	type InventoryItem struct {
		Name            string  `json:"name"`
		Category        string  `json:"category"`
		CurrentStock    int     `json:"current_stock"`
		SellingPrice    float64 `json:"selling_price"`
		CostPrice       float64 `json:"cost_price"`
		StockValue      float64 `json:"stock_value"`
		PotentialProfit float64 `json:"potential_profit"`
	}

	inventory := make([]InventoryItem, len(products))
	totalValue := 0.0
	totalCost := 0.0

	for i, p := range products {
		stockValue := p.SellingPrice * float64(p.CurrentStock)
		costValue := p.CostPrice * float64(p.CurrentStock)
		totalValue += stockValue
		totalCost += costValue

		inventory[i] = InventoryItem{
			Name:            p.Name,
			Category:        p.Category,
			CurrentStock:    p.CurrentStock,
			SellingPrice:    p.SellingPrice,
			CostPrice:       p.CostPrice,
			StockValue:      stockValue,
			PotentialProfit: stockValue - costValue,
		}
	}

	if query.Format == "json" {
		return c.JSON(fiber.Map{
			"inventory":         inventory,
			"total_stock_value": totalValue,
			"total_cost_value":  totalCost,
			"potential_profit":  totalValue - totalCost,
			"product_count":     len(products),
		})
	}

	var result string
	result += "NAME,CATEGORY,STOCK,SELLING PRICE,COST PRICE,STOCK VALUE,POTENTIAL PROFIT\n"
	for _, item := range inventory {
		result += fmt.Sprintf("%s,%s,%d,%.2f,%.2f,%.2f,%.2f\n",
			item.Name, item.Category, item.CurrentStock,
			item.SellingPrice, item.CostPrice,
			item.StockValue, item.PotentialProfit)
	}
	result += fmt.Sprintf("\nTOTAL,,,,,,%.2f,%.2f\n", totalValue, totalValue-totalCost)

	filename := fmt.Sprintf("inventory_%s.csv", time.Now().Format("20060102"))
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Set("Content-Type", "text/csv")

	return c.Send([]byte(result))
}

func parseUint(s string) uint {
	i, _ := strconv.ParseUint(s, 10, 32)
	return uint(i)
}
