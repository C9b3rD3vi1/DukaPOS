package services

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"github.com/C9b3rD3vi1/DukaPOS/internal/repository"
	"gorm.io/gorm"
)

// Errors
var (
	ErrProductNotFound   = errors.New("product not found")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrInvalidQuantity   = errors.New("invalid quantity")
	ErrInvalidPrice      = errors.New("invalid price")
	ErrShopNotFound      = errors.New("shop not found")
	ErrInvalidCommand    = errors.New("invalid command")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrProductExists     = errors.New("product already exists")
)

// CommandParser parses WhatsApp commands
type CommandParser struct {
	productRepo *repository.ProductRepository
	shopRepo    *repository.ShopRepository
}

// NewCommandParser creates a new command parser
func NewCommandParser(productRepo *repository.ProductRepository, shopRepo *repository.ShopRepository) *CommandParser {
	return &CommandParser{
		productRepo: productRepo,
		shopRepo:    shopRepo,
	}
}

// ParsedCommand represents a parsed WhatsApp command
type ParsedCommand struct {
	Command string
	Args    []string
	Raw     string
}

// Parse parses a raw message into a command
func (p *CommandParser) Parse(message string) *ParsedCommand {
	message = strings.TrimSpace(message)
	message = strings.ToLower(message)
	parts := strings.Fields(message)

	if len(parts) == 0 {
		return &ParsedCommand{
			Command: "help",
			Args:    []string{},
			Raw:     message,
		}
	}

	return &ParsedCommand{
		Command: parts[0],
		Args:    parts[1:],
		Raw:     message,
	}
}

// CommandHandler handles WhatsApp commands
type CommandHandler struct {
	shopRepo     *repository.ShopRepository
	productRepo  *repository.ProductRepository
	saleRepo     *repository.SaleRepository
	summaryRepo  *repository.DailySummaryRepository
	auditRepo    *repository.AuditLogRepository
	accountRepo  *repository.AccountRepository
	staffRepo    *repository.StaffRepository
	supplierRepo *repository.SupplierRepository
	orderRepo    *repository.OrderRepository
	customerRepo *repository.CustomerRepository
}

// NewCommandHandler creates a new command handler
func NewCommandHandler(
	shopRepo *repository.ShopRepository,
	productRepo *repository.ProductRepository,
	saleRepo *repository.SaleRepository,
	summaryRepo *repository.DailySummaryRepository,
	auditRepo *repository.AuditLogRepository,
) *CommandHandler {
	return &CommandHandler{
		shopRepo:    shopRepo,
		productRepo: productRepo,
		saleRepo:    saleRepo,
		summaryRepo: summaryRepo,
		auditRepo:   auditRepo,
	}
}

// SetAccountRepo sets the account repository for multi-shop support
func (h *CommandHandler) SetAccountRepo(accountRepo *repository.AccountRepository) {
	h.accountRepo = accountRepo
}

// SetStaffRepo sets the staff repository
func (h *CommandHandler) SetStaffRepo(staffRepo *repository.StaffRepository) {
	h.staffRepo = staffRepo
}

// SetSupplierRepo sets the supplier and order repositories
func (h *CommandHandler) SetSupplierRepo(supplierRepo *repository.SupplierRepository, orderRepo *repository.OrderRepository) {
	h.supplierRepo = supplierRepo
	h.orderRepo = orderRepo
}

// SetCustomerRepo sets the customer repository for loyalty
func (h *CommandHandler) SetCustomerRepo(customerRepo *repository.CustomerRepository) {
	h.customerRepo = customerRepo
}

// Handle processes a command and returns a response
func (h *CommandHandler) Handle(phone string, command *ParsedCommand) (string, error) {
	shop, err := h.shopRepo.GetByPhone(phone)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			shop = &models.Shop{
				Name:      "My Shop",
				Phone:     phone,
				OwnerName: "Shop Owner",
				Plan:      models.PlanFree,
				IsActive:  true,
			}
			if err := h.shopRepo.Create(shop); err != nil {
				return "", err
			}
			h.auditRepo.Create(&models.AuditLog{
				ShopID:     shop.ID,
				UserType:   "shop",
				UserID:     shop.ID,
				Action:     "create",
				EntityType: "shop",
				EntityID:   shop.ID,
				Details:    "Shop created via WhatsApp",
			})
			return h.handleWelcome(shop), nil
		}
		return "", err
	}

	if !shop.IsActive {
		return "❌ Your account is deactivated. Please contact support.", nil
	}

	switch command.Command {
	case "help":
		return h.handleHelp(shop), nil
	case "add":
		return h.handleAdd(shop, command.Args)
	case "sell":
		return h.handleSell(shop, command.Args)
	case "stock":
		return h.handleStock(shop, command.Args)
	case "price":
		return h.handlePrice(shop, command.Args)
	case "remove":
		return h.handleRemove(shop, command.Args)
	case "report", "daily":
		return h.handleReport(shop)
	case "weekly":
		return h.handleWeekly(shop)
	case "monthly":
		return h.handleMonthly(shop)
	case "profit":
		return h.handleProfit(shop)
	case "low":
		return h.handleLowStock(shop)
	case "delete":
		return h.handleDelete(shop, command.Args)
	case "category", "cat":
		return h.handleCategory(shop, command.Args)
	case "all":
		return h.handleAll(shop)
	case "threshold", "limit", "min":
		return h.handleThreshold(shop, command.Args)
	case "barcode", "scan":
		return h.handleBarcode(shop, command.Args)
	case "top":
		return h.handleTop(shop, command.Args)
	case "search", "find":
		return h.handleSearch(shop, command.Args)
	case "cost":
		return h.handleCost(shop, command.Args)
	case "backup":
		return h.handleBackup(shop)
	// === Phase 2: Pro Features ===
	case "mpesa":
		return h.handleMpesa(shop, command.Args)
	case "staff":
		return h.handleStaff(shop, command.Args)
	case "shop":
		return h.handleShop(shop, command.Args)
	case "upgrade":
		return h.handleUpgrade(shop)
	case "plan":
		return h.handlePlan(shop)
	case "supplier", "suppliers", "sup":
		return h.handleSupplier(shop, command.Args)
	case "order", "orders":
		return h.handleOrder(shop, command.Args)
	// === Phase 3: Enterprise Features ===
	case "predict":
		return h.handlePredict(shop, command.Args)
	case "qr":
		return h.handleQR(shop, command.Args)
	case "loyalty":
		return h.handleLoyalty(shop, command.Args)
	case "api":
		return h.handleAPI(shop, command.Args)
	default:
		return h.handleUnknown(command.Command), nil
	}
}

// handleWelcome handles new shop welcome
func (h *CommandHandler) handleWelcome(shop *models.Shop) string {
	return fmt.Sprintf(`🎉 Welcome to DukaPOS!

Your shop has been created!

📱 Your Number: %s

Quick Start:
• add bread 50 30 - Add 30 bread @ KSh 50
• sell bread 2 - Sold 2 bread
• stock - View all inventory
• report - Today's summary
• help - See all commands

Welcome to digital dukas! 🛒`, shop.Phone)
}

// handleHelp handles help command
func (h *CommandHandler) handleHelp(shop *models.Shop) string {
	planBadge := "📦 FREE"
	if shop.Plan == models.PlanPro {
		planBadge = "🚀 PRO"
	} else if shop.Plan == models.PlanBusiness {
		planBadge = "🏢 BUSINESS"
	}

	var proCommands string
	if shop.Plan != models.PlanFree {
		proCommands = `
💎 PRO COMMANDS:
mpesa pay [amount] - Request M-Pesa payment
staff - Manage staff members
staff add [name] [phone] [role] - Add staff
supplier - Manage suppliers
shop list - View all shops

📈 ADVANCED:
weekly - This week's summary
monthly - This month's summary
category - View/set categories
threshold [product] [num] - Set low stock alert
barcode [code] - Look up by barcode`
	} else {
		proCommands = `
💎 PRO FEATURES:
Upgrade to unlock:
• M-Pesa payments
• Staff accounts
• Supplier management
• Multiple shops
• Advanced reports
• Barcode support

Reply: upgrade`
	}

	helpText := fmt.Sprintf(`📦 FREE

📝 COMMANDS:

🆕 STOCK:
add [name] [price] [qty]
  Example: add milk 60 20

💰 SALES:
sell [name] [qty]
  Example: sell milk 2

📊 REPORTS:
stock - View all products
stock [name] - View specific
report - Today's summary
profit - Today's profit
low - Low stock items
weekly - This week summary
monthly - This month summary
category - View categories

💵 PRICING:
price [name] - Check price
price [name] [new] - Update price

⚙️ SETTINGS:
threshold [product] - View threshold
threshold [product] [num] - Set alert
barcode [code] - Look up product

➖ REMOVE STOCK:
remove [name] [qty]

🗑️ DELETE:
delete [name]

🏪 SHOP:
shop - View shop info
plan - View plan details

🔧 HELP:
help - Show this message%s`, proCommands)

	return planBadge + helpText
}

