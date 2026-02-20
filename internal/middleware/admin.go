package middleware

import (
	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"github.com/gofiber/fiber/v2"
)

func RequireAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		account, ok := c.Locals("account").(*models.Account)
		if !ok || account == nil {
			return c.Status(401).JSON(fiber.Map{
				"error": "Unauthorized",
				"code":  "UNAUTHORIZED",
			})
		}

		if !account.IsAdmin {
			return c.Status(403).JSON(fiber.Map{
				"error": "Admin access required",
				"code":  "FORBIDDEN",
			})
		}

		return c.Next()
	}
}
