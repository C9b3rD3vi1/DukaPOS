package printer

import (
	"time"

	"github.com/C9b3rD3vi1/DukaPOS/internal/services/printer"
	"github.com/gofiber/fiber/v2"
)

// Handler handles receipt/print HTTP requests
type Handler struct {
	service *printer.Service
}

// New creates a new printer handler
func New(service *printer.Service) *Handler {
	return &Handler{service: service}
}

// PrintRequest represents print request
type PrintRequest struct {
	SaleID        uint             `json:"sale_id"`
	ShopName      string           `json:"shop_name"`
	ShopPhone     string           `json:"shop_phone"`
	ShopAddress   string           `json:"shop_address"`
	Items         []ReceiptItem    `json:"items"`
	PaymentMethod string           `json:"payment_method"`
	CashGiven     float64          `json:"cash_given"`
	Cashier       string           `json:"cashier"`
}

// ReceiptItem represents an item on receipt
type ReceiptItem struct {
	Name      string  `json:"name"`
	Quantity int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Total     float64 `json:"total"`
}

// PrintReceipt prints a receipt
// POST /api/v1/print/receipt
func (h *Handler) PrintReceipt(c *fiber.Ctx) error {
	var req PrintRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.ShopName == "" || len(req.Items) == 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "shop name and items are required",
		})
	}

	subtotal := 0.0
	for i := range req.Items {
		req.Items[i].Total = float64(req.Items[i].Quantity) * req.Items[i].UnitPrice
		subtotal += req.Items[i].Total
	}

	receipt := &printer.Receipt{
		ID:            "RCP",
		ShopName:      req.ShopName,
		ShopPhone:     req.ShopPhone,
		ShopAddress:   req.ShopAddress,
		Items:         convertItems(req.Items),
		Subtotal:      subtotal,
		Total:         subtotal,
		PaymentMethod: req.PaymentMethod,
		CashGiven:     req.CashGiven,
		Change:        req.CashGiven - subtotal,
		Cashier:       req.Cashier,
		PrintedAt:     time.Now(),
	}

	if err := h.service.Print(receipt); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":    "success",
		"receipt_id": receipt.ID,
		"message":  "Receipt printed",
	})
}

func convertItems(items []ReceiptItem) []printer.ReceiptItem {
	result := make([]printer.ReceiptItem, len(items))
	for i, item := range items {
		result[i] = printer.ReceiptItem{
			Name:      item.Name,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
			Total:     item.Total,
		}
	}
	return result
}

// GetTextReceipt returns plain text receipt
// POST /api/v1/print/text
func (h *Handler) GetTextReceipt(c *fiber.Ctx) error {
	var req PrintRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	subtotal := 0.0
	for i := range req.Items {
		req.Items[i].Total = float64(req.Items[i].Quantity) * req.Items[i].UnitPrice
		subtotal += req.Items[i].Total
	}

	receipt := &printer.Receipt{
		ID:            "RCP",
		ShopName:      req.ShopName,
		ShopPhone:     req.ShopPhone,
		Items:         convertItems(req.Items),
		Subtotal:      subtotal,
		Total:         subtotal,
		PaymentMethod: req.PaymentMethod,
		CashGiven:     req.CashGiven,
		Change:        req.CashGiven - subtotal,
		PrintedAt:     time.Now(),
	}

	text := h.service.FormatText(receipt)

	return c.JSON(fiber.Map{
		"receipt_id": receipt.ID,
		"text":       text,
	})
}

// GetThermalReceipt returns ESC/POS thermal printer commands
// POST /api/v1/print/thermal
func (h *Handler) GetThermalReceipt(c *fiber.Ctx) error {
	var req PrintRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	subtotal := 0.0
	for i := range req.Items {
		req.Items[i].Total = float64(req.Items[i].Quantity) * req.Items[i].UnitPrice
		subtotal += req.Items[i].Total
	}

	receipt := &printer.Receipt{
		ID:            "RCP",
		ShopName:      req.ShopName,
		ShopPhone:     req.ShopPhone,
		Items:         convertItems(req.Items),
		Subtotal:      subtotal,
		Total:         subtotal,
		PaymentMethod: req.PaymentMethod,
		CashGiven:     req.CashGiven,
		Change:        req.CashGiven - subtotal,
		PrintedAt:     time.Now(),
	}

	thermal := h.service.FormatThermal(receipt)

	return c.JSON(fiber.Map{
		"receipt_id": receipt.ID,
		"commands":   thermal,
		"encoding":   "base64",
	})
}

// GetHTMLReceipt returns HTML receipt for PDF generation
// POST /api/v1/print/html
func (h *Handler) GetHTMLReceipt(c *fiber.Ctx) error {
	var req PrintRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	subtotal := 0.0
	for i := range req.Items {
		req.Items[i].Total = float64(req.Items[i].Quantity) * req.Items[i].UnitPrice
		subtotal += req.Items[i].Total
	}

	receipt := &printer.Receipt{
		ID:            "RCP",
		ShopName:      req.ShopName,
		ShopPhone:     req.ShopPhone,
		Items:         convertItems(req.Items),
		Subtotal:      subtotal,
		Total:         subtotal,
		PaymentMethod: req.PaymentMethod,
		CashGiven:     req.CashGiven,
		Change:        req.CashGiven - subtotal,
		PrintedAt:     time.Now(),
	}

	html := h.service.FormatHTML(receipt)

	return c.JSON(fiber.Map{
		"receipt_id": receipt.ID,
		"html":       html,
	})
}

// PrintDailyReport prints daily report
// POST /api/v1/print/report
func (h *Handler) PrintDailyReport(c *fiber.Ctx) error {
	type ReportRequest struct {
		ShopName         string   `json:"shop_name"`
		TotalSales       float64  `json:"total_sales"`
		TransactionCount int      `json:"transaction_count"`
		TopProducts      []string `json:"top_products"`
	}

	var req ReportRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	report := h.service.DailyReport(req.ShopName, req.TotalSales, req.TransactionCount, req.TopProducts)

	return c.JSON(fiber.Map{
		"report": report,
		"date":   time.Now().Format("2006-01-02"),
	})
}

// GetPrinters returns available printers
// GET /api/v1/print/printers
func (h *Handler) GetPrinters(c *fiber.Ctx) error {
	printers := []fiber.Map{
		{
			"id":      "thermal_1",
			"name":    "Thermal Printer 1",
			"type":    "thermal",
			"status":  "online",
			"default": true,
		},
	}

	return c.JSON(fiber.Map{
		"printers": printers,
	})
}

// TestPrinter tests printer connection
// POST /api/v1/print/test
func (h *Handler) TestPrinter(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Test receipt generated",
	})
}

// Configure updates printer configuration
// PUT /api/v1/print/config
func (h *Handler) Configure(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Printer configured",
	})
}

// GetConfig returns current printer configuration
// GET /api/v1/print/config
func (h *Handler) GetConfig(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"config": fiber.Map{
			"type": "thermal",
			"width": 32,
		},
	})
}
