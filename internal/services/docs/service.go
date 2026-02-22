package docs

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// OpenAPI generates OpenAPI 3.0 documentation
func OpenAPI() map[string]interface{} {
	return map[string]interface{}{
		"openapi": "3.0.3",
		"info": map[string]interface{}{
			"title":       "DukaPOS API",
			"description": "REST API for Kenyan Duka (Shop) Management - WhatsApp, USSD, and REST",
			"version":     "1.0.0",
			"contact": map[string]interface{}{
				"name":  "DukaPOS Support",
				"email": "support@dukapos.io",
			},
			"license": map[string]interface{}{
				"name": "MIT",
			},
		},
		"servers": []map[string]interface{}{
			{"url": "https://api.dukapos.io/v1", "description": "Production server"},
			{"url": "https://sandbox.dukapos.io/v1", "description": "Sandbox server"},
		},
		"paths":      Paths(),
		"components": Components(),
		"tags": []map[string]interface{}{
			{"name": "Auth", "description": "Authentication endpoints"},
			{"name": "Products", "description": "Product management"},
			{"name": "Sales", "description": "Sales transactions"},
			{"name": "Shop", "description": "Shop management"},
			{"name": "Staff", "description": "Staff management (Pro)"},
			{"name": "Payments", "description": "M-Pesa payments (Pro)"},
			{"name": "Reports", "description": "Reports and analytics"},
		},
	}
}

