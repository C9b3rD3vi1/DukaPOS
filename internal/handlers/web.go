package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"github.com/C9b3rD3vi1/DukaPOS/internal/repository"
	"github.com/gofiber/fiber/v2"
)

// getShopID returns shop_id from JWT token or URL params
func getShopID(c *fiber.Ctx) (uint, error) {
	if sid, ok := c.Locals("shop_id").(uint); ok && sid > 0 {
		return sid, nil
	}
	id, err := strconv.ParseUint(c.Params("shop_id"), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// DashboardData holds all dashboard information
type DashboardData struct {
	Shop        *models.Shop
	Products    []models.Product
	Sales       []models.Sale
	LowStock    []models.Product
	Stats       DashboardStats
	RecentSales []SaleSummary
	TopProducts []ProductSummary
	WeeklyData  []DailyData
}

type DashboardStats struct {
	TotalSales       float64 `json:"total_sales"`
	TotalProfit      float64 `json:"total_profit"`
	TransactionCount int     `json:"transaction_count"`
	ProductCount     int     `json:"product_count"`
	LowStockCount    int     `json:"low_stock_count"`
	AvgTransaction   float64 `json:"avg_transaction"`
	ProfitMargin     float64 `json:"profit_margin"`
}

type SaleSummary struct {
	ID            uint      `json:"id"`
	ProductName   string    `json:"product_name"`
	Quantity      int       `json:"quantity"`
	TotalAmount   float64   `json:"total_amount"`
	PaymentMethod string    `json:"payment_method"`
	CreatedAt     time.Time `json:"created_at"`
}

type ProductSummary struct {
	Name         string  `json:"name"`
	TotalSold    int     `json:"total_sold"`
	TotalRevenue float64 `json:"total_revenue"`
}

type DailyData struct {
	Date   string  `json:"date"`
	Sales  float64 `json:"sales"`
	Profit float64 `json:"profit"`
	Count  int     `json:"count"`
}

// WebHandler handles web dashboard requests
type WebHandler struct {
	shopRepo     *repository.ShopRepository
	productRepo  *repository.ProductRepository
	saleRepo     *repository.SaleRepository
	summaryRepo  *repository.DailySummaryRepository
	customerRepo *repository.CustomerRepository
	staffRepo    *repository.StaffRepository
}

// NewWebHandler creates a new web handler
func NewWebHandler(
	shopRepo *repository.ShopRepository,
	productRepo *repository.ProductRepository,
	saleRepo *repository.SaleRepository,
) *WebHandler {
	return &WebHandler{
		shopRepo:    shopRepo,
		productRepo: productRepo,
		saleRepo:    saleRepo,
	}
}

// SetAdditionalRepos sets additional repositories
func (h *WebHandler) SetAdditionalRepos(
	summaryRepo *repository.DailySummaryRepository,
	customerRepo *repository.CustomerRepository,
	staffRepo *repository.StaffRepository,
) {
	h.summaryRepo = summaryRepo
	h.customerRepo = customerRepo
	h.staffRepo = staffRepo
}

func (h *WebHandler) GetDashboardData(shopID uint) (*DashboardData, error) {
	shop, err := h.shopRepo.GetByID(shopID)
	if err != nil {
		return nil, err
	}

	products, err := h.productRepo.GetByShopID(shop.ID)
	if err != nil {
		products = []models.Product{}
	}

	sales, err := h.saleRepo.GetTodaySales(shop.ID)
	if err != nil {
		sales = []models.Sale{}
	}

	var totalSales, totalProfit float64
	for _, s := range sales {
		totalSales += s.TotalAmount
		totalProfit += s.Profit
	}

	lowStock := []models.Product{}
	for _, p := range products {
		if p.CurrentStock <= p.LowStockThreshold {
			lowStock = append(lowStock, p)
		}
	}

	avgTransaction := 0.0
	if len(sales) > 0 {
		avgTransaction = totalSales / float64(len(sales))
	}

	profitMargin := 0.0
	if totalSales > 0 {
		profitMargin = (totalProfit / totalSales) * 100
	}

	recentSales := make([]SaleSummary, 0, len(sales))
	if len(sales) > 10 {
		sales = sales[:10]
	}
	for _, s := range sales {
		recentSales = append(recentSales, SaleSummary{
			ID:            s.ID,
			ProductName:   s.Product.Name,
			Quantity:      s.Quantity,
			TotalAmount:   s.TotalAmount,
			PaymentMethod: string(s.PaymentMethod),
			CreatedAt:     s.CreatedAt,
		})
	}

	topProducts := h.calculateTopProducts(shop.ID, 5)
	weeklyData := h.calculateWeeklyData(shop.ID)

	return &DashboardData{
		Shop:     shop,
		Products: products,
		Sales:    sales,
		LowStock: lowStock,
		Stats: DashboardStats{
			TotalSales:       totalSales,
			TotalProfit:      totalProfit,
			TransactionCount: len(sales),
			ProductCount:     len(products),
			LowStockCount:    len(lowStock),
			AvgTransaction:   avgTransaction,
			ProfitMargin:     profitMargin,
		},
		RecentSales: recentSales,
		TopProducts: topProducts,
		WeeklyData:  weeklyData,
	}, nil
}

func (h *WebHandler) calculateTopProducts(shopID uint, limit int) []ProductSummary {
	end := time.Now()
	start := end.AddDate(0, 0, -30)
	sales, err := h.saleRepo.GetByDateRange(shopID, start, end)
	if err != nil {
		return []ProductSummary{}
	}

	productMap := make(map[string]ProductSummary)
	for _, s := range sales {
		if existing, ok := productMap[s.Product.Name]; ok {
			existing.TotalSold += s.Quantity
			existing.TotalRevenue += s.TotalAmount
			productMap[s.Product.Name] = existing
		} else {
			productMap[s.Product.Name] = ProductSummary{
				Name:         s.Product.Name,
				TotalSold:    s.Quantity,
				TotalRevenue: s.TotalAmount,
			}
		}
	}

	summaries := make([]ProductSummary, 0, len(productMap))
	for _, p := range productMap {
		summaries = append(summaries, p)
	}

	for i := 0; i < len(summaries)-1; i++ {
		for j := i + 1; j < len(summaries); j++ {
			if summaries[j].TotalRevenue > summaries[i].TotalRevenue {
				summaries[i], summaries[j] = summaries[j], summaries[i]
			}
		}
	}

	if len(summaries) > limit {
		summaries = summaries[:limit]
	}

	return summaries
}

func (h *WebHandler) calculateWeeklyData(shopID uint) []DailyData {
	data := make([]DailyData, 7)
	now := time.Now()

	for i := 6; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
		endOfDay := startOfDay.Add(24 * time.Hour)

		sales, err := h.saleRepo.GetByDateRange(shopID, startOfDay, endOfDay)
		if err != nil {
			sales = []models.Sale{}
		}

		var daySales, dayProfit float64
		for _, s := range sales {
			daySales += s.TotalAmount
			dayProfit += s.Profit
		}

		data[6-i] = DailyData{
			Date:   date.Format("Mon"),
			Sales:  daySales,
			Profit: dayProfit,
			Count:  len(sales),
		}
	}

	return data
}

// Dashboard renders the main dashboard
func (h *WebHandler) Dashboard(c *fiber.Ctx) error {
	shopID, err := getShopID(c)
	if err != nil {
		return c.Status(400).Render("error", fiber.Map{
			"Error": "Invalid shop ID",
		})
	}

	data, err := h.GetDashboardData(shopID)
	if err != nil {
		return c.Status(404).Render("error", fiber.Map{
			"Error": "Shop not found",
		})
	}

	return c.Render("dashboard", fiber.Map{
		"Shop":           data.Shop,
		"Products":       data.Products,
		"ProductCount":   data.Stats.ProductCount,
		"Sales":          data.RecentSales,
		"SaleCount":      data.Stats.TransactionCount,
		"TotalSales":     data.Stats.TotalSales,
		"TotalProfit":    data.Stats.TotalProfit,
		"LowStock":       data.LowStock,
		"LowStockCount":  data.Stats.LowStockCount,
		"AvgTransaction": data.Stats.AvgTransaction,
		"ProfitMargin":   data.Stats.ProfitMargin,
		"TopProducts":    data.TopProducts,
		"WeeklyData":     data.WeeklyData,
		"Title":          fmt.Sprintf("Dashboard - %s", data.Shop.Name),
	})
}

// DashboardJSON returns dashboard data as JSON
func (h *WebHandler) DashboardJSON(c *fiber.Ctx) error {
	shopID, err := getShopID(c)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid shop ID",
		})
	}

	data, err := h.GetDashboardData(shopID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Shop not found",
		})
	}

	return c.JSON(fiber.Map{
		"shop":         data.Shop,
		"stats":        data.Stats,
		"low_stock":    data.LowStock,
		"recent_sales": data.RecentSales,
		"top_products": data.TopProducts,
		"weekly_data":  data.WeeklyData,
		"timestamp":    time.Now().Unix(),
	})
}

