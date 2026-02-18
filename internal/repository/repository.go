package repository

import (
	"time"

	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"gorm.io/gorm"
)

// ShopRepository handles shop database operations
type ShopRepository struct {
	db *gorm.DB
}

// NewShopRepository creates a new shop repository
func NewShopRepository(db *gorm.DB) *ShopRepository {
	return &ShopRepository{db: db}
}

// Create creates a new shop
func (r *ShopRepository) Create(shop *models.Shop) error {
	return r.db.Create(shop).Error
}

// GetByID gets a shop by ID
func (r *ShopRepository) GetByID(id uint) (*models.Shop, error) {
	var shop models.Shop
	err := r.db.First(&shop, id).Error
	if err != nil {
		return nil, err
	}
	return &shop, nil
}

// GetByPhone gets a shop by phone number
func (r *ShopRepository) GetByPhone(phone string) (*models.Shop, error) {
	var shop models.Shop
	err := r.db.Where("phone = ?", phone).First(&shop).Error
	if err != nil {
		return nil, err
	}
	return &shop, nil
}

// GetByEmail gets a shop by email
func (r *ShopRepository) GetByEmail(email string) (*models.Shop, error) {
	var shop models.Shop
	err := r.db.Where("email = ?", email).First(&shop).Error
	if err != nil {
		return nil, err
	}
	return &shop, nil
}

// Update updates a shop
func (r *ShopRepository) Update(shop *models.Shop) error {
	return r.db.Save(shop).Error
}

// Delete soft deletes a shop
func (r *ShopRepository) Delete(id uint) error {
	return r.db.Delete(&models.Shop{}, id).Error
}

// List lists all shops with pagination
func (r *ShopRepository) List(limit, offset int) ([]models.Shop, int64, error) {
	var shops []models.Shop
	var total int64

	r.db.Model(&models.Shop{}).Count(&total)
	err := r.db.Limit(limit).Offset(offset).Find(&shops).Error

	return shops, total, err
}

// ProductRepository handles product database operations
type ProductRepository struct {
	db *gorm.DB
}

// NewProductRepository creates a new product repository
func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// Create creates a new product
func (r *ProductRepository) Create(product *models.Product) error {
	return r.db.Create(product).Error
}

