package loyalty

import (
	"fmt"
	"time"

	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"github.com/C9b3rD3vi1/DukaPOS/internal/repository"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Handler struct {
	db           *gorm.DB
	customerRepo *repository.CustomerRepository
	saleRepo     *repository.SaleRepository
}

func NewHandler(customerRepo *repository.CustomerRepository, saleRepo *repository.SaleRepository, db *gorm.DB) *Handler {
	return &Handler{
		db:           db,
		customerRepo: customerRepo,
		saleRepo:     saleRepo,
	}
}

func (h *Handler) RegisterRoutes(app fiber.Router) {
	loyalty := app.Group("/loyalty")
	loyalty.Get("/points/:customer_id", h.GetCustomerPoints)
	loyalty.Get("/stats/:customer_id", h.GetCustomerStats)
	loyalty.Post("/redeem", h.RedeemPoints)
	loyalty.Post("/earn", h.EarnPoints)
	loyalty.Get("/transactions/:customer_id", h.ListTransactions)

	// Shop-level endpoints
	loyalty.Get("/stats/shop/:shop_id", h.GetShopLoyaltyStats)
	loyalty.Get("/members", h.ListLoyaltyMembers)
	loyalty.Post("/points/add", h.AddPoints)
	loyalty.Post("/points/redeem", h.RedeemPoints)
}

func (h *Handler) GetCustomerPoints(c *fiber.Ctx) error {
	customerID, err := c.ParamsInt("customer_id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid customer_id"})
	}

	customer, err := h.customerRepo.GetByID(uint(customerID))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "customer not found"})
	}

	validPoints := h.getValidPoints(uint(customerID))

	return c.JSON(fiber.Map{
		"customer_id":     customer.ID,
		"loyalty_points":  validPoints,
		"tier":            customer.Tier,
		"total_spent":     customer.TotalSpent,
		"total_purchases": customer.TotalPurchases,
	})
}

func (h *Handler) getValidPoints(customerID uint) int {
	var total int
	h.db.Model(&models.LoyaltyTransaction{}).
		Where("customer_id = ? AND points > 0 AND (expires_at IS NULL OR expires_at > ?)", customerID, time.Now()).
		Select("COALESCE(SUM(points), 0)").
		Scan(&total)

	var redeemed int
	h.db.Model(&models.LoyaltyTransaction{}).
		Where("customer_id = ? AND points < 0", customerID).
		Select("COALESCE(SUM(ABS(points)), 0)").
		Scan(&redeemed)

	return total - redeemed
}

func (h *Handler) GetCustomerStats(c *fiber.Ctx) error {
	customerID, err := c.ParamsInt("customer_id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid customer_id"})
	}

	customer, err := h.customerRepo.GetByID(uint(customerID))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "customer not found"})
	}

	validPoints := h.getValidPoints(uint(customerID))

	var transactions []models.LoyaltyTransaction
	h.db.Where("customer_id = ?", customerID).
		Order("created_at DESC").
		Limit(10).
		Find(&transactions)

	tierConfig := models.DefaultTierConfigs[customer.Tier]

	return c.JSON(fiber.Map{
		"customer": map[string]interface{}{
			"id":              customer.ID,
			"name":            customer.Name,
			"phone":           customer.Phone,
			"email":           customer.Email,
			"tier":            customer.Tier,
			"total_spent":     customer.TotalSpent,
			"total_purchases": customer.TotalPurchases,
		},
		"points": map[string]interface{}{
			"available":         validPoints,
			"lifetime_earned":   customer.PointsEarned,
			"lifetime_redeemed": customer.PointsRedeemed,
			"points_rate":       tierConfig.PointsRate,
		},
		"tier": map[string]interface{}{
			"current":   customer.Tier,
			"min_spend": tierConfig.MinSpend,
			"perks":     tierConfig.Perks,
		},
		"recent_transactions": transactions,
	})
}

func (h *Handler) RedeemPoints(c *fiber.Ctx) error {
	type RedeemRequest struct {
		CustomerID  uint   `json:"customer_id"`
		Points      int    `json:"points"`
		Description string `json:"description"`
	}

	var req RedeemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.Points <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "points must be greater than 0"})
	}

	customer, err := h.customerRepo.GetByID(req.CustomerID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "customer not found"})
	}

	validPoints := h.getValidPoints(req.CustomerID)
	if validPoints < req.Points {
		return c.Status(400).JSON(fiber.Map{"error": "insufficient points"})
	}

	pointsBefore := customer.LoyaltyPoints
	customer.LoyaltyPoints -= req.Points
	customer.PointsRedeemed += req.Points

	if err := h.customerRepo.Update(customer); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	discountAmount := float64(req.Points) * 0.5

	transaction := &models.LoyaltyTransaction{
		CustomerID:   req.CustomerID,
		ShopID:       customer.ShopID,
		Type:         models.LoyaltyRedeemed,
		Points:       -req.Points,
		PointsBefore: pointsBefore,
		PointsAfter:  customer.LoyaltyPoints,
		Amount:       discountAmount,
		Description:  req.Description,
	}

	h.db.Create(transaction)

	return c.JSON(fiber.Map{
		"message":         "points redeemed successfully",
		"points_used":     req.Points,
		"discount_amount": discountAmount,
	})
}

