package handlers

import (
	"fmt"
	"time"

	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"github.com/C9b3rD3vi1/DukaPOS/internal/repository"
	"github.com/gofiber/fiber/v2"
)

// ShopHandler handles shop-related HTTP requests
type ShopHandler struct {
	shopRepo    *repository.ShopRepository
	productRepo *repository.ProductRepository
	saleRepo    *repository.SaleRepository
	accountRepo *repository.AccountRepository
}

// NewShopHandler creates a new shop handler
func NewShopHandler(
	shopRepo *repository.ShopRepository,
	productRepo *repository.ProductRepository,
	saleRepo *repository.SaleRepository,
) *ShopHandler {
	return &ShopHandler{
		shopRepo:    shopRepo,
		productRepo: productRepo,
		saleRepo:    saleRepo,
	}
}

// NewShopHandlerWithAccount creates a new shop handler with account repository
func NewShopHandlerWithAccount(
	shopRepo *repository.ShopRepository,
	productRepo *repository.ProductRepository,
	saleRepo *repository.SaleRepository,
	accountRepo *repository.AccountRepository,
) *ShopHandler {
	return &ShopHandler{
		shopRepo:    shopRepo,
		productRepo: productRepo,
		saleRepo:    saleRepo,
		accountRepo: accountRepo,
	}
}

// GetProfile returns the shop's profile
func (h *ShopHandler) GetProfile(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)

	shop, err := h.shopRepo.GetByID(shopID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Shop not found",
		})
	}

	return c.JSON(shop)
}

// GetAccount returns the account with all shops
func (h *ShopHandler) GetAccount(c *fiber.Ctx) error {
	if h.accountRepo == nil {
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
			"error": "Account feature not available",
		})
	}

	shopID := c.Locals("shop_id").(uint)
	shop, err := h.shopRepo.GetByID(shopID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Shop not found",
		})
	}

	if shop.AccountID == 0 {
		return c.JSON(fiber.Map{
			"id":         shop.ID,
			"account_id": 0,
			"shops":      []models.Shop{*shop},
		})
	}

	account, err := h.accountRepo.GetByID(shop.AccountID)
	if err != nil {
		return c.JSON(fiber.Map{
			"id":         shop.ID,
			"account_id": shop.AccountID,
			"shops":      []models.Shop{*shop},
		})
	}

	shops, _ := h.accountRepo.GetShops(account.ID)
	if shops == nil {
		shops = []models.Shop{*shop}
	}

	return c.JSON(fiber.Map{
		"id":         account.ID,
		"email":      account.Email,
		"name":       account.Name,
		"phone":      account.Phone,
		"plan":       account.Plan,
		"account_id": account.ID,
		"shops":      shops,
	})
}

// UpdateProfile updates the shop's profile
func (h *ShopHandler) UpdateProfile(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)

	shop, err := h.shopRepo.GetByID(shopID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Shop not found",
		})
	}

	type UpdateRequest struct {
		Name      string `json:"name"`
		OwnerName string `json:"owner_name"`
		Address   string `json:"address"`
		Email     string `json:"email"`
	}

	var req UpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name != "" {
		shop.Name = req.Name
	}
	if req.OwnerName != "" {
		shop.OwnerName = req.OwnerName
	}
	if req.Address != "" {
		shop.Address = req.Address
	}
	if req.Email != "" {
		shop.Email = req.Email
	}

	if err := h.shopRepo.Update(shop); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update profile",
		})
	}

	return c.JSON(shop)
}

// GetDashboard returns dashboard statistics
func (h *ShopHandler) GetDashboard(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)

	// Get product count
	products, err := h.productRepo.GetByShopID(shopID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get products",
		})
	}

	// Get today's sales
	sales, err := h.saleRepo.GetTodaySales(shopID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get sales",
		})
	}

	// Calculate totals
	var totalSales, totalProfit float64
	for _, sale := range sales {
		totalSales += sale.TotalAmount
		totalProfit += sale.Profit
	}

	// Get low stock items
	lowStock, _ := h.productRepo.GetLowStock(shopID)

	// Get recent sales
	recentSales, _ := h.saleRepo.GetByShopID(shopID, 10)

	return c.JSON(fiber.Map{
		"total_products":  len(products),
		"total_sales":     totalSales,
		"total_profit":    totalProfit,
		"today_sales":     len(sales),
		"low_stock_count": len(lowStock),
		"low_stock":       lowStock,
		"recent_sales":    recentSales,
	})
}

