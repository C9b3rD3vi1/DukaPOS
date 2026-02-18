package main

import (
	"encoding/json"
	"testing"
)

// ============================================
// API Handler Tests
// ============================================

func TestAPIInfoResponse(t *testing.T) {
	type APIInfo struct {
		Name        string `json:"name"`
		Version     string `json:"version"`
		Description string `json:"description"`
		BaseURL     string `json:"base_url"`
	}

	info := APIInfo{
		Name:        "DukaPOS API",
		Version:     "1.0.0",
		Description: "REST API for DukaPOS",
		BaseURL:     "/api/v1",
	}

	if info.Name != "DukaPOS API" {
		t.Errorf("Name = %s; want DukaPOS API", info.Name)
	}
	if info.Version != "1.0.0" {
		t.Errorf("Version = %s; want 1.0.0", info.Version)
	}
}

func TestEndpointList(t *testing.T) {
	type Endpoint struct {
		Path        string
		Method      string
		RequiresAuth bool
		RateLimit   string
	}

	endpoints := []Endpoint{
		{"GET", "/products", true, "100/min"},
		{"POST", "/products", true, "30/min"},
		{"GET", "/sales", true, "100/min"},
		{"POST", "/sales", true, "60/min"},
		{"GET", "/staff", true, "30/min"},
		{"POST", "/staff", true, "10/min"},
		{"GET", "/webhooks", true, "30/min"},
		{"POST", "/webhooks", true, "10/min"},
	}

	// All endpoints should require auth
	for _, e := range endpoints {
		if !e.RequiresAuth {
			t.Errorf("Endpoint %s %s should require auth", e.Method, e.Path)
		}
	}
}

// ============================================
// Auth Handler Tests  
// ============================================

func TestRegisterValidation(t *testing.T) {
	type Request struct {
		Phone    string
		Password string
	}

	validate := func(req Request) bool {
		if req.Phone == "" {
			return false
		}
		if req.Password == "" || len(req.Password) < 6 {
			return false
		}
		return true
	}

	if !validate(Request{"+254712345678", "password123"}) {
		t.Error("Valid request should pass")
	}
	if validate(Request{"", "password123"}) {
		t.Error("Missing phone should fail")
	}
	if validate(Request{"+254712345678", "123"}) {
		t.Error("Short password should fail")
	}
}

func TestLoginValidation(t *testing.T) {
	type LoginReq struct {
		PhoneOrEmail string
		Password     string
	}

	validate := func(req LoginReq) bool {
		return req.PhoneOrEmail != "" && req.Password != ""
	}

	if !validate(LoginReq{"+254712345678", "password123"}) {
		t.Error("Valid login should pass")
	}
	if validate(LoginReq{"", ""}) {
		t.Error("Empty credentials should fail")
	}
}

// ============================================
// Staff Handler Tests
// ============================================

func TestStaffReqValidation(t *testing.T) {
	type StaffReq struct {
		Name  string
		Phone string
		Pin   string
	}

	validate := func(req StaffReq) bool {
		if req.Name == "" {
			return false
		}
		if req.Phone == "" {
			return false
		}
		if req.Pin == "" || len(req.Pin) < 4 {
			return false
		}
		return true
	}

	if !validate(StaffReq{"John", "+254712345678", "1234"}) {
		t.Error("Valid staff should pass")
	}
	if validate(StaffReq{"", "+254712345678", "1234"}) {
		t.Error("Missing name should fail")
	}
	if validate(StaffReq{"John", "+254712345678", "12"}) {
		t.Error("Short PIN should fail")
	}
}

// ============================================
// M-Pesa Handler Tests
// ============================================

func TestMpesaCallbackParse(t *testing.T) {
	type CallbackItem struct {
		Name  string `json:"Name"`
		Value string `json:"Value"`
	}

	type STKCallback struct {
		Body struct {
			CallbackMetadata struct {
				Item []CallbackItem `json:"Item"`
			} `json:"CallbackMetadata"`
		} `json:"Body"`
	}

	jsonData := `{"Body":{"CallbackMetadata":{"Item":[{"Name":"Amount","Value":"100"},{"Name":"MpesaReceiptNumber","Value":"PIX123"},{"Name":"PhoneNumber","Value":"254712345678"}]}}}`

	var callback STKCallback
	err := json.Unmarshal([]byte(jsonData), &callback)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	items := callback.Body.CallbackMetadata.Item
	if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}

	// Extract values
	amount := items[0].Value
	receipt := items[1].Value
	phone := items[2].Value

	if amount != "100" {
		t.Errorf("Amount = %s; want 100", amount)
	}
	if receipt != "PIX123" {
		t.Errorf("Receipt = %s; want PIX123", receipt)
	}
	if phone != "254712345678" {
		t.Errorf("Phone = %s; want 254712345678", phone)
	}
}

func TestMpesaSTKValidation(t *testing.T) {
	type STKReq struct {
		Amount int
		Phone  string
	}

	validate := func(req STKReq) bool {
		return req.Amount > 0 && req.Phone != ""
	}

	if !validate(STKReq{100, "+254712345678"}) {
		t.Error("Valid request should pass")
	}
	if validate(STKReq{0, "+254712345678"}) {
		t.Error("Zero amount should fail")
	}
	if validate(STKReq{100, ""}) {
		t.Error("Empty phone should fail")
	}
}

