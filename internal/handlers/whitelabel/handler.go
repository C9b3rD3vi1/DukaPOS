package whitelabel

import (
	"github.com/C9b3rD3vi1/DukaPOS/internal/services/whitelabel"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *whitelabel.Service
}

func NewHandler(service *whitelabel.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(app fiber.Router) {
	branding := app.Group("/branding")
	branding.Get("/", h.GetBranding)
	branding.Put("/", h.UpdateBranding)
	branding.Delete("/reset", h.ResetBranding)
	branding.Get("/css-vars", h.GetCSSVariables)
	branding.Get("/preview/invoice", h.PreviewInvoice)
}

func (h *Handler) GetBranding(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)

	branding, err := h.service.GetBranding(shopID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get branding",
		})
	}

	return c.JSON(branding)
}

func (h *Handler) UpdateBranding(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)

	var config whitelabel.BrandingConfig
	if err := c.BodyParser(&config); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err := h.service.UpdateBranding(shopID, &config)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Branding updated successfully",
	})
}

func (h *Handler) ResetBranding(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)

	err := h.service.ResetBranding(shopID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to reset branding",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Branding reset successfully",
	})
}

func (h *Handler) GetCSSVariables(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)

	cssVars, err := h.service.GenerateCSSVariables(shopID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate CSS variables",
		})
	}

	return c.JSON(cssVars)
}

func (h *Handler) PreviewInvoice(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)

	html, err := h.service.GenerateInvoiceHTML(shopID, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate invoice preview",
		})
	}

	c.Set("Content-Type", "text/html")
	return c.SendString(html)
}
