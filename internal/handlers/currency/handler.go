package currency

import (
	"github.com/C9b3rD3vi1/DukaPOS/internal/config"
	currencyservice "github.com/C9b3rD3vi1/DukaPOS/internal/services/currency"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Handler struct {
	service *currencyservice.Service
}

func NewHandler(db *gorm.DB, cfg *config.Config) *Handler {
	return &Handler{
		service: currencyservice.NewService(db, cfg),
	}
}

func (h *Handler) RegisterRoutes(app fiber.Router) {
	currency := app.Group("/currency")
	currency.Get("/list", h.ListCurrencies)
	currency.Get("/:code", h.GetCurrency)
	currency.Post("/convert", h.Convert)
	currency.Post("/format", h.Format)
	currency.Put("/:code/default", h.SetDefault)
}

func (h *Handler) ListCurrencies(c *fiber.Ctx) error {
	currencies, err := h.service.ListCurrencies()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"currencies": currencies,
		"total":      len(currencies),
	})
}

func (h *Handler) GetCurrency(c *fiber.Ctx) error {
	code := c.Params("code")

	currency, err := h.service.GetCurrency(code)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "currency not found"})
	}

	return c.JSON(currency)
}

func (h *Handler) Convert(c *fiber.Ctx) error {
	type ConvertRequest struct {
		Amount float64 `json:"amount"`
		From   string  `json:"from"`
		To     string  `json:"to"`
	}

	var req ConvertRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.Amount <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "amount must be greater than 0"})
	}

	if req.From == "" || req.To == "" {
		return c.Status(400).JSON(fiber.Map{"error": "from and to currencies are required"})
	}

	result, err := h.service.Convert(req.Amount, req.From, req.To)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"amount":    req.Amount,
		"from":      req.From,
		"to":        req.To,
		"result":    result,
		"formatted": h.service.Format(result, req.To),
	})
}

func (h *Handler) Format(c *fiber.Ctx) error {
	type FormatRequest struct {
		Amount   float64 `json:"amount"`
		Currency string  `json:"currency"`
	}

	var req FormatRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	formatted := h.service.Format(req.Amount, req.Currency)

	return c.JSON(fiber.Map{
		"amount":    req.Amount,
		"currency":  req.Currency,
		"formatted": formatted,
	})
}

func (h *Handler) SetDefault(c *fiber.Ctx) error {
	code := c.Params("code")

	if err := h.service.SetDefault(code); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "default currency updated",
		"code":    code,
	})
}
