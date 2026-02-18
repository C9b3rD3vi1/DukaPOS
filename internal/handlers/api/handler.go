package api

import (
	"github.com/gofiber/fiber/v2"
)

// Handler handles API v1 requests
type Handler struct{}

// New creates a new API handler
func New() *Handler {
	return &Handler{}
}

// APIInfo represents API information
type APIInfo struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	BaseURL     string   `json:"base_url"`
	Endpoints   []Endpoint `json:"endpoints"`
}

// Endpoint represents an API endpoint
type Endpoint struct {
	Path        string   `json:"path"`
	Method      string   `json:"method"`
	Description string   `json:"description"`
	Auth        bool    `json:"auth"`
	RateLimit   string   `json:"rate_limit"`
}

// GetAPIInfo returns API information
// GET /api/v1
func (h *Handler) GetAPIInfo(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"name":        "DukaPOS API",
		"version":     "1.0.0",
		"description": "REST API for DukaPOS - Digital Kenyan Duka Management",
		"base_url":    "/api/v1",
		"docs":        "https://dukapos.io/docs/api",
		"auth": fiber.Map{
			"type":        "Bearer Token",
			"header":      "Authorization",
			"description": "Use your JWT token or API key",
		},
	})
}

// ListEndpoints lists all available endpoints
// GET /api/v1/endpoints
func (h *Handler) ListEndpoints(c *fiber.Ctx) error {
	endpoints := []Endpoint{
		// Products
		{"GET", "/products", "List all products", true, "100/min"},
		{"GET", "/products/:id", "Get product by ID", true, "100/min"},
		{"POST", "/products", "Create new product", true, "30/min"},
		{"PUT", "/products/:id", "Update product", true, "30/min"},
		{"DELETE", "/products/:id", "Delete product", true, "30/min"},

		// Sales
		{"GET", "/sales", "List sales", true, "100/min"},
		{"GET", "/sales/:id", "Get sale by ID", true, "100/min"},
		{"POST", "/sales", "Create new sale", true, "60/min"},
		{"POST", "/sales/from-whatsapp", "Record sale via WhatsApp", false, "60/min"},

		// Reports
		{"GET", "/reports/daily", "Daily sales report", true, "30/min"},
		{"GET", "/reports/weekly", "Weekly sales report", true, "30/min"},
		{"GET", "/reports/monthly", "Monthly sales report", true, "30/min"},
		{"GET", "/reports/products", "Product performance", true, "30/min"},

		// Staff (Pro+)
		{"GET", "/staff", "List staff members", true, "30/min"},
		{"GET", "/staff/:id", "Get staff member", true, "30/min"},
		{"POST", "/staff", "Create staff member", true, "10/min"},
		{"PUT", "/staff/:id", "Update staff member", true, "10/min"},
		{"DELETE", "/staff/:id", "Delete staff member", true, "10/min"},

		// Shops (Pro+)
		{"GET", "/shops", "List shops", true, "30/min"},
		{"GET", "/shops/:id", "Get shop details", true, "30/min"},
		{"POST", "/shops", "Create new shop", true, "10/min"},
		{"PUT", "/shops/:id", "Update shop", true, "10/min"},

		// M-Pesa (Pro+)
		{"POST", "/payments/mpesa/stk", "Initiate STK push", true, "10/min"},
		{"GET", "/payments/mpesa/status/:checkoutId", "Check payment status", true, "30/min"},
		{"POST", "/webhooks/mpesa/validate", "Validate M-Pesa payment", false, "N/A"},

		// Webhooks
		{"GET", "/webhooks", "List webhooks", true, "30/min"},
		{"POST", "/webhooks", "Create webhook", true, "10/min"},
		{"PUT", "/webhooks/:id", "Update webhook", true, "10/min"},
		{"DELETE", "/webhooks/:id", "Delete webhook", true, "10/min"},

		// API Keys (Business)
		{"GET", "/api-keys", "List API keys", true, "30/min"},
		{"POST", "/api-keys", "Create API key", true, "10/min"},
		{"DELETE", "/api-keys/:id", "Revoke API key", true, "10/min"},

		// Loyalty (Business)
		{"GET", "/customers", "List loyalty customers", true, "30/min"},
		{"GET", "/customers/:id/points", "Get customer points", true, "30/min"},
		{"POST", "/customers/:id/redeem", "Redeem points", true, "10/min"},

		// AI Predictions (Business)
		{"GET", "/predictions/inventory", "Inventory predictions", true, "10/min"},
		{"GET", "/predictions/restock", "Restock recommendations", true, "10/min"},
		{"GET", "/predictions/trends", "Sales trends", true, "10/min"},
	}

	return c.JSON(fiber.Map{
		"endpoints": endpoints,
		"total":     len(endpoints),
	})
}

// ErrorResponse represents an API error
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// RateLimitResponse represents rate limit info
type RateLimitResponse struct {
	Limit     int    `json:"limit"`
	Remaining int    `json:"remaining"`
	Reset     int64  `json:"reset"`
}

// FormatError formats error response
func FormatError(c *fiber.Ctx, status int, err string) error {
	return c.Status(status).JSON(ErrorResponse{Error: err})
}

// FormatSuccess formats success response
func FormatSuccess(c *fiber.Ctx, status int, message string, data interface{}) error {
	return c.Status(status).JSON(SuccessResponse{
		Message: message,
		Data:    data,
	})
}
