package handlers

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"github.com/C9b3rD3vi1/DukaPOS/internal/repository"
	"github.com/gofiber/fiber/v2"
)

// CustomerHandler handles customer-related HTTP requests
type CustomerHandler struct {
	customerRepo *repository.CustomerRepository
	shopRepo     *repository.ShopRepository
}

// NewCustomerHandler creates a new customer handler
func NewCustomerHandler(
	customerRepo *repository.CustomerRepository,
	shopRepo *repository.ShopRepository,
) *CustomerHandler {
	return &CustomerHandler{
		customerRepo: customerRepo,
		shopRepo:     shopRepo,
	}
}

// List returns all customers for a shop
// GET /api/v1/customers
func (h *CustomerHandler) List(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)

	customers, err := h.customerRepo.GetByShopID(shopID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":  customers,
		"total": len(customers),
	})
}

// Get returns a single customer
// GET /api/v1/customers/:id
func (h *CustomerHandler) Get(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid customer ID",
		})
	}

	customer, err := h.customerRepo.GetByID(uint(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Customer not found",
		})
	}

	return c.JSON(customer)
}

// Create creates a new customer
// POST /api/v1/customers
func (h *CustomerHandler) Create(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)

	type Request struct {
		Name    string `json:"name"`
		Phone   string `json:"phone"`
		Email   string `json:"email"`
		Address string `json:"address"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name == "" || req.Phone == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Name and phone are required",
		})
	}

	// Generate unique referral code
	referralCode := generateReferralCode(req.Name, req.Phone)

	customer := &models.Customer{
		ShopID:       shopID,
		Name:         req.Name,
		Phone:        req.Phone,
		Email:        req.Email,
		Address:      req.Address,
		Tier:         "bronze",
		IsActive:     true,
		ReferralCode: referralCode,
	}

	if err := h.customerRepo.Create(customer); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Customer created",
		"data":    customer,
	})
}

// generateReferralCode creates a unique referral code from name and phone
func generateReferralCode(name, phone string) string {
	// Use first 3 letters of name + last 4 digits of phone
	namePart := "CUS"
	if len(name) >= 3 {
		namePart = strings.ToUpper(name[:3])
	} else if len(name) > 0 {
		namePart = strings.ToUpper(name)
	}

	phonePart := "0000"
	phoneDigits := strings.ReplaceAll(phone, " ", "")
	phoneDigits = strings.ReplaceAll(phoneDigits, "-", "")
	if len(phoneDigits) >= 4 {
		phonePart = phoneDigits[len(phoneDigits)-4:]
	}

	// Add random suffix to ensure uniqueness
	rand.Seed(time.Now().UnixNano())
	suffix := rand.Intn(9000) + 1000

	return fmt.Sprintf("%s%s%d", namePart, phonePart, suffix)
}

// Update updates a customer
// PUT /api/v1/customers/:id
func (h *CustomerHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid customer ID",
		})
	}

	customer, err := h.customerRepo.GetByID(uint(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Customer not found",
		})
	}

	type Request struct {
		Name  string `json:"name"`
		Phone string `json:"phone"`
		Email string `json:"email"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name != "" {
		customer.Name = req.Name
	}
	if req.Phone != "" {
		customer.Phone = req.Phone
	}
	if req.Email != "" {
		customer.Email = req.Email
	}

	if err := h.customerRepo.Update(customer); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Customer updated",
		"data":    customer,
	})
}

// Delete deletes a customer
// DELETE /api/v1/customers/:id
func (h *CustomerHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid customer ID",
		})
	}

	if err := h.customerRepo.Delete(uint(id)); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Customer deleted",
	})
}
