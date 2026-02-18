package middleware

import (
	"strings"

	"github.com/C9b3rD3vi1/DukaPOS/internal/services"
	"github.com/gofiber/fiber/v2"
)

// JWT returns a JWT authentication middleware
func JWT(authService *services.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header required",
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		tokenString := parts[1]
		shop, err := authService.ValidateToken(tokenString)
		if err != nil {
			if err == services.ErrTokenExpired {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Token has expired",
				})
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		c.Locals("shop_id", shop.ID)
		c.Locals("shop", shop)

		return c.Next()
	}
}

// OptionalJWT returns an optional JWT authentication middleware
func OptionalJWT(authService *services.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Next()
		}

		tokenString := parts[1]
		shop, err := authService.ValidateToken(tokenString)
		if err != nil {
			return c.Next()
		}

		c.Locals("shop_id", shop.ID)
		c.Locals("shop", shop)

		return c.Next()
	}
}

// CORSMiddleware returns a CORS middleware
func CORSMiddleware(allowedOrigins []string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		origin := c.Get("Origin")

		allowed := false
		for _, o := range allowedOrigins {
			if o == "*" || o == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Set("Access-Control-Allow-Origin", origin)
		}

		if c.Method() == "OPTIONS" {
			c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			c.Set("Access-Control-Max-Age", "86400")
			return c.SendStatus(fiber.StatusOK)
		}

		return c.Next()
	}
}

// RequestLogger returns a request logging middleware
func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Next()
	}
}

// Recover returns a panic recovery middleware
func Recover() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Internal server error",
				})
			}
		}()
		return c.Next()
	}
}

// GetShopID extracts shop ID from context
func GetShopID(c *fiber.Ctx) uint {
	if shopID, ok := c.Locals("shop_id").(uint); ok {
		return shopID
	}
	return 0
}