// APIProducts handles products API
func (h *WebHandler) APIProducts(c *fiber.Ctx) error {
	// Try JWT shop_id first, fall back to URL param
	var shopID uint
	if sid, ok := c.Locals("shop_id").(uint); ok && sid > 0 {
		shopID = sid
	} else {
		id, err := strconv.ParseUint(c.Params("shop_id"), 10, 32)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid shop ID"})
		}
		shopID = uint(id)
	}

	products, err := h.productRepo.GetByShopID(shopID)
	if err != nil {
		products = []models.Product{}
	}

	return c.JSON(fiber.Map{
		"data":    products,
		"total":   len(products),
		"shop_id": shopID,
	})
}

// APIProductCreate creates a new product via API
func (h *WebHandler) APIProductCreate(c *fiber.Ctx) error {
	// Try JWT shop_id first, fall back to URL param
	var shopID uint
	if sid, ok := c.Locals("shop_id").(uint); ok && sid > 0 {
		shopID = sid
	} else {
		id, err := strconv.ParseUint(c.Params("shop_id"), 10, 32)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid shop ID"})
		}
		shopID = uint(id)
	}

	var req struct {
		Name              string  `json:"name"`
		SellingPrice      float64 `json:"selling_price"`
		CostPrice         float64 `json:"cost_price"`
		CurrentStock      int     `json:"current_stock"`
		LowStockThreshold int     `json:"low_stock_threshold"`
		Category          string  `json:"category"`
		Unit              string  `json:"unit"`
		Barcode           string  `json:"barcode"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Product name is required"})
	}

	if req.SellingPrice <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Selling price must be greater than 0"})
	}

	threshold := req.LowStockThreshold
	if threshold == 0 {
		threshold = 10
	}

	unit := req.Unit
	if unit == "" {
		unit = "pcs"
	}

	product := &models.Product{
		ShopID:            uint(shopID),
		Name:              req.Name,
		SellingPrice:      req.SellingPrice,
		CostPrice:         req.CostPrice,
		CurrentStock:      req.CurrentStock,
		LowStockThreshold: threshold,
		Category:          req.Category,
		Unit:              unit,
		Barcode:           req.Barcode,
		IsActive:          true,
	}

	if err := h.productRepo.Create(product); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create product"})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "Product created successfully",
		"product": product,
	})
}

// APIProductUpdate updates a product
func (h *WebHandler) APIProductUpdate(c *fiber.Ctx) error {
	productID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid product ID"})
	}

	product, err := h.productRepo.GetByID(uint(productID))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Product not found"})
	}

	var req struct {
		Name              *string  `json:"name"`
		SellingPrice      *float64 `json:"selling_price"`
		CostPrice         *float64 `json:"cost_price"`
		CurrentStock      *int     `json:"current_stock"`
		LowStockThreshold *int     `json:"low_stock_threshold"`
		Category          *string  `json:"category"`
		Unit              *string  `json:"unit"`
		Barcode           *string  `json:"barcode"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Name != nil && *req.Name != "" {
		product.Name = *req.Name
	}
	if req.SellingPrice != nil && *req.SellingPrice > 0 {
		product.SellingPrice = *req.SellingPrice
	}
	if req.CostPrice != nil {
		product.CostPrice = *req.CostPrice
	}
	if req.CurrentStock != nil {
		product.CurrentStock = *req.CurrentStock
	}
	if req.LowStockThreshold != nil && *req.LowStockThreshold > 0 {
		product.LowStockThreshold = *req.LowStockThreshold
	}
	if req.Category != nil {
		product.Category = *req.Category
	}
	if req.Unit != nil && *req.Unit != "" {
		product.Unit = *req.Unit
	}
	if req.Barcode != nil {
		product.Barcode = *req.Barcode
	}

	if err := h.productRepo.Update(product); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update product"})
	}

	return c.JSON(fiber.Map{
		"message": "Product updated successfully",
		"product": product,
	})
}

