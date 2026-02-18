package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/C9b3rD3vi1/DukaPOS/internal/config"
	"github.com/C9b3rD3vi1/DukaPOS/internal/services"
	"github.com/gofiber/fiber/v2"
)

// WhatsAppHandler handles WhatsApp webhooks from Twilio
type WhatsAppHandler struct {
	cmdHandler *services.CommandHandler
	cfg        *config.Config
	httpClient *http.Client
}

// NewWhatsAppHandler creates a new WhatsApp handler
func NewWhatsAppHandler(cmdHandler *services.CommandHandler, cfg *config.Config) *WhatsAppHandler {
	return &WhatsAppHandler{
		cmdHandler: cmdHandler,
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// HandleWebhook handles incoming WhatsApp messages
func (h *WhatsAppHandler) HandleWebhook(c *fiber.Ctx) error {
	from := c.FormValue("From")
	body := c.FormValue("Body")

	if from == "" || body == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing required fields: From, Body",
		})
	}

	phone := extractPhoneFromWhatsApp(from)
	fmt.Printf("📱 WhatsApp message from %s: %s\n", phone, body)

	// Create a simple parser
	parser := services.NewCommandParser(nil, nil)
	cmd := parser.Parse(body)

	response, err := h.cmdHandler.Handle(phone, cmd)
	if err != nil {
		fmt.Printf("❌ Error handling message: %v\n", err)
		response = "❌ An error occurred. Please try again."
	}

	// Return TwiML XML response for Twilio WhatsApp
	return c.Type("xml").SendString(h.generateTwiML(response))
}

// generateTwiML generates Twilio's TwiML XML response
func (h *WhatsAppHandler) generateTwiML(message string) string {
	escapedMessage := escapeXML(message)
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<Response>
    <Message>%s</Message>
</Response>`, escapedMessage)
}

// escapeXML escapes special characters for XML
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// HandleStatusCallback handles message status callbacks
func (h *WhatsAppHandler) HandleStatusCallback(c *fiber.Ctx) error {
	messageSid := c.FormValue("MessageSid")
	messageStatus := c.FormValue("MessageStatus")

	fmt.Printf("📊 Message Status Update: %s - %s\n", messageSid, messageStatus)

	return c.SendStatus(fiber.StatusOK)
}

// SendWhatsAppMessage sends a WhatsApp message (for notifications)
func (h *WhatsAppHandler) SendWhatsAppMessage(to, message string) error {
	// First try to send via Twilio API
	if h.cfg.TwilioAccountSID != "" && h.cfg.TwilioAuthToken != "" && h.cfg.TwilioWhatsAppNumber != "" {
		return h.SendWhatsAppMessageWithTwilio(to, message)
	}
	// Fallback to console log
	fmt.Printf("📤 [SCHEDULER] Would send WhatsApp to %s: %s\n", to, message)
	return nil
}

// SendWhatsAppMessageWithTwilio sends actual WhatsApp message via Twilio API
func (h *WhatsAppHandler) SendWhatsAppMessageWithTwilio(to, message string) error {
	if h.cfg.TwilioAccountSID == "" || h.cfg.TwilioAuthToken == "" || h.cfg.TwilioWhatsAppNumber == "" {
		return fmt.Errorf("Twilio credentials not configured")
	}

	twilioURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json",
		h.cfg.TwilioAccountSID)

	from := h.cfg.TwilioWhatsAppNumber
	if !strings.HasPrefix(to, "whatsapp:") {
		to = "whatsapp:" + to
	}

	data := url.Values{}
	data.Set("From", from)
	data.Set("To", to)
	data.Set("Body", message)

	req, err := http.NewRequest("POST", twilioURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(h.cfg.TwilioAccountSID, h.cfg.TwilioAuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("Twilio error (%d): %s", resp.StatusCode, errResp.Message)
	}

	fmt.Printf("✅ WhatsApp message sent to %s\n", to)
	return nil
}

// extractPhoneFromWhatsApp extracts phone from WhatsApp format
func extractPhoneFromWhatsApp(whatsapp string) string {
	if strings.HasPrefix(whatsapp, "whatsapp:") {
		return strings.TrimPrefix(whatsapp, "whatsapp:")
	}
	return whatsapp
}

// WebhookVerification handles Twilio's webhook verification
func (h *WhatsAppHandler) WebhookVerification(c *fiber.Ctx) error {
	mode := c.Query("hub.mode")
	challenge := c.Query("hub.challenge")
	token := c.Query("hub.verify_token")

	expectedToken := h.cfg.TwilioAuthTokenConfirm
	if expectedToken == "" {
		fmt.Printf("⚠️ Webhook verification skipped: no token configured\n")
		return c.SendString(challenge)
	}

	if token != expectedToken {
		fmt.Printf("⚠️ Webhook verification failed: token mismatch\n")
		return c.Status(fiber.StatusForbidden).SendString("Verification failed")
	}

	if mode == "subscribe" {
		fmt.Println("✅ Webhook verified!")
		return c.SendString(challenge)
	}

	return c.Status(fiber.StatusBadRequest).SendString("Invalid mode")
}
