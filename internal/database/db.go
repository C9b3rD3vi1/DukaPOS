package database

import (
	"fmt"
	"log"
	"time"

	"github.com/C9b3rD3vi1/DukaPOS/internal/config"
	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func GetDB() *gorm.DB {
	return DB
}

func GetTestDB() *gorm.DB {
	return DB
}

func Connect(cfg *config.Config) error {
	var err error
	logLevel := logger.Silent
	if cfg.Debug {
		logLevel = logger.Info
	}

	var dialector gorm.Dialector

	switch cfg.DBType {
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
			cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode)
		dialector = postgres.Open(dsn)
		log.Printf("📦 Connecting to PostgreSQL: %s:%d/%s", cfg.DBHost, cfg.DBPort, cfg.DBName)
	default:
		dsn := cfg.DBPath + "?_foreign_keys=on"
		dialector = sqlite.Open(dsn)
		log.Printf("📦 Connecting to SQLite: %s", cfg.DBPath)
	}

	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})

	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConnections)
	sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConnections)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("✅ Database connected successfully")
	return nil
}

func Migrate() error {
	log.Println("🔄 Running database migrations...")

	migrator := DB.Migrator()

	modelsToMigrate := []interface{}{
		&models.Account{},
		&models.Shop{},
		&models.Product{},
		&models.Sale{},
		&models.DailySummary{},
		&models.Staff{},
		&models.Customer{},
		&models.Supplier{},
		&models.Order{},
		&models.OrderItem{},
		&models.AuditLog{},
		&models.Webhook{},
		&models.APIKey{},
		&models.LoyaltyTransaction{},
	}

	for _, model := range modelsToMigrate {
		if !migrator.HasTable(model) {
			if err := DB.AutoMigrate(model); err != nil {
				log.Printf("⚠️ Failed to migrate %v: %v", model, err)
				continue
			}
		} else {
			if err := DB.AutoMigrate(model); err != nil {
				log.Printf("⚠️ Failed to migrate %v: %v", model, err)
			}
		}
	}

	log.Println("✅ Database migrations completed")
	return nil
}

func Seed() error {
	log.Println("🌱 Checking for seed data...")

	var count int64
	DB.Model(&models.Account{}).Count(&count)

	// Upgrade all existing accounts to Business plan for testing
	if count > 0 {
		log.Println("⬆️  Upgrading existing accounts to Business plan for testing...")
		DB.Model(&models.Account{}).Update("plan", models.PlanBusiness)
		DB.Model(&models.Shop{}).Update("plan", models.PlanBusiness)
		log.Println("✅ All accounts upgraded to Business plan")
	}

	// Always create an admin account
	adminEmail := "admin@dukapos.com"
	var adminCount int64
	DB.Model(&models.Account{}).Where("email = ?", adminEmail).Count(&adminCount)

	if adminCount == 0 {
		admin := models.Account{
			Email:        adminEmail,
			PasswordHash: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			Name:         "Admin User",
			Phone:        "+254700000000",
			IsActive:     true,
			IsVerified:   true,
			Plan:         models.PlanBusiness,
			IsAdmin:      true,
		}
		if err := DB.Create(&admin).Error; err != nil {
			log.Printf("Failed to create admin account: %v", err)
		} else {
			log.Println("✅ Admin account created: admin@dukapos.com / password")
		}
	}

	if count > 0 {
		log.Println("⏭️  Skipping seed - data already exists")
		return nil
	}

	account := models.Account{
		Email:        "test@dukapos.com",
		PasswordHash: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
		Name:         "John Doe",
		Phone:        "+254700000001",
		IsActive:     true,
		IsVerified:   true,
		Plan:         models.PlanBusiness,
	}

	if err := DB.Create(&account).Error; err != nil {
		return fmt.Errorf("failed to seed account: %w", err)
	}

	shop := models.Shop{
		AccountID: account.ID,
		Name:      "Test Duka",
		Phone:     "+254700000001",
		OwnerName: "John Doe",
		Plan:      models.PlanBusiness,
		IsActive:  true,
		Email:     "test@dukapos.com",
	}

	if err := DB.Create(&shop).Error; err != nil {
		return fmt.Errorf("failed to seed shop: %w", err)
	}

	products := []models.Product{
		{ShopID: shop.ID, Name: "Milk", Category: "Drinks", Unit: "pcs", CostPrice: 45, SellingPrice: 60, CurrentStock: 50, LowStockThreshold: 10},
		{ShopID: shop.ID, Name: "Bread", Category: "Bakery", Unit: "loaf", CostPrice: 35, SellingPrice: 50, CurrentStock: 30, LowStockThreshold: 5},
		{ShopID: shop.ID, Name: "Eggs", Category: "Dairy", Unit: "tray", CostPrice: 200, SellingPrice: 250, CurrentStock: 20, LowStockThreshold: 5},
		{ShopID: shop.ID, Name: "Soda", Category: "Drinks", Unit: "bottle", CostPrice: 35, SellingPrice: 50, CurrentStock: 100, LowStockThreshold: 20},
		{ShopID: shop.ID, Name: "Water", Category: "Drinks", Unit: "bottle", CostPrice: 15, SellingPrice: 25, CurrentStock: 200, LowStockThreshold: 30},
		{ShopID: shop.ID, Name: "Mandazi", Category: "Bakery", Unit: "pcs", CostPrice: 10, SellingPrice: 15, CurrentStock: 50, LowStockThreshold: 10},
		{ShopID: shop.ID, Name: "Sugar", Category: "Groceries", Unit: "kg", CostPrice: 100, SellingPrice: 130, CurrentStock: 25, LowStockThreshold: 5},
		{ShopID: shop.ID, Name: "Rice", Category: "Groceries", Unit: "kg", CostPrice: 120, SellingPrice: 150, CurrentStock: 30, LowStockThreshold: 10},
	}

	for _, p := range products {
		if err := DB.Create(&p).Error; err != nil {
			log.Printf("Failed to seed product %s: %v", p.Name, err)
		}
	}

	now := time.Now()
	for i := 0; i < 10; i++ {
		sale := models.Sale{
			ShopID:        shop.ID,
			ProductID:     1,
			Quantity:      2,
			UnitPrice:     60,
			TotalAmount:   120,
			CostAmount:    90,
			Profit:        30,
			PaymentMethod: models.PaymentCash,
			CreatedAt:     now.Add(-time.Duration(i) * time.Hour),
		}
		if err := DB.Create(&sale).Error; err != nil {
			log.Printf("Failed to seed sale: %v", err)
		}
	}

	log.Println("✅ Seed data created successfully")
	log.Println("📱 Test shop phone: +254700000001")
	log.Println("🔑 Test password: password")
	return nil
}

func Close() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
