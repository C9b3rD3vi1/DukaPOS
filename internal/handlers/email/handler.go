package emailhandler

import (
	"github.com/C9b3rD3vi1/DukaPOS/internal/services/email"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	emailSvc *email.Service
}

func New(emailSvc *email.Service) *Handler {
	return &Handler{emailSvc: emailSvc}
}

func (h *Handler) SendEmail(c *fiber.Ctx) error {
	type SendRequest struct {
		To      string `json:"to"`
		ToName  string `json:"to_name"`
		Subject string `json:"subject"`
		Body    string `json:"body"`
		HTML    string `json:"html"`
	}

	var req SendRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	if req.To == "" || req.Subject == "" || req.Body == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "to, subject, and body required"})
	}

	err := h.emailSvc.SendEmail(&email.Email{
		To:      req.To,
		ToName:  req.ToName,
		Subject: req.Subject,
		Body:    req.Body,
		HTML:    req.HTML,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) SendWelcomeEmail(c *fiber.Ctx) error {
	type Request struct {
		To       string `json:"to"`
		ShopName string `json:"shop_name"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	if req.To == "" || req.ShopName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "to and shop_name required"})
	}

	err := h.emailSvc.SendWelcomeEmail(req.To, req.ShopName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true})
}

func (h *Handler) GetHistory(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"data": []interface{}{}})
}

func (h *Handler) RegisterRoutes(protected fiber.Router) {
	emailRoutes := protected.Group("/email")
	emailRoutes.Post("/send", h.SendEmail)
	emailRoutes.Post("/welcome", h.SendWelcomeEmail)
}