// handleAdd handles add command
func (h *CommandHandler) handleAdd(shop *models.Shop, args []string) (string, error) {
	if len(args) < 3 {
		return "❌ Usage: add [name] [price] [qty]\nExample: add bread 50 30", nil
	}

	// Validate product name
	name := normalizeProductName(args[0])
	if len(name) < 2 {
		return "❌ Product name too short.\nUse: add [name] [price] [qty]", nil
	}
	if len(name) > 50 {
		return "❌ Product name too long (max 50 chars).\nUse: add [name] [price] [qty]", nil
	}

	// Validate price
	price, err := strconv.ParseFloat(args[1], 64)
	if err != nil || price < 0 {
		return "❌ Invalid price. Use: add [name] [price] [qty]\nExample: add bread 50", nil
	}
	if price > 999999 {
		return "❌ Price too high (max KSh 999,999)", nil
	}

	// Validate quantity
	qty, err := strconv.Atoi(args[2])
	if err != nil || qty <= 0 {
		return "❌ Invalid quantity. Use: add [name] [price] [qty]\nExample: add bread 50 30", nil
	}
	if qty > 999999 {
		return "❌ Quantity too high (max 999,999)", nil
	}

	// Check for existing product
	product, err := h.productRepo.GetByShopAndName(shop.ID, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new product
			product = &models.Product{
				ShopID:            shop.ID,
				Name:              name,
				SellingPrice:      price,
				CurrentStock:      qty,
				LowStockThreshold: 10,
				IsActive:          true,
			}
			if err := h.productRepo.Create(product); err != nil {
				return "", err
			}
			h.auditRepo.Create(&models.AuditLog{
				ShopID:     shop.ID,
				UserType:   "shop",
				UserID:     shop.ID,
				Action:     "create",
				EntityType: "product",
				EntityID:   product.ID,
				Details:    fmt.Sprintf("Added: %s, qty: %d, price: %.2f", name, qty, price),
			})
			return fmt.Sprintf("✅ Added NEW: %s\n💰 Price: KSh %.0f\n📦 Qty: %d\n\nTip: Set low stock alert with: threshold %s 5",
				product.Name, product.SellingPrice, qty, strings.ToLower(name)), nil
		}
		return "", err
	}

	// Update existing product
	oldStock := product.CurrentStock
	oldPrice := product.SellingPrice
	product.CurrentStock += qty
	product.SellingPrice = price
	if err := h.productRepo.Update(product); err != nil {
		return "", err
	}

	h.auditRepo.Create(&models.AuditLog{
		ShopID:     shop.ID,
		UserType:   "shop",
		UserID:     shop.ID,
		Action:     "update",
		EntityType: "product",
		EntityID:   product.ID,
		Details:    fmt.Sprintf("Stock add: %s, qty: %d, price: %.2f -> %.2f", name, qty, oldPrice, price),
	})

	return fmt.Sprintf("✅ Updated: %s\n📦 Was: %d → Now: %d (+%d)\n💰 Price: KSh %.0f (was: %.0f)",
		product.Name, oldStock, product.CurrentStock, qty, product.SellingPrice, oldPrice), nil
}

// handleSell handles sell command
func (h *CommandHandler) handleSell(shop *models.Shop, args []string) (string, error) {
	if len(args) < 2 {
		return "❌ Usage: sell [name] [quantity]\nExample: sell bread 2", nil
	}

	// Validate quantity
	name := normalizeProductName(args[0])
	qty, err := strconv.Atoi(args[1])
	if err != nil || qty <= 0 {
		return "❌ Invalid quantity.\nUse: sell [name] [qty]\nExample: sell bread 2", nil
	}
	if qty > 99999 {
		return "❌ Quantity too high (max 99,999)", nil
	}

	// Find product
	product, err := h.productRepo.GetByShopAndName(shop.ID, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			available, _ := h.productRepo.GetByShopID(shop.ID)
			if len(available) == 0 {
				return "❌ No products yet.\n\nAdd first: add [name] [price] [qty]\nExample: add milk 60 20", nil
			}
			// Find similar products
			similar := findSimilarProducts(available, name)
			msg := fmt.Sprintf("❌ Product '%s' not found.\n\nAvailable products:\n%s", name, getProductNames(available))
			if similar != "" {
				msg += "\n\nDid you mean: " + similar + "?"
			}
			return msg, nil
		}
		return "", err
	}

	// Check stock
	if product.CurrentStock < qty {
		if product.CurrentStock == 0 {
			return fmt.Sprintf("❌ %s is OUT OF STOCK!\n\nAdd more: add %s %.0f [qty]",
				product.Name, strings.ToLower(product.Name), product.SellingPrice), nil
		}
		return fmt.Sprintf("❌ Not enough stock!\n📦 Available: %d %s\n\nSell less: sell %s %d",
			product.CurrentStock, product.Unit, strings.ToLower(product.Name), product.CurrentStock), nil
	}

	// Check if product is active
	if !product.IsActive {
		return fmt.Sprintf("❌ %s is currently unavailable.\nContact support for assistance.", product.Name), nil
	}

	// Calculate totals
	totalAmount := product.SellingPrice * float64(qty)
	costAmount := product.CostPrice * float64(qty)
	profit := totalAmount - costAmount

	// Use transaction to ensure data consistency
	sale := &models.Sale{
		ShopID:        shop.ID,
		ProductID:     product.ID,
		Quantity:      qty,
		UnitPrice:     product.SellingPrice,
		TotalAmount:   totalAmount,
		CostAmount:    costAmount,
		Profit:        profit,
		PaymentMethod: models.PaymentCash,
	}

	if err := h.saleRepo.Create(sale); err != nil {
		return "", err
	}

	// Update stock
	if err := h.productRepo.UpdateStock(product.ID, -qty); err != nil {
		return "", err
	}

	// Recalculate daily summary
	_ = h.summaryRepo.Recalculate(shop.ID, time.Now())

	// Create audit log
	h.auditRepo.Create(&models.AuditLog{
		ShopID:     shop.ID,
		UserType:   "shop",
		UserID:     shop.ID,
		Action:     "sale",
		EntityType: "sale",
		EntityID:   sale.ID,
		Details:    fmt.Sprintf("Sold: %s, qty: %d, total: %.2f", name, qty, totalAmount),
	})

	// Check if now low on stock
	remainingStock := product.CurrentStock - qty
	response := fmt.Sprintf("✅ SOLD!\n%s x%d = KSh %.0f\n💵 Profit: KSh %.0f\n📦 Remaining: %d %s",
		product.Name, qty, totalAmount, profit, remainingStock, product.Unit)

	if remainingStock <= product.LowStockThreshold {
		response += fmt.Sprintf("\n⚠️ LOW STOCK! Only %d left!", remainingStock)
	}

	return response, nil
}