// Paths returns API paths
func Paths() map[string]interface{} {
	return map[string]interface{}{
		"/api/auth/register": map[string]interface{}{
			"post": map[string]interface{}{
				"tags":        []string{"Auth"},
				"summary":     "Register new shop account",
				"description": "Create a new account with shop",
				"requestBody": map[string]interface{}{
					"required": true,
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"email":    map[string]interface{}{"type": "string"},
									"password": map[string]interface{}{"type": "string"},
									"name":     map[string]interface{}{"type": "string"},
									"phone":    map[string]interface{}{"type": "string"},
								},
							},
						},
					},
				},
				"responses": map[string]interface{}{
					"201": map[string]interface{}{"description": "Account created"},
					"400": map[string]interface{}{"description": "Invalid request"},
				},
			},
		},
		"/api/auth/login": map[string]interface{}{
			"post": map[string]interface{}{
				"tags":        []string{"Auth"},
				"summary":     "Login to account",
				"description": "Authenticate and get JWT token",
				"requestBody": map[string]interface{}{
					"required": true,
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"email":    map[string]interface{}{"type": "string"},
									"password": map[string]interface{}{"type": "string"},
								},
							},
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{"description": "Login successful"},
					"401": map[string]interface{}{"description": "Invalid credentials"},
				},
			},
		},
		"/api/v1/products": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"Products"},
				"summary":     "List all products",
				"description": "Get all products for the authenticated shop",
				"security":    []map[string]string{{"BearerAuth": ""}},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{"description": "List of products"},
				},
			},
			"post": map[string]interface{}{
				"tags":        []string{"Products"},
				"summary":     "Create product",
				"description": "Add a new product to inventory",
				"security":    []map[string]string{{"BearerAuth": ""}},
				"requestBody": map[string]interface{}{
					"required": true,
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"type":     "object",
								"required": []string{"name", "selling_price"},
								"properties": map[string]interface{}{
									"name":                map[string]interface{}{"type": "string"},
									"category":            map[string]interface{}{"type": "string"},
									"selling_price":       map[string]interface{}{"type": "number"},
									"cost_price":          map[string]interface{}{"type": "number"},
									"current_stock":       map[string]interface{}{"type": "integer"},
									"low_stock_threshold": map[string]interface{}{"type": "integer"},
									"barcode":             map[string]interface{}{"type": "string"},
								},
							},
						},
					},
				},
				"responses": map[string]interface{}{
					"201": map[string]interface{}{"description": "Product created"},
				},
			},
		},
		"/api/v1/products/{id}": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"Products"},
				"summary":     "Get product",
				"description": "Get a specific product by ID",
				"security":    []map[string]string{{"BearerAuth": ""}},
				"parameters": []map[string]interface{}{
					{"name": "id", "in": "path", "required": true, "schema": map[string]interface{}{"type": "integer"}},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{"description": "Product details"},
					"404": map[string]interface{}{"description": "Product not found"},
				},
			},
			"put": map[string]interface{}{
				"tags":        []string{"Products"},
				"summary":     "Update product",
				"description": "Update product details",
				"security":    []map[string]string{{"BearerAuth": ""}},
				"parameters": []map[string]interface{}{
					{"name": "id", "in": "path", "required": true, "schema": map[string]interface{}{"type": "integer"}},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{"description": "Product updated"},
				},
			},
			"delete": map[string]interface{}{
				"tags":        []string{"Products"},
				"summary":     "Delete product",
				"description": "Remove a product from inventory",
				"security":    []map[string]string{{"BearerAuth": ""}},
				"parameters": []map[string]interface{}{
					{"name": "id", "in": "path", "required": true, "schema": map[string]interface{}{"type": "integer"}},
				},
				"responses": map[string]interface{}{
					"204": map[string]interface{}{"description": "Product deleted"},
				},
			},
		},
		"/api/v1/sales": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"Sales"},
				"summary":     "List sales",
				"description": "Get all sales for the authenticated shop",
				"security":    []map[string]string{{"BearerAuth": ""}},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{"description": "List of sales"},
				},
			},
			"post": map[string]interface{}{
				"tags":        []string{"Sales"},
				"summary":     "Create sale",
				"description": "Record a new sale transaction",
				"security":    []map[string]string{{"BearerAuth": ""}},
				"requestBody": map[string]interface{}{
					"required": true,
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"type":     "object",
								"required": []string{"product_id", "quantity"},
								"properties": map[string]interface{}{
									"product_id":     map[string]interface{}{"type": "integer"},
									"quantity":       map[string]interface{}{"type": "integer"},
									"payment_method": map[string]interface{}{"type": "string"},
								},
							},
						},
					},
				},
				"responses": map[string]interface{}{
					"201": map[string]interface{}{"description": "Sale recorded"},
				},
			},
		},
		"/api/v1/sales/{id}": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"Sales"},
				"summary":     "Get sale",
				"description": "Get a specific sale by ID",
				"security":    []map[string]string{{"BearerAuth": ""}},
				"parameters": []map[string]interface{}{
					{"name": "id", "in": "path", "required": true, "schema": map[string]interface{}{"type": "integer"}},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{"description": "Sale details"},
				},
			},
		},
		"/api/v1/shop/profile": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"Shop"},
				"summary":     "Get shop profile",
				"description": "Get current shop information",
				"security":    []map[string]string{{"BearerAuth": ""}},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{"description": "Shop profile"},
				},
			},
			"put": map[string]interface{}{
				"tags":        []string{"Shop"},
				"summary":     "Update shop profile",
				"description": "Update shop information",
				"security":    []map[string]string{{"BearerAuth": ""}},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{"description": "Shop updated"},
				},
			},
		},
		"/api/v1/shop/dashboard": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"Shop"},
				"summary":     "Get dashboard data",
				"description": "Get daily summary and statistics",
				"security":    []map[string]string{{"BearerAuth": ""}},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{"description": "Dashboard data"},
				},
			},
		},
		"/api/v1/mpesa/stk-push": map[string]interface{}{
			"post": map[string]interface{}{
				"tags":        []string{"Payments"},
				"summary":     "Initiate STK Push",
				"description": "Send M-Pesa STK push payment request",
				"security":    []map[string]string{{"BearerAuth": ""}},
				"requestBody": map[string]interface{}{
					"required": true,
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"type":     "object",
								"required": []string{"phone", "amount"},
								"properties": map[string]interface{}{
									"phone":  map[string]interface{}{"type": "string"},
									"amount": map[string]interface{}{"type": "number"},
								},
							},
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{"description": "STK push initiated"},
				},
			},
		},
		"/api/v1/staff": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"Staff"},
				"summary":     "List staff",
				"description": "Get all staff members (Pro feature)",
				"security":    []map[string]string{{"BearerAuth": ""}},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{"description": "List of staff"},
				},
			},
			"post": map[string]interface{}{
				"tags":        []string{"Staff"},
				"summary":     "Add staff",
				"description": "Create a new staff account (Pro feature)",
				"security":    []map[string]string{{"BearerAuth": ""}},
				"responses": map[string]interface{}{
					"201": map[string]interface{}{"description": "Staff created"},
				},
			},
		},
		"/api/v1/customers": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"Customers"},
				"summary":     "List customers",
				"description": "Get all customers (Business feature)",
				"security":    []map[string]string{{"BearerAuth": ""}},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{"description": "List of customers"},
				},
			},
		},
		"/api/v1/loyalty/points/{customer_id}": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"Loyalty"},
				"summary":     "Get customer points",
				"description": "Get loyalty points for a customer",
				"security":    []map[string]string{{"BearerAuth": ""}},
				"parameters": []map[string]interface{}{
					{"name": "customer_id", "in": "path", "required": true, "schema": map[string]interface{}{"type": "integer"}},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{"description": "Points balance"},
				},
			},
		},
		"/api/v1/currency/convert": map[string]interface{}{
			"post": map[string]interface{}{
				"tags":        []string{"Currency"},
				"summary":     "Convert currency",
				"description": "Convert amount between currencies",
				"security":    []map[string]string{{"BearerAuth": ""}},
				"requestBody": map[string]interface{}{
					"required": true,
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"type":     "object",
								"required": []string{"amount", "from", "to"},
								"properties": map[string]interface{}{
									"amount": map[string]interface{}{"type": "number"},
									"from":   map[string]interface{}{"type": "string"},
									"to":     map[string]interface{}{"type": "string"},
								},
							},
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{"description": "Conversion result"},
				},
			},
		},
		"/api/v1/billing/plans": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"Billing"},
				"summary":     "Get subscription plans",
				"description": "List available subscription plans",
				"security":    []map[string]string{{"BearerAuth": ""}},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{"description": "List of plans"},
				},
			},
		},
		"/webhook/twilio": map[string]interface{}{
			"post": map[string]interface{}{
				"tags":        []string{"Webhooks"},
				"summary":     "Twilio WhatsApp webhook",
				"description": "Receive WhatsApp messages from Twilio",
				"responses": map[string]interface{}{
					"200": map[string]interface{}{"description": "Message processed"},
				},
			},
		},
		"/webhook/mpesa/stk": map[string]interface{}{
			"post": map[string]interface{}{
				"tags":        []string{"Webhooks"},
				"summary":     "M-Pesa STK callback",
				"description": "Receive M-Pesa payment callbacks",
				"responses": map[string]interface{}{
					"200": map[string]interface{}{"description": "Callback processed"},
				},
			},
		},
	}
}

