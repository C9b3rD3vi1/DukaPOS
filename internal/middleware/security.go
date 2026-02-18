package middleware

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validation *validator.Validate

func init() {
	validation = validator.New()
}

type ValidationRules struct {
	Email    string `validate:"required,email"`
	Phone    string `validate:"required,phone"`
	Name     string `validate:"required,min=2,max=100"`
	Amount   string `validate:"required,min=1,max=150000"`
	Password string `validate:"required,min=6,max=72"`
	Pin      string `validate:"required,len=4,numeric"`
}

func ValidateStruct(s interface{}) error {
	return validation.Struct(s)
}

func ValidateField(field interface{}, tag string) error {
	return validation.Var(field, tag)
}

func InputValidation() fiber.Handler {
	return func(c *fiber.Ctx) error {
		contentType := c.Get("Content-Type")

		if strings.Contains(contentType, "application/json") {
			var body map[string]interface{}
			if err := c.BodyParser(&body); err != nil {
				return c.Status(400).JSON(fiber.Map{
					"error": "Invalid JSON body",
					"code":  "INVALID_JSON",
				})
			}

			for key, value := range body {
				if err := validateField(key, value); err != nil {
					return c.Status(400).JSON(fiber.Map{
						"error":   "Validation failed",
						"field":   key,
						"message": err.Error(),
						"code":    "VALIDATION_ERROR",
					})
				}
			}
		}

		return c.Next()
	}
}

func validateField(key string, value interface{}) error {
	if value == nil {
		return nil
	}

	strValue, ok := value.(string)
	if !ok {
		return nil
	}

	switch key {
	case "email":
		if !isValidEmail(strValue) {
			return fiber.NewError(400, "Invalid email format")
		}
	case "phone":
		if !isValidPhone(strValue) {
			return fiber.NewError(400, "Invalid phone format")
		}
	case "name", "product_name", "customer_name":
		if len(strValue) < 2 || len(strValue) > 100 {
			return fiber.NewError(400, "Name must be between 2 and 100 characters")
		}
		if containsSQLInjection(strValue) {
			return fiber.NewError(400, "Invalid characters in name")
		}
	case "amount", "price", "total":
		if !isValidAmount(strValue) {
			return fiber.NewError(400, "Invalid amount")
		}
	case "password":
		if len(strValue) < 6 || len(strValue) > 72 {
			return fiber.NewError(400, "Password must be between 6 and 72 characters")
		}
	case "pin":
		if !isValidPin(strValue) {
			return fiber.NewError(400, "PIN must be exactly 4 digits")
		}
	}

	return nil
}

func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func isValidPhone(phone string) bool {
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.TrimPrefix(phone, "+")

	phoneRegex := regexp.MustCompile(`^(?:254|0)?[1-9]\d{8}$`)
	return phoneRegex.MatchString(phone)
}

func isValidAmount(amount string) bool {
	amountRegex := regexp.MustCompile(`^\d+(\.\d{1,2})?$`)
	if !amountRegex.MatchString(amount) {
		return false
	}

	var value float64
	_, err := fmt.Sscanf(amount, "%f", &value)
	if err != nil {
		return false
	}

	return value >= 0 && value <= 150000
}

func isValidPin(pin string) bool {
	pinRegex := regexp.MustCompile(`^\d{4}$`)
	return pinRegex.MatchString(pin)
}

func containsSQLInjection(value string) bool {
	sqlKeywords := []string{
		"'", "\"", ";", "--", "/*", "*/",
		"xp_", "sp_", "exec", "execute",
		"union", "select", "insert", "update", "delete",
		"drop", "create", "alter", "truncate",
		"script", "<script", "javascript:",
	}

	lowerValue := strings.ToLower(value)
	for _, keyword := range sqlKeywords {
		if strings.Contains(lowerValue, keyword) {
			return true
		}
	}

	return false
}

func SanitizeInput() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Method() == "POST" || c.Method() == "PUT" {
			body := string(c.Body())

			if containsSQLInjection(body) {
				return c.Status(400).JSON(fiber.Map{
					"error": "Invalid input detected",
					"code":  "INVALID_INPUT",
				})
			}
		}

		return c.Next()
	}
}

type RequestValidator struct {
	Rules map[string][]string
}

func NewValidator(rules map[string][]string) *RequestValidator {
	return &RequestValidator{
		Rules: rules,
	}
}

func (v *RequestValidator) Validate() fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()

		if rules, exists := v.Rules[path]; exists {
			for _, rule := range rules {
				switch rule {
				case "auth":
					if c.Locals("user_id") == nil && c.Locals("api_key") == nil {
						return c.Status(401).JSON(fiber.Map{
							"error": "Authentication required",
							"code":  "UNAUTHORIZED",
						})
					}
				case "shop_owner":
					shopID := c.Locals("shop_id")
					if shopID == nil {
						return c.Status(403).JSON(fiber.Map{
							"error": "Shop access required",
							"code":  "FORBIDDEN",
						})
					}
				case "staff":
					if c.Locals("staff_id") == nil {
						return c.Status(403).JSON(fiber.Map{
							"error": "Staff access required",
							"code":  "FORBIDDEN",
						})
					}
				}
			}
		}

		return c.Next()
	}
}
