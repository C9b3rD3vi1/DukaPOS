package billing

import (
	"github.com/C9b3rD3vi1/DukaPOS/internal/config"
	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Handler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewHandler(db *gorm.DB, cfg *config.Config) *Handler {
	return &Handler{
		db:  db,
		cfg: cfg,
	}
}

func (h *Handler) RegisterRoutes(app fiber.Router) {
	billing := app.Group("/billing")
	billing.Get("/plans", h.GetPlans)
	billing.Get("/current", h.GetCurrentPlan)
	billing.Post("/upgrade", h.UpgradePlan)
}

type Plan struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Price        float64  `json:"price"`
	Interval     string   `json:"interval"`
	Features     []string `json:"features"`
	ProductLimit int      `json:"product_limit"`
	ShopLimit    int      `json:"shop_limit"`
	StaffLimit   int      `json:"staff_limit"`
	IsPopular    bool     `json:"is_popular"`
}

var plans = []Plan{
	{
		ID:           "free",
		Name:         "Free",
		Price:        0,
		Interval:     "forever",
		Features:     []string{"WhatsApp bot", "1 Shop", "50 Products", "Basic reports"},
		ProductLimit: 50,
		ShopLimit:    1,
		StaffLimit:   0,
	},
	{
		ID:           "pro",
		Name:         "Pro",
		Price:        500,
		Interval:     "month",
		Features:     []string{"Everything in Free", "Unlimited Products", "3 Shops", "2 Staff", "M-Pesa integration", "QR Payments", "Loyalty program"},
		ProductLimit: -1,
		ShopLimit:    3,
		StaffLimit:   2,
		IsPopular:    true,
	},
	{
		ID:           "business",
		Name:         "Business",
		Price:        1500,
		Interval:     "month",
		Features:     []string{"Everything in Pro", "Unlimited Shops", "Unlimited Staff", "API Access", "Webhooks", "AI Predictions", "Priority Support"},
		ProductLimit: -1,
		ShopLimit:    -1,
		StaffLimit:   -1,
	},
}

func (h *Handler) GetPlans(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"plans": plans,
	})
}

func (h *Handler) GetCurrentPlan(c *fiber.Ctx) error {
	accountID := c.Locals("account_id")
	if accountID == nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	var account models.Account
	if err := h.db.First(&account, accountID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "account not found"})
	}

	var currentPlan Plan
	for _, p := range plans {
		if p.ID == string(account.Plan) {
			currentPlan = p
			break
		}
	}

	var shopCount int64
	h.db.Model(&models.Shop{}).Where("account_id = ?", account.ID).Count(&shopCount)

	var productCount int64
	h.db.Model(&models.Product{}).Where("account_id = ?", account.ID).Count(&productCount)

	return c.JSON(fiber.Map{
		"plan":      currentPlan,
		"is_active": account.IsActive,
		"created":   account.CreatedAt,
		"usage": map[string]interface{}{
			"shops":    shopCount,
			"products": productCount,
		},
		"limits": map[string]interface{}{
			"shops":    currentPlan.ShopLimit,
			"products": currentPlan.ProductLimit,
		},
	})
}

func (h *Handler) UpgradePlan(c *fiber.Ctx) error {
	type UpgradeRequest struct {
		PlanID string `json:"plan_id"`
	}

	var req UpgradeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	accountID := c.Locals("account_id")
	if accountID == nil {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}

	validPlan := false
	for _, p := range plans {
		if p.ID == req.PlanID {
			validPlan = true
			break
		}
	}

	if !validPlan {
		return c.Status(400).JSON(fiber.Map{"error": "invalid plan_id"})
	}

	var account models.Account
	if err := h.db.First(&account, accountID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "account not found"})
	}

	oldPlan := account.Plan
	account.Plan = models.PlanType(req.PlanID)

	if err := h.db.Save(&account).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to upgrade plan"})
	}

	return c.JSON(fiber.Map{
		"message":  "plan upgraded successfully",
		"new_plan": req.PlanID,
		"old_plan": oldPlan,
	})
}

func (h *Handler) GetHistory(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"data": []interface{}{},
	})
}
