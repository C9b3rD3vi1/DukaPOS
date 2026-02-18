package ai

import (
	"strconv"

	aiservice "github.com/C9b3rD3vi1/DukaPOS/internal/services/ai"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	predictionService *aiservice.PredictionService
}

func New(predictionService *aiservice.PredictionService) *Handler {
	return &Handler{
		predictionService: predictionService,
	}
}

func (h *Handler) GetPredictions(c *fiber.Ctx) error {
	shopID, err := strconv.ParseUint(c.Params("shop_id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid shop ID",
		})
	}

	predictions, err := h.predictionService.GetPredictions(uint(shopID))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to generate predictions",
		})
	}

	return c.JSON(predictions)
}

func (h *Handler) GetRestockRecommendations(c *fiber.Ctx) error {
	shopID, err := strconv.ParseUint(c.Params("shop_id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid shop ID",
		})
	}

	recommendations, err := h.predictionService.GetRestockRecommendations(uint(shopID))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get recommendations",
		})
	}

	return c.JSON(fiber.Map{
		"data":    recommendations,
		"count":   len(recommendations),
		"shop_id": shopID,
	})
}

func (h *Handler) GetSalesAnalytics(c *fiber.Ctx) error {
	shopID, err := strconv.ParseUint(c.Params("shop_id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid shop ID",
		})
	}

	days, _ := strconv.Atoi(c.Query("days", "30"))
	if days < 1 {
		days = 30
	}
	if days > 365 {
		days = 365
	}

	analytics, err := h.predictionService.GetSalesAnalytics(uint(shopID), days)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get analytics",
		})
	}

	return c.JSON(analytics)
}

func (h *Handler) GetInventoryValue(c *fiber.Ctx) error {
	shopID, err := strconv.ParseUint(c.Params("shop_id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid shop ID",
		})
	}

	value, err := h.predictionService.GetInventoryValue(uint(shopID))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get inventory value",
		})
	}

	return c.JSON(fiber.Map{
		"shop_id": shopID,
		"value":   value,
	})
}

func (h *Handler) GenerateForecast(c *fiber.Ctx) error {
	shopID, err := strconv.ParseUint(c.Params("shop_id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid shop ID",
		})
	}

	var req struct {
		ProductID   uint `json:"product_id"`
		Days        int  `json:"days"`
		TargetStock int  `json:"target_stock"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.ProductID == 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "Product ID is required",
		})
	}

	forecast, err := h.predictionService.GenerateForecast(
		c.Context(),
		uint(shopID),
		&aiservice.ForecastRequest{
			ProductID:   req.ProductID,
			Days:        req.Days,
			TargetStock: req.TargetStock,
		},
	)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to generate forecast",
			"details": err.Error(),
		})
	}

	return c.JSON(forecast)
}

func (h *Handler) GetTrends(c *fiber.Ctx) error {
	shopID, err := strconv.ParseUint(c.Params("shop_id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid shop ID",
		})
	}

	predictions, err := h.predictionService.GetPredictions(uint(shopID))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get trends",
		})
	}

	trendingUp := []string{}
	trendingDown := []string{}
	stable := []string{}

	for _, p := range predictions.Predictions {
		if p.Trend == "up" {
			trendingUp = append(trendingUp, p.ProductName)
		} else if p.Trend == "down" {
			trendingDown = append(trendingDown, p.ProductName)
		} else {
			stable = append(stable, p.ProductName)
		}
	}

	return c.JSON(fiber.Map{
		"shop_id":       shopID,
		"trending_up":   trendingUp,
		"trending_down": trendingDown,
		"stable":        stable,
	})
}