// APIProductDelete deletes a product
func (h *WebHandler) APIProductDelete(c *fiber.Ctx) error {
	productID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid product ID"})
	}

	if err := h.productRepo.Delete(uint(productID)); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete product"})
	}

	return c.JSON(fiber.Map{
		"message": "Product deleted successfully",
	})
}

// APISales handles sales API
func (h *WebHandler) APISales(c *fiber.Ctx) error {
	shopID, err := getShopID(c)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid shop ID"})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	sales, err := h.saleRepo.GetByShopID(shopID, limit)
	if err != nil {
		sales = []models.Sale{}
	}

	return c.JSON(fiber.Map{
		"data":    sales,
		"total":   len(sales),
		"shop_id": shopID,
		"limit":   limit,
		"offset":  offset,
	})
}

// APISaleCreate creates a new sale
func (h *WebHandler) APISaleCreate(c *fiber.Ctx) error {
	shopID, err := getShopID(c)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid shop ID"})
	}

	var req struct {
		ProductID     uint   `json:"product_id"`
		Quantity      int    `json:"quantity"`
		PaymentMethod string `json:"payment_method"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.ProductID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Product ID is required"})
	}

	if req.Quantity <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Quantity must be greater than 0"})
	}

	product, err := h.productRepo.GetByID(req.ProductID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Product not found"})
	}

	if product.CurrentStock < req.Quantity {
		return c.Status(400).JSON(fiber.Map{
			"error":           "Insufficient stock",
			"available_stock": product.CurrentStock,
		})
	}

	totalAmount := product.SellingPrice * float64(req.Quantity)
	costAmount := product.CostPrice * float64(req.Quantity)
	profit := totalAmount - costAmount

	paymentMethod := models.PaymentCash
	if req.PaymentMethod != "" {
		paymentMethod = models.PaymentMethod(req.PaymentMethod)
	}

	sale := &models.Sale{
		ShopID:        uint(shopID),
		ProductID:     req.ProductID,
		Quantity:      req.Quantity,
		UnitPrice:     product.SellingPrice,
		TotalAmount:   totalAmount,
		CostAmount:    costAmount,
		Profit:        profit,
		PaymentMethod: paymentMethod,
	}

	if err := h.saleRepo.Create(sale); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create sale"})
	}

	if err := h.productRepo.UpdateStock(req.ProductID, -req.Quantity); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update stock"})
	}

	return c.Status(201).JSON(fiber.Map{
		"message":   "Sale created successfully",
		"sale":      sale,
		"product":   product.Name,
		"total":     totalAmount,
		"profit":    profit,
		"new_stock": product.CurrentStock - req.Quantity,
	})
}

// APIReports handles reports API
func (h *WebHandler) APIReports(c *fiber.Ctx) error {
	shopID, err := getShopID(c)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid shop ID"})
	}

	reportType := c.Query("type", "daily")
	days := 1
	if reportType == "weekly" {
		days = 7
	} else if reportType == "monthly" {
		days = 30
	}

	end := time.Now()
	start := end.AddDate(0, 0, -days)

	sales, err := h.saleRepo.GetByDateRange(uint(shopID), start, end)
	if err != nil {
		sales = []models.Sale{}
	}

	var totalSales, totalProfit, totalCost float64
	productSales := make(map[string]struct {
		quantity int
		revenue  float64
	})

	for _, s := range sales {
		totalSales += s.TotalAmount
		totalProfit += s.Profit
		totalCost += s.CostAmount

		if existing, ok := productSales[s.Product.Name]; ok {
			productSales[s.Product.Name] = struct {
				quantity int
				revenue  float64
			}{
				quantity: existing.quantity + s.Quantity,
				revenue:  existing.revenue + s.TotalAmount,
			}
		} else {
			productSales[s.Product.Name] = struct {
				quantity int
				revenue  float64
			}{
				quantity: s.Quantity,
				revenue:  s.TotalAmount,
			}
		}
	}

	topProducts := make([]map[string]interface{}, 0)
	for name, data := range productSales {
		topProducts = append(topProducts, map[string]interface{}{
			"name":     name,
			"quantity": data.quantity,
			"revenue":  data.revenue,
		})
	}

	return c.JSON(fiber.Map{
		"type":              reportType,
		"period_days":       days,
		"total_sales":       totalSales,
		"total_profit":      totalProfit,
		"total_cost":        totalCost,
		"transaction_count": len(sales),
		"avg_transaction":   0,
		"profit_margin":     0,
		"top_products":      topProducts,
		"start_date":        start.Format("2006-01-02"),
		"end_date":          end.Format("2006-01-02"),
	})
}

// ProductsList renders the products list page
func (h *WebHandler) ProductsList(c *fiber.Ctx) error {
	shopID, err := getShopID(c)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid shop ID",
		})
	}

	products, err := h.productRepo.GetByShopID(shopID)
	if err != nil {
		products = []models.Product{}
	}

	shop, _ := h.shopRepo.GetByID(uint(shopID))

	lowStockCount := 0
	for _, p := range products {
		if p.CurrentStock <= p.LowStockThreshold {
			lowStockCount++
		}
	}

	return c.Render("products", fiber.Map{
		"Shop":          shop,
		"Products":      products,
		"ProductCount":  len(products),
		"LowStockCount": lowStockCount,
		"Title":         "Products - DukaPOS",
	})
}

// SalesList renders the sales list page
func (h *WebHandler) SalesList(c *fiber.Ctx) error {
	shopID, err := getShopID(c)
	if err != nil {
		return c.Status(400).Render("error", fiber.Map{
			"Error": "Invalid shop ID",
		})
	}

	sales, err := h.saleRepo.GetByShopID(shopID, 100)
	if err != nil {
		sales = []models.Sale{}
	}

	shop, _ := h.shopRepo.GetByID(uint(shopID))

	totalSales := 0.0
	for _, s := range sales {
		totalSales += s.TotalAmount
	}

	return c.Render("sales", fiber.Map{
		"Shop":       shop,
		"Sales":      sales,
		"TotalSales": totalSales,
		"Title":      "Sales - DukaPOS",
	})
}

// Index renders the home page
func (h *WebHandler) Index(c *fiber.Ctx) error {
	return c.Render("index", fiber.Map{
		"Title": "DukaPOS - WhatsApp POS for Kenyan Businesses",
	})
}

// Health renders health check page
func (h *WebHandler) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "healthy",
		"service":   "DukaPOS",
		"version":   "1.0.0",
		"timestamp": fmt.Sprintf("%v", c.Context().Time()),
	})
}

// NotFound renders 404 page
func (h *WebHandler) NotFound(c *fiber.Ctx) error {
	return c.Status(404).Render("error", fiber.Map{
		"Error":   "Page not found",
		"Code":    404,
		"Message": "The page you are looking for does not exist.",
	})
}