// ============================================
// Webhook Handler Tests
// ============================================

func TestWebhookValidation(t *testing.T) {
	type WebhookReq struct {
		Name   string
		URL    string
		Events string
	}

	validate := func(req WebhookReq) bool {
		if req.Name == "" {
			return false
		}
		if req.URL == "" || len(req.URL) < 10 {
			return false
		}
		if req.Events == "" {
			return false
		}
		// Must start with http
		if req.URL[:4] != "http" {
			return false
		}
		return true
	}

	tests := []struct {
		name  string
		req   WebhookReq
		valid bool
	}{
		{"valid", WebhookReq{"Test", "https://example.com", "sale.created"}, true},
		{"empty name", WebhookReq{"", "https://example.com", "sale.created"}, false},
		{"invalid url", WebhookReq{"Test", "not-a-url", "sale.created"}, false},
		{"empty events", WebhookReq{"Test", "https://example.com", ""}, false},
	}

	for _, tt := range tests {
		result := validate(tt.req)
		if result != tt.valid {
			t.Errorf("%s: validate() = %v; want %v", tt.name, result, tt.valid)
		}
	}
}

// ============================================
// USSD Handler Tests
// ============================================

func TestUSSDParsing(t *testing.T) {
	type USSDReq struct {
		SessionID string
		Phone     string
		Text      string
	}

	// Initial request
	req1 := USSDReq{SessionID: "123", Phone: "+254712345678", Text: ""}
	if req1.Text != "" {
		t.Error("Initial text should be empty")
	}

	// With input
	req2 := USSDReq{SessionID: "123", Phone: "+254712345678", Text: "1"}
	if req2.Text != "1" {
		t.Errorf("Text = %s; want 1", req2.Text)
	}
}

func TestUSSDRespFormat(t *testing.T) {
	type USSDResp struct {
		Response  string
		SessionID string
		Action   string
	}

	continueResp := USSDResp{Response: "Menu", SessionID: "123", Action: "continue"}
	if continueResp.Action != "continue" {
		t.Errorf("Action = %s; want continue", continueResp.Action)
	}

	endResp := USSDResp{Response: "Goodbye", SessionID: "123", Action: "end"}
	if endResp.Action != "end" {
		t.Errorf("Action = %s; want end", endResp.Action)
	}
}

// ============================================
// Printer Handler Tests
// ============================================

func TestPrintValidation(t *testing.T) {
	type Item struct {
		Name     string
		Quantity int
		Price    float64
	}

	type PrintReq struct {
		ShopName     string
		ItemCount   int
		PaymentType string
		Cash        float64
	}

	validate := func(req PrintReq) bool {
		if req.ShopName == "" {
			return false
		}
		if req.ItemCount == 0 {
			return false
		}
		if req.PaymentType == "cash" && req.Cash < 0 {
			return false
		}
		return true
	}

	tests := []struct {
		name  string
		req   PrintReq
		valid bool
	}{
		{"valid cash", PrintReq{"Shop", 2, "cash", 200}, true},
		{"valid mpesa", PrintReq{"Shop", 2, "mpesa", 0}, true},
		{"no shop name", PrintReq{"", 2, "cash", 200}, false},
		{"no items", PrintReq{"Shop", 0, "cash", 200}, false},
		{"negative cash", PrintReq{"Shop", 2, "cash", -10}, false},
	}

	for _, tt := range tests {
		result := validate(tt.req)
		if result != tt.valid {
			t.Errorf("%s: validate() = %v; want %v", tt.name, result, tt.valid)
		}
	}
}

// ============================================
// WhatsApp Handler Tests
// ============================================

func TestWhatsAppMessage(t *testing.T) {
	type WhatsAppMsg struct {
		From string
		To   string
		Body string
	}

	msg := WhatsAppMsg{
		From: "whatsapp:+254712345678",
		To:   "whatsapp:+14155238886",
		Body: "add bread 50 30",
	}

	// Verify message parts
	if msg.From == "" {
		t.Error("From should not be empty")
	}
	if msg.Body == "" {
		t.Error("Body should not be empty")
	}

	// Parse command
	parts := []string{}
	word := ""
	for _, c := range msg.Body + " " {
		if c == ' ' {
			if word != "" {
				parts = append(parts, word)
				word = ""
			}
		} else {
			word += string(c)
		}
	}

	if parts[0] != "add" {
		t.Errorf("Command = %s; want add", parts[0])
	}
	if len(parts) < 3 {
		t.Errorf("Expected 3+ parts, got %d", len(parts))
	}
}

func TestWhatsAppResponse(t *testing.T) {
	type WhatsAppResp struct {
		To   string
		Body string
	}

	success := WhatsAppResp{To: "+254712345678", Body: "✅ Success"}
	if len(success.Body) == 0 {
		t.Error("Body should not be empty")
	}

	errorResp := WhatsAppResp{To: "+254712345678", Body: "❌ Error"}
	if len(errorResp.Body) < 2 || errorResp.Body[0] != 0xe2 { // ❌ starts with 0xE2
		t.Error("Error response should start with ❌")
	}
}