// GetByID gets a product by ID
func (r *ProductRepository) GetByID(id uint) (*models.Product, error) {
	var product models.Product
	err := r.db.First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// GetByShopAndName gets a product by shop ID and name
func (r *ProductRepository) GetByShopAndName(shopID uint, name string) (*models.Product, error) {
	var product models.Product
	err := r.db.Where("shop_id = ? AND name = ? AND is_active = ?", shopID, name, true).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// GetByShopID gets all products for a shop
func (r *ProductRepository) GetByShopID(shopID uint) ([]models.Product, error) {
	var products []models.Product
	err := r.db.Where("shop_id = ? AND is_active = ?", shopID, true).
		Order("name ASC").
		Find(&products).Error
	return products, err
}

// GetLowStock gets products below threshold
func (r *ProductRepository) GetLowStock(shopID uint) ([]models.Product, error) {
	var products []models.Product
	err := r.db.Where("shop_id = ? AND is_active = ? AND current_stock <= low_stock_threshold", shopID, true).
		Find(&products).Error
	return products, err
}

// GetByCategory gets products by category
func (r *ProductRepository) GetByCategory(shopID uint, category string) ([]models.Product, error) {
	var products []models.Product
	err := r.db.Where("shop_id = ? AND category = ? AND is_active = ?", shopID, category, true).
		Order("name ASC").
		Find(&products).Error
	return products, err
}

// GetCategories gets all unique categories for a shop
func (r *ProductRepository) GetCategories(shopID uint) ([]string, error) {
	var categories []string
	err := r.db.Model(&models.Product{}).
		Where("shop_id = ? AND category != '' AND is_active = ?", shopID, true).
		Distinct("category").
		Pluck("category", &categories).Error
	return categories, err
}

// GetByBarcode gets a product by barcode
func (r *ProductRepository) GetByBarcode(shopID uint, barcode string) (*models.Product, error) {
	var product models.Product
	err := r.db.Where("shop_id = ? AND barcode = ? AND is_active = ?", shopID, barcode, true).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// UpdateThreshold updates product low stock threshold
func (r *ProductRepository) UpdateThreshold(id uint, threshold int) error {
	return r.db.Model(&models.Product{}).Where("id = ?", id).Update("low_stock_threshold", threshold).Error
}

// Update updates a product
func (r *ProductRepository) Update(product *models.Product) error {
	return r.db.Save(product).Error
}

// Delete soft deletes a product
func (r *ProductRepository) Delete(id uint) error {
	return r.db.Delete(&models.Product{}, id).Error
}

// UpdateStock updates product stock
func (r *ProductRepository) UpdateStock(id uint, quantity int) error {
	return r.db.Model(&models.Product{}).
		Where("id = ?", id).
		Update("current_stock", gorm.Expr("current_stock + ?", quantity)).Error
}

// SaleRepository handles sale database operations
type SaleRepository struct {
	db *gorm.DB
}

// NewSaleRepository creates a new sale repository
func NewSaleRepository(db *gorm.DB) *SaleRepository {
	return &SaleRepository{db: db}
}

// Create creates a new sale
func (r *SaleRepository) Create(sale *models.Sale) error {
	return r.db.Create(sale).Error
}

// GetByID gets a sale by ID
func (r *SaleRepository) GetByID(id uint) (*models.Sale, error) {
	var sale models.Sale
	err := r.db.First(&sale, id).Error
	if err != nil {
		return nil, err
	}
	return &sale, nil
}

// GetByShopID gets all sales for a shop
func (r *SaleRepository) GetByShopID(shopID uint, limit int) ([]models.Sale, error) {
	var sales []models.Sale
	query := r.db.Where("shop_id = ?", shopID).Preload("Product")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Order("created_at DESC").Find(&sales).Error
	return sales, err
}

// GetByProductAndDateRange gets sales for a specific product within a date range
func (r *SaleRepository) GetByProductAndDateRange(productID, shopID uint, start, end time.Time) ([]models.Sale, error) {
	var sales []models.Sale
	err := r.db.Where("product_id = ? AND shop_id = ? AND created_at BETWEEN ? AND ?", productID, shopID, start, end).
		Order("created_at DESC").
		Find(&sales).Error
	return sales, err
}

// GetByDateRange gets sales within a date range
func (r *SaleRepository) GetByDateRange(shopID uint, start, end time.Time) ([]models.Sale, error) {
	var sales []models.Sale
	err := r.db.Where("shop_id = ? AND created_at BETWEEN ? AND ?", shopID, start, end).
		Preload("Product").
		Order("created_at DESC").
		Find(&sales).Error
	return sales, err
}

// GetTodaySales gets today's sales for a shop
func (r *SaleRepository) GetTodaySales(shopID uint) ([]models.Sale, error) {
	startOfDay := time.Now().Truncate(24 * time.Hour)
	endOfDay := startOfDay.Add(24 * time.Hour)
	return r.GetByDateRange(shopID, startOfDay, endOfDay)
}

// GetTotalSales gets total sales amount for a shop
func (r *SaleRepository) GetTotalSales(shopID uint, start, end time.Time) (float64, int, error) {
	var result struct {
		Total float64
		Count int
	}
	err := r.db.Model(&models.Sale{}).
		Select("COALESCE(SUM(total_amount), 0) as total, COUNT(*) as count").
		Where("shop_id = ? AND created_at BETWEEN ? AND ?", shopID, start, end).
		Scan(&result).Error
	return result.Total, result.Count, err
}

// DailySummaryRepository handles daily summary database operations
type DailySummaryRepository struct {
	db *gorm.DB
}

// NewDailySummaryRepository creates a new daily summary repository
func NewDailySummaryRepository(db *gorm.DB) *DailySummaryRepository {
	return &DailySummaryRepository{db: db}
}

// GetOrCreate gets or creates a daily summary
func (r *DailySummaryRepository) GetOrCreate(shopID uint, date time.Time) (*models.DailySummary, error) {
	date = date.Truncate(24 * time.Hour)
	var summary models.DailySummary
	err := r.db.Where("shop_id = ? AND date = ?", shopID, date).First(&summary).Error
	if err == gorm.ErrRecordNotFound {
		summary = models.DailySummary{
			ShopID: shopID,
			Date:   date,
		}
		err = r.db.Create(&summary).Error
	}
	return &summary, err
}

// Update updates a daily summary
func (r *DailySummaryRepository) Update(summary *models.DailySummary) error {
	return r.db.Save(summary).Error
}

// Recalculate recalculates daily summary from sales
func (r *DailySummaryRepository) Recalculate(shopID uint, date time.Time) error {
	date = date.Truncate(24 * time.Hour)
	start := date
	end := date.Add(24 * time.Hour)

	var result struct {
		TotalSales        float64
		TotalTransactions int
		TotalCost         float64
		TotalProfit       float64
	}

	err := r.db.Model(&models.Sale{}).
		Select(
			"COALESCE(SUM(total_amount), 0) as total_sales",
			"COUNT(*) as total_transactions",
			"COALESCE(SUM(cost_amount), 0) as total_cost",
			"COALESCE(SUM(profit), 0) as total_profit",
		).
		Where("shop_id = ? AND created_at BETWEEN ? AND ?", shopID, start, end).
		Scan(&result).Error

	if err != nil {
		return err
	}

	summary, err := r.GetOrCreate(shopID, date)
	if err != nil {
		return err
	}

	summary.TotalSales = result.TotalSales
	summary.TotalTransactions = result.TotalTransactions
	summary.TotalCost = result.TotalCost
	summary.TotalProfit = result.TotalProfit

	return r.db.Save(summary).Error
}

// GetByDateRange gets daily summaries within a date range
func (r *DailySummaryRepository) GetByDateRange(shopID uint, start, end time.Time) ([]models.DailySummary, error) {
	var summaries []models.DailySummary
	err := r.db.Where("shop_id = ? AND date BETWEEN ? AND ?", shopID, start, end).
		Order("date DESC").
		Find(&summaries).Error
	return summaries, err
}

// AuditLogRepository handles audit log database operations
type AuditLogRepository struct {
	db *gorm.DB
}

// NewAuditLogRepository creates a new audit log repository
func NewAuditLogRepository(db *gorm.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// Create creates a new audit log
func (r *AuditLogRepository) Create(log *models.AuditLog) error {
	return r.db.Create(log).Error
}

// GetByShopID gets audit logs for a shop
func (r *AuditLogRepository) GetByShopID(shopID uint, limit int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := r.db.Where("shop_id = ?", shopID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// StaffRepository handles staff database operations
type StaffRepository struct {
	db *gorm.DB
}

// NewStaffRepository creates a new staff repository
func NewStaffRepository(db *gorm.DB) *StaffRepository {
	return &StaffRepository{db: db}
}

// Create creates a new staff member
func (r *StaffRepository) Create(staff *models.Staff) error {
	return r.db.Create(staff).Error
}

// GetByID gets a staff member by ID
func (r *StaffRepository) GetByID(id uint) (*models.Staff, error) {
	var staff models.Staff
	err := r.db.First(&staff, id).Error
	if err != nil {
		return nil, err
	}
	return &staff, nil
}

// GetByPhone gets a staff member by phone number
func (r *StaffRepository) GetByPhone(shopID uint, phone string) (*models.Staff, error) {
	var staff models.Staff
	err := r.db.Where("shop_id = ? AND phone = ?", shopID, phone).First(&staff).Error
	if err != nil {
		return nil, err
	}
	return &staff, nil
}

// GetByShopID gets all staff for a shop
func (r *StaffRepository) GetByShopID(shopID uint) ([]models.Staff, error) {
	var staff []models.Staff
	err := r.db.Where("shop_id = ?", shopID).
		Order("name ASC").
		Find(&staff).Error
	return staff, err
}

// Update updates a staff member
func (r *StaffRepository) Update(staff *models.Staff) error {
	return r.db.Save(staff).Error
}

// Delete soft deletes a staff member
func (r *StaffRepository) Delete(id uint) error {
	return r.db.Delete(&models.Staff{}, id).Error
}

// ============================================
// Account Repository - Multiple Shops Support
// ============================================

// AccountRepository handles account database operations
type AccountRepository struct {
	db *gorm.DB
}

// NewAccountRepository creates a new account repository
func NewAccountRepository(db *gorm.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

// Create creates a new account
func (r *AccountRepository) Create(account *models.Account) error {
	return r.db.Create(account).Error
}

// GetByID gets an account by ID
func (r *AccountRepository) GetByID(id uint) (*models.Account, error) {
	var account models.Account
	err := r.db.First(&account, id).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// GetByEmail gets an account by email
func (r *AccountRepository) GetByEmail(email string) (*models.Account, error) {
	var account models.Account
	err := r.db.Where("email = ?", email).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// GetByPhone gets an account by phone
func (r *AccountRepository) GetByPhone(phone string) (*models.Account, error) {
	var account models.Account
	err := r.db.Where("phone = ?", phone).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// Update updates an account
func (r *AccountRepository) Update(account *models.Account) error {
	return r.db.Save(account).Error
}

// GetShops gets all shops for an account
func (r *AccountRepository) GetShops(accountID uint) ([]models.Shop, error) {
	var shops []models.Shop
	err := r.db.Where("account_id = ?", accountID).
		Order("created_at DESC").
		Find(&shops).Error
	return shops, err
}

// CountShops counts shops for an account
func (r *AccountRepository) CountShops(accountID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Shop{}).
		Where("account_id = ?", accountID).
		Count(&count).Error
	return count, err
}

// Delete soft deletes an account
func (r *AccountRepository) Delete(id uint) error {
	return r.db.Delete(&models.Account{}, id).Error
}

// ShopRepositoryWithAccount handles shop operations with account support
type ShopRepositoryWithAccount struct {
	db *gorm.DB
}

// NewShopRepositoryWithAccount creates a new shop repository with account support
func NewShopRepositoryWithAccount(db *gorm.DB) *ShopRepositoryWithAccount {
	return &ShopRepositoryWithAccount{db: db}
}

// GetByAccountAndID gets a shop by account ID and shop ID
func (r *ShopRepositoryWithAccount) GetByAccountAndID(accountID, shopID uint) (*models.Shop, error) {
	var shop models.Shop
	err := r.db.Where("id = ? AND account_id = ?", shopID, accountID).First(&shop).Error
	if err != nil {
		return nil, err
	}
	return &shop, nil
}

// GetByAccountID gets all shops for an account
func (r *ShopRepositoryWithAccount) GetByAccountID(accountID uint) ([]models.Shop, error) {
	var shops []models.Shop
	err := r.db.Where("account_id = ?", accountID).
		Order("name ASC").
		Find(&shops).Error
	return shops, err
}

// Create creates a new shop for an account
func (r *ShopRepositoryWithAccount) Create(shop *models.Shop) error {
	return r.db.Create(shop).Error
}

// Update updates a shop
func (r *ShopRepositoryWithAccount) Update(shop *models.Shop) error {
	return r.db.Save(shop).Error
}

// Delete soft deletes a shop
func (r *ShopRepositoryWithAccount) Delete(id uint) error {
	return r.db.Delete(&models.Shop{}, id).Error
}

// GetByID gets a shop by ID (legacy compatibility)
func (r *ShopRepositoryWithAccount) GetByID(id uint) (*models.Shop, error) {
	var shop models.Shop
	err := r.db.Preload("Account").First(&shop, id).Error
	if err != nil {
		return nil, err
	}
	return &shop, nil
}

// GetByPhone gets a shop by phone
func (r *ShopRepositoryWithAccount) GetByPhone(phone string) (*models.Shop, error) {
	var shop models.Shop
	err := r.db.Where("phone = ?", phone).First(&shop).Error
	if err != nil {
		return nil, err
	}
	return &shop, nil
}

// ============================================
// Webhook Repository
// ============================================

// WebhookRepository handles webhook database operations
type WebhookRepository struct {
	db *gorm.DB
}

// NewWebhookRepository creates a new webhook repository
func NewWebhookRepository(db *gorm.DB) *WebhookRepository {
	return &WebhookRepository{db: db}
}

// Create creates a new webhook
func (r *WebhookRepository) Create(webhook *models.Webhook) error {
	return r.db.Create(webhook).Error
}

// GetByID gets a webhook by ID
func (r *WebhookRepository) GetByID(id uint) (*models.Webhook, error) {
	var webhook models.Webhook
	err := r.db.First(&webhook, id).Error
	if err != nil {
		return nil, err
	}
	return &webhook, nil
}

// GetByShopID gets all webhooks for a shop
func (r *WebhookRepository) GetByShopID(shopID uint) ([]models.Webhook, error) {
	var webhooks []models.Webhook
	err := r.db.Where("shop_id = ?", shopID).Find(&webhooks).Error
	return webhooks, err
}

// GetActive gets active webhooks for a shop and event
func (r *WebhookRepository) GetActive(shopID uint, event string) ([]models.Webhook, error) {
	var webhooks []models.Webhook
	err := r.db.Where("shop_id = ? AND is_active = ? AND events LIKE ?", shopID, true, "%"+event+"%").Find(&webhooks).Error
	return webhooks, err
}

// Update updates a webhook
func (r *WebhookRepository) Update(webhook *models.Webhook) error {
	return r.db.Save(webhook).Error
}

// Delete soft deletes a webhook
func (r *WebhookRepository) Delete(id uint) error {
	return r.db.Delete(&models.Webhook{}, id).Error
}

// ============================================
// API Key Repository
// ============================================

// APIKeyRepository handles API key database operations
type APIKeyRepository struct {
	db *gorm.DB
}

// NewAPIKeyRepository creates a new API key repository
func NewAPIKeyRepository(db *gorm.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

// Create creates a new API key
func (r *APIKeyRepository) Create(apiKey *models.APIKey) error {
	return r.db.Create(apiKey).Error
}

// GetByID gets an API key by ID
func (r *APIKeyRepository) GetByID(id uint) (*models.APIKey, error) {
	var apiKey models.APIKey
	err := r.db.First(&apiKey, id).Error
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

// GetByKey gets an API key by key string
func (r *APIKeyRepository) GetByKey(key string) (*models.APIKey, error) {
	var apiKey models.APIKey
	err := r.db.Where("key = ? AND is_active = ?", key, true).First(&apiKey).Error
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

// GetByShopID gets all API keys for a shop
func (r *APIKeyRepository) GetByShopID(shopID uint) ([]models.APIKey, error) {
	var apiKeys []models.APIKey
	err := r.db.Where("shop_id = ?", shopID).Find(&apiKeys).Error
	return apiKeys, err
}

// Update updates an API key
func (r *APIKeyRepository) Update(apiKey *models.APIKey) error {
	return r.db.Save(apiKey).Error
}

// UpdateLastUsed updates the last used timestamp
func (r *APIKeyRepository) UpdateLastUsed(id uint) error {
	return r.db.Model(&models.APIKey{}).Where("id = ?", id).Update("last_used_at", time.Now()).Error
}

// Delete soft deletes an API key
func (r *APIKeyRepository) Delete(id uint) error {
	return r.db.Delete(&models.APIKey{}, id).Error
}

// ============================================
// Customer Repository
// ============================================

// CustomerRepository handles customer database operations
type CustomerRepository struct {
	db *gorm.DB
}

// NewCustomerRepository creates a new customer repository
func NewCustomerRepository(db *gorm.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

// Create creates a new customer
func (r *CustomerRepository) Create(customer *models.Customer) error {
	return r.db.Create(customer).Error
}

// GetByID gets a customer by ID
func (r *CustomerRepository) GetByID(id uint) (*models.Customer, error) {
	var customer models.Customer
	err := r.db.First(&customer, id).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

// GetByPhone gets a customer by phone
func (r *CustomerRepository) GetByPhone(shopID uint, phone string) (*models.Customer, error) {
	var customer models.Customer
	err := r.db.Where("shop_id = ? AND phone = ?", shopID, phone).First(&customer).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

// GetByShopID gets all customers for a shop
func (r *CustomerRepository) GetByShopID(shopID uint) ([]models.Customer, error) {
	var customers []models.Customer
	err := r.db.Where("shop_id = ?", shopID).Order("created_at DESC").Find(&customers).Error
	return customers, err
}

// GetByTier gets customers by tier
func (r *CustomerRepository) GetByTier(shopID uint, tier string) ([]models.Customer, error) {
	var customers []models.Customer
	err := r.db.Where("shop_id = ? AND tier = ?", shopID, tier).Find(&customers).Error
	return customers, err
}

// Update updates a customer
func (r *CustomerRepository) Update(customer *models.Customer) error {
	return r.db.Save(customer).Error
}

// AddPoints adds loyalty points to a customer
func (r *CustomerRepository) AddPoints(id uint, points int) error {
	return r.db.Model(&models.Customer{}).Where("id = ?", id).
		UpdateColumn("loyalty_points", gorm.Expr("loyalty_points + ?", points)).Error
}

// DeductPoints deducts loyalty points from a customer
func (r *CustomerRepository) DeductPoints(id uint, points int) error {
	return r.db.Model(&models.Customer{}).Where("id = ? AND loyalty_points >= ?", id, points).
		UpdateColumn("loyalty_points", gorm.Expr("loyalty_points - ?", points)).Error
}

// UpdateTier updates customer tier based on total spent
func (r *CustomerRepository) UpdateTier(id uint) error {
	var customer models.Customer
	if err := r.db.First(&customer, id).Error; err != nil {
		return err
	}

	var newTier string
	switch {
	case customer.TotalSpent >= 100000:
		newTier = "platinum"
	case customer.TotalSpent >= 50000:
		newTier = "gold"
	case customer.TotalSpent >= 20000:
		newTier = "silver"
	default:
		newTier = "bronze"
	}

	return r.db.Model(&customer).Update("tier", newTier).Error
}

// Delete soft deletes a customer
func (r *CustomerRepository) Delete(id uint) error {
	return r.db.Delete(&models.Customer{}, id).Error
}

// ============================================
// Loyalty Transaction Repository
// ============================================

// LoyaltyTransactionRepository handles loyalty transaction database operations
type LoyaltyTransactionRepository struct {
	db *gorm.DB
}

// NewLoyaltyTransactionRepository creates a new loyalty transaction repository
func NewLoyaltyTransactionRepository(db *gorm.DB) *LoyaltyTransactionRepository {
	return &LoyaltyTransactionRepository{db: db}
}

// Create creates a new loyalty transaction
func (r *LoyaltyTransactionRepository) Create(tx *models.LoyaltyTransaction) error {
	return r.db.Create(tx).Error
}

// GetByCustomerID gets all transactions for a customer
func (r *LoyaltyTransactionRepository) GetByCustomerID(customerID uint) ([]models.LoyaltyTransaction, error) {
	var transactions []models.LoyaltyTransaction
	err := r.db.Where("customer_id = ?", customerID).Order("created_at DESC").Find(&transactions).Error
	return transactions, err
}

// GetByShopID gets all transactions for a shop
func (r *LoyaltyTransactionRepository) GetByShopID(shopID uint) ([]models.LoyaltyTransaction, error) {
	var transactions []models.LoyaltyTransaction
	err := r.db.Joins("JOIN customers ON customers.id = loyalty_transactions.customer_id").
		Where("customers.shop_id = ?", shopID).
		Order("loyalty_transactions.created_at DESC").
		Find(&transactions).Error
	return transactions, err
}

// GetByDateRange gets transactions within a date range
func (r *LoyaltyTransactionRepository) GetByDateRange(shopID uint, start, end time.Time) ([]models.LoyaltyTransaction, error) {
	var transactions []models.LoyaltyTransaction
	err := r.db.Joins("JOIN customers ON customers.id = loyalty_transactions.customer_id").
		Where("customers.shop_id = ? AND loyalty_transactions.created_at BETWEEN ? AND ?", shopID, start, end).
		Order("loyalty_transactions.created_at DESC").
		Find(&transactions).Error
	return transactions, err
}

// ============================================
// Supplier Repository
// ============================================

// SupplierRepository handles supplier database operations
type SupplierRepository struct {
	db *gorm.DB
}

// NewSupplierRepository creates a new supplier repository
func NewSupplierRepository(db *gorm.DB) *SupplierRepository {
	return &SupplierRepository{db: db}
}

// Create creates a new supplier
func (r *SupplierRepository) Create(supplier *models.Supplier) error {
	return r.db.Create(supplier).Error
}

// GetByID gets a supplier by ID
func (r *SupplierRepository) GetByID(id uint) (*models.Supplier, error) {
	var supplier models.Supplier
	err := r.db.First(&supplier, id).Error
	if err != nil {
		return nil, err
	}
	return &supplier, nil
}

// GetByShopID gets all suppliers for a shop
func (r *SupplierRepository) GetByShopID(shopID uint) ([]models.Supplier, error) {
	var suppliers []models.Supplier
	err := r.db.Where("shop_id = ?", shopID).Order("name ASC").Find(&suppliers).Error
	return suppliers, err
}

// GetByName gets a supplier by name
func (r *SupplierRepository) GetByName(shopID uint, name string) (*models.Supplier, error) {
	var supplier models.Supplier
	err := r.db.Where("shop_id = ? AND name ILIKE ?", shopID, "%"+name+"%").First(&supplier).Error
	if err != nil {
		return nil, err
	}
	return &supplier, nil
}

// Update updates a supplier
func (r *SupplierRepository) Update(supplier *models.Supplier) error {
	return r.db.Save(supplier).Error
}

// Delete soft deletes a supplier
func (r *SupplierRepository) Delete(id uint) error {
	return r.db.Delete(&models.Supplier{}, id).Error
}

// ============================================
// Order Repository
// ============================================

// OrderRepository handles order database operations
type OrderRepository struct {
	db *gorm.DB
}

// NewOrderRepository creates a new order repository
func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// Create creates a new order
func (r *OrderRepository) Create(order *models.Order) error {
	return r.db.Create(order).Error
}

// GetByID gets an order by ID
func (r *OrderRepository) GetByID(id uint) (*models.Order, error) {
	var order models.Order
	err := r.db.Preload("Supplier").Preload("Items").First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// GetByShopID gets all orders for a shop
func (r *OrderRepository) GetByShopID(shopID uint) ([]models.Order, error) {
	var orders []models.Order
	err := r.db.Preload("Supplier").Where("shop_id = ?", shopID).Order("created_at DESC").Find(&orders).Error
	return orders, err
}

// GetByStatus gets orders by status
func (r *OrderRepository) GetByStatus(shopID uint, status string) ([]models.Order, error) {
	var orders []models.Order
	err := r.db.Preload("Supplier").Where("shop_id = ? AND status = ?", shopID, status).Order("created_at DESC").Find(&orders).Error
	return orders, err
}

// Update updates an order
func (r *OrderRepository) Update(order *models.Order) error {
	return r.db.Save(order).Error
}

// Delete soft deletes an order
func (r *OrderRepository) Delete(id uint) error {
	return r.db.Delete(&models.Order{}, id).Error
}

// CreateItem creates an order item
func (r *OrderRepository) CreateItem(item *models.OrderItem) error {
	return r.db.Create(item).Error
}

// GetItems gets all items for an order
func (r *OrderRepository) GetItems(orderID uint) ([]models.OrderItem, error) {
	var items []models.OrderItem
	err := r.db.Where("order_id = ?", orderID).Find(&items).Error
	return items, err
}

// DeleteItems deletes all items for an order
func (r *OrderRepository) DeleteItems(orderID uint) error {
	return r.db.Where("order_id = ?", orderID).Delete(&models.OrderItem{}).Error
}
