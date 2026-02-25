package handlers

import (
	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type StaffRoleHandler struct {
	db *gorm.DB
}

func NewStaffRoleHandler(db *gorm.DB) *StaffRoleHandler {
	return &StaffRoleHandler{db: db}
}

func (h *StaffRoleHandler) List(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)

	var roles []models.StaffRole
	if err := h.db.Where("shop_id = ?", shopID).Order("created_at DESC").Find(&roles).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": roles})
}

func (h *StaffRoleHandler) Get(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)
	id := c.Params("id")

	var role models.StaffRole
	if err := h.db.Where("id = ? AND shop_id = ?", id, shopID).First(&role).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "role not found"})
	}

	return c.JSON(fiber.Map{"data": role})
}

func (h *StaffRoleHandler) Create(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)

	type Request struct {
		Name        string `json:"name"`
		Permissions string `json:"permissions"`
		Description string `json:"description"`
		IsDefault   bool   `json:"is_default"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "name is required"})
	}

	// If setting as default, unset other defaults
	if req.IsDefault {
		h.db.Model(&models.StaffRole{}).Where("shop_id = ? AND is_default = ?", shopID, true).Update("is_default", false)
	}

	role := models.StaffRole{
		ShopID:      shopID,
		Name:        req.Name,
		Permissions: req.Permissions,
		Description: req.Description,
		IsDefault:   req.IsDefault,
	}

	if err := h.db.Create(&role).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": role})
}

func (h *StaffRoleHandler) Update(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)
	id := c.Params("id")

	type Request struct {
		Name        string `json:"name"`
		Permissions string `json:"permissions"`
		Description string `json:"description"`
		IsDefault   bool   `json:"is_default"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	var role models.StaffRole
	if err := h.db.Where("id = ? AND shop_id = ?", id, shopID).First(&role).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "role not found"})
	}

	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Permissions != "" {
		role.Permissions = req.Permissions
	}
	if req.Description != "" {
		role.Description = req.Description
	}

	// If setting as default, unset other defaults
	if req.IsDefault && !role.IsDefault {
		h.db.Model(&models.StaffRole{}).Where("shop_id = ? AND is_default = ?", shopID, true).Update("is_default", false)
		role.IsDefault = true
	}

	if err := h.db.Save(&role).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": role})
}

func (h *StaffRoleHandler) Delete(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)
	id := c.Params("id")

	// Check if this is the last role
	var count int64
	h.db.Model(&models.StaffRole{}).Where("shop_id = ?", shopID).Count(&count)
	if count <= 1 {
		return c.Status(400).JSON(fiber.Map{"error": "cannot delete the last role"})
	}

	if err := h.db.Where("id = ? AND shop_id = ?", id, shopID).Delete(&models.StaffRole{}).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true})
}
