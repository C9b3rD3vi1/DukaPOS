package handlers

import (
	"github.com/C9b3rD3vi1/DukaPOS/internal/services/twofactor"
	"github.com/gofiber/fiber/v2"
)

type TwoFactorHandler struct {
	twoFactorSvc *twofactor.Service
}

func NewTwoFactorHandler(twoFactorSvc *twofactor.Service) *TwoFactorHandler {
	return &TwoFactorHandler{
		twoFactorSvc: twoFactorSvc,
	}
}

func (h *TwoFactorHandler) RegisterRoutes(app fiber.Router) {
	twofa := app.Group("/twofactor")
	twofa.Post("/setup", h.SetupTwoFactor)
	twofa.Post("/verify", h.VerifyTwoFactor)
	twofa.Post("/disable", h.DisableTwoFactor)
	twofa.Post("/backup-codes/generate", h.GenerateBackupCodes)
	twofa.Post("/backup-codes/verify", h.VerifyBackupCode)
}

type SetupTwoFactorRequest struct {
	AccountID   uint   `json:"account_id"`
	AccountName string `json:"account_name"`
	Phone       string `json:"phone"`
}

func (h *TwoFactorHandler) SetupTwoFactor(c *fiber.Ctx) error {
	var req SetupTwoFactorRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	secret, err := h.twoFactorSvc.EnableTwoFactor(req.AccountID, req.AccountName, req.Phone)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to setup two-factor authentication",
		})
	}

	return c.JSON(fiber.Map{
		"message":      "Two-factor authentication setup initiated",
		"secret":       secret.Secret,
		"qr_code_url":  secret.QRCodeURL,
		"account_name": secret.AccountName,
		"issuer":       secret.Issuer,
	})
}

type VerifyTwoFactorRequest struct {
	Token  string `json:"token"`
	Secret string `json:"secret"`
}

func (h *TwoFactorHandler) VerifyTwoFactor(c *fiber.Ctx) error {
	var req VerifyTwoFactorRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Token == "" || req.Secret == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Token and secret are required",
		})
	}

	isValid := h.twoFactorSvc.VerifyToken(req.Secret, req.Token)
	if !isValid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "Invalid two-factor authentication token",
			"valid":   false,
			"message": "The token you entered is invalid or expired",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Two-factor authentication verified successfully",
		"valid":   true,
	})
}

type DisableTwoFactorRequest struct {
	AccountID uint   `json:"account_id"`
	Token     string `json:"token"`
	Secret    string `json:"secret"`
}

func (h *TwoFactorHandler) DisableTwoFactor(c *fiber.Ctx) error {
	var req DisableTwoFactorRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	isValid := h.twoFactorSvc.VerifyToken(req.Secret, req.Token)
	if !isValid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token - cannot disable two-factor authentication",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Two-factor authentication disabled successfully",
	})
}

func (h *TwoFactorHandler) GenerateBackupCodes(c *fiber.Ctx) error {
	codes := twofactor.GenerateBackupCodes(10)

	return c.JSON(fiber.Map{
		"backup_codes": codes,
		"message":      "Backup codes generated. Store these securely.",
	})
}

type VerifyBackupCodeRequest struct {
	Code string   `json:"code"`
	Hash []string `json:"hash"`
}

func (h *TwoFactorHandler) VerifyBackupCode(c *fiber.Ctx) error {
	var req VerifyBackupCodeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	isValid := twofactor.VerifyBackupCode(req.Hash, req.Code)
	if !isValid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid backup code",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Backup code verified successfully",
	})
}
