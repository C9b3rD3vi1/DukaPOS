package staff

import (
	"net/http"
	"strconv"

	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"github.com/C9b3rD3vi1/DukaPOS/internal/repository"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// Handler handles staff HTTP requests
type Handler struct {
	staffRepo *repository.StaffRepository
	shopRepo  *repository.ShopRepository
}

// New creates a new staff handler
func New(staffRepo *repository.StaffRepository, shopRepo *repository.ShopRepository) *Handler {
	return &Handler{
		staffRepo: staffRepo,
		shopRepo:  shopRepo,
	}
}

// List returns all staff for a shop
// GET /api/v1/staff
func (h *Handler) List(c *fiber.Ctx) error {
	shopID, err := strconv.ParseUint(c.Params("shopId"), 10, 32)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid shop ID",
		})
	}

	staff, err := h.staffRepo.GetByShopID(uint(shopID))
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": staff,
		"meta": fiber.Map{
			"total": len(staff),
		},
	})
}

// Get returns a single staff member
// GET /api/v1/staff/:id
func (h *Handler) Get(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid staff ID",
		})
	}

	staff, err := h.staffRepo.GetByID(uint(id))
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "staff not found",
		})
	}

	return c.JSON(fiber.Map{"data": staff})
}

// Create creates a new staff member
// POST /api/v1/staff
func (h *Handler) Create(c *fiber.Ctx) error {
	type Request struct {
		ShopID uint   `json:"shop_id"`
		Name   string `json:"name"`
		Phone  string `json:"phone"`
		Role   string `json:"role"`
		Pin    string `json:"pin"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Validation
	if req.Name == "" || req.Phone == "" || req.Pin == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "name, phone, and pin are required",
		})
	}

	if req.Role == "" {
		req.Role = "staff"
	}

	// Check if shop exists
	_, err := h.shopRepo.GetByID(req.ShopID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "shop not found",
		})
	}

	// Check if staff with phone exists
	existing, _ := h.staffRepo.GetByPhone(req.ShopID, req.Phone)
	if existing != nil {
		return c.Status(http.StatusConflict).JSON(fiber.Map{
			"error": "staff with this phone already exists",
		})
	}

	// Create staff with hashed PIN
	hashedPin, err := bcrypt.GenerateFromPassword([]byte(req.Pin), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to hash PIN",
		})
	}

	staff := &models.Staff{
		ShopID:   req.ShopID,
		Name:     req.Name,
		Phone:    req.Phone,
		Role:     req.Role,
		Pin:      string(hashedPin),
		IsActive: true,
	}

	if err := h.staffRepo.Create(staff); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"data":    staff,
		"message": "staff created successfully",
	})
}

// Update updates a staff member
// PUT /api/v1/staff/:id
func (h *Handler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid staff ID",
		})
	}

	type Request struct {
		Name     string `json:"name"`
		Phone    string `json:"phone"`
		Role     string `json:"role"`
		IsActive *bool  `json:"is_active"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	staff, err := h.staffRepo.GetByID(uint(id))
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "staff not found",
		})
	}

	// Update fields
	if req.Name != "" {
		staff.Name = req.Name
	}
	if req.Phone != "" {
		// Check if phone is taken by another staff
		existing, _ := h.staffRepo.GetByPhone(staff.ShopID, req.Phone)
		if existing != nil && existing.ID != staff.ID {
			return c.Status(http.StatusConflict).JSON(fiber.Map{
				"error": "phone number already in use",
			})
		}
		staff.Phone = req.Phone
	}
	if req.Role != "" {
		staff.Role = req.Role
	}
	if req.IsActive != nil {
		staff.IsActive = *req.IsActive
	}

	if err := h.staffRepo.Update(staff); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":    staff,
		"message": "staff updated successfully",
	})
}

// Delete soft deletes a staff member
// DELETE /api/v1/staff/:id
func (h *Handler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid staff ID",
		})
	}

	staff, err := h.staffRepo.GetByID(uint(id))
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "staff not found",
		})
	}

	if err := h.staffRepo.Delete(staff.ID); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "staff deleted successfully",
	})
}

// UpdatePin updates staff PIN
// PUT /api/v1/staff/:id/pin
func (h *Handler) UpdatePin(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid staff ID",
		})
	}

	type Request struct {
		CurrentPin string `json:"current_pin"`
		NewPin     string `json:"new_pin"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.NewPin == "" || len(req.NewPin) < 4 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "new pin must be at least 4 digits",
		})
	}

	staff, err := h.staffRepo.GetByID(uint(id))
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "staff not found",
		})
	}

	// Verify current PIN
	if err := bcrypt.CompareHashAndPassword([]byte(staff.Pin), []byte(req.CurrentPin)); err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "current PIN is incorrect",
		})
	}

	// Hash new PIN
	hashedPin, err := bcrypt.GenerateFromPassword([]byte(req.NewPin), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to hash new PIN",
		})
	}

	staff.Pin = string(hashedPin)
	if err := h.staffRepo.Update(staff); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "pin updated successfully",
	})
}
