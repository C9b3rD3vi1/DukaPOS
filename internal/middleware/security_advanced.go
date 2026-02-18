package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

		c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		c.Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data:;")

		return c.Next()
	}
}

func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Locals("request_id", requestID)
		c.Set("X-Request-ID", requestID)

		return c.Next()
	}
}

func generateRequestID() string {
	return fmt.Sprintf("%d-%s", time.Now().UnixNano(), randomString(8))
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

func TimeoutMiddleware(timeout time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		done := make(chan error, 1)

		go func() {
			done <- c.Next()
		}()

		select {
		case err := <-done:
			return err
		case <-time.After(timeout):
			return c.Status(503).JSON(fiber.Map{
				"error":      "Request timeout",
				"code":       "TIMEOUT",
				"request_id": c.Locals("request_id"),
			})
		}
	}
}

func IPBlocklist(blockedIPs []string) fiber.Handler {
	blocklist := make(map[string]bool)
	for _, ip := range blockedIPs {
		blocklist[ip] = true
	}

	return func(c *fiber.Ctx) error {
		ip := c.IP()
		if blocklist[ip] {
			return c.Status(403).JSON(fiber.Map{
				"error": "Access denied",
				"code":  "BLOCKED_IP",
			})
		}
		return c.Next()
	}
}

func TrustedProxies(proxies []string) fiber.Handler {
	trusted := make(map[string]bool)
	for _, proxy := range proxies {
		trusted[proxy] = true
	}

	return func(c *fiber.Ctx) error {
		ip := c.IP()

		forwarded := c.Get("X-Forwarded-For")
		if forwarded != "" && len(proxies) > 0 {
			ips := strings.Split(forwarded, ",")
			if len(ips) > 0 {
				clientIP := strings.TrimSpace(ips[0])
				proxyIP := strings.TrimSpace(ips[len(ips)-1])

				if trusted[proxyIP] {
					ip = clientIP
				}
			}
		}

		c.Locals("client_ip", ip)
		return c.Next()
	}
}

func AccountLockout(maxAttempts int, lockoutDuration time.Duration) fiber.Handler {
	attempts := make(map[string]int)
	lockedUntil := make(map[string]time.Time)

	return func(c *fiber.Ctx) error {
		identifier := c.IP() + c.Path()

		if locked, ok := lockedUntil[identifier]; ok {
			if time.Now().Before(locked) {
				remaining := time.Until(locked)
				return c.Status(423).JSON(fiber.Map{
					"error":       "Account temporarily locked",
					"code":        "LOCKED",
					"retry_after": remaining.Seconds(),
				})
			}
			delete(lockedUntil, identifier)
			attempts[identifier] = 0
		}

		err := c.Next()

		if err != nil {
			attempts[identifier]++
			if attempts[identifier] >= maxAttempts {
				lockedUntil[identifier] = time.Now().Add(lockoutDuration)
				return c.Status(423).JSON(fiber.Map{
					"error":      "Too many failed attempts",
					"code":       "LOCKED",
					"lock_until": lockedUntil[identifier].Format(time.RFC3339),
				})
			}
		} else {
			delete(attempts, identifier)
		}

		return err
	}
}