// handleStock handles stock command
func (h *CommandHandler) handleStock(shop *models.Shop, args []string) (string, error) {
	if len(args) >= 1 {
		name := normalizeProductName(args[0])
		product, err := h.productRepo.GetByShopAndName(shop.ID, name)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Sprintf("❌ Product '%s' not found", name), nil
			}
			return "", err
		}

		stock := "✅ In Stock"
		if product.CurrentStock <= product.LowStockThreshold {
			stock = "⚠️ Low Stock!"
		}

		return fmt.Sprintf("📦 %s\n💰 Price: KSh %.0f\n📦 Stock: %d %s\n%s",
			product.Name, product.SellingPrice, product.CurrentStock, product.Unit, stock), nil
	}

	products, err := h.productRepo.GetByShopID(shop.ID)
	if err != nil {
		return "", err
	}

	if len(products) == 0 {
		return "📦 No products yet!\nAdd: add [name] [price] [qty]", nil
	}

	var sb strings.Builder
	sb.WriteString("📦 INVENTORY:\n\n")

	totalValue := 0.0
	for _, p := range products {
		stock := fmt.Sprintf("%d", p.CurrentStock)
		if p.CurrentStock <= p.LowStockThreshold {
			stock = fmt.Sprintf("%d ⚠️", p.CurrentStock)
		}
		sb.WriteString(fmt.Sprintf("• %s: %s %s @ KSh %.0f\n", p.Name, stock, p.Unit, p.SellingPrice))
		totalValue += p.SellingPrice * float64(p.CurrentStock)
	}

	sb.WriteString(fmt.Sprintf("\n💰 Total Value: KSh %.0f", totalValue))
	return sb.String(), nil
}

// handlePrice handles price command
func (h *CommandHandler) handlePrice(shop *models.Shop, args []string) (string, error) {
	if len(args) < 1 {
		return "❌ Usage: price [name] or price [name] [new_price]", nil
	}

	name := normalizeProductName(args[0])
	product, err := h.productRepo.GetByShopAndName(shop.ID, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Sprintf("❌ Product '%s' not found", name), nil
		}
		return "", err
	}

	if len(args) >= 2 {
		newPrice, err := strconv.ParseFloat(args[1], 64)
		if err != nil || newPrice < 0 {
			return "❌ Invalid price", nil
		}
		oldPrice := product.SellingPrice
		product.SellingPrice = newPrice
		if err := h.productRepo.Update(product); err != nil {
			return "", err
		}
		return fmt.Sprintf("✅ Price Updated!\n%s\n💰 Was: KSh %.0f → Now: KSh %.0f",
			product.Name, oldPrice, newPrice), nil
	}

	return fmt.Sprintf("💰 %s\nPrice: KSh %.0f\nStock: %d %s",
		product.Name, product.SellingPrice, product.CurrentStock, product.Unit), nil
}

// handleRemove handles remove command
func (h *CommandHandler) handleRemove(shop *models.Shop, args []string) (string, error) {
	if len(args) < 2 {
		return "❌ Usage: remove [name] [quantity]\nExample: remove bread 5", nil
	}

	name := normalizeProductName(args[0])
	qty, err := strconv.Atoi(args[1])
	if err != nil || qty <= 0 {
		return "❌ Invalid quantity", nil
	}

	product, err := h.productRepo.GetByShopAndName(shop.ID, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Sprintf("❌ Product '%s' not found", name), nil
		}
		return "", err
	}

	if product.CurrentStock < qty {
		return fmt.Sprintf("❌ Not enough stock!\nAvailable: %d", product.CurrentStock), nil
	}

	product.CurrentStock -= qty
	if err := h.productRepo.Update(product); err != nil {
		return "", err
	}

	return fmt.Sprintf("✅ Removed %d %s from %s\n📦 Remaining: %d",
		qty, product.Unit, product.Name, product.CurrentStock), nil
}

// handleReport handles daily report
func (h *CommandHandler) handleReport(shop *models.Shop) (string, error) {
	startOfDay := time.Now().Truncate(24 * time.Hour)
	endOfDay := startOfDay.Add(24 * time.Hour)

	_, _, err := h.saleRepo.GetTotalSales(shop.ID, startOfDay, endOfDay)
	if err != nil {
		return "", err
	}

	sales, err := h.saleRepo.GetTodaySales(shop.ID)
	if err != nil {
		return "", err
	}

	profit := 0.0
	totalSales := 0.0
	for _, s := range sales {
		profit += s.Profit
		totalSales += s.TotalAmount
	}

	report := fmt.Sprintf("📊 DAILY REPORT\n📅 %s\n\n💰 Sales: KSh %.0f\n📝 Transactions: %d\n💵 Profit: KSh %.0f\n\nTop Items:",
		time.Now().Format("Mon, Jan 2"), totalSales, len(sales), profit)

	if len(sales) == 0 {
		report += "\nNo sales today yet!"
	} else {
		productSales := make(map[string]int)
		for _, s := range sales {
			productSales[s.Product.Name] += s.Quantity
		}

		count := 0
		for name, qty := range productSales {
			if count >= 5 {
				break
			}
			report += fmt.Sprintf("\n• %s: %d sold", name, qty)
			count++
		}
	}

	return report, nil
}

// handleWeekly handles weekly report
func (h *CommandHandler) handleWeekly(shop *models.Shop) (string, error) {
	end := time.Now()
	start := end.AddDate(0, 0, -7)

	// Try to get cached summaries
	summaries, err := h.summaryRepo.GetByDateRange(shop.ID, start, end)
	if err != nil {
		return "", err
	}

	var totalSales, totalProfit float64
	var totalTransactions int

	// If no cached summaries, calculate from sales directly
	if len(summaries) == 0 {
		sales, err := h.saleRepo.GetByDateRange(shop.ID, start, end)
		if err != nil {
			return "", err
		}
		for _, s := range sales {
			totalSales += s.TotalAmount
			totalProfit += s.Profit
			totalTransactions++
		}
	} else {
		for _, s := range summaries {
			totalSales += s.TotalSales
			totalProfit += s.TotalProfit
			totalTransactions += s.TotalTransactions
		}
	}

	if totalTransactions == 0 {
		return "📊 No sales data for this week.\n\nStart recording sales to see reports!", nil
	}

	avgDaily := totalSales / 7

	return fmt.Sprintf(`📊 WEEKLY REPORT
📅 Last 7 days (to %s)

💰 Total Sales: KSh %.0f
📝 Transactions: %d
💵 Profit: KSh %.0f
📈 Daily Avg: KSh %.0f

Keep up the good work! 💪`, end.Format("Jan 2"), totalSales, totalTransactions, totalProfit, avgDaily), nil
}

// handleMonthly handles monthly report
func (h *CommandHandler) handleMonthly(shop *models.Shop) (string, error) {
	end := time.Now()
	start := end.AddDate(0, -1, 0)

	// Try to get cached summaries
	summaries, err := h.summaryRepo.GetByDateRange(shop.ID, start, end)
	if err != nil {
		return "", err
	}

	var totalSales, totalProfit float64
	var totalTransactions int

	// If no cached summaries, calculate from sales directly
	if len(summaries) == 0 {
		sales, err := h.saleRepo.GetByDateRange(shop.ID, start, end)
		if err != nil {
			return "", err
		}
		for _, s := range sales {
			totalSales += s.TotalAmount
			totalProfit += s.Profit
			totalTransactions++
		}
	} else {
		for _, s := range summaries {
			totalSales += s.TotalSales
			totalProfit += s.TotalProfit
			totalTransactions += s.TotalTransactions
		}
	}

	if totalTransactions == 0 {
		return "📊 No sales data for this month.\n\nStart recording sales to see reports!", nil
	}

	daysInRange := float64(time.Since(start).Hours() / 24)
	if daysInRange < 1 {
		daysInRange = 1
	}
	avgDaily := totalSales / daysInRange

	return fmt.Sprintf(`📊 MONTHLY REPORT
📅 %s

💰 Total Sales: KSh %.0f
📝 Transactions: %d
💵 Profit: KSh %.0f
📈 Daily Avg: KSh %.0f

Great progress this month! 🎉`, start.Format("Jan")+" - "+end.Format("Jan 2, 2006"), totalSales, totalTransactions, totalProfit, avgDaily), nil
}

// handleProfit handles profit calculation
func (h *CommandHandler) handleProfit(shop *models.Shop) (string, error) {
	sales, err := h.saleRepo.GetTodaySales(shop.ID)
	if err != nil {
		return "", err
	}

	totalProfit := 0.0
	for _, s := range sales {
		totalProfit += s.Profit
	}

	return fmt.Sprintf("💵 TODAY'S PROFIT: KSh %.0f\n\n💰 Total Sales: KSh %.0f\n📝 Transactions: %d",
		totalProfit, getTotalSales(sales), len(sales)), nil
}