// ProductHandler handles product-related HTTP requests
type ProductHandler struct {
	productRepo *repository.ProductRepository
}

// NewProductHandler creates a new product handler
func NewProductHandler(productRepo *repository.ProductRepository) *ProductHandler {
	return &ProductHandler{productRepo: productRepo}
}

// ListProducts returns all products for a shop
func (h *ProductHandler) ListProducts(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)

	products, err := h.productRepo.GetByShopID(shopID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get products",
		})
	}

	return c.JSON(products)
}

// GetProduct returns a single product
func (h *ProductHandler) GetProduct(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)
	productID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid product ID",
		})
	}

	product, err := h.productRepo.GetByID(uint(productID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Product not found",
		})
	}

	// Verify ownership
	if product.ShopID != shopID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	return c.JSON(product)
}

// CreateProduct creates a new product
func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)

	type CreateRequest struct {
		Name              string  `json:"name"`
		Category          string  `json:"category"`
		Unit              string  `json:"unit"`
		CostPrice         float64 `json:"cost_price"`
		SellingPrice      float64 `json:"selling_price"`
		CurrentStock      int     `json:"current_stock"`
		LowStockThreshold int     `json:"low_stock_threshold"`
		Barcode           string  `json:"barcode"`
	}

	var req CreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Product name is required",
		})
	}
	if req.SellingPrice <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Selling price must be greater than 0",
		})
	}

	product := &models.Product{
		ShopID:            shopID,
		Name:              req.Name,
		Category:          req.Category,
		Unit:              req.Unit,
		CostPrice:         req.CostPrice,
		SellingPrice:      req.SellingPrice,
		CurrentStock:      req.CurrentStock,
		LowStockThreshold: req.LowStockThreshold,
		Barcode:           req.Barcode,
		IsActive:          true,
	}

	if product.Unit == "" {
		product.Unit = "pcs"
	}
	if product.LowStockThreshold == 0 {
		product.LowStockThreshold = 10
	}

	if err := h.productRepo.Create(product); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create product",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(product)
}

// UpdateProduct updates a product
func (h *ProductHandler) UpdateProduct(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)
	productID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid product ID",
		})
	}

	product, err := h.productRepo.GetByID(uint(productID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Product not found",
		})
	}

	if product.ShopID != shopID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	type UpdateRequest struct {
		Name              string  `json:"name"`
		Category          string  `json:"category"`
		Unit              string  `json:"unit"`
		CostPrice         float64 `json:"cost_price"`
		SellingPrice      float64 `json:"selling_price"`
		CurrentStock      *int    `json:"current_stock"`
		LowStockThreshold int     `json:"low_stock_threshold"`
		Barcode           string  `json:"barcode"`
	}

	var req UpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name != "" {
		product.Name = req.Name
	}
	if req.Category != "" {
		product.Category = req.Category
	}
	if req.Unit != "" {
		product.Unit = req.Unit
	}
	if req.CostPrice > 0 {
		product.CostPrice = req.CostPrice
	}
	if req.SellingPrice > 0 {
		product.SellingPrice = req.SellingPrice
	}
	if req.CurrentStock != nil {
		product.CurrentStock = *req.CurrentStock
	}
	if req.LowStockThreshold > 0 {
		product.LowStockThreshold = req.LowStockThreshold
	}
	if req.Barcode != "" {
		product.Barcode = req.Barcode
	}

	if err := h.productRepo.Update(product); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update product",
		})
	}

	return c.JSON(product)
}

// DeleteProduct deletes a product
func (h *ProductHandler) DeleteProduct(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)
	productID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid product ID",
		})
	}

	product, err := h.productRepo.GetByID(uint(productID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Product not found",
		})
	}

	if product.ShopID != shopID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	if err := h.productRepo.Delete(product.ID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete product",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Product deleted successfully",
	})
}

// SaleHandler handles sale-related HTTP requests
type SaleHandler struct {
	saleRepo    *repository.SaleRepository
	productRepo *repository.ProductRepository
}

