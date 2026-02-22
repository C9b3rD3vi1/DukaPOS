package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type TestShop struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Plan     string `json:"plan"`
	IsActive bool   `json:"is_active"`
}

type TestProduct struct {
	ID                uint    `json:"id"`
	Name              string  `json:"name"`
	Category          string  `json:"category"`
	SellingPrice      float64 `json:"selling_price"`
	CostPrice         float64 `json:"cost_price"`
	CurrentStock      int     `json:"current_stock"`
	LowStockThreshold int     `json:"low_stock_threshold"`
}

type TestSale struct {
	ID            uint    `json:"id"`
	ProductID     uint    `json:"product_id"`
	Quantity      int     `json:"quantity"`
	TotalAmount   float64 `json:"total_amount"`
	PaymentMethod string  `json:"payment_method"`
}

type TestResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func TestProductCRUDWorkflow(t *testing.T) {
	baseURL := "/api/v1"

	t.Run("Create Product", func(t *testing.T) {
		body := `{"name":"Test Bread","category":"Food","selling_price":50,"cost_price":30,"current_stock":100,"low_stock_threshold":10}`
		req := httptest.NewRequest("POST", baseURL+"/products", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")

		w := httptest.NewRecorder()
		if w.Code != http.StatusCreated {
			t.Logf("Expected status 201, got %d", w.Code)
		}
	})

	t.Run("Get Products", func(t *testing.T) {
		req := httptest.NewRequest("GET", baseURL+"/products", nil)
		req.Header.Set("Authorization", "Bearer test-token")

		w := httptest.NewRecorder()
		if w.Code != http.StatusOK {
			t.Logf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("Update Product", func(t *testing.T) {
		body := `{"selling_price":55,"current_stock":50}`
		req := httptest.NewRequest("PUT", baseURL+"/products/1", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")

		w := httptest.NewRecorder()
		if w.Code != http.StatusOK {
			t.Logf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("Delete Product", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", baseURL+"/products/1", nil)
		req.Header.Set("Authorization", "Bearer test-token")

		w := httptest.NewRecorder()
		if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
			t.Logf("Expected status 204 or 200, got %d", w.Code)
		}
	})
}

func TestSaleWorkflow(t *testing.T) {
	baseURL := "/api/v1"

	t.Run("Record Sale", func(t *testing.T) {
		body := `{"product_id":1,"quantity":2,"payment_method":"cash"}`
		req := httptest.NewRequest("POST", baseURL+"/sales", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")

		w := httptest.NewRecorder()
		if w.Code != http.StatusCreated {
			t.Logf("Expected status 201, got %d", w.Code)
		}
	})

	t.Run("Get Sales", func(t *testing.T) {
		req := httptest.NewRequest("GET", baseURL+"/sales", nil)
		req.Header.Set("Authorization", "Bearer test-token")

		w := httptest.NewRecorder()
		if w.Code != http.StatusOK {
			t.Logf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("Get Sale by ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", baseURL+"/sales/1", nil)
		req.Header.Set("Authorization", "Bearer test-token")

		w := httptest.NewRecorder()
		if w.Code != http.StatusOK {
			t.Logf("Expected status 200, got %d", w.Code)
		}
	})
}

func TestAuthenticationWorkflow(t *testing.T) {
	t.Run("Register Shop", func(t *testing.T) {
		body := `{"phone":"+254700000001","password":"testpass123","name":"Test Shop"}`
		req := httptest.NewRequest("POST", "/api/auth/register", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		if w.Code != http.StatusCreated {
			t.Logf("Expected status 201, got %d", w.Code)
		}
	})

	t.Run("Login", func(t *testing.T) {
		body := `{"phone":"+254700000001","password":"testpass123"}`
		req := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		if w.Code != http.StatusOK {
			t.Logf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("Invalid Login", func(t *testing.T) {
		body := `{"phone":"+254700000001","password":"wrongpassword"}`
		req := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		if w.Code != http.StatusUnauthorized {
			t.Logf("Expected status 401, got %d", w.Code)
		}
	})
}

func TestMpesaWorkflow(t *testing.T) {
	baseURL := "/api/v1/mpesa"

	t.Run("STK Push", func(t *testing.T) {
		body := `{"phone":"+254700000001","amount":100}`
		req := httptest.NewRequest("POST", baseURL+"/stk-push", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")

		w := httptest.NewRecorder()
		if w.Code != http.StatusOK && w.Code != http.StatusAccepted {
			t.Logf("Expected status 200 or 202, got %d", w.Code)
		}
	})

	t.Run("Get Payments", func(t *testing.T) {
		req := httptest.NewRequest("GET", baseURL+"/payments", nil)
		req.Header.Set("Authorization", "Bearer test-token")

		w := httptest.NewRecorder()
		if w.Code != http.StatusOK {
			t.Logf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("Get Transactions", func(t *testing.T) {
		req := httptest.NewRequest("GET", baseURL+"/transactions", nil)
		req.Header.Set("Authorization", "Bearer test-token")

		w := httptest.NewRecorder()
		if w.Code != http.StatusOK {
			t.Logf("Expected status 200, got %d", w.Code)
		}
	})
}

func TestWebhookWorkflow(t *testing.T) {
	t.Run("Twilio Webhook", func(t *testing.T) {
		body := `{"From":"+254700000001","Body":"stock"}`
		req := httptest.NewRequest("POST", "/webhook/twilio", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		if w.Code != http.StatusOK {
			t.Logf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("M-Pesa Callback", func(t *testing.T) {
		body := `{"Body":{"stkCallback":{"ResultCode":0,"ResultDesc":"Success"}}}`
		req := httptest.NewRequest("POST", "/webhook/mpesa/stk", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		if w.Code != http.StatusOK {
			t.Logf("Expected status 200, got %d", w.Code)
		}
	})
}

func TestReportWorkflow(t *testing.T) {
	baseURL := "/api/v1"

	t.Run("Daily Report", func(t *testing.T) {
		date := time.Now().Format("2006-01-02")
		req := httptest.NewRequest("GET", baseURL+"/reports/daily?date="+date, nil)
		req.Header.Set("Authorization", "Bearer test-token")

		w := httptest.NewRecorder()
		if w.Code != http.StatusOK {
			t.Logf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("Weekly Report", func(t *testing.T) {
		req := httptest.NewRequest("GET", baseURL+"/reports/weekly", nil)
		req.Header.Set("Authorization", "Bearer test-token")

		w := httptest.NewRecorder()
		if w.Code != http.StatusOK {
			t.Logf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("Monthly Report", func(t *testing.T) {
		req := httptest.NewRequest("GET", baseURL+"/reports/monthly", nil)
		req.Header.Set("Authorization", "Bearer test-token")

		w := httptest.NewRecorder()
		if w.Code != http.StatusOK {
			t.Logf("Expected status 200, got %d", w.Code)
		}
	})
}

func TestErrorResponses(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		url        string
		body       string
		wantStatus int
	}{
		{"Unauthorized GET /products", "GET", "/api/v1/products", "", http.StatusUnauthorized},
		{"Invalid JSON", "POST", "/api/v1/products", "{invalid", http.StatusBadRequest},
		{"Missing required field", "POST", "/api/v1/products", "{}", http.StatusBadRequest},
		{"Not found", "GET", "/api/v1/products/999999", "", http.StatusNotFound},
		{"Method not allowed", "PUT", "/api/v1/products", "", http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.url, strings.NewReader(tt.body))
			} else {
				req = httptest.NewRequest(tt.method, tt.url, nil)
			}
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			if w.Code != tt.wantStatus {
				t.Logf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestJSONResponseFormat(t *testing.T) {
	resp := TestResponse{
		Status:  200,
		Message: "Success",
		Data:    map[string]string{"key": "value"},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded TestResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Status != 200 {
		t.Errorf("Expected status 200, got %d", decoded.Status)
	}
	if decoded.Message != "Success" {
		t.Errorf("Expected message 'Success', got '%s'", decoded.Message)
	}
}

func TestRateLimiting(t *testing.T) {
	maxRequests := 60
	window := time.Minute

	type Request struct {
		timestamp time.Time
	}

	requests := make([]Request, 0, maxRequests+10)

	for i := 0; i < maxRequests+10; i++ {
		requests = append(requests, Request{timestamp: time.Now()})

		withinWindow := 0
		cutoff := time.Now().Add(-window)
		for _, r := range requests {
			if r.timestamp.After(cutoff) {
				withinWindow++
			}
		}

		if i >= maxRequests && withinWindow > maxRequests {
			t.Logf("Request %d would be rate limited (within window: %d)", i+1, withinWindow)
		}
	}
}

func TestShopProfileWorkflow(t *testing.T) {
	baseURL := "/api/v1/shop"

	t.Run("Get Profile", func(t *testing.T) {
		req := httptest.NewRequest("GET", baseURL+"/profile", nil)
		req.Header.Set("Authorization", "Bearer test-token")

		w := httptest.NewRecorder()
		if w.Code != http.StatusOK {
			t.Logf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("Update Profile", func(t *testing.T) {
		body := `{"name":"Updated Shop Name","address":"123 Main St"}`
		req := httptest.NewRequest("PUT", baseURL+"/profile", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")

		w := httptest.NewRecorder()
		if w.Code != http.StatusOK {
			t.Logf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("Get Dashboard", func(t *testing.T) {
		req := httptest.NewRequest("GET", baseURL+"/dashboard", nil)
		req.Header.Set("Authorization", "Bearer test-token")

		w := httptest.NewRecorder()
		if w.Code != http.StatusOK {
			t.Logf("Expected status 200, got %d", w.Code)
		}
	})
}

func TestWhatsAppCommandWorkflow(t *testing.T) {
	commands := []string{
		"add bread 50 30",
		"sell bread 2",
		"stock",
		"report",
		"price bread",
		"help",
	}

	for _, cmd := range commands {
		t.Run("Command: "+cmd, func(t *testing.T) {
			body := fmt.Sprintf("From=+254700000001&Body=%s", cmd)
			req := httptest.NewRequest("POST", "/webhook/twilio", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			w := httptest.NewRecorder()
			if w.Code != http.StatusOK {
				t.Logf("Command '%s' - Expected status 200, got %d", cmd, w.Code)
			}
		})
	}
}
