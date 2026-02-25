package handlers

import (
	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type WhiteLabelHandler struct {
	db *gorm.DB
}

func NewWhiteLabelHandler(db *gorm.DB) *WhiteLabelHandler {
	return &WhiteLabelHandler{db: db}
}

func (h *WhiteLabelHandler) Get(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)

	var config models.WhiteLabelConfig
	if err := h.db.Where("shop_id = ?", shopID).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(fiber.Map{"data": nil})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": config})
}

func (h *WhiteLabelHandler) Update(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)

	type Request struct {
		BrandName      string `json:"brand_name"`
		BrandColor     string `json:"brand_color"`
		LogoURL        string `json:"logo_url"`
		CustomCSS      string `json:"custom_css"`
		CustomDomain   string `json:"custom_domain"`
		WhatsAppNumber string `json:"whatsapp_number"`
		Enabled        bool   `json:"enabled"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	var config models.WhiteLabelConfig
	if err := h.db.Where("shop_id = ?", shopID).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			config = models.WhiteLabelConfig{ShopID: shopID}
		} else {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
	}

	if req.BrandName != "" {
		config.BrandName = req.BrandName
	}
	if req.BrandColor != "" {
		config.BrandColor = req.BrandColor
	}
	if req.LogoURL != "" {
		config.LogoURL = req.LogoURL
	}
	if req.CustomCSS != "" {
		config.CustomCSS = req.CustomCSS
	}
	if req.CustomDomain != "" {
		config.CustomDomain = req.CustomDomain
	}
	if req.WhatsAppNumber != "" {
		config.WhatsAppNumber = req.WhatsAppNumber
	}
	config.Enabled = req.Enabled

	if config.ID == 0 {
		if err := h.db.Create(&config).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
	} else {
		if err := h.db.Save(&config).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.JSON(fiber.Map{"data": config})
}