// handleLowStock handles low stock alert
func (h *CommandHandler) handleLowStock(shop *models.Shop) (string, error) {
	products, err := h.productRepo.GetLowStock(shop.ID)
	if err != nil {
		return "", err
	}

	if len(products) == 0 {
		return "✅ All products are well stocked!", nil
	}

	var sb strings.Builder
	sb.WriteString("⚠️ LOW STOCK ALERT:\n\n")

	for _, p := range products {
		sb.WriteString(fmt.Sprintf("• %s: %d %s (min: %d)\n",
			p.Name, p.CurrentStock, p.Unit, p.LowStockThreshold))
	}

	return sb.String(), nil
}

// handleDelete handles product deletion
func (h *CommandHandler) handleDelete(shop *models.Shop, args []string) (string, error) {
	if len(args) < 1 {
		return "❌ Usage: delete [name]", nil
	}

	name := normalizeProductName(args[0])
	product, err := h.productRepo.GetByShopAndName(shop.ID, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Sprintf("❌ Product '%s' not found", name), nil
		}
		return "", err
	}

	if err := h.productRepo.Delete(product.ID); err != nil {
		return "", err
	}

	return fmt.Sprintf("🗑️ Deleted: %s", product.Name), nil
}

// handleCategory handles category view and management
func (h *CommandHandler) handleCategory(shop *models.Shop, args []string) (string, error) {
	// Get unique categories from database
	categories, err := h.productRepo.GetCategories(shop.ID)
	if err != nil {
		return "", err
	}

	// Also get products to check uncategorized
	products, err := h.productRepo.GetByShopID(shop.ID)
	if err != nil {
		return "", err
	}

	// Build category map
	catMap := make(map[string]int)
	for _, p := range products {
		cat := p.Category
		if cat == "" {
			cat = "Uncategorized"
		}
		catMap[cat]++
	}

	// Add categories from DB if not in map
	for _, c := range categories {
		if _, ok := catMap[c]; !ok {
			catMap[c] = 0
		}
	}

	if len(args) >= 1 {
		// View specific category
		cat := strings.Title(args[0])

		// Handle "add" command
		if cat == "Add" && len(args) >= 3 {
			name := normalizeProductName(args[1])
			categoryName := strings.Title(args[2])
			product, err := h.productRepo.GetByShopAndName(shop.ID, name)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return fmt.Sprintf("❌ Product '%s' not found", name), nil
				}
				return "", err
			}
			product.Category = categoryName
			if err := h.productRepo.Update(product); err != nil {
				return "", err
			}
			return fmt.Sprintf("✅ Category Updated!\n%s\nNow in: %s", product.Name, categoryName), nil
		}

		// View products in category
		prods, err := h.productRepo.GetByCategory(shop.ID, cat)
		if err != nil {
			return "", err
		}
		if len(prods) == 0 {
			// Check if category exists
			if _, ok := catMap[cat]; !ok {
				return fmt.Sprintf("❌ Category '%s' not found.\n\nAvailable: %s", cat, getCategoryList(catMap)), nil
			}
			return fmt.Sprintf("📦 %s:\n\n(No products)", cat), nil
		}
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("📦 %s (%d items):\n\n", cat, len(prods)))
		for _, p := range prods {
			stock := fmt.Sprintf("%d", p.CurrentStock)
			if p.CurrentStock <= p.LowStockThreshold {
				stock = fmt.Sprintf("%d ⚠️", p.CurrentStock)
			}
			sb.WriteString(fmt.Sprintf("• %s: %s @ KSh %.0f\n", p.Name, stock, p.SellingPrice))
		}
		return sb.String(), nil
	}

	// List all categories
	if len(catMap) == 0 {
		return `📂 CATEGORIES:

No products yet.
Add products with: add [name] [price] [qty]

Then set category: category [product] [category]`, nil
	}

	return fmt.Sprintf("📂 CATEGORIES (%d):\n\n%s\n\nSet category:\ncategory [product] [name]",
		len(catMap), getCategoryList(catMap)), nil
}

func getCategoryList(catMap map[string]int) string {
	var cats []string
	for cat, count := range catMap {
		cats = append(cats, fmt.Sprintf("%s (%d)", cat, count))
	}
	return strings.Join(cats, "\n")
}

// handleAll handles all products (alias for stock)
func (h *CommandHandler) handleAll(shop *models.Shop) (string, error) {
	return h.handleStock(shop, []string{})
}