// Components returns OpenAPI components
func Components() map[string]interface{} {
	return map[string]interface{}{
		"schemas": Schemas(),
		"securitySchemes": map[string]interface{}{
			"BearerAuth": map[string]interface{}{
				"type":         "http",
				"scheme":       "bearer",
				"bearerFormat": "JWT",
				"description":  "Enter JWT token",
			},
			"ApiKeyAuth": map[string]interface{}{
				"type": "apiKey",
				"name": "X-API-Key",
				"in":   "header",
			},
		},
	}
}

// Schemas returns API schemas
func Schemas() map[string]interface{} {
	return map[string]interface{}{
		"Product": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id":            map[string]interface{}{"type": "integer"},
				"name":          map[string]interface{}{"type": "string"},
				"category":      map[string]interface{}{"type": "string"},
				"selling_price": map[string]interface{}{"type": "number"},
				"cost_price":    map[string]interface{}{"type": "number"},
				"current_stock": map[string]interface{}{"type": "integer"},
			},
		},
		"Sale": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id":             map[string]interface{}{"type": "integer"},
				"product_id":     map[string]interface{}{"type": "integer"},
				"quantity":       map[string]interface{}{"type": "integer"},
				"total_amount":   map[string]interface{}{"type": "number"},
				"payment_method": map[string]interface{}{"type": "string"},
			},
		},
		"Error": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"error": map[string]interface{}{"type": "string"},
				"code":  map[string]interface{}{"type": "string"},
			},
		},
	}
}

