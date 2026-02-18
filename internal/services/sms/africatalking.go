package sms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Config holds Africa Talking configuration
type Config struct {
	APIKey   string
	Username string
	BaseURL  string
}

// Service handles SMS sending via Africa Talking
type Service struct {
	config *Config
	client *http.Client
}

// New creates a new Africa Talking SMS service
func New(config *Config) *Service {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.africastalking.com"
	}
	return &Service{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SendSMS sends an SMS message
func (s *Service) SendSMS(to, message string) (string, error) {
	if s.config.APIKey == "" || s.config.Username == "" {
		return "", fmt.Errorf("Africa Talking credentials not configured")
	}

	// Format phone number
	to = formatPhone(to)

	data := map[string]string{
		"username": s.config.Username,
		"to":       to,
		"message":  message,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", s.config.BaseURL+"/version1/messaging", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apiKey", s.config.APIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("SMS send failed: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	// Check for errors in response
	if recipients, ok := result["recipients"].(map[string]interface{}); ok {
		if recipient, ok := recipients[to].(map[string]interface{}); ok {
			if status, ok := recipient["status"].(string); ok {
				if status == "Success" {
					return fmt.Sprintf("SMS sent to %s", to), nil
				}
				return "", fmt.Errorf("SMS failed: %s", recipient["errorMessage"])
			}
		}
	}

	return fmt.Sprintf("SMS sent to %s", to), nil
}

// SendBulkSMS sends SMS to multiple recipients
func (s *Service) SendBulkSMS(recipients []string, message string) (map[string]string, error) {
	results := make(map[string]string)

	for _, to := range recipients {
		result, err := s.SendSMS(to, message)
		if err != nil {
			results[to] = err.Error()
		} else {
			results[to] = result
		}
		// Rate limiting - sleep between sends
		time.Sleep(100 * time.Millisecond)
	}

	return results, nil
}

// GetBalance gets SMS balance
func (s *Service) GetBalance() (string, error) {
	if s.config.APIKey == "" || s.config.Username == "" {
		return "", fmt.Errorf("Africa Talking credentials not configured")
	}

	req, err := http.NewRequest("GET", s.config.BaseURL+"/version1/user", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("apiKey", s.config.APIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if userData, ok := result["UserData"].(map[string]interface{}); ok {
		if balance, ok := userData["smsBalance"].(string); ok {
			return balance, nil
		}
	}

	return "", fmt.Errorf("Could not get balance")
}

// formatPhone formats phone number for Africa Talking
func formatPhone(phone string) string {
	// Remove all non-digits
	var digits string
	for _, c := range phone {
		if c >= '0' && c <= '9' {
			digits += string(c)
		}
	}

	// Handle different formats
	if len(digits) == 10 && digits[0] == '0' {
		return "+254" + digits[1:]
	} else if len(digits) == 9 {
		return "+254" + digits
	} else if len(digits) == 12 && digits[:3] == "254" {
		return "+" + digits
	}

	return phone
}