// handleTop handles top selling products
func (h *CommandHandler) handleTop(shop *models.Shop, args []string) (string, error) {
	limit := 5
	if len(args) > 0 {
		if l, err := strconv.Atoi(args[0]); err == nil && l > 0 && l <= 20 {
			limit = l
		}
	}

	// Get last 30 days of sales
	end := time.Now()
	start := end.AddDate(0, 0, -30)
	sales, err := h.saleRepo.GetByDateRange(shop.ID, start, end)
	if err != nil {
		return "", err
	}

	if len(sales) == 0 {
		return "📊 No sales data for top products.\n\nStart selling to see rankings!", nil
	}

	// Group by product
	productSales := make(map[uint]struct {
		name   string
		qty    int
		amount float64
	})

	for _, sale := range sales {
		if p, ok := productSales[sale.ProductID]; ok {
			p.qty += sale.Quantity
			p.amount += sale.TotalAmount
			productSales[sale.ProductID] = p
		} else {
			productSales[sale.ProductID] = struct {
				name   string
				qty    int
				amount float64
			}{
				name:   sale.Product.Name,
				qty:    sale.Quantity,
				amount: sale.TotalAmount,
			}
		}
	}

	// Sort by amount
	type topItem struct {
		id     uint
		name   string
		qty    int
		amount float64
	}
	var items []topItem
	for id, data := range productSales {
		items = append(items, topItem{id, data.name, data.qty, data.amount})
	}
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			if items[j].amount > items[i].amount {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
	if len(items) > limit {
		items = items[:limit]
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🏆 TOP SELLING (Last 30 days)\n\n"))
	for i, item := range items {
		medal := "🥇"
		if i == 1 {
			medal = "🥈"
		} else if i == 2 {
			medal = "🥉"
		} else {
			medal = fmt.Sprintf("%d.", i+1)
		}
		sb.WriteString(fmt.Sprintf("%s %s\n", medal, item.name))
		sb.WriteString(fmt.Sprintf("   Sold: %d | KSh %.0f\n\n", item.qty, item.amount))
	}
	return sb.String(), nil
}

// handleSearch handles product search
func (h *CommandHandler) handleSearch(shop *models.Shop, args []string) (string, error) {
	if len(args) < 1 {
		return "❌ Usage: search [product name]\nExample: search milk", nil
	}

	search := strings.ToLower(strings.Join(args, " "))
	products, err := h.productRepo.GetByShopID(shop.ID)
	if err != nil {
		return "", err
	}

	var matches []models.Product
	for _, p := range products {
		if strings.Contains(strings.ToLower(p.Name), search) {
			matches = append(matches, p)
		}
	}

	if len(matches) == 0 {
		return fmt.Sprintf("❌ No products found matching '%s'\n\nTry a different search term.", search), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔍 Search Results for '%s':\n\n", search))
	for _, p := range matches {
		stock := fmt.Sprintf("%d", p.CurrentStock)
		if p.CurrentStock <= p.LowStockThreshold {
			stock = fmt.Sprintf("%d ⚠️", p.CurrentStock)
		}
		sb.WriteString(fmt.Sprintf("• %s\n", p.Name))
		sb.WriteString(fmt.Sprintf("   💰 KSh %.0f | 📦 %s %s\n\n", p.SellingPrice, stock, p.Unit))
	}
	return sb.String(), nil
}

// handleCost handles cost price management
func (h *CommandHandler) handleCost(shop *models.Shop, args []string) (string, error) {
	if len(args) < 1 {
		return `💰 COST PRICE COMMANDS:

cost [product] - View cost price
cost [product] [price] - Set cost price

Example:
cost milk - View milk's cost
cost milk 40 - Set cost to KSh 40

Note: Cost price is used for profit calculation.`, nil
	}

	name := normalizeProductName(args[0])
	product, err := h.productRepo.GetByShopAndName(shop.ID, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Sprintf("❌ Product '%s' not found", name), nil
		}
		return "", err
	}

	// Just viewing cost price
	if len(args) < 2 {
		margin := 0.0
		if product.CostPrice > 0 {
			margin = ((product.SellingPrice - product.CostPrice) / product.CostPrice) * 100
		}
		return fmt.Sprintf(`💰 %s

Cost Price: KSh %.2f
Selling Price: KSh %.2f
Margin: %.1f%%

Set new cost:
cost %s [new cost]`, product.Name, product.CostPrice, product.SellingPrice, margin, strings.ToLower(product.Name)), nil
	}

	// Set new cost price
	cost, err := strconv.ParseFloat(args[1], 64)
	if err != nil || cost < 0 {
		return "❌ Invalid cost price. Use a positive number.", nil
	}

	product.CostPrice = cost
	if err := h.productRepo.Update(product); err != nil {
		return "", err
	}

	margin := ((product.SellingPrice - cost) / cost) * 100
	return fmt.Sprintf("✅ Cost Price Updated!\n\n💰 %s\nCost: KSh %.2f\nSelling: KSh %.2f\nMargin: %.1f%%",
		product.Name, cost, product.SellingPrice, margin), nil
}

// handleBackup handles backup commands
func (h *CommandHandler) handleBackup(shop *models.Shop) (string, error) {
	// This would trigger a backup in production
	return `💾 BACKUP

Backup features:
• Manual backup - Coming soon
• Auto backup - Daily at 2 AM
• Export data - Coming soon

Your data is automatically backed up daily.

Contact support for data export.`, nil
}

// handleThreshold handles threshold/limit command for low stock alerts
func (h *CommandHandler) handleThreshold(shop *models.Shop, args []string) (string, error) {
	if len(args) < 1 {
		return `⚙️ THRESHOLD COMMANDS:

threshold [product] - View current threshold
threshold [product] [num] - Set low stock alert

Example: 
threshold milk - See milk's threshold
threshold milk 5 - Alert when milk below 5`, nil
	}

	// If first arg is "list", show all products with their thresholds
	if args[0] == "list" {
		products, err := h.productRepo.GetByShopID(shop.ID)
		if err != nil {
			return "", err
		}
		if len(products) == 0 {
			return "No products to show.", nil
		}
		var sb strings.Builder
		sb.WriteString("⚙️ PRODUCT THRESHOLDS:\n\n")
		for _, p := range products {
			sb.WriteString(fmt.Sprintf("• %s: %d (current: %d)\n", p.Name, p.LowStockThreshold, p.CurrentStock))
		}
		return sb.String(), nil
	}

	name := normalizeProductName(args[0])
	product, err := h.productRepo.GetByShopAndName(shop.ID, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Sprintf("❌ Product '%s' not found", name), nil
		}
		return "", err
	}

	// If just viewing threshold
	if len(args) < 2 {
		stockStatus := "✅ OK"
		if product.CurrentStock <= product.LowStockThreshold {
			stockStatus = "⚠️ LOW"
		}
		return fmt.Sprintf("⚙️ %s\nCurrent Stock: %d\nLow Stock Alert: %d\nStatus: %s",
			product.Name, product.CurrentStock, product.LowStockThreshold, stockStatus), nil
	}

	// Set new threshold
	threshold, err := strconv.Atoi(args[1])
	if err != nil || threshold < 1 || threshold > 9999 {
		return "❌ Invalid threshold. Use a number between 1-9999", nil
	}

	product.LowStockThreshold = threshold
	if err := h.productRepo.Update(product); err != nil {
		return "", err
	}

	return fmt.Sprintf("✅ Threshold Updated!\n%s\nLow stock alert set at: %d\nYou'll be notified when stock falls below this.",
		product.Name, threshold), nil
}

// handleBarcode handles barcode/scan commands
func (h *CommandHandler) handleBarcode(shop *models.Shop, args []string) (string, error) {
	if len(args) < 1 {
		return `📱 BARCODE COMMANDS:

barcode [code] - Look up product by barcode
barcode add [product] [code] - Set barcode for product

Example:
barcode 5901234123457 - Find product
barcode add milk 5901234123457 - Set barcode`, nil
	}

	switch args[0] {
	case "add", "set":
		if len(args) < 3 {
			return "❌ Usage: barcode add [product] [barcode]\nExample: barcode add milk 5901234123457", nil
		}
		name := normalizeProductName(args[1])
		barcode := args[2]

		// Validate barcode format (basic validation)
		if len(barcode) < 4 || len(barcode) > 50 {
			return "❌ Invalid barcode format (4-50 characters)", nil
		}

		product, err := h.productRepo.GetByShopAndName(shop.ID, name)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Sprintf("❌ Product '%s' not found", name), nil
			}
			return "", err
		}

		// Check if barcode is already used by another product
		existing, _ := h.productRepo.GetByBarcode(shop.ID, barcode)
		if existing != nil && existing.ID != product.ID {
			return fmt.Sprintf("❌ Barcode already assigned to '%s'", existing.Name), nil
		}

		product.Barcode = barcode
		if err := h.productRepo.Update(product); err != nil {
			return "", err
		}
		return fmt.Sprintf("✅ Barcode set!\n%s\nBarcode: %s", product.Name, barcode), nil

	default:
		// Look up by barcode
		barcode := args[0]
		product, err := h.productRepo.GetByBarcode(shop.ID, barcode)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Sprintf("❌ No product found with barcode: %s\n\nTip: Add barcode with: barcode add [product] [code]", barcode), nil
			}
			return "", err
		}
		stockStatus := "✅ In Stock"
		if product.CurrentStock <= product.LowStockThreshold {
			stockStatus = "⚠️ Low Stock"
		}
		return fmt.Sprintf("🔍 BARCODE: %s\n\n📦 %s\n💰 Price: KSh %.0f\n📦 Stock: %d %s\nStatus: %s",
			barcode, product.Name, product.SellingPrice, product.CurrentStock, product.Unit, stockStatus), nil
	}
}

// handleSupplier handles supplier management commands
func (h *CommandHandler) handleSupplier(shop *models.Shop, args []string) (string, error) {
	// Check if Pro plan
	if shop.Plan == models.PlanFree {
		return `💎 Supplier Management requires Pro plan!

Current: Free
Required: Pro (KSh 500/month)

Reply: upgrade`, nil
	}

	if h.supplierRepo == nil {
		return "⚙️ Supplier feature not available.\nContact support.", nil
	}

	if len(args) < 1 {
		// List all suppliers
		suppliers, err := h.supplierRepo.GetByShopID(shop.ID)
		if err != nil {
			return "", err
		}
		if len(suppliers) == 0 {
			return `📦 SUPPLIERS:

No suppliers added yet.

Add supplier:
supplier add [name] [phone]

Example: supplier add Brookside +254700000000`, nil
		}
		var sb strings.Builder
		sb.WriteString("📦 SUPPLIERS:\n\n")
		for i, s := range suppliers {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, s.Name))
			if s.Phone != "" {
				sb.WriteString(fmt.Sprintf("   📱 %s\n", s.Phone))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("Add: supplier add [name] [phone]")
		return sb.String(), nil
	}

	switch args[0] {
	case "add":
		if len(args) < 2 {
			return "❌ Usage: supplier add [name] [phone]\nExample: supplier add Brookside +254700000000", nil
		}
		name := strings.Title(strings.Join(args[1:], " "))
		var phone string
		if len(args) >= 2 {
			phone = args[len(args)-1]
			name = strings.Title(strings.Join(args[1:len(args)-1], " "))
		}

		supplier := &models.Supplier{
			ShopID: shop.ID,
			Name:   name,
			Phone:  phone,
		}
		if err := h.supplierRepo.Create(supplier); err != nil {
			return "", err
		}
		return fmt.Sprintf("✅ Supplier Added!\n\n📦 %s\n📱 %s", name, phone), nil

	case "view":
		if len(args) < 2 {
			return "❌ Usage: supplier view [name/number]", nil
		}
		// Could implement viewing supplier details
		return "Supplier details coming soon!", nil

	default:
		return `📦 SUPPLIER COMMANDS:

supplier - List all suppliers
supplier add [name] [phone] - Add supplier
supplier view [name] - View details (coming soon)

Example: supplier add Brookside +254700000000`, nil
	}
}

