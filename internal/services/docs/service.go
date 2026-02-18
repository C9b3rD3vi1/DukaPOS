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
		"paths":     Paths(),
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
	// Stub - returns empty paths until fixed properly
	return map[string]interface{}{}
}

// Components returns OpenAPI components
func Components() map[string]interface{} {
	return map[string]interface{}{
		"schemas": Schemas(),
		"securitySchemes": map[string]interface{}{
			"BearerAuth": map[string]interface{}{
				"type":        "http",
				"scheme":      "bearer",
				"bearerFormat": "JWT",
				"description": "Enter JWT token",
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
				"id":              map[string]interface{}{"type": "integer"},
				"name":            map[string]interface{}{"type": "string"},
				"category":        map[string]interface{}{"type": "string"},
				"selling_price":   map[string]interface{}{"type": "number"},
				"cost_price":     map[string]interface{}{"type": "number"},
				"current_stock":  map[string]interface{}{"type": "integer"},
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

func Get(op map[string]interface{}) *OperationBuilder { return &OperationBuilder{op: op} }
func Put(op map[string]interface{}) *OperationBuilder { return &OperationBuilder{op: op} }
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
		"tags":       []string{"Products"},
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
		"tags":       []string{"Products"},
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
		"tags":       []string{"Sales"},
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

func ShopUpdate() map[string]interface{} { return ShopProfile() }
func ShopDashboard() map[string]interface{} { return ShopProfile() }

func STKPush() map[string]interface{} {
	return map[string]interface{}{
		"summary":     "Initiate M-Pesa STK Push",
		"description": "Send payment request to customer phone",
		"tags":        []string{"Payments"},
		"security":    []interface{}{map[string]interface{}{"BearerAuth": []interface{}{}}},
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

func WeeklyReport() map[string]interface{} { return DailyReport() }
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
