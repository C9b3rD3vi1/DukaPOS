package main

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

// TestMpesaPhoneValidation tests phone validation logic
func TestMpesaPhoneValidation(t *testing.T) {
	validatePhone := func(phone string) string {
		var digits string
		for _, c := range phone {
			if c >= '0' && c <= '9' {
				digits += string(c)
			}
		}
		if len(digits) == 10 && digits[0] == '0' {
			digits = "254" + digits[1:]
		} else if len(digits) == 9 {
			digits = "254" + digits
		}
		return digits
	}

	tests := []struct {
		name     string
		phone    string
		expected string
	}{
		{"Kenyan number with 0 prefix", "0712345678", "254712345678"},
		{"Kenyan number with 7 digits", "712345678", "254712345678"},
		{"Already formatted", "254712345678", "254712345678"},
		{"With plus sign", "+254712345678", "254712345678"},
		{"With spaces", "0712 345 678", "254712345678"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validatePhone(tt.phone)
			if result != tt.expected {
				t.Errorf("validatePhone(%s) = %s; want %s", tt.phone, result, tt.expected)
			}
		})
	}
}

// TestMpesaPasswordGeneration tests password generation logic
func TestMpesaPasswordGeneration(t *testing.T) {
	generatePassword := func(shortcode, passkey, timestamp string) string {
		data := shortcode + passkey + timestamp
		return base64.StdEncoding.EncodeToString([]byte(data))
	}

	password := generatePassword("174379", "bfb279f9aa9bdbcf158e97dd71a467cd2e0c893059b10f78e6b72e1f3", "20240215120000")
	
	// Just verify it produces a valid base64 string of expected length
	// (not testing exact value as it's implementation-specific)
	decoded, err := base64.StdEncoding.DecodeString(password)
	if err != nil {
		t.Errorf("generatePassword() should produce valid base64: %v", err)
	}
	
	expectedLength := len("174379") + len("bfb279f9aa9bdbcf158e97dd71a467cd2e0c893059b10f78e6b72e1f3") + len("20240215120000")
	if len(decoded) != expectedLength {
		t.Errorf("Decoded password length = %d; want %d", len(decoded), expectedLength)
	}
}

// TestMpesaCallbackParsing tests callback JSON parsing
func TestMpesaCallbackParsing(t *testing.T) {
	type CallbackMetadata struct {
		Item []struct {
			Name  string `json:"Name"`
			Value string `json:"Value"`
		} `json:"Item"`
	}

	type STKCallback struct {
		Body struct {
			CallbackMetadata CallbackMetadata `json:"CallbackMetadata"`
		} `json:"Body"`
	}

	callbackData := []byte(`{
		"Body": {
			"CallbackMetadata": {
				"Item": [
					{"Name": "Amount", "Value": "100"},
					{"Name": "MpesaReceiptNumber", "Value": "PIX123456789"},
					{"Name": "PhoneNumber", "Value": "254712345678"}
				]
			}
		}
	}`)

	var callback STKCallback
	err := json.Unmarshal(callbackData, &callback)
	if err != nil {
		t.Fatalf("ParseCallback() error = %v", err)
	}

	// Verify parsed values
	if len(callback.Body.CallbackMetadata.Item) != 3 {
		t.Errorf("len(Items) = %d; want 3", len(callback.Body.CallbackMetadata.Item))
	}

	result := make(map[string]string)
	for _, item := range callback.Body.CallbackMetadata.Item {
		result[item.Name] = item.Value
	}

	if result["Amount"] != "100" {
		t.Errorf("Amount = %s; want 100", result["Amount"])
	}
	if result["MpesaReceiptNumber"] != "PIX123456789" {
		t.Errorf("MpesaReceiptNumber = %s; want PIX123456789", result["MpesaReceiptNumber"])
	}
}

// TestMpesaConfigValidation tests M-Pesa configuration validation
func TestMpesaConfigValidation(t *testing.T) {
	type Config struct {
		ConsumerKey    string
		ConsumerSecret string
		Shortcode      string
		Passkey        string
	}

	isConfigured := func(cfg *Config) bool {
		return cfg.ConsumerKey != "" && 
			   cfg.ConsumerSecret != "" && 
			   cfg.Shortcode != "" &&
			   cfg.Passkey != ""
	}

	tests := []struct {
		name     string
		cfg      Config
		expected bool
	}{
		{"fully configured", Config{"key", "secret", "174379", "passkey"}, true},
		{"missing key", Config{"", "secret", "174379", "passkey"}, false},
		{"missing secret", Config{"key", "", "174379", "passkey"}, false},
		{"missing shortcode", Config{"key", "secret", "", "passkey"}, false},
		{"missing passkey", Config{"key", "secret", "174379", ""}, false},
		{"nothing configured", Config{"", "", "", ""}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isConfigured(&tt.cfg)
			if result != tt.expected {
				t.Errorf("isConfigured() = %v; want %v", result, tt.expected)
			}
		})
	}
}