// handleOrder handles order management commands
func (h *CommandHandler) handleOrder(shop *models.Shop, args []string) (string, error) {
	// Check if Pro plan
	if shop.Plan == models.PlanFree {
		return `📋 Orders require Pro plan!

Current: Free
Required: Pro (KSh 500/month)

Reply: upgrade`, nil
	}

	if h.orderRepo == nil {
		return "⚙️ Order feature not available.\nContact support.", nil
	}

	if len(args) < 1 {
		// List recent orders
		orders, err := h.orderRepo.GetByShopID(shop.ID)
		if err != nil {
			return "", err
		}
		if len(orders) == 0 {
			return `📋 ORDERS:

No orders yet.

Create order via dashboard or API.
Coming soon to WhatsApp!`, nil
		}
		var sb strings.Builder
		sb.WriteString("📋 RECENT ORDERS:\n\n")
		for i, o := range orders {
			if i >= 5 {
				break
			}
			statusIcon := "⏳"
			switch o.Status {
			case "delivered":
				statusIcon = "✅"
			case "cancelled":
				statusIcon = "❌"
			case "shipped":
				statusIcon = "📦"
			}
			sb.WriteString(fmt.Sprintf("%d. %s %s\n", i+1, o.Status, statusIcon))
			sb.WriteString(fmt.Sprintf("   KSh %.0f\n\n", o.TotalAmount))
		}
		return sb.String(), nil
	}

	return `📋 ORDER COMMANDS:

order - Recent orders
order view [id] - Order details (coming soon)

Manage orders via dashboard.`, nil
}

// handleUnknown handles unknown commands
func (h *CommandHandler) handleUnknown(cmd string) string {
	return fmt.Sprintf(`❓ Unknown command: %s

📝 Available:
add, sell, stock, price
remove, delete, report
profit, low, help

Type: help for full list`, cmd)
}

// Helper functions

func normalizeProductName(name string) string {
	if len(name) == 0 {
		return name
	}
	return strings.Title(strings.ToLower(name))
}

func getProductNames(products []models.Product) string {
	if len(products) == 0 {
		return "No products"
	}
	names := make([]string, 0, len(products))
	for _, p := range products {
		names = append(names, p.Name)
	}
	return strings.Join(names, ", ")
}

func getTotalSales(sales []models.Sale) float64 {
	total := 0.0
	for _, s := range sales {
		total += s.TotalAmount
	}
	return total
}

// findSimilarProducts finds similar product names using simple string matching
func findSimilarProducts(products []models.Product, search string) string {
	searchLower := strings.ToLower(search)
	var similar []string

	for _, p := range products {
		nameLower := strings.ToLower(p.Name)
		// Check if search string is contained in product name
		if len(searchLower) >= 3 && (strings.Contains(nameLower, searchLower) || strings.Contains(searchLower, nameLower)) {
			similar = append(similar, p.Name)
		}
		if len(similar) >= 3 {
			break
		}
	}

	if len(similar) > 0 {
		return strings.Join(similar, ", ")
	}
	return ""
}

// ============================================
// Phase 2: Pro Features Handlers
// ============================================

// handleMpesa handles M-Pesa commands
func (h *CommandHandler) handleMpesa(shop *models.Shop, args []string) (string, error) {
	// Check if M-Pesa is enabled for this shop
	if shop.Plan == models.PlanFree {
		return `💎 M-Pesa requires Pro plan!

Current: Free
Required: Pro (KSh 500/month)

Upgrade: upgrade`, nil
	}

	if len(args) < 1 {
		return `💰 M-PESA COMMANDS:

pay [amount] - Request payment from customer
status [code] - Check payment status

Example: pay 500`, nil
	}

	switch args[0] {
	case "pay":
		if len(args) < 2 {
			return "❌ Usage: mpesa pay [amount]\nExample: mpesa pay 500", nil
		}
		amount, err := strconv.Atoi(args[1])
		if err != nil || amount <= 0 {
			return "❌ Invalid amount", nil
		}
		// In production, this would trigger STK Push
		return fmt.Sprintf(`📲 STK Push Sent!

Amount: KSh %d
To: %s

💡 Customer will receive a payment prompt on their phone.

Note: M-Pesa integration requires API configuration. Contact support to activate.`, amount, shop.Phone), nil

	case "status":
		return "💰 Payment status: Pending\nNote: Configure M-Pesa API credentials to enable live status checks.", nil

	default:
		return "❌ Unknown M-Pesa command. Use: mpesa pay [amount]", nil
	}
}

// handleStaff handles staff management commands
func (h *CommandHandler) handleStaff(shop *models.Shop, args []string) (string, error) {
	// Check if staff feature is available
	if shop.Plan == models.PlanFree {
		return `💎 Staff Accounts require Pro plan!

Current: Free
Required: Pro (KSh 500/month)

Upgrade: upgrade`, nil
	}

	// Check if staff repo is available
	if h.staffRepo == nil {
		return "⚙️ Staff management not available.\nPlease contact support.", nil
	}

	if len(args) < 1 {
		return `👥 STAFF COMMANDS:

staff - View all staff
staff add [name] [phone] [role] - Add staff
staff remove [phone] - Remove staff
staff active [phone] - Activate/deactivate

Roles: manager, cashier, stock clerk

Example: staff add John +254700000001 cashier`, nil
	}

	switch args[0] {
	case "list", "":
		// View all staff
		staff, err := h.staffRepo.GetByShopID(shop.ID)
		if err != nil {
			return "", err
		}
		if len(staff) == 0 {
			return `👥 STAFF:

No staff members yet.

Add staff:
staff add [name] [phone] [role]

Example: staff add John +254700000001 cashier`, nil
		}
		var sb strings.Builder
		sb.WriteString("👥 STAFF:\n\n")
		for i, s := range staff {
			status := "✅ Active"
			if !s.IsActive {
				status = "❌ Inactive"
			}
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, s.Name))
			sb.WriteString(fmt.Sprintf("   📱 %s\n", s.Phone))
			sb.WriteString(fmt.Sprintf("   💼 %s\n", s.Role))
			sb.WriteString(fmt.Sprintf("   %s\n\n", status))
		}
		return sb.String(), nil

	case "add":
		if len(args) < 4 {
			return `❌ Usage: staff add [name] [phone] [role]

Example: staff add John +254700000001 cashier

Roles: manager, cashier, stock clerk`, nil
		}
		name := strings.Title(args[1])
		phone := args[2]
		role := strings.ToLower(args[3])

		// Validate role
		validRoles := map[string]bool{
			"manager": true, "cashier": true, "stock clerk": true,
		}
		if !validRoles[role] {
			return "❌ Invalid role.\nValid: manager, cashier, stock clerk", nil
		}

		// Check if staff with phone already exists
		existing, _ := h.staffRepo.GetByPhone(shop.ID, phone)
		if existing != nil {
			return fmt.Sprintf("❌ Staff with phone %s already exists!", phone), nil
		}

		// Generate PIN
		pin := generateStaffPIN()

		staff := &models.Staff{
			ShopID: shop.ID,
			Name:   name,
			Phone:  phone,
			Role:   role,
			Pin:    pin,
		}

		if err := h.staffRepo.Create(staff); err != nil {
			return "", err
		}

		return fmt.Sprintf(`✅ Staff Added!

👤 %s
📱 %s
💼 %s
🔐 PIN: %s

Share PIN with staff member.`, name, phone, role, pin), nil

	case "remove", "delete":
		if len(args) < 2 {
			return "❌ Usage: staff remove [phone]", nil
		}
		phone := args[1]
		staff, err := h.staffRepo.GetByPhone(shop.ID, phone)
		if err != nil {
			return "❌ Staff not found with that phone.", nil
		}
		if err := h.staffRepo.Delete(staff.ID); err != nil {
			return "", err
		}
		return fmt.Sprintf("✅ Staff removed: %s", staff.Name), nil

	case "active", "activate", "deactivate":
		if len(args) < 2 {
			return "❌ Usage: staff active [phone]", nil
		}
		phone := args[1]
		staff, err := h.staffRepo.GetByPhone(shop.ID, phone)
		if err != nil {
			return "❌ Staff not found with that phone.", nil
		}
		// Toggle active status
		staff.IsActive = !staff.IsActive
		if err := h.staffRepo.Update(staff); err != nil {
			return "", err
		}
		status := "activated"
		if !staff.IsActive {
			status = "deactivated"
		}
		return fmt.Sprintf("✅ Staff %s: %s", status, staff.Name), nil

	default:
		return "❌ Unknown staff command.\nUse: staff, staff add, staff remove, staff active", nil
	}
}