// NewSaleHandler creates a new sale handler
func NewSaleHandler(saleRepo *repository.SaleRepository, productRepo *repository.ProductRepository) *SaleHandler {
	return &SaleHandler{
		saleRepo:    saleRepo,
		productRepo: productRepo,
	}
}

// GetSale returns a single sale by ID
func (h *SaleHandler) GetSale(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)
	saleID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid sale ID",
		})
	}

	sale, err := h.saleRepo.GetByID(uint(saleID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Sale not found",
		})
	}

	if sale.ShopID != shopID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	return c.JSON(sale)
}

// ListSales returns all sales for a shop
func (h *SaleHandler) ListSales(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)
	limit := 50 // default

	sales, err := h.saleRepo.GetByShopID(shopID, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get sales",
		})
	}

	return c.JSON(sales)
}

// CreateSale creates a new sale
func (h *SaleHandler) CreateSale(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)

	type CreateRequest struct {
		ProductID     uint    `json:"product_id"`
		Quantity      int     `json:"quantity"`
		UnitPrice     float64 `json:"unit_price"`
		PaymentMethod string  `json:"payment_method"`
	}

	var req CreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate
	if req.ProductID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Product ID is required",
		})
	}
	if req.Quantity <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Quantity must be greater than 0",
		})
	}

	// Get product
	product, err := h.productRepo.GetByID(req.ProductID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Product not found",
		})
	}

	if product.ShopID != shopID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	// Check stock
	if product.CurrentStock < req.Quantity {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":     "Insufficient stock",
			"available": product.CurrentStock,
		})
	}

	// Calculate totals
	totalAmount := product.SellingPrice * float64(req.Quantity)
	costAmount := product.CostPrice * float64(req.Quantity)
	profit := totalAmount - costAmount

	// Use provided price if different from product price
	if req.UnitPrice > 0 {
		totalAmount = req.UnitPrice * float64(req.Quantity)
		costAmount = product.CostPrice * float64(req.Quantity)
		profit = totalAmount - costAmount
	}

	paymentMethod := models.PaymentCash
	if req.PaymentMethod == "mpesa" {
		paymentMethod = models.PaymentMpesa
	}

	sale := &models.Sale{
		ShopID:        shopID,
		ProductID:     product.ID,
		Quantity:      req.Quantity,
		UnitPrice:     product.SellingPrice,
		TotalAmount:   totalAmount,
		CostAmount:    costAmount,
		Profit:        profit,
		PaymentMethod: paymentMethod,
	}

	if err := h.saleRepo.Create(sale); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create sale",
		})
	}

	// Update stock
	h.productRepo.UpdateStock(product.ID, -req.Quantity)

	return c.Status(fiber.StatusCreated).JSON(sale)
}

// ReportHandler handles report-related HTTP requests
type ReportHandler struct {
	saleRepo    *repository.SaleRepository
	productRepo *repository.ProductRepository
	summaryRepo *repository.DailySummaryRepository
}

// NewReportHandler creates a new report handler
func NewReportHandler(
	saleRepo *repository.SaleRepository,
	productRepo *repository.ProductRepository,
	summaryRepo *repository.DailySummaryRepository,
) *ReportHandler {
	return &ReportHandler{
		saleRepo:    saleRepo,
		productRepo: productRepo,
		summaryRepo: summaryRepo,
	}
}

// GetDailyReport returns daily report
func (h *ReportHandler) GetDailyReport(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)

	sales, err := h.saleRepo.GetTodaySales(shopID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get sales",
		})
	}

	var totalSales, totalProfit, totalCost float64
	var transactionCount int
	productSales := make(map[string]float64)

	for _, sale := range sales {
		totalSales += sale.TotalAmount
		totalProfit += sale.Profit
		totalCost += sale.CostAmount
		transactionCount++
		productSales[sale.Product.Name] += sale.TotalAmount
	}

	// Find top product
	var topProduct string
	var topAmount float64
	for name, amount := range productSales {
		if amount > topAmount {
			topProduct = name
			topAmount = amount
		}
	}

	// Get low stock
	lowStock, _ := h.productRepo.GetLowStock(shopID)

	return c.JSON(fiber.Map{
		"type":            "daily",
		"date":            "today",
		"total_sales":     totalSales,
		"total_profit":    totalProfit,
		"total_cost":      totalCost,
		"transactions":    transactionCount,
		"top_product":     topProduct,
		"top_amount":      topAmount,
		"low_stock_count": len(lowStock),
		"low_stock":       lowStock,
	})
}

