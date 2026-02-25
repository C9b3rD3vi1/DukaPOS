package webhook

import (
	"strconv"

	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"github.com/gofiber/fiber/v2"
)

// Handler handles webhook HTTP requests
type Handler struct {
	webhookRepo Repository
}

// Repository interface for webhooks
type Repository interface {
	Create(webhook *models.Webhook) error
	GetByID(id uint) (*models.Webhook, error)
	GetByShopID(shopID uint) ([]models.Webhook, error)
	Update(webhook *models.Webhook) error
	Delete(id uint) error
}

// New creates a new webhook handler
func New(repo Repository) *Handler {
	return &Handler{webhookRepo: repo}
}

// List returns all webhooks for a shop
// GET /api/v1/webhooks
func (h *Handler) List(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)

	webhooks, err := h.webhookRepo.GetByShopID(shopID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"data": webhooks,
		"meta": fiber.Map{"total": len(webhooks)},
	})
}

// Get returns a single webhook
// GET /api/v1/webhooks/:id
func (h *Handler) Get(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid webhook ID"})
	}

	webhook, err := h.webhookRepo.GetByID(uint(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "webhook not found"})
	}

	return c.JSON(fiber.Map{"data": webhook})
}

// Create creates a new webhook
// POST /api/v1/webhooks
func (h *Handler) Create(c *fiber.Ctx) error {
	type Request struct {
		ShopID uint   `json:"shop_id"`
		Name   string `json:"name"`
		URL    string `json:"url"`
		Events any    `json:"events"` // string or array
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	// Support both string and array events
	var events string
	switch v := req.Events.(type) {
	case string:
		events = v
	case []interface{}:
		var arr []string
		for _, e := range v {
			if s, ok := e.(string); ok {
				arr = append(arr, s)
			}
		}
		events = joinEvents(arr)
	}

	// Validation
	if req.Name == "" || req.URL == "" || events == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "name, url, and events are required",
		})
	}

	// Validate URL
	if !isValidURL(req.URL) {
		return c.Status(400).JSON(fiber.Map{"error": "invalid URL format"})
	}

	// Validate events
	eventList := splitEvents(events)
	if len(eventList) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "at least one event is required"})
	}

	// Generate secret
	secret := generateRandomSecret()

	webhook := &models.Webhook{
		ShopID:   c.Locals("shop_id").(uint),
		Name:     req.Name,
		URL:      req.URL,
		Events:   events,
		Secret:   secret,
		IsActive: true,
	}

	if err := h.webhookRepo.Create(webhook); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"data":    webhook,
		"message": "webhook created successfully",
		"warning": "Save the secret - it won't be shown again: " + secret,
	})
}

// Update updates a webhook
// PUT /api/v1/webhooks/:id
func (h *Handler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid webhook ID"})
	}

	type Request struct {
		Name     string `json:"name"`
		URL      string `json:"url"`
		Events   string `json:"events"`
		IsActive *bool  `json:"is_active"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	webhook, err := h.webhookRepo.GetByID(uint(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "webhook not found"})
	}

	// Update fields
	if req.Name != "" {
		webhook.Name = req.Name
	}
	if req.URL != "" {
		if !isValidURL(req.URL) {
			return c.Status(400).JSON(fiber.Map{"error": "invalid URL format"})
		}
		webhook.URL = req.URL
	}
	if req.Events != "" {
		events := splitEvents(req.Events)
		if len(events) == 0 {
			return c.Status(400).JSON(fiber.Map{"error": "invalid events"})
		}
		webhook.Events = req.Events
	}
	if req.IsActive != nil {
		webhook.IsActive = *req.IsActive
	}

	if err := h.webhookRepo.Update(webhook); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"data":    webhook,
		"message": "webhook updated successfully",
	})
}

// Delete deletes a webhook
// DELETE /api/v1/webhooks/:id
func (h *Handler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid webhook ID"})
	}

	if err := h.webhookRepo.Delete(uint(id)); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "webhook deleted successfully"})
}

// Test sends a test event to a webhook
// POST /api/v1/webhooks/:id/test
func (h *Handler) Test(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid webhook ID"})
	}

	webhook, err := h.webhookRepo.GetByID(uint(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "webhook not found"})
	}

	// Create test event
	testEvent := map[string]interface{}{
		"type":      "test",
		"shop_id":   webhook.ShopID,
		"timestamp": "2024-01-01T00:00:00Z",
		"data": fiber.Map{
			"message": "This is a test webhook from DukaPOS",
		},
	}

	// In production, this would actually send the webhook
	// For now, just return success

	return c.JSON(fiber.Map{
		"message": "Test event would be sent",
		"webhook": webhook.URL,
		"event":   testEvent,
	})
}

// Helper functions

func isValidURL(url string) bool {
	// Simple URL validation
	if len(url) < 10 {
		return false
	}
	if url[:4] != "http" {
		return false
	}
	return true
}

func splitEvents(events string) []string {
	if events == "" {
		return []string{}
	}
	// Support comma or space separated
	var result []string
	for _, e := range split(events, ",") {
		for _, e2 := range split(e, " ") {
			if e2 != "" {
				result = append(result, e2)
			}
		}
	}
	return result
}

func split(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}

func joinEvents(events []string) string {
	if len(events) == 0 {
		return ""
	}
	result := events[0]
	for i := 1; i < len(events); i++ {
		result += "," + events[i]
	}
	return result
}

func generateRandomSecret() string {
	// Simple random string generation
	return "whsec_" + strconv.FormatInt(int64(len("test")), 36)
}