func (h *Handler) EarnPoints(c *fiber.Ctx) error {
	type EarnRequest struct {
		CustomerID  uint    `json:"customer_id"`
		ShopID      uint    `json:"shop_id"`
		Amount      float64 `json:"amount"`
		Description string  `json:"description"`
	}

	var req EarnRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.Amount <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "amount must be greater than 0"})
	}

	customer, err := h.customerRepo.GetByID(req.CustomerID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "customer not found"})
	}

	pointsRate := 1.0
	if customer.Tier == models.TierSilver {
		pointsRate = 1.5
	} else if customer.Tier == models.TierGold {
		pointsRate = 2.0
	} else if customer.Tier == models.TierPlatinum {
		pointsRate = 3.0
	}

	pointsEarned := int(req.Amount * pointsRate)
	if pointsEarned < 1 {
		pointsEarned = 1
	}

	pointsBefore := customer.LoyaltyPoints
	customer.LoyaltyPoints += pointsEarned
	customer.PointsEarned += pointsEarned
	customer.TotalSpent += req.Amount
	customer.TotalPurchases++

	if err := h.customerRepo.Update(customer); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	transaction := &models.LoyaltyTransaction{
		CustomerID:   req.CustomerID,
		ShopID:       req.ShopID,
		Type:         models.LoyaltyEarned,
		Points:       pointsEarned,
		PointsBefore: pointsBefore,
		PointsAfter:  customer.LoyaltyPoints,
		Amount:       req.Amount,
		Description:  req.Description,
	}

	h.db.Create(transaction)

	return c.JSON(fiber.Map{
		"message":       "points earned successfully",
		"points_earned": pointsEarned,
		"total_points":  customer.LoyaltyPoints,
	})
}

func (h *Handler) ListTransactions(c *fiber.Ctx) error {
	customerID, err := c.ParamsInt("customer_id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid customer_id"})
	}

	var transactions []models.LoyaltyTransaction
	h.db.Where("customer_id = ?", customerID).
		Order("created_at DESC").
		Limit(50).
		Find(&transactions)

	return c.JSON(fiber.Map{
		"transactions": transactions,
		"total":        len(transactions),
	})
}

func (h *Handler) GetShopLoyaltyStats(c *fiber.Ctx) error {
	shopID, err := c.ParamsInt("shop_id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid shop_id"})
	}

	var totalMembers int64
	h.db.Model(&models.Customer{}).Where("shop_id = ? AND loyalty_points > 0", shopID).Count(&totalMembers)

	var totalPoints int64
	h.db.Model(&models.Customer{}).Where("shop_id = ?", shopID).Select("COALESCE(SUM(loyalty_points), 0)").Scan(&totalPoints)

	var totalRedemptions int64
	h.db.Model(&models.LoyaltyTransaction{}).Where("shop_id = ? AND points < 0", shopID).Count(&totalRedemptions)

	var activeThisMonth int64
	startOfMonth := time.Now().Truncate(24*time.Hour).AddDate(0, 0, -30)
	h.db.Model(&models.LoyaltyTransaction{}).Where("shop_id = ? AND created_at > ?", shopID, startOfMonth).Count(&activeThisMonth)

	return c.JSON(fiber.Map{
		"total_members":     totalMembers,
		"total_points":      totalPoints,
		"total_redemptions": totalRedemptions,
		"active_this_month": activeThisMonth,
	})
}

func (h *Handler) ListLoyaltyMembers(c *fiber.Ctx) error {
	shopIDStr := c.Query("shop_id")
	if shopIDStr == "" {
		return c.Status(400).JSON(fiber.Map{"error": "shop_id is required"})
	}
	var shopID uint
	if _, err := fmt.Sscanf(shopIDStr, "%d", &shopID); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid shop_id"})
	}

	var members []models.Customer
	h.db.Where("shop_id = ? AND loyalty_points > 0", shopID).
		Order("loyalty_points DESC").
		Find(&members)

	type MemberResponse struct {
		ID         uint       `json:"id"`
		CustomerID uint       `json:"customer_id"`
		ShopID     uint       `json:"shop_id"`
		Points     int        `json:"points"`
		Tier       string     `json:"tier"`
		TotalSpent float64    `json:"total_spent"`
		Visits     int        `json:"visits"`
		LastVisit  *time.Time `json:"last_visit,omitempty"`
	}

	response := make([]MemberResponse, len(members))
	for i, m := range members {
		response[i] = MemberResponse{
			ID:         m.ID,
			CustomerID: m.ID,
			ShopID:     m.ShopID,
			Points:     m.LoyaltyPoints,
			Tier:       string(m.Tier),
			TotalSpent: m.TotalSpent,
			Visits:     m.TotalPurchases,
		}
	}

	return c.JSON(fiber.Map{
		"data": response,
	})
}

func (h *Handler) AddPoints(c *fiber.Ctx) error {
	type AddPointsRequest struct {
		CustomerID  uint   `json:"customer_id"`
		ShopID      uint   `json:"shop_id"`
		Points      int    `json:"points"`
		Description string `json:"description"`
	}

	var req AddPointsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.Points <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "points must be greater than 0"})
	}

	customer, err := h.customerRepo.GetByID(req.CustomerID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "customer not found"})
	}

	pointsBefore := customer.LoyaltyPoints
	customer.LoyaltyPoints += req.Points
	customer.PointsEarned += req.Points
	customer.TotalSpent += float64(req.Points)

	if err := h.customerRepo.Update(customer); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	transaction := &models.LoyaltyTransaction{
		CustomerID:   req.CustomerID,
		ShopID:       req.ShopID,
		Type:         models.LoyaltyEarned,
		Points:       req.Points,
		PointsBefore: pointsBefore,
		PointsAfter:  customer.LoyaltyPoints,
		Amount:       float64(req.Points),
		Description:  req.Description,
	}

	h.db.Create(transaction)

	return c.JSON(fiber.Map{
		"message":      "points added successfully",
		"points_added": req.Points,
		"total_points": customer.LoyaltyPoints,
	})
}