// GetWeeklyReport returns weekly report
func (h *ReportHandler) GetWeeklyReport(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)

	end := time.Now()
	start := end.AddDate(0, 0, -7)

	sales, err := h.saleRepo.GetByDateRange(shopID, start, end)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get sales",
		})
	}

	var totalSales, totalProfit, totalCost float64
	var transactionCount int

	for _, sale := range sales {
		totalSales += sale.TotalAmount
		totalProfit += sale.Profit
		totalCost += sale.CostAmount
		transactionCount++
	}

	dailyAvg := totalSales / 7

	return c.JSON(fiber.Map{
		"type":         "weekly",
		"start_date":   start.Format("2006-01-02"),
		"end_date":     end.Format("2006-01-02"),
		"total_sales":  totalSales,
		"total_profit": totalProfit,
		"total_cost":   totalCost,
		"transactions": transactionCount,
		"daily_avg":    dailyAvg,
	})
}

// GetMonthlyReport returns monthly report
func (h *ReportHandler) GetMonthlyReport(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)

	end := time.Now()
	start := end.AddDate(0, -1, 0)

	sales, err := h.saleRepo.GetByDateRange(shopID, start, end)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get sales",
		})
	}

	var totalSales, totalProfit, totalCost float64
	var transactionCount int

	for _, sale := range sales {
		totalSales += sale.TotalAmount
		totalProfit += sale.Profit
		totalCost += sale.CostAmount
		transactionCount++
	}

	daysInRange := float64(time.Since(start).Hours() / 24)
	if daysInRange < 1 {
		daysInRange = 1
	}
	dailyAvg := totalSales / daysInRange

	return c.JSON(fiber.Map{
		"type":         "monthly",
		"start_date":   start.Format("2006-01-02"),
		"end_date":     end.Format("2006-01-02"),
		"total_sales":  totalSales,
		"total_profit": totalProfit,
		"total_cost":   totalCost,
		"transactions": transactionCount,
		"daily_avg":    dailyAvg,
	})
}

// GetAnalytics returns analytics data
func (h *ReportHandler) GetAnalytics(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)

	// Get last 30 days sales
	end := time.Now()
	start := end.AddDate(0, 0, -30)

	sales, err := h.saleRepo.GetByDateRange(shopID, start, end)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get sales",
		})
	}

	// Group by day
	dailySales := make(map[string]float64)
	productSales := make(map[string]float64)

	for _, sale := range sales {
		day := sale.CreatedAt.Format("2006-01-02")
		dailySales[day] += sale.TotalAmount
		productSales[sale.Product.Name] += sale.TotalAmount
	}

	// Find top products
	type productStat struct {
		name   string
		amount float64
	}
	var topProducts []productStat
	for name, amount := range productSales {
		topProducts = append(topProducts, productStat{name, amount})
	}
	// Sort by amount descending
	for i := 0; i < len(topProducts)-1; i++ {
		for j := i + 1; j < len(topProducts); j++ {
			if topProducts[j].amount > topProducts[i].amount {
				topProducts[i], topProducts[j] = topProducts[j], topProducts[i]
			}
		}
	}
	if len(topProducts) > 5 {
		topProducts = topProducts[:5]
	}

	return c.JSON(fiber.Map{
		"period":       "last_30_days",
		"total_sales":  len(sales),
		"daily_sales":  dailySales,
		"top_products": topProducts,
	})
}