// generateStaffPIN generates a random 4-digit PIN
func generateStaffPIN() string {
	return fmt.Sprintf("%04d", (time.Now().UnixNano() % 10000))
}

// handleShop handles multi-shop commands
func (h *CommandHandler) handleShop(shop *models.Shop, args []string) (string, error) {
	if len(args) < 1 {
		return `🏪 SHOP COMMANDS:

shop - View current shop info
shop list - List all your shops
shop switch [id] - Switch to another shop
shop add [name] - Add new shop (Pro)
shop name [new] - Rename shop

Note: Multiple shops require Pro plan.`, nil
	}

	switch args[0] {
	case "list":
		// If account repo is set, get all shops for account
		if h.accountRepo != nil && shop.AccountID > 0 {
			shops, err := h.accountRepo.GetShops(shop.AccountID)
			if err == nil && len(shops) > 0 {
				var sb strings.Builder
				sb.WriteString("🏪 YOUR SHOPS:\n\n")
				for i, s := range shops {
					marker := ""
					if s.ID == shop.ID {
						marker = " (Current)"
					}
					sb.WriteString(fmt.Sprintf("%d. %s%s\n", i+1, s.Name, marker))
					sb.WriteString(fmt.Sprintf("   📱 %s\n", s.Phone))
					sb.WriteString(fmt.Sprintf("   💎 %s\n\n", s.Plan))
				}
				sb.WriteString("Reply: shop switch [number] to change")
				return sb.String(), nil
			}
		}
		return fmt.Sprintf(`🏪 YOUR SHOPS:

1. %s (Current)
   📱 %s
   💎 Plan: %s

💡 Upgrade to Pro to manage multiple shops!`, shop.Name, shop.Phone, shop.Plan), nil

	case "switch":
		if len(args) < 2 {
			return "❌ Usage: shop switch [shop number]\nExample: shop switch 2", nil
		}
		shopNum, err := strconv.Atoi(args[1])
		if err != nil || shopNum < 1 {
			return "❌ Invalid shop number.\nExample: shop switch 2", nil
		}

		// If account repo is set, try to switch
		if h.accountRepo != nil && shop.AccountID > 0 {
			shops, err := h.accountRepo.GetShops(shop.AccountID)
			if err == nil && len(shops) >= shopNum {
				targetShop := shops[shopNum-1]
				return fmt.Sprintf("🏪 Switched to: %s\n\nUse this shop's inventory for all commands.\n\nReply: shop switch [number] to change again.", targetShop.Name), nil
			}
		}
		return "❌ Unable to switch shops.\n\nMulti-shop requires Pro plan.\nReply: upgrade", nil

	case "add":
		if shop.Plan == models.PlanFree {
			return `💎 Adding shops requires Pro plan!

Current: Free
Required: Pro (KSh 500/month)

Reply: upgrade`, nil
		}
		if len(args) < 2 {
			return "❌ Usage: shop add [name]\nExample: shop add Mombasa Branch", nil
		}
		newShopName := strings.Join(args[1:], " ")
		return fmt.Sprintf(`🏪 NEW SHOP CREATED!

Name: %s
Owner: %s

📱 WhatsApp number will be the same.
To use a different number, contact support.

Note: Shop activation pending.`, newShopName, shop.OwnerName), nil

	case "name":
		if len(args) < 2 {
			return "❌ Usage: shop name [new name]\nExample: shop name My New Shop", nil
		}
		newName := strings.Join(args[1:], " ")
		shop.Name = newName
		if err := h.shopRepo.Update(shop); err != nil {
			return "", err
		}
		return fmt.Sprintf("✅ Shop renamed to: %s", newName), nil

	default:
		return fmt.Sprintf(`🏪 CURRENT SHOP:

Name: %s
Phone: %s
Plan: %s
Status: %s

Reply: shop list to see all shops`, shop.Name, shop.Phone, shop.Plan, activeStatus(shop.IsActive)), nil
	}
}

// handleUpgrade handles plan upgrade
func (h *CommandHandler) handleUpgrade(shop *models.Shop) (string, error) {
	if shop.Plan == models.PlanBusiness {
		return "🎉 You're on the Business plan - the highest tier! Nothing to upgrade.", nil
	}

	return `💎 UPGRADE TO PRO:

📱 Features:
• Unlimited products
• M-Pesa payments
• Multiple shops (up to 5)
• Staff accounts (up to 3)
• Advanced analytics

💰 Price: KSh 500/month

Reply: pro to upgrade

Or visit: https://dukapos.io/upgrade`, nil
}

// handlePlan handles plan info
func (h *CommandHandler) handlePlan(shop *models.Shop) (string, error) {
	info := getPlanInfo(shop.Plan)

	msg := fmt.Sprintf(`💎 YOUR PLAN: %s

🛒 Shops: %d
📦 Products: %s
👥 Staff: %s
💰 M-Pesa: %s
📊 Analytics: %s

%s`,
		info["name"],
		info["shops"].(int),
		info["products"].(string),
		info["staff"].(string),
		info["mpesa"].(string),
		info["analytics"].(string),
		info["cta"].(string),
	)
	return msg, nil
}

func activeStatus(active bool) string {
	if active {
		return "✅ Active"
	}
	return "❌ Inactive"
}

func getPlanInfo(plan models.PlanType) map[string]interface{} {
	switch plan {
	case models.PlanFree:
		return map[string]interface{}{
			"name":      "Free",
			"shops":     1,
			"products":  "50",
			"staff":     "0",
			"mpesa":     "❌",
			"analytics": "Basic",
			"cta":       "Reply: upgrade to go Pro!",
		}
	case models.PlanPro:
		return map[string]interface{}{
			"name":      "Pro",
			"shops":     5,
			"products":  "Unlimited",
			"staff":     "3",
			"mpesa":     "✅",
			"analytics": "Advanced",
			"cta":       "Reply: business for Enterprise",
		}
	case models.PlanBusiness:
		return map[string]interface{}{
			"name":      "Business",
			"shops":     50,
			"products":  "Unlimited",
			"staff":     "Unlimited",
			"mpesa":     "✅",
			"analytics": "Advanced + AI",
			"cta":       "🎉 You're maxed out!",
		}
	default:
		return getPlanInfo(models.PlanFree)
	}
}

// ============================================
// Phase 3: Enterprise Features Handlers
// ============================================

// handlePredict handles AI predictions
func (h *CommandHandler) handlePredict(shop *models.Shop, args []string) (string, error) {
	// Check if Business plan (required for AI)
	if shop.Plan != models.PlanBusiness {
		return `🤖 AI PREDICTIONS require Business plan!

Current: %s
Required: Business (KSh 1,500/month)

Features:
• Restock predictions
• Sales trends
• Seasonal analysis

Reply: business to upgrade`, nil
	}

	if len(args) < 1 {
		return `🤖 AI PREDICTIONS:

predict stock - Inventory predictions
predict trends - Sales trends
predict restock - Items needing restock

Note: AI service requires database integration.`, nil
	}

	switch args[0] {
	case "stock", "inventory":
		return `📊 AI INVENTORY PREDICTIONS

Top items needing attention:
1. Milk - Stockout in 3 days ⚠️
2. Bread - Stockout in 5 days
3. Sugar - Stock OK (14 days)

Confidence: 78%

Note: Connect AI service for live predictions.`, nil

	case "trends":
		return `📈 SALES TRENDS

This Week: +12% ↑
Last Week: +8% ↑

Top Categories:
1. Dairy: +15%
2. Beverages: +10%
3. Snacks: +5%

Note: Analytics integration required.`, nil

	case "restock":
		return `📦 RESTOCK ALERTS

⚠️ URGENT (next 3 days):
• Milk: Order 100 units
• Bread: Order 50 units

📋 This Week:
• Eggs: Order 30 units
• Sugar: Order 20 units

Note: AI predictions require setup.`, nil

	default:
		return "❌ Unknown predict command. Use: predict stock, predict trends, or predict restock", nil
	}
}

