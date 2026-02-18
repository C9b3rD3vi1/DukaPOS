package api

import (
	"strconv"

	"github.com/C9b3rD3vi1/DukaPOS/internal/services/api"
	"github.com/gofiber/fiber/v2"
)

// APIKeyHandler handles API key HTTP requests
type APIKeyHandler struct {
	service *api.Service
}

// NewAPIKeyHandler creates a new API key handler
func NewAPIKeyHandler(service *api.Service) *APIKeyHandler {
	return &APIKeyHandler{service: service}
}

// List returns all API keys for a shop
// GET /api/v1/api-keys
func (h *APIKeyHandler) List(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)

	keys, err := h.service.ListByShop(shopID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": keys,
		"total": len(keys),
	})
}

// Create creates a new API key
// POST /api/v1/api-keys
func (h *APIKeyHandler) Create(c *fiber.Ctx) error {
	type Request struct {
		Name        string `json:"name"`
		Permissions string `json:"permissions"`
		RateLimit   int    `json:"rate_limit"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "name is required",
		})
	}

	shopID := c.Locals("shop_id").(uint)
	rateLimit := req.RateLimit
	if rateLimit == 0 {
		rateLimit = 60
	}

	key, err := h.service.CreateKey(shopID, req.Name, req.Permissions, rateLimit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "API key created successfully",
		"warning": "Save the secret - it won't be shown again!",
		"data": fiber.Map{
			"id":          key.ID,
			"name":        key.Name,
			"key":         key.Key,
			"secret":      key.SecretHash,
			"permissions": key.Permissions,
			"rate_limit":  key.RateLimit,
			"is_active":   key.IsActive,
			"created_at": key.CreatedAt,
		},
	})
}

// Revoke revokes an API key
// DELETE /api/v1/api-keys/:id
func (h *APIKeyHandler) Revoke(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid API key ID",
		})
	}

	err = h.service.RevokeKey(uint(id))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "API key revoked successfully",
	})
}