// BulkCreateProducts creates multiple products at once
func (h *ProductHandler) BulkCreateProducts(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)

	type BulkProduct struct {
		Name              string  `json:"name"`
		Category          string  `json:"category"`
		Unit              string  `json:"unit"`
		CostPrice         float64 `json:"cost_price"`
		SellingPrice      float64 `json:"selling_price"`
		CurrentStock      int     `json:"current_stock"`
		LowStockThreshold int     `json:"low_stock_threshold"`
		Barcode           string  `json:"barcode"`
	}

	var products []BulkProduct
	if err := c.BodyParser(&products); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if len(products) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No products provided",
		})
	}

	if len(products) > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Maximum 100 products per request",
		})
	}

	var created []models.Product
	var errors []string

	for i, p := range products {
		if p.Name == "" {
			errors = append(errors, fmt.Sprintf("Row %d: name required", i+1))
			continue
		}
		if p.SellingPrice <= 0 {
			errors = append(errors, fmt.Sprintf("Row %d: invalid price", i+1))
			continue
		}

		unit := p.Unit
		if unit == "" {
			unit = "pcs"
		}
		threshold := p.LowStockThreshold
		if threshold == 0 {
			threshold = 10
		}

		product := &models.Product{
			ShopID:            shopID,
			Name:              p.Name,
			Category:          p.Category,
			Unit:              unit,
			CostPrice:         p.CostPrice,
			SellingPrice:      p.SellingPrice,
			CurrentStock:      p.CurrentStock,
			LowStockThreshold: threshold,
			Barcode:           p.Barcode,
			IsActive:          true,
		}

		if err := h.productRepo.Create(product); err != nil {
			errors = append(errors, fmt.Sprintf("Row %d: %s", i+1, err.Error()))
			continue
		}
		created = append(created, *product)
	}

	return c.JSON(fiber.Map{
		"created":  len(created),
		"total":    len(products),
		"errors":   errors,
		"products": created,
	})
}

// ListCategories returns all unique categories
func (h *ProductHandler) ListCategories(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)

	categories, err := h.productRepo.GetCategories(shopID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get categories",
		})
	}

	// Also get uncategorized count
	products, _ := h.productRepo.GetByShopID(shopID)
	uncategorized := 0
	for _, p := range products {
		if p.Category == "" {
			uncategorized++
		}
	}

	return c.JSON(fiber.Map{
		"categories":    categories,
		"uncategorized": uncategorized,
		"total":         len(categories) + 1, // +1 for uncategorized
	})
}

// CreateCategory creates a new category (as string in products)
func (h *ProductHandler) CreateCategory(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)

	type Request struct {
		Name string `json:"name"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Category name is required"})
	}

	// Check if category already exists
	existing, _ := h.productRepo.GetCategories(shopID)
	for _, cat := range existing {
		if cat == req.Name {
			return c.Status(400).JSON(fiber.Map{"error": "Category already exists"})
		}
	}

	// Create a placeholder product with this category to "register" it
	placeholder := &models.Product{
		ShopID:   shopID,
		Name:     "__category_" + req.Name,
		Category: req.Name,
		IsActive: false, // Inactive = not shown but category exists
	}

	if err := h.productRepo.Create(placeholder); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create category"})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Category created",
		"id":      placeholder.ID,
		"name":    req.Name,
	})
}

// UpdateCategory renames a category
func (h *ProductHandler) UpdateCategory(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)
	categoryID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid category ID"})
	}

	type Request struct {
		Name string `json:"name"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Category name is required"})
	}

	// Get the placeholder product
	product, err := h.productRepo.GetByID(uint(categoryID))
	if err != nil || product.ShopID != shopID {
		return c.Status(404).JSON(fiber.Map{"error": "Category not found"})
	}

	oldName := product.Category
	product.Category = req.Name
	if err := h.productRepo.Update(product); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update category"})
	}

	// Update all products with old category name
	products, _ := h.productRepo.GetByShopID(shopID)
	for i := range products {
		if products[i].Category == oldName {
			products[i].Category = req.Name
			h.productRepo.Update(&products[i])
		}
	}

	return c.JSON(fiber.Map{
		"message": "Category updated",
		"name":    req.Name,
	})
}

// DeleteCategory deletes a category
func (h *ProductHandler) DeleteCategory(c *fiber.Ctx) error {
	shopID := c.Locals("shop_id").(uint)
	categoryID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid category ID"})
	}

	// Get the placeholder product
	product, err := h.productRepo.GetByID(uint(categoryID))
	if err != nil || product.ShopID != shopID {
		return c.Status(404).JSON(fiber.Map{"error": "Category not found"})
	}

	categoryName := product.Category

	// Update all products with this category to uncategorized
	products, _ := h.productRepo.GetByShopID(shopID)
	for i := range products {
		if products[i].Category == categoryName {
			products[i].Category = ""
			h.productRepo.Update(&products[i])
		}
	}

	// Delete the placeholder product
	h.productRepo.Delete(uint(categoryID))

	return c.JSON(fiber.Map{
		"message": "Category deleted",
	})
}