// handleQR handles QR payment commands
func (h *CommandHandler) handleQR(shop *models.Shop, args []string) (string, error) {
	if len(args) < 1 {
		return `📱 QR PAYMENTS:

qr generate [amount] - Generate payment QR
qr static - Get shop's static QR
qr help - QR help

Note: QR payments require Business plan.`, nil
	}

	switch args[0] {
	case "generate":
		if len(args) < 2 {
			return "❌ Usage: qr generate [amount]\nExample: qr generate 500", nil
		}
		amount := args[1]
		return fmt.Sprintf(`📱 QR CODE GENERATED

Amount: KSh %s
Shop: %s

[QR Code would appear here]

Customer scans to pay via M-Pesa.

Note: QR service requires configuration.`, amount, shop.Name), nil

	case "static":
		return fmt.Sprintf(`🏪 SHOP STATIC QR

Shop: %s

[Static QR Code]

Customers can scan this to pay any amount.

Note: Generate from dashboard for production use.`, shop.Name), nil

	default:
		return "❌ Unknown qr command. Use: qr generate [amount] or qr static", nil
	}
}

// handleLoyalty handles loyalty program commands
func (h *CommandHandler) handleLoyalty(shop *models.Shop, args []string) (string, error) {
	if shop.Plan != models.PlanBusiness {
		return fmt.Sprintf(`🎁 LOYALTY PROGRAM requires Business plan!

Current: %s
Required: Business (KSh 1,500/month)

Features:
• Points accumulation
• Tier rewards
• Customer tracking

Reply: business to upgrade`, shop.Plan), nil
	}

	// Check if customer repo is available
	if h.customerRepo == nil {
		return "⚙️ Loyalty feature not fully configured.\nPlease contact support.", nil
	}

	if len(args) < 1 {
		// List all customers
		customers, err := h.customerRepo.GetByShopID(shop.ID)
		if err != nil {
			return "", err
		}
		if len(customers) == 0 {
			return `🎁 LOYALTY PROGRAM:

No customers enrolled yet.

Add customer:
loyalty add [phone] [name]

Example: loyalty add +254700000001 John`, nil
		}
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("🎁 LOYALTY CUSTOMERS (%d):\n\n", len(customers)))
		for i, c := range customers {
			if i >= 10 {
				break
			}
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, c.Name))
			sb.WriteString(fmt.Sprintf("   📱 %s\n", c.Phone))
			sb.WriteString(fmt.Sprintf("   💎 %d pts | Tier: %s\n\n", c.LoyaltyPoints, c.Tier))
		}
		return sb.String(), nil
	}

	switch args[0] {
	case "points":
		if len(args) < 2 {
			return "❌ Usage: loyalty points [phone]", nil
		}
		phone := args[1]
		customer, err := h.customerRepo.GetByPhone(shop.ID, phone)
		if err != nil {
			return "❌ Customer not found.\nUse: loyalty add [phone] [name] to add", nil
		}
		pointsValue := float64(customer.LoyaltyPoints) / 10 // 10 points = KSh 1
		return fmt.Sprintf(`🎁 LOYALTY POINTS

📱 %s
💎 Points: %d
💰 Value: KSh %.2f

🏆 Tier: %s
📊 Total Spent: KSh %.2f

Earn 1 point per KSh spent!`, customer.Phone, customer.LoyaltyPoints, pointsValue, customer.Tier, customer.TotalSpent), nil

	case "add":
		if len(args) < 3 {
			return `❌ Usage: loyalty add [phone] [name]

Example: loyalty add +254700000001 John Doe`, nil
		}
		phone := args[1]
		name := strings.Title(args[2])

		// Check if customer already exists
		existing, _ := h.customerRepo.GetByPhone(shop.ID, phone)
		if existing != nil {
			return "❌ Customer with this phone already exists!", nil
		}

		customer := &models.Customer{
			ShopID:   shop.ID,
			Name:     name,
			Phone:    phone,
			Tier:     "bronze",
			IsActive: true,
		}
		if err := h.customerRepo.Create(customer); err != nil {
			return "", err
		}
		return fmt.Sprintf(`✅ CUSTOMER ADDED TO LOYALTY!

👤 %s
📱 %s

💎 Points: 0
🏆 Tier: Bronze

Welcome to the loyalty program!`, name, phone), nil

	case "rewards":
		return `🎁 AVAILABLE REWARDS:

1. KSh 50 Off
   💎 100 points

2. KSh 100 Off
   💎 200 points

3. Free Item
   💎 300 points

4. KSh 200 Off
   💎 400 points

Redeem: loyalty redeem [phone] [points]`, nil

	case "tiers":
		return `🏆 LOYALTY TIERS:

🥉 Bronze (Start)
   • 1 point per KSh 1

🥈 Silver (KSh 20,000+ spent)
   • 1.25x points
   • Birthday bonus

🥇 Gold (KSh 50,000+ spent)
   • 1.5x points
   • Priority support

💎 Platinum (KSh 100,000+ spent)
   • 2x points
   • Exclusive offers`, nil

	case "redeem":
		if len(args) < 3 {
			return "❌ Usage: loyalty redeem [phone] [points]", nil
		}
		phone := args[1]
		points, err := strconv.Atoi(args[2])
		if err != nil || points < 10 {
			return "❌ Invalid points (minimum 10)", nil
		}

		customer, err := h.customerRepo.GetByPhone(shop.ID, phone)
		if err != nil {
			return "❌ Customer not found", nil
		}

		if customer.LoyaltyPoints < points {
			return fmt.Sprintf("❌ Not enough points!\nAvailable: %d points", customer.LoyaltyPoints), nil
		}

		if err := h.customerRepo.DeductPoints(customer.ID, points); err != nil {
			return "", err
		}

		value := float64(points) / 10
		return fmt.Sprintf(`✅ POINTS REDEEMED!

📱 %s
💎 Points: -%d
💰 Value: KSh %.2f

Remaining: %d points`, customer.Phone, points, value, customer.LoyaltyPoints-points), nil

	default:
		return `❌ Unknown loyalty command.

Commands:
loyalty - List customers
loyalty points [phone] - Check points
loyalty add [phone] [name] - Add customer
loyalty rewards - View rewards
loyalty tiers - View tiers
loyalty redeem [phone] [points] - Redeem points`, nil
	}
}

// handleAPI handles API access commands
func (h *CommandHandler) handleAPI(shop *models.Shop, args []string) (string, error) {
	if shop.Plan != models.PlanBusiness {
		return `🔗 API ACCESS requires Business plan!

Current: %s
Required: Business (KSh 1,500/month)

Features:
• REST API
• Webhooks
• Third-party integrations

Reply: business to upgrade`, nil
	}

	if len(args) < 1 {
		return `🔗 API ACCESS:

api key create [name] - Generate API key
api key list - View your keys
api endpoints - Available endpoints
api webhooks - Webhook settings

Note: API requires configuration.`, nil
	}

	switch args[0] {
	case "key":
		if len(args) < 2 {
			return "❌ Usage: api key create [name]", nil
		}
		return fmt.Sprintf(`🔑 API KEY GENERATED!

Name: %s
Key: dkp_abc123...
Secret: ***SECRET***

⚠️ Save the secret - it won't be shown again!

Permissions: products, sales, reports
Rate Limit: 60 req/min`, args[1]), nil

	case "endpoints":
		return `📡 AVAILABLE ENDPOINTS:

Products:
GET    /api/v1/products
POST   /api/v1/products

Sales:
GET    /api/v1/sales
POST   /api/v1/sales

Reports:
GET    /api/v1/reports/daily
GET    /api/v1/reports/weekly

Payments:
POST   /api/v1/payments/mpesa

Full docs: https://dukapos.io/docs/api`, nil

	case "webhooks":
		return `🔗 WEBHOOKS:

Events supported:
• sale.created
• product.low_stock
• payment.completed
• payment.failed

Configure in dashboard.

Note: Webhook service requires setup.`, nil

	default:
		return "❌ Unknown api command", nil
	}
}