// =========================================
// Builder helpers
// =========================================

type OperationBuilder struct {
	op map[string]interface{}
}

func Post(op map[string]interface{}) *OperationBuilder {
	op["post"] = map[string]interface{}{"tags": []string{""}, "responses": map[string]interface{}{"200": map[string]interface{}{"description": "Success"}}}
	return &OperationBuilder{op: op}
}

func Get(op map[string]interface{}) *OperationBuilder    { return &OperationBuilder{op: op} }
func Put(op map[string]interface{}) *OperationBuilder    { return &OperationBuilder{op: op} }
func Delete(op map[string]interface{}) *OperationBuilder { return &OperationBuilder{op: op} }

func (b *OperationBuilder) Tag(name string) *OperationBuilder {
	return b
}

func (b *OperationBuilder) Summary(s string) *OperationBuilder {
	return b
}

func (b *OperationBuilder) Description(d string) *OperationBuilder {
	return b
}

func (b *OperationBuilder) Operation() map[string]interface{} {
	return b.op
}

// =========================================
// Operation definitions
// =========================================

func AuthRegister() map[string]interface{} {
	return map[string]interface{}{
		"summary":     "Register a new shop",
		"description": "Create a new shop account",
		"tags":        []string{"Auth"},
		"requestBody": map[string]interface{}{
			"required": true,
			"content": map[string]interface{}{
				"application/json": map[string]interface{}{
					"schema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"phone":    map[string]interface{}{"type": "string"},
							"password": map[string]interface{}{"type": "string"},
							"name":     map[string]interface{}{"type": "string"},
						},
						"required": []string{"phone", "password"},
					},
				},
			},
		},
		"responses": map[string]interface{}{
			"201": map[string]interface{}{"description": "Shop created"},
			"400": map[string]interface{}{"description": "Validation error"},
		},
	}
}

func AuthLogin() map[string]interface{} {
	return map[string]interface{}{
		"summary":     "Login to shop",
		"description": "Authenticate and get JWT token",
		"tags":        []string{"Auth"},
		"requestBody": map[string]interface{}{
			"required": true,
			"content": map[string]interface{}{
				"application/json": map[string]interface{}{
					"schema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"phone":    map[string]interface{}{"type": "string"},
							"password": map[string]interface{}{"type": "string"},
						},
						"required": []string{"phone", "password"},
					},
				},
			},
		},
		"responses": map[string]interface{}{
			"200": map[string]interface{}{"description": "Login successful"},
			"401": map[string]interface{}{"description": "Invalid credentials"},
		},
	}
}

func ListProducts() map[string]interface{} {
	return map[string]interface{}{
		"summary":     "List all products",
		"description": "Get paginated list of products",
		"tags":        []string{"Products"},
		"parameters": []map[string]interface{}{
			{"name": "page", "in": "query", "schema": map[string]interface{}{"type": "integer"}},
			{"name": "limit", "in": "query", "schema": map[string]interface{}{"type": "integer"}},
			{"name": "category", "in": "query", "schema": map[string]interface{}{"type": "string"}},
		},
		"responses": map[string]interface{}{
			"200": map[string]interface{}{"description": "Product list"},
		},
	}
}

func CreateProduct() map[string]interface{} {
	return map[string]interface{}{
		"summary":     "Create product",
		"description": "Add a new product to inventory",
		"tags":        []string{"Products"},
		"requestBody": map[string]interface{}{
			"required": true,
			"content": map[string]interface{}{
				"application/json": map[string]interface{}{
					"schema": map[string]interface{}{"$ref": "#/components/schemas/Product"},
				},
			},
		},
		"responses": map[string]interface{}{
			"201": map[string]interface{}{"description": "Product created"},
		},
	}
}

