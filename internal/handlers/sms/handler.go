package smshandler

import (
	"github.com/C9b3rD3vi1/DukaPOS/internal/services/sms"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	smsSvc *sms.Service
}

func New(smsSvc *sms.Service) *Handler {
	return &Handler{smsSvc: smsSvc}
}

func (h *Handler) SendSMS(c *fiber.Ctx) error {
	type SendRequest struct {
		Phone   string `json:"phone"`
		Message string `json:"message"`
	}

	var req SendRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	if req.Phone == "" || req.Message == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "phone and message required"})
	}

	result, err := h.smsSvc.SendSMS(req.Phone, req.Message)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "result": result})
}

func (h *Handler) SendBulkSMS(c *fiber.Ctx) error {
	type BulkRequest struct {
		Recipients []string `json:"recipients"`
		Message    string   `json:"message"`
	}

	var req BulkRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	if len(req.Recipients) == 0 || req.Message == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "recipients and message required"})
	}

	results, err := h.smsSvc.SendBulkSMS(req.Recipients, req.Message)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "results": results})
}

func (h *Handler) GetBalance(c *fiber.Ctx) error {
	if h.smsSvc == nil {
		return c.Status(503).JSON(fiber.Map{"error": "SMS service not configured"})
	}

	balance, err := h.smsSvc.GetBalance()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "balance": balance})
}

func (h *Handler) GetHistory(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"data": []interface{}{}})
}

func (h *Handler) RegisterRoutes(app *fiber.App, protected fiber.Router) {
	if h.smsSvc == nil {
		return
	}

	smsRoutes := protected.Group("/sms")
	smsRoutes.Post("/send", h.SendSMS)
	smsRoutes.Post("/bulk", h.SendBulkSMS)
	smsRoutes.Get("/balance", h.GetBalance)
	smsRoutes.Get("/history", h.GetHistory)
}
