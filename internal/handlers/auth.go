package handlers

import (
	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"github.com/C9b3rD3vi1/DukaPOS/internal/services"
	"github.com/gofiber/fiber/v2"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	OwnerName string `json:"owner_name"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register handles shop registration
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if req.Phone == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Phone number is required",
		})
	}
	if req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Password is required",
		})
	}
	if len(req.Password) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Password must be at least 6 characters",
		})
	}

	shop := &models.Shop{
		Name:      req.Name,
		Phone:     req.Phone,
		Email:     req.Email,
		OwnerName: req.OwnerName,
	}

	if shop.Name == "" {
		shop.Name = "My Shop"
	}
	if shop.OwnerName == "" {
		shop.OwnerName = "Shop Owner"
	}

	err := h.authService.Register(shop, req.Password)
	if err != nil {
		if err == services.ErrShopExists {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Shop already exists with this phone or email",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create shop",
		})
	}

	// Generate token
	_, token, err := h.authService.Login(req.Phone, req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"shop":  shop,
		"token": token,
	})
}

// Login handles shop login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if req.Phone == "" && req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Phone or email is required",
		})
	}
	if req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Password is required",
		})
	}

	// Determine login identifier
	identifier := req.Phone
	if identifier == "" {
		identifier = req.Email
	}

	shop, token, err := h.authService.Login(identifier, req.Password)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid phone/email or password",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Login failed",
		})
	}

	// Get account info for admin check
	account, _ := h.authService.GetAccountByID(shop.AccountID)

	return c.JSON(fiber.Map{
		"shop":    shop,
		"token":   token,
		"account": account,
	})
}

// Logout handles shop logout
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// In a stateless JWT system, logout is handled client-side
	// For stateful systems, you would invalidate the token here
	return c.JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}

// ChangePassword handles password change
func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)

	type ChangePasswordRequest struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	var req ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Old and new passwords are required",
		})
	}

	if len(req.NewPassword) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "New password must be at least 6 characters",
		})
	}

	err := h.authService.ChangePassword(shopID, req.OldPassword, req.NewPassword)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Current password is incorrect",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to change password",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Password changed successfully",
	})
}