func GetProduct() map[string]interface{} {
	return map[string]interface{}{
		"summary": "Get product",
		"tags":    []string{"Products"},
		"parameters": []map[string]interface{}{
			{"name": "id", "in": "path", "required": true, "schema": map[string]interface{}{"type": "integer"}},
		},
		"responses": map[string]interface{}{
			"200": map[string]interface{}{"description": "Product found"},
			"404": map[string]interface{}{"description": "Product not found"},
		},
	}
}

func UpdateProduct() map[string]interface{} { return GetProduct() }
func DeleteProduct() map[string]interface{} { return GetProduct() }

func LowStock() map[string]interface{} {
	return map[string]interface{}{
		"summary": "Get low stock products",
		"tags":    []string{"Products"},
		"responses": map[string]interface{}{
			"200": map[string]interface{}{"description": "Low stock products"},
		},
	}
}

func ListSales() map[string]interface{} {
	return map[string]interface{}{
		"summary":     "List sales",
		"description": "Get sales history",
		"tags":        []string{"Sales"},
		"parameters": []map[string]interface{}{
			{"name": "date", "in": "query", "schema": map[string]interface{}{"type": "string"}},
			{"name": "page", "in": "query", "schema": map[string]interface{}{"type": "integer"}},
		},
		"responses": map[string]interface{}{
			"200": map[string]interface{}{"description": "Sales list"},
		},
	}
}

func CreateSale() map[string]interface{} {
	return map[string]interface{}{
		"summary": "Record a sale",
		"tags":    []string{"Sales"},
		"requestBody": map[string]interface{}{
			"required": true,
			"content": map[string]interface{}{
				"application/json": map[string]interface{}{
					"schema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"product_id": map[string]interface{}{"type": "integer"},
							"quantity":   map[string]interface{}{"type": "integer"},
						},
					},
				},
			},
		},
		"responses": map[string]interface{}{
			"201": map[string]interface{}{"description": "Sale recorded"},
		},
	}
}

func GetSale() map[string]interface{} {
	return map[string]interface{}{
		"summary": "Get sale details",
		"tags":    []string{"Sales"},
		"parameters": []map[string]interface{}{
			{"name": "id", "in": "path", "required": true, "schema": map[string]interface{}{"type": "integer"}},
		},
		"responses": map[string]interface{}{
			"200": map[string]interface{}{"description": "Sale details"},
		},
	}
}

func ShopProfile() map[string]interface{} {
	return map[string]interface{}{
		"summary": "Get shop profile",
		"tags":    []string{"Shop"},
		"responses": map[string]interface{}{
			"200": map[string]interface{}{"description": "Shop profile"},
		},
	}
}

func ShopUpdate() map[string]interface{}    { return ShopProfile() }
func ShopDashboard() map[string]interface{} { return ShopProfile() }

func STKPush() map[string]interface{} {
	return map[string]interface{}{
		"summary":     "Initiate M-Pesa STK Push",
		"description": "Send payment request to customer phone",
		"tags":        []string{"Payments"},
		"security":    []map[string]string{{"BearerAuth": ""}},
		"requestBody": map[string]interface{}{
			"required": true,
			"content": map[string]interface{}{
				"application/json": map[string]interface{}{
					"schema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"phone":   map[string]interface{}{"type": "string"},
							"amount":  map[string]interface{}{"type": "integer"},
							"account": map[string]interface{}{"type": "string"},
						},
					},
				},
			},
		},
		"responses": map[string]interface{}{
			"200": map[string]interface{}{"description": "STK Push initiated"},
		},
	}
}

func MpesaCallback() map[string]interface{} {
	return map[string]interface{}{
		"summary":     "M-Pesa webhook callback",
		"description": "Receive payment callbacks from M-Pesa",
		"tags":        []string{"Payments"},
		"responses": map[string]interface{}{
			"200": map[string]interface{}{"description": "Callback received"},
		},
	}
}

