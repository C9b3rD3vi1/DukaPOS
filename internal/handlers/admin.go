package handlers

import (
	"time"

	"github.com/C9b3rD3vi1/DukaPOS/internal/database"
	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"github.com/gofiber/fiber/v2"
)

type AdminHandler struct{}

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{}
}

func (h *AdminHandler) requireAdmin(c *fiber.Ctx) error {
	account, ok := c.Locals("account").(*models.Account)
	if !ok || account == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized - Please login"})
	}
	if !account.IsAdmin {
		return c.Status(403).JSON(fiber.Map{"error": "Forbidden - Admin access required"})
	}
	return nil
}

func (h *AdminHandler) Dashboard(c *fiber.Ctx) error {
	if err := h.requireAdmin(c); err != nil {
		return err
	}

	db := database.GetDB()

	var stats struct {
		TotalAccounts    int64
		TotalShops       int64
		TotalProducts    int64
		TotalSales       int64
		TotalRevenue     float64
		ActiveAccounts   int64
		ProAccounts      int64
		BusinessAccounts int64
		TodaySales       int64
		TodayRevenue     float64
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	db.Model(&models.Account{}).Count(&stats.TotalAccounts)
	db.Model(&models.Account{}).Where("is_active = ?", true).Count(&stats.ActiveAccounts)
	db.Model(&models.Account{}).Where("plan = ?", models.PlanPro).Count(&stats.ProAccounts)
	db.Model(&models.Account{}).Where("plan = ?", models.PlanBusiness).Count(&stats.BusinessAccounts)
	db.Model(&models.Shop{}).Count(&stats.TotalShops)
	db.Model(&models.Product{}).Count(&stats.TotalProducts)
	db.Model(&models.Sale{}).Count(&stats.TotalSales)
	db.Model(&models.Sale{}).Select("COALESCE(SUM(total_amount), 0)").Scan(&stats.TotalRevenue)
	db.Model(&models.Sale{}).Where("created_at >= ?", today).Count(&stats.TodaySales)
	db.Model(&models.Sale{}).Where("created_at >= ?", today).Select("COALESCE(SUM(total_amount), 0)").Scan(&stats.TodayRevenue)

	return c.JSON(stats)
}

func (h *AdminHandler) GetAccounts(c *fiber.Ctx) error {
	if err := h.requireAdmin(c); err != nil {
		return err
	}

	db := database.GetDB()

	var accounts []models.Account
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	search := c.Query("search", "")
	plan := c.Query("plan", "")

	query := db.Model(&models.Account{})

	if search != "" {
		query = query.Where("name LIKE ? OR email LIKE ? OR phone LIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	if plan != "" {
		query = query.Where("plan = ?", plan)
	}

	var total int64
	query.Count(&total)

	offset := (page - 1) * limit
	query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&accounts)

	return c.JSON(fiber.Map{
		"accounts": accounts,
		"total":    total,
		"page":     page,
		"limit":    limit,
		"pages":    (total + int64(limit) - 1) / int64(limit),
	})
}

func (h *AdminHandler) GetAccount(c *fiber.Ctx) error {
	if err := h.requireAdmin(c); err != nil {
		return err
	}

	db := database.GetDB()
	id := c.Params("id")

	var account models.Account
	if err := db.Preload("Shops").First(&account, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Account not found"})
	}

	var shopCount int64
	var productCount int64
	var saleCount int64
	var totalRevenue float64

	db.Model(&models.Shop{}).Where("account_id = ?", id).Count(&shopCount)
	db.Model(&models.Product{}).Where("shop_id IN (SELECT id FROM shops WHERE account_id = ?)", id).Count(&productCount)
	db.Model(&models.Sale{}).Where("shop_id IN (SELECT id FROM shops WHERE account_id = ?)", id).Count(&saleCount)
	db.Model(&models.Sale{}).Where("shop_id IN (SELECT id FROM shops WHERE account_id = ?)", id).Select("COALESCE(SUM(total_amount), 0)").Scan(&totalRevenue)

	return c.JSON(fiber.Map{
		"account":       account,
		"shop_count":    shopCount,
		"product_count": productCount,
		"sale_count":    saleCount,
		"total_revenue": totalRevenue,
	})
}

func (h *AdminHandler) UpdateAccountPlan(c *fiber.Ctx) error {
	if err := h.requireAdmin(c); err != nil {
		return err
	}

	db := database.GetDB()
	id := c.Params("id")

	type PlanUpdate struct {
		Plan string `json:"plan"`
	}

	var input PlanUpdate
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if input.Plan != string(models.PlanFree) && input.Plan != string(models.PlanPro) && input.Plan != string(models.PlanBusiness) {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid plan"})
	}

	var account models.Account
	if err := db.First(&account, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Account not found"})
	}

	account.Plan = models.PlanType(input.Plan)
	db.Save(&account)

	db.Model(&models.Shop{}).Where("account_id = ?", id).Update("plan", input.Plan)

	return c.JSON(fiber.Map{"message": "Plan updated successfully", "plan": input.Plan})
}

func (h *AdminHandler) UpdateAccountStatus(c *fiber.Ctx) error {
	if err := h.requireAdmin(c); err != nil {
		return err
	}

	db := database.GetDB()
	id := c.Params("id")

	type StatusUpdate struct {
		IsActive bool `json:"is_active"`
	}

	var input StatusUpdate
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	var account models.Account
	if err := db.First(&account, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Account not found"})
	}

	account.IsActive = input.IsActive
	db.Save(&account)

	db.Model(&models.Shop{}).Where("account_id = ?", id).Update("is_active", input.IsActive)

	return c.JSON(fiber.Map{"message": "Status updated successfully", "is_active": input.IsActive})
}

func (h *AdminHandler) GetShops(c *fiber.Ctx) error {
	if err := h.requireAdmin(c); err != nil {
		return err
	}

	db := database.GetDB()

	var shops []models.Shop
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	search := c.Query("search", "")

	query := db.Model(&models.Shop{}).Preload("Account")

	if search != "" {
		query = query.Where("name LIKE ? OR phone LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	query.Count(&total)

	offset := (page - 1) * limit
	query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&shops)

	return c.JSON(fiber.Map{
		"shops": shops,
		"total": total,
		"page":  page,
		"limit": limit,
		"pages": (total + int64(limit) - 1) / int64(limit),
	})
}

func (h *AdminHandler) GetSystemStats(c *fiber.Ctx) error {
	if err := h.requireAdmin(c); err != nil {
		return err
	}

	type SystemStats struct {
		DBSize       string `json:"db_size"`
		Uptime       string `json:"uptime"`
		GoVersion    string `json:"go_version"`
		FiberVersion string `json:"fiber_version"`
	}

	stats := SystemStats{
		GoVersion:    "1.21",
		FiberVersion: "2.52",
	}

	return c.JSON(stats)
}

func (h *AdminHandler) GetRevenueStats(c *fiber.Ctx) error {
	if err := h.requireAdmin(c); err != nil {
		return err
	}

	db := database.GetDB()

	days := c.QueryInt("days", 30)
	startDate := time.Now().AddDate(0, 0, -days)

	type DailyRevenue struct {
		Date    string  `json:"date"`
		Revenue float64 `json:"revenue"`
		Sales   int64   `json:"sales"`
	}

	var results []DailyRevenue

	db.Raw(`
		SELECT 
			DATE(created_at) as date,
			COALESCE(SUM(total_amount), 0) as revenue,
			COUNT(*) as sales
		FROM sales
		WHERE created_at >= ?
		GROUP BY DATE(created_at)
		ORDER BY date ASC
	`, startDate).Scan(&results)

	if results == nil {
		results = []DailyRevenue{}
	}

	return c.JSON(results)
}

func (h *AdminHandler) UpgradeAllAccounts(c *fiber.Ctx) error {
	if err := h.requireAdmin(c); err != nil {
		return err
	}

	db := database.GetDB()

	type UpgradeRequest struct {
		Plan string `json:"plan"`
	}

	var input UpgradeRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if input.Plan != string(models.PlanFree) && input.Plan != string(models.PlanPro) && input.Plan != string(models.PlanBusiness) {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid plan"})
	}

	db.Model(&models.Account{}).Update("plan", input.Plan)
	db.Model(&models.Shop{}).Update("plan", input.Plan)

	return c.JSON(fiber.Map{"message": "All accounts upgraded to " + input.Plan})
}
