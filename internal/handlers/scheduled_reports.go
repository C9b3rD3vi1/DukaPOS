package handlers

import (
	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type ScheduledReportHandler struct {
	db *gorm.DB
}

func NewScheduledReportHandler(db *gorm.DB) *ScheduledReportHandler {
	return &ScheduledReportHandler{db: db}
}

func (h *ScheduledReportHandler) List(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)

	var reports []models.ScheduledReport
	if err := h.db.Where("shop_id = ?", shopID).Order("created_at DESC").Find(&reports).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": reports})
}

func (h *ScheduledReportHandler) Get(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)
	id := c.Params("id")

	var report models.ScheduledReport
	if err := h.db.Where("id = ? AND shop_id = ?", id, shopID).First(&report).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "report not found"})
	}

	return c.JSON(fiber.Map{"data": report})
}

func (h *ScheduledReportHandler) Create(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)

	type Request struct {
		Type       string `json:"type"`
		Frequency  string `json:"frequency"`
		Time       string `json:"time"`
		Enabled    bool   `json:"enabled"`
		Recipients string `json:"recipients"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.Type == "" || req.Time == "" {
		return c.Status(400).JSON(fiber.Map{"error": "type and time are required"})
	}

	report := models.ScheduledReport{
		ShopID:     shopID,
		Type:       req.Type,
		Frequency:  req.Frequency,
		Time:       req.Time,
		Enabled:    req.Enabled,
		Recipients: req.Recipients,
	}

	if err := h.db.Create(&report).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": report})
}

func (h *ScheduledReportHandler) Update(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)
	id := c.Params("id")

	type Request struct {
		Type       string `json:"type"`
		Frequency  string `json:"frequency"`
		Time       string `json:"time"`
		Enabled    bool   `json:"enabled"`
		Recipients string `json:"recipients"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	var report models.ScheduledReport
	if err := h.db.Where("id = ? AND shop_id = ?", id, shopID).First(&report).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "report not found"})
	}

	if req.Type != "" {
		report.Type = req.Type
	}
	if req.Frequency != "" {
		report.Frequency = req.Frequency
	}
	if req.Time != "" {
		report.Time = req.Time
	}
	report.Enabled = req.Enabled
	if req.Recipients != "" {
		report.Recipients = req.Recipients
	}

	if err := h.db.Save(&report).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": report})
}

func (h *ScheduledReportHandler) Delete(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)
	id := c.Params("id")

	if err := h.db.Where("id = ? AND shop_id = ?", id, shopID).Delete(&models.ScheduledReport{}).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true})
}