func DailyReport() map[string]interface{} {
	return map[string]interface{}{
		"summary": "Daily sales report",
		"tags":    []string{"Reports"},
		"parameters": []map[string]interface{}{
			{"name": "date", "in": "query", "schema": map[string]interface{}{"type": "string"}},
		},
		"responses": map[string]interface{}{
			"200": map[string]interface{}{"description": "Daily report"},
		},
	}
}

func WeeklyReport() map[string]interface{}  { return DailyReport() }
func MonthlyReport() map[string]interface{} { return DailyReport() }

// GenerateDocsJSON returns JSON documentation
func GenerateDocsJSON() (string, error) {
	docs := OpenAPI()
	data, err := json.MarshalIndent(docs, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Markdown generates Markdown API documentation
func Markdown() string {
	var md strings.Builder

	md.WriteString("# DukaPOS API Documentation\n\n")
	md.WriteString("Version: 1.0.0\n\n")
	md.WriteString("## Base URL\n")
	md.WriteString("- Production: `https://api.dukapos.io/v1`\n")
	md.WriteString("- Sandbox: `https://sandbox.dukapos.io/v1`\n\n")

	md.WriteString("## Authentication\n\n")
	md.WriteString("### JWT Token\n")
	md.WriteString("```\nAuthorization: Bearer <token>\n```\n\n")
	md.WriteString("### API Key\n")
	md.WriteString("```\nX-API-Key: <key>\n```\n\n")

	md.WriteString("## Endpoints\n\n")

	// Auth
	md.WriteString("### Auth\n")
	md.WriteString("| Method | Endpoint | Description |\n")
	md.WriteString("|--------|----------|-------------|\n")
	md.WriteString("| POST | /auth/register | Register new shop |\n")
	md.WriteString("| POST | /auth/login | Login |\n\n")

	// Products
	md.WriteString("### Products\n")
	md.WriteString("| Method | Endpoint | Description |\n")
	md.WriteString("|--------|----------|-------------|\n")
	md.WriteString("| GET | /products | List products |\n")
	md.WriteString("| POST | /products | Create product |\n")
	md.WriteString("| GET | /products/:id | Get product |\n")
	md.WriteString("| PUT | /products/:id | Update product |\n")
	md.WriteString("| DELETE | /products/:id | Delete product |\n\n")

	// Sales
	md.WriteString("### Sales\n")
	md.WriteString("| Method | Endpoint | Description |\n")
	md.WriteString("|--------|----------|-------------|\n")
	md.WriteString("| GET | /sales | List sales |\n")
	md.WriteString("| POST | /sales | Record sale |\n")
	md.WriteString("| GET | /sales/:id | Get sale |\n\n")

	// Payments
	md.WriteString("### Payments (Pro)\n")
	md.WriteString("| Method | Endpoint | Description |\n")
	md.WriteString("|--------|----------|-------------|\n")
	md.WriteString("| POST | /mpesa/stk-push | Initiate payment |\n")
	md.WriteString("| POST | /mpesa/callback | Payment callback |\n\n")

	// Reports
	md.WriteString("### Reports\n")
	md.WriteString("| Method | Endpoint | Description |\n")
	md.WriteString("|--------|----------|-------------|\n")
	md.WriteString("| GET | /reports/daily | Daily report |\n")
	md.WriteString("| GET | /reports/weekly | Weekly report |\n")
	md.WriteString("| GET | /reports/monthly | Monthly report |\n\n")

	md.WriteString("## Rate Limits\n")
	md.WriteString("- Default: 60 requests/minute\n")
	md.WriteString("- Products: 100 requests/minute\n")
	md.WriteString("- Sales: 60 requests/minute\n")
	md.WriteString("- Reports: 30 requests/minute\n\n")

	md.WriteString("---\n")
	md.WriteString(fmt.Sprintf("Generated: %s\n", time.Now().Format("2006-01-02")))

	return md.String()
}
