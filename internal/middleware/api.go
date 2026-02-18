package middleware

import (
	"strings"

	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	apiservice "github.com/C9b3rD3vi1/DukaPOS/internal/services/api"
	"github.com/gofiber/fiber/v2"
)

func APIKeyAuth(apiService *apiservice.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		apiKey := c.Get("X-API-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}

		if apiKey == "" {
			return c.Status(401).JSON(fiber.Map{
				"error": "API key required",
				"hint":  "Use X-API-Key header or api_key query parameter",
			})
		}

		key, err := apiService.ValidateKey(apiKey)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"error": "Invalid API key",
				"code":  "INVALID_KEY",
			})
		}

		if !apiService.CheckRateLimit(key) {
			return c.Status(429).JSON(fiber.Map{
				"error":       "Rate limit exceeded",
				"code":        "RATE_LIMITED",
				"retry_after": 60,
			})
		}

		apiService.UpdateLastUsed(key.ID)

		c.Locals("api_key", key)
		c.Locals("shop_id", key.ShopID)

		return c.Next()
	}
}

func APIKeyPermissionCheck(apiService *apiservice.Service, permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := c.Locals("api_key")
		if key == nil {
			return c.Status(403).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		apiKey := key.(*models.APIKey)
		if !apiService.HasPermission(apiKey, permission) {
			return c.Status(403).JSON(fiber.Map{
				"error":      "Insufficient permissions",
				"code":       "FORBIDDEN",
				"permission": permission,
			})
		}

		return c.Next()
	}
}

func APIKeyRateLimitInfo(apiService *apiservice.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := c.Locals("api_key")
		if key != nil {
			apiKey := key.(*models.APIKey)
			status := apiService.GetRateLimitStatus(apiKey)
			c.Set("X-RateLimit-Limit", status["limit"].(string))
			c.Set("X-RateLimit-Remaining", status["remaining"].(string))
			c.Set("X-RateLimit-Reset", status["reset"].(string))
		}
		return c.Next()
	}
}

func RequirePermission(permissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := c.Locals("api_key")
		if key == nil {
			return c.Status(403).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		apiKey := key.(*models.APIKey)

		for _, perm := range permissions {
			if strings.Contains(apiKey.Permissions, perm) || apiKey.Permissions == "*" {
				return c.Next()
			}
		}

		return c.Status(403).JSON(fiber.Map{
			"error":    "Insufficient permissions",
			"required": permissions,
			"current":  apiKey.Permissions,
		})
	}
}
