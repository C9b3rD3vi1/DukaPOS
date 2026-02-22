package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/C9b3rD3vi1/DukaPOS/internal/config"
	"github.com/C9b3rD3vi1/DukaPOS/internal/database"
	"github.com/C9b3rD3vi1/DukaPOS/internal/handlers"
	aihandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/ai"
	apihandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/api"
	billinghandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/billing"
	currhandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/currency"
	docshandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/docs"
	emailhandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/email"
	exporthandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/export"
	loyaltyhandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/loyalty"
	mpesahandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/mpesa"
	printerhandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/printer"
	qrhandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/qr"
	smshandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/sms"
	staffhandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/staff"
	supplierhandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/supplier"
	ussdhandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/ussd"
	webhookhandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/webhook"
	"github.com/C9b3rD3vi1/DukaPOS/internal/middleware"
	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"github.com/C9b3rD3vi1/DukaPOS/internal/repository"
	"github.com/C9b3rD3vi1/DukaPOS/internal/services"
	ai "github.com/C9b3rD3vi1/DukaPOS/internal/services/ai"
	apiservice "github.com/C9b3rD3vi1/DukaPOS/internal/services/api"
	cacheservice "github.com/C9b3rD3vi1/DukaPOS/internal/services/cache"
	email "github.com/C9b3rD3vi1/DukaPOS/internal/services/email"
	encryption "github.com/C9b3rD3vi1/DukaPOS/internal/services/encryption"
	mpesaservice "github.com/C9b3rD3vi1/DukaPOS/internal/services/mpesa"
	printerservice "github.com/C9b3rD3vi1/DukaPOS/internal/services/printer"
	qrservice "github.com/C9b3rD3vi1/DukaPOS/internal/services/qr"
	scheduler "github.com/C9b3rD3vi1/DukaPOS/internal/services/scheduler"
	smsservice "github.com/C9b3rD3vi1/DukaPOS/internal/services/sms"
	ussdservice "github.com/C9b3rD3vi1/DukaPOS/internal/services/ussd"
	webhookservice "github.com/C9b3rD3vi1/DukaPOS/internal/services/webhook"
	websocket "github.com/C9b3rD3vi1/DukaPOS/internal/services/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize encryption service for sensitive data (AES-256-GCM)
	var encryptSvc *encryption.EncryptionService
	_ = encryptSvc // Used for encrypting sensitive data throughout the app
	if cfg.EncryptionKey != "" {
		encryptSvc, err = encryption.NewEncryptionServiceWithKey([]byte(cfg.EncryptionKey))
		if err != nil {
			log.Printf("Warning: Failed to initialize encryption service: %v", err)
		} else {
			log.Println("✅ Encryption service initialized (AES-256-GCM)")
		}
	} else {
		log.Println("⚠️ Encryption key not set - sensitive data will not be encrypted")
	}

	// Connect to database
	if err := database.Connect(cfg); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	if err := database.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Seed data (for development)
	if cfg.IsDevelopment() {
		if err := database.Seed(); err != nil {
			log.Printf("Warning: Failed to seed data: %v", err)
		}
	}

	// Get database instance
	db := database.GetDB()

	// Initialize webhook service
	webhookservice.Init(db, 3, 5)
	log.Println("✅ Webhook service initialized")

	// ========== Initialize Repositories ==========
	shopRepo := repository.NewShopRepository(db)
	productRepo := repository.NewProductRepository(db)
	saleRepo := repository.NewSaleRepository(db)
	summaryRepo := repository.NewDailySummaryRepository(db)
	auditRepo := repository.NewAuditLogRepository(db)
	staffRepo := repository.NewStaffRepository(db)
	apiKeyRepo := repository.NewAPIKeyRepository(db)
	webhookRepo := repository.NewWebhookRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	accountRepo := repository.NewAccountRepository(db)
	supplierRepo := repository.NewSupplierRepository(db)
	orderRepo := repository.NewOrderRepository(db)

	// ========== Initialize Services ==========
	authService := services.NewAuthService(shopRepo, cfg)
	authService.SetAccountRepo(accountRepo)
	cmdHandler := services.NewCommandHandler(db, shopRepo, productRepo, saleRepo, summaryRepo, auditRepo)

	// Set account repo for multi-shop support
	if cfg.FeatureMultipleShopsEnabled {
		cmdHandler.SetAccountRepo(accountRepo)
	}

	// Set staff repo for staff commands
	if cfg.FeatureStaffAccountsEnabled {
		cmdHandler.SetStaffRepo(staffRepo)
	}

	// Set supplier repo for supplier commands (Pro feature)
	if cfg.FeatureStaffAccountsEnabled {
		cmdHandler.SetSupplierRepo(supplierRepo, orderRepo)
	}

	// Set customer repo for loyalty commands (Business feature)
	if cfg.FeatureAnalyticsEnabled {
		cmdHandler.SetCustomerRepo(customerRepo)
	}

	// Initialize M-Pesa repositories
	mpesaPaymentRepo := repository.NewMpesaPaymentRepository(db)
	mpesaTransactionRepo := repository.NewMpesaTransactionRepository(db)

	// M-Pesa Service (if enabled)
	var mpesaSvc *mpesaservice.Service
	if cfg.FeatureMpesaEnabled {
		if cfg.MPesaConsumerKey != "" && cfg.MPesaShortcode != "" {
			mpesaSvc = mpesaservice.New(&mpesaservice.Config{
				ConsumerKey:    cfg.MPesaConsumerKey,
				ConsumerSecret: cfg.MPesaConsumerSecret,
				Shortcode:      cfg.MPesaShortcode,
				Passkey:        cfg.MPesaPasskey,
				CallbackURL:    cfg.MPesaCallbackURL,
				Environment:    cfg.MPesaEnvironment,
			}, mpesaPaymentRepo, mpesaTransactionRepo)
			log.Println("✅ M-Pesa service initialized")
		} else {
			log.Println("⚠️ M-Pesa enabled but not configured (missing credentials)")
		}
	}

	// Set M-Pesa service for WhatsApp payments
	if mpesaSvc != nil {
		cmdHandler.SetMpesaService(mpesaSvc)
	}

	// SMS Service (Africa Talking)
	var smsSvc *smsservice.Service
	if cfg.AfricaTalkingAPIKey != "" && cfg.AfricaTalkingUsername != "" {
		smsSvc = smsservice.New(&smsservice.Config{
			APIKey:   cfg.AfricaTalkingAPIKey,
			Username: cfg.AfricaTalkingUsername,
			BaseURL:  "https://api.africastalking.com",
		})
		log.Println("✅ SMS service (Africa Talking) initialized")
	} else {
		log.Println("⚠️ Africa Talking SMS not configured")
	}

	// Email Service (SendGrid)
	var emailSvc *email.Service
	if cfg.SendGridAPIKey != "" {
		emailSvc = email.New(&email.Config{
			APIKey:    cfg.SendGridAPIKey,
			FromEmail: cfg.SendGridFromEmail,
			FromName:  cfg.SendGridFromName,
		})
		log.Println("✅ Email service (SendGrid) initialized")
	} else {
		log.Println("⚠️ SendGrid email not configured")
	}

	// API Service (if enabled)
	var apiSvc *apiservice.Service
	if cfg.FeatureAnalyticsEnabled {
		apiSvc = apiservice.New(apiKeyRepo)
		log.Println("✅ API service initialized")
	}

	// USSD Service (if multiple shops enabled)
	var ussdSvc *ussdservice.Service
	var ussdHandler *ussdhandler.Handler
	if cfg.FeatureMultipleShopsEnabled {
		ussdSvc = ussdservice.New()
		ussdSvc.SetRepositories(shopRepo, productRepo, saleRepo, summaryRepo)
		ussdHandler = ussdhandler.New(ussdSvc)
		log.Println("✅ USSD service initialized")
	}

	// Printer Service
	printerSvc := printerservice.New(&printerservice.PrinterConfig{})
	log.Println("✅ Printer service initialized")

	// Cache Service (Redis)
	var cacheSvc *cacheservice.CacheService
	_ = cacheSvc // Used for caching daily summaries
	if cfg.RedisURL != "" {
		var err error
		cacheSvc, err = cacheservice.NewCacheService(&cacheservice.Config{
			URL:      cfg.RedisURL,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		})
		if err != nil {
			log.Printf("⚠️ Redis connection failed: %v (using in-memory fallback)", err)
		} else {
			log.Println("✅ Cache service (Redis) initialized")
		}
	}

	// ========== Initialize Handlers ==========
	whatsappHandler := handlers.NewWhatsAppHandler(cmdHandler, cfg)
	authHandler := handlers.NewAuthHandler(authService)
	shopHandler := handlers.NewShopHandlerWithAccount(shopRepo, productRepo, saleRepo, accountRepo)
	productHandler := handlers.NewProductHandler(productRepo)
	saleHandler := handlers.NewSaleHandler(saleRepo, productRepo)
	staffHandler := staffhandler.New(staffRepo, shopRepo)
	webhookHandler := webhookhandler.New(webhookRepo)

	// Export Handler
	exportHandler := exporthandler.NewExportHandler(productRepo, saleRepo, summaryRepo)
	log.Println("✅ Export handler initialized")

	// QR Handler
	var qrHandler *qrhandler.QRHandler
	if mpesaSvc != nil {
		qrSvc := qrservice.NewQRPaymentService(db, mpesaSvc, shopRepo, saleRepo, productRepo)
		qrHandler = qrhandler.NewQRHandler(qrSvc)
		cmdHandler.SetQRService(qrSvc)
		log.Println("✅ QR handler initialized")
	}

	// SMS Handler (Africa Talking)
	var smsHandler *smshandler.Handler
	if smsSvc != nil {
		smsHandler = smshandler.New(smsSvc)
		log.Println("✅ SMS handler initialized")
	}

	// M-Pesa Handler
	var mpesaHandler *mpesahandler.Handler
	if mpesaSvc != nil {
		mpesaHandler = mpesahandler.New(mpesaSvc, shopRepo, productRepo, saleRepo, mpesaPaymentRepo, mpesaTransactionRepo)
		log.Println("✅ M-Pesa handler initialized")
	}

	// ========== Initialize Scheduler ==========
	scheduler := scheduler.New()

	// Daily report task - runs every 24 hours
	scheduler.AddTask("daily_reports", 24*time.Hour, func() error {
		log.Println("📊 Running daily reports task...")

		shops, _, err := shopRepo.List(1000, 0)
		if err != nil {
			log.Printf("❌ Failed to get shops: %v", err)
			return err
		}

		for _, shop := range shops {
			if !shop.IsActive {
				continue
			}

			sales, err := saleRepo.GetTodaySales(shop.ID)
			if err != nil {
				continue
			}

			totalSales := 0.0
			totalProfit := 0.0
			for _, s := range sales {
				totalSales += s.TotalAmount
				totalProfit += s.Profit
			}

			if len(sales) > 0 {
				reportMsg := fmt.Sprintf("📊 DAILY REPORT - %s\n\n💰 Today's Sales: KSh %.0f\n💵 Profit: KSh %.0f\n📝 Transactions: %d\n\nSent automatically by DukaPOS", shop.Name, totalSales, totalProfit, len(sales))

				if err := whatsappHandler.SendWhatsAppMessage(shop.Phone, reportMsg); err != nil {
					log.Printf("❌ Failed to send daily report to shop %s: %v", shop.Name, err)
				} else {
					log.Printf("✅ Daily report sent to shop %s", shop.Name)
				}
			}
		}

		log.Println("✅ Daily reports task completed")
		return nil
	})

	// Low stock check - runs every 6 hours
	scheduler.AddTask("low_stock_check", 6*time.Hour, func() error {
		log.Println("⚠️ Running low stock check...")

		shops, _, err := shopRepo.List(1000, 0)
		if err != nil {
			return err
		}

		for _, shop := range shops {
			if !shop.IsActive {
				continue
			}

			lowStock, err := productRepo.GetLowStock(shop.ID)
			if err != nil {
				continue
			}

			if len(lowStock) > 0 {
				var productList strings.Builder
				productList.WriteString("⚠️ LOW STOCK ALERT\n\n")
				for _, p := range lowStock {
					productList.WriteString(fmt.Sprintf("• %s: %d (min: %d)\n", p.Name, p.CurrentStock, p.LowStockThreshold))
				}
				productList.WriteString("\nAdd stock: add [name] [price] [qty]")

				if err := whatsappHandler.SendWhatsAppMessage(shop.Phone, productList.String()); err != nil {
					log.Printf("❌ Failed to send low stock alert to shop %s: %v", shop.Name, err)
				} else {
					log.Printf("✅ Low stock alert sent to shop %s", shop.Name)
				}
			}
		}

		log.Println("✅ Low stock check completed")
		return nil
	})

	// Weekly report task - runs every 7 days
	scheduler.AddTask("weekly_reports", 7*24*time.Hour, func() error {
		log.Println("📊 Running weekly reports task...")

		shops, _, err := shopRepo.List(1000, 0)
		if err != nil {
			return err
		}

		for _, shop := range shops {
			if !shop.IsActive {
				continue
			}

			end := time.Now()
			start := end.AddDate(0, 0, -7)
			sales, err := saleRepo.GetByDateRange(shop.ID, start, end)
			if err != nil {
				continue
			}

			if len(sales) > 0 {
				totalSales := 0.0
				totalProfit := 0.0
				for _, s := range sales {
					totalSales += s.TotalAmount
					totalProfit += s.Profit
				}

				reportMsg := fmt.Sprintf("📊 WEEKLY REPORT\n\n💰 Weekly Sales: KSh %.0f\n💵 Profit: KSh %.0f\n📝 Transactions: %d\n\nHave a great week!", totalSales, totalProfit, len(sales))

				if err := whatsappHandler.SendWhatsAppMessage(shop.Phone, reportMsg); err != nil {
					log.Printf("❌ Failed to send weekly report to shop %s: %v", shop.Name, err)
				}
			}
		}

		log.Println("✅ Weekly reports task completed")
		return nil
	})

	// Monthly report task - runs every 30 days
	scheduler.AddTask("monthly_reports", 30*24*time.Hour, func() error {
		log.Println("📊 Running monthly reports task...")

		shops, _, err := shopRepo.List(1000, 0)
		if err != nil {
			return err
		}

		for _, shop := range shops {
			if !shop.IsActive {
				continue
			}

			end := time.Now()
			start := end.AddDate(0, -1, 0)
			sales, err := saleRepo.GetByDateRange(shop.ID, start, end)
			if err != nil {
				continue
			}

			if len(sales) > 0 {
				totalSales := 0.0
				totalProfit := 0.0
				for _, s := range sales {
					totalSales += s.TotalAmount
					totalProfit += s.Profit
				}

				avgDaily := totalSales / 30

				reportMsg := fmt.Sprintf("📊 MONTHLY REPORT\n\n💰 Monthly Sales: KSh %.0f\n💵 Profit: KSh %.0f\n📝 Transactions: %d\n📈 Daily Avg: KSh %.0f\n\nGreat progress this month! 🎉", totalSales, totalProfit, len(sales), avgDaily)

				if err := whatsappHandler.SendWhatsAppMessage(shop.Phone, reportMsg); err != nil {
					log.Printf("❌ Failed to send monthly report to shop %s: %v", shop.Name, err)
				}
			}
		}

		log.Println("✅ Monthly reports task completed")
		return nil
	})

	// Start scheduler
	scheduler.Start()
	log.Println("✅ Scheduler initialized")

	// ========== Create Fiber App ==========
	var emailHandler *emailhandler.Handler
	if emailSvc != nil {
		emailHandler = emailhandler.New(emailSvc)
		log.Println("✅ Email handler initialized")
	}

	// API Key Handler
	var apiKeyHandler *apihandler.APIKeyHandler
	if apiSvc != nil {
		apiKeyHandler = apihandler.NewAPIKeyHandler(apiSvc)
	}

	// AI Handler
	var aiHandler *aihandler.Handler
	if cfg.FeatureAnalyticsEnabled {
		aiPredService := ai.NewPredictionService(productRepo, saleRepo, summaryRepo)
		aiHandler = aihandler.New(aiPredService)
		cmdHandler.SetPredictionService(aiPredService)
		log.Println("✅ AI Predictions service initialized")
	}

	// ========== Create Fiber App ==========
	app := fiber.New(fiber.Config{
		AppName:      "DukaPOS",
		ServerHeader: "DukaPOS/1.0.0",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${method} ${path} ${latency}\n",
		TimeFormat: "2006-01-02 15:04:05",
	}))
	app.Use(compress.New())

	// CORS
	if cfg.CORSEnabled {
		app.Use(middleware.CORSMiddleware(cfg.GetAllowedOrigins()))
	}

	// Rate Limiter
	app.Use(middleware.RateLimiter(60, 60))

	// Serve static files
	app.Static("/static", "./static")

	// Template engine
	if cfg.FeatureWebDashboardEnabled {
		// Load templates
		app.Static("/templates", "./templates")

		// Serve the JS dashboard
		app.Static("/dashboard", "./static")

		// Initialize web handler
		webHandler := handlers.NewWebHandler(shopRepo, productRepo, saleRepo)
		webHandler.SetAdditionalRepos(summaryRepo, customerRepo, staffRepo)

		// Web routes - serve templates from templates directory
		web := app.Group("")

		// Landing page
		web.Get("/", func(c *fiber.Ctx) error {
			return c.SendFile("./templates/landing.html")
		})

		// Login page
		web.Get("/login", func(c *fiber.Ctx) error {
			return c.SendFile("./templates/login.html")
		})

		// Register page
		web.Get("/register", func(c *fiber.Ctx) error {
			return c.SendFile("./templates/register.html")
		})

		// Logout
		web.Get("/logout", func(c *fiber.Ctx) error {
			c.ClearCookie("token")
			return c.Redirect("/login")
		})

		// Dashboard
		web.Get("/dashboard", func(c *fiber.Ctx) error {
			return c.SendFile("./templates/dashboard.html")
		})
		web.Get("/dashboard/:shop_id", func(c *fiber.Ctx) error {
			return c.SendFile("./templates/dashboard.html")
		})

		// Products
		web.Get("/products", func(c *fiber.Ctx) error {
			return c.SendFile("./templates/products.html")
		})
		web.Get("/products/:shop_id", webHandler.ProductsList)

		// Sales
		web.Get("/sales", func(c *fiber.Ctx) error {
			return c.SendFile("./templates/sales.html")
		})
		web.Get("/sales/:shop_id", webHandler.SalesList)

		// Customers
		web.Get("/customers", func(c *fiber.Ctx) error {
			return c.SendFile("./templates/customers.html")
		})

		// Suppliers
		web.Get("/suppliers", func(c *fiber.Ctx) error {
			return c.SendFile("./templates/suppliers.html")
		})

		// Orders
		web.Get("/orders", func(c *fiber.Ctx) error {
			return c.SendFile("./templates/orders.html")
		})

		// M-Pesa
		web.Get("/mpesa", func(c *fiber.Ctx) error {
			return c.SendFile("./templates/mpesa.html")
		})

		// Staff
		web.Get("/staff", func(c *fiber.Ctx) error {
			return c.SendFile("./templates/staff.html")
		})

		// AI Insights
		web.Get("/ai", func(c *fiber.Ctx) error {
			return c.SendFile("./templates/ai.html")
		})

		// API Keys
		web.Get("/apikeys", func(c *fiber.Ctx) error {
			return c.SendFile("./templates/apikeys.html")
		})

		// Webhooks
		web.Get("/webhooks", func(c *fiber.Ctx) error {
			return c.SendFile("./templates/webhooks.html")
		})

		// SMS
		web.Get("/sms", func(c *fiber.Ctx) error {
			return c.SendFile("./templates/sms.html")
		})

		// Email
		web.Get("/email", func(c *fiber.Ctx) error {
			return c.SendFile("./templates/email.html")
		})

		// Printer
		web.Get("/printer", func(c *fiber.Ctx) error {
			return c.SendFile("./templates/printer.html")
		})

		// Reports
		web.Get("/reports", func(c *fiber.Ctx) error {
			return c.SendFile("./templates/reports.html")
		})

		// Export
		web.Get("/export", func(c *fiber.Ctx) error {
			return c.SendFile("./templates/export.html")
		})

		// Settings
		web.Get("/settings", func(c *fiber.Ctx) error {
			return c.SendFile("./templates/settings.html")
		})

		// Billing
		web.Get("/billing", func(c *fiber.Ctx) error {
			return c.SendFile("./templates/billing.html")
		})

		// Admin Routes - with auth middleware
		adminWeb := web.Group("/admin")
		adminWeb.Use(middleware.JWT(authService))
		adminWeb.Use(func(c *fiber.Ctx) error {
			account, ok := c.Locals("account").(*models.Account)
			if !ok || account == nil || !account.IsAdmin {
				return c.Redirect("/admin/login?error=unauthorized")
			}
			return c.Next()
		})

		adminWeb.Get("/", func(c *fiber.Ctx) error {
			return c.Redirect("/admin/dashboard")
		})
		adminWeb.Get("/dashboard", func(c *fiber.Ctx) error {
			return c.SendFile("./templates/admin/dashboard.html")
		})
		adminWeb.Get("/users", func(c *fiber.Ctx) error {
			return c.SendFile("./templates/admin/users.html")
		})
		adminWeb.Get("/accounts", func(c *fiber.Ctx) error {
			return c.SendFile("./templates/admin/accounts.html")
		})
		adminWeb.Get("/shops", func(c *fiber.Ctx) error {
			return c.SendFile("./templates/admin/shops.html")
		})
		adminWeb.Get("/settings", func(c *fiber.Ctx) error {
			return c.SendFile("./templates/admin/settings.html")
		})
		adminWeb.Get("/subscriptions", func(c *fiber.Ctx) error {
			return c.SendFile("./templates/admin/subscriptions.html")
		})

		// Admin login page (no auth required)
		web.Get("/admin/login", func(c *fiber.Ctx) error {
			return c.SendFile("./templates/admin/login.html")
		})

		log.Println("✅ Web dashboard enabled")

		// Dashboard API routes - use JWT auth like protected routes
		webAPI := app.Group("/api/v1")
		webAPI.Use(middleware.JWT(authService))
		webAPI.Get("/shop/dashboard-json/:shop_id", webHandler.DashboardJSON)
		webAPI.Get("/shop/dashboard/:shop_id", webHandler.Dashboard)
		// Put specific routes BEFORE parameterized routes
		webAPI.Get("/products/categories", productHandler.ListCategories)
		webAPI.Post("/products/bulk", productHandler.BulkCreateProducts)
		webAPI.Post("/products", webHandler.APIProductCreate)
		webAPI.Get("/products", productHandler.ListProducts)
		webAPI.Get("/products/:id", productHandler.GetProduct)
		webAPI.Put("/products/:id", webHandler.APIProductUpdate)
		webAPI.Delete("/products/:id", webHandler.APIProductDelete)
		webAPI.Get("/sales/:shop_id", webHandler.APISales)
		webAPI.Post("/sales", webHandler.APISaleCreate)
		webAPI.Get("/reports/:shop_id", webHandler.APIReports)
	}

	// ========== API Routes ==========
	api := app.Group("/api")

	// Health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "DukaPOS",
			"version": "1.0.0",
		})
	})

	// API Documentation
	docsHandler := docshandler.New()
	docsHandler.RegisterRoutes(app)

	// API Info
	api.Get("/", func(c *fiber.Ctx) error {
		features := []string{"inventory", "sales"}
		if cfg.FeatureMpesaEnabled && mpesaSvc != nil {
			features = append(features, "mpesa")
		}
		if cfg.FeatureStaffAccountsEnabled {
			features = append(features, "staff")
		}
		if cfg.FeatureAnalyticsEnabled {
			features = append(features, "api", "webhooks")
		}
		if cfg.FeatureMultipleShopsEnabled {
			features = append(features, "multi-shop")
		}

		return c.JSON(fiber.Map{
			"name":        "DukaPOS API",
			"version":     "1.0.0",
			"description": "REST API for Kenyan Duka POS",
			"base_url":    "/api/v1",
			"channels":    []string{"WhatsApp", "USSD", "REST"},
			"features":    features,
		})
	})

	// Auth routes (public)
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/otp/send", authHandler.SendOTP)
	auth.Post("/otp/verify", authHandler.VerifyOTP)

	// Protected routes
	protected := api.Group("/v1")
	protected.Use(middleware.JWT(authService))

	// Plan/Subscription routes
	planHandler := middleware.NewPlanInfoHandler()
	api.Get("/plans", planHandler.GetAllPlans)
	protected.Get("/plan", planHandler.GetPlanInfo)

	// ========== Admin Routes ==========
	adminHandler := handlers.NewAdminHandler()
	admin := protected.Group("/admin")
	admin.Use(middleware.RequireAdmin())
	admin.Get("/dashboard", adminHandler.Dashboard)
	admin.Get("/accounts", adminHandler.GetAccounts)
	admin.Get("/accounts/:id", adminHandler.GetAccount)
	admin.Put("/accounts/:id/plan", adminHandler.UpdateAccountPlan)
	admin.Put("/accounts/:id/status", adminHandler.UpdateAccountStatus)
	admin.Get("/shops", adminHandler.GetShops)
	admin.Get("/revenue", adminHandler.GetRevenueStats)
	admin.Post("/upgrade-all", adminHandler.UpgradeAllAccounts)

	// Public admin fix endpoint (call once to create admin shop)
	api.Post("/admin/fix", adminHandler.FixAdmin)

	// Shop routes
	protected.Get("/shop/profile", shopHandler.GetProfile)
	protected.Put("/shop/profile", shopHandler.UpdateProfile)
	protected.Get("/shop/dashboard", shopHandler.GetDashboard)
	protected.Get("/shop/account", shopHandler.GetAccount)

	// Product routes
	protected.Get("/products", productHandler.ListProducts)
	protected.Get("/products/:id", productHandler.GetProduct)
	protected.Post("/products", productHandler.CreateProduct)
	protected.Put("/products/:id", productHandler.UpdateProduct)
	protected.Delete("/products/:id", productHandler.DeleteProduct)
	protected.Post("/products/bulk", productHandler.BulkCreateProducts)
	protected.Get("/products/categories", productHandler.ListCategories)
	protected.Post("/products/categories", productHandler.CreateCategory)
	protected.Put("/products/categories/:id", productHandler.UpdateCategory)
	protected.Delete("/products/categories/:id", productHandler.DeleteCategory)

	// Sale routes
	protected.Get("/sales", saleHandler.ListSales)
	protected.Get("/sales/:id", saleHandler.GetSale)
	protected.Post("/sales", saleHandler.CreateSale)

	// Export routes
	protected.Get("/export/products", exportHandler.ExportProducts)
	protected.Get("/export/sales", exportHandler.ExportSales)
	protected.Get("/export/report", exportHandler.ExportReport)
	protected.Get("/export/inventory", exportHandler.ExportInventory)

	// QR Payment routes
	if qrHandler != nil {
		protected.Post("/qr/generate", qrHandler.GenerateDynamicQR)
		protected.Post("/qr/static", qrHandler.GenerateStaticQR)
		protected.Get("/qr/status/:id", qrHandler.GetPaymentStatus)
		protected.Post("/qr/callback", qrHandler.HandleCallback)
	}

	// ========== Staff Routes (Feature Flag) ==========
	if cfg.FeatureStaffAccountsEnabled {
		staffRoutes := protected.Group("/staff")
		staffRoutes.Use(middleware.RequirePro())
		staffRoutes.Get("/", staffHandler.List)
		staffRoutes.Get("/:id", staffHandler.Get)
		staffRoutes.Post("/", staffHandler.Create)
		staffRoutes.Put("/:id", staffHandler.Update)
		staffRoutes.Delete("/:id", staffHandler.Delete)
		staffRoutes.Put("/:id/pin", staffHandler.UpdatePin)
		log.Println("✅ Staff routes enabled (Plan: Pro+)")
	}

	// ========== M-Pesa Routes (Feature Flag) ==========
	if mpesaHandler != nil {
		mpesaRoutes := protected.Group("/mpesa")
		mpesaRoutes.Use(middleware.RequirePro())
		mpesaRoutes.Post("/stk-push", mpesaHandler.STKPush)
		mpesaRoutes.Get("/status/:checkoutId", mpesaHandler.GetStatus)
		mpesaRoutes.Get("/payments", mpesaHandler.ListPayments)
		mpesaRoutes.Post("/payments/:id/retry", mpesaHandler.RetryPayment)
		mpesaRoutes.Get("/transactions", mpesaHandler.GetTransactions)
		mpesaRoutes.Get("/balance", mpesaHandler.GetBalance)
		mpesaRoutes.Post("/b2c", mpesaHandler.B2CSend)
		log.Println("✅ M-Pesa routes enabled (Plan: Pro+)")
	}

	// ========== API Keys Routes (Feature Flag) ==========
	if apiKeyHandler != nil {
		apiKeyRoutes := protected.Group("/api-keys")
		apiKeyRoutes.Use(middleware.RequireBusiness())
		apiKeyRoutes.Get("/", apiKeyHandler.List)
		apiKeyRoutes.Post("/", apiKeyHandler.Create)
		apiKeyRoutes.Delete("/:id", apiKeyHandler.Revoke)
		log.Println("✅ API Keys routes enabled")
	}

	// ========== Webhook Routes (Feature Flag) ==========
	if cfg.FeatureAnalyticsEnabled {
		webhookRoutes := protected.Group("/webhooks")
		webhookRoutes.Use(middleware.RequireBusiness())
		webhookRoutes.Get("/", webhookHandler.List)
		webhookRoutes.Get("/:id", webhookHandler.Get)
		webhookRoutes.Post("/", webhookHandler.Create)
		webhookRoutes.Put("/:id", webhookHandler.Update)
		webhookRoutes.Delete("/:id", webhookHandler.Delete)
		webhookRoutes.Post("/:id/test", webhookHandler.Test)
		log.Println("✅ Webhook routes enabled (Plan: Business)")
	}

	// ========== SMS Routes ==========
	if smsHandler != nil {
		smsHandler.RegisterRoutes(app, protected)
		log.Println("✅ SMS routes enabled")
	}

	// ========== Email Routes ==========
	if emailHandler != nil {
		emailHandler.RegisterRoutes(protected)
		log.Println("✅ Email routes enabled")
	}

	// ========== AI Routes (Feature Flag) ==========
	if aiHandler != nil {
		aiRoutes := protected.Group("/ai")
		aiRoutes.Use(middleware.RequireBusiness())
		aiRoutes.Get("/predictions/:shop_id", aiHandler.GetPredictions)
		aiRoutes.Get("/restock/:shop_id", aiHandler.GetRestockRecommendations)
		aiRoutes.Get("/analytics/:shop_id", aiHandler.GetSalesAnalytics)
		aiRoutes.Get("/inventory-value/:shop_id", aiHandler.GetInventoryValue)
		aiRoutes.Get("/trends/:shop_id", aiHandler.GetTrends)
		aiRoutes.Post("/forecast/:shop_id", aiHandler.GenerateForecast)
		log.Println("✅ AI Predictions routes enabled (Plan: Business)")
	}

	// ========== Customer/Loyalty Routes ==========
	if cfg.FeatureAnalyticsEnabled {
		customerHandler := handlers.NewCustomerHandler(customerRepo, shopRepo)
		customerRoutes := protected.Group("/customers")
		customerRoutes.Get("/", customerHandler.List)
		customerRoutes.Get("/:id", customerHandler.Get)
		customerRoutes.Post("/", customerHandler.Create)
		customerRoutes.Put("/:id", customerHandler.Update)
		customerRoutes.Delete("/:id", customerHandler.Delete)
		log.Println("✅ Customer routes enabled")

		loyaltyHandler := loyaltyhandler.NewHandler(customerRepo, saleRepo, db)
		loyaltyRoutes := protected.Group("/loyalty")
		loyaltyRoutes.Get("/points/:customer_id", loyaltyHandler.GetCustomerPoints)
		loyaltyRoutes.Get("/stats/:customer_id", loyaltyHandler.GetCustomerStats)
		loyaltyRoutes.Post("/redeem", loyaltyHandler.RedeemPoints)
		loyaltyRoutes.Post("/earn", loyaltyHandler.EarnPoints)
		loyaltyRoutes.Get("/transactions/:customer_id", loyaltyHandler.ListTransactions)
		log.Println("✅ Loyalty routes enabled")
	}

	// Currency Routes
	currencyHandler := currhandler.NewHandler(db, cfg)
	currencyRoutes := protected.Group("/currency")
	currencyRoutes.Get("/list", currencyHandler.ListCurrencies)
	currencyRoutes.Get("/:code", currencyHandler.GetCurrency)
	currencyRoutes.Post("/convert", currencyHandler.Convert)
	currencyRoutes.Post("/format", currencyHandler.Format)
	log.Println("✅ Currency routes enabled")

	// ========== Billing Routes ==========
	billingHandler := billinghandler.NewHandler(db, cfg)
	billingRoutes := protected.Group("/billing")
	billingRoutes.Get("/plans", billingHandler.GetPlans)
	billingRoutes.Get("/current", billingHandler.GetCurrentPlan)
	billingRoutes.Post("/upgrade", billingHandler.UpgradePlan)
	log.Println("✅ Billing routes enabled")

	// ========== Supplier/Order Routes (Pro Feature) ==========
	supplierHandler := supplierhandler.New(supplierRepo, orderRepo, productRepo)
	supplierRoutes := protected.Group("/suppliers")
	supplierRoutes.Get("/", supplierHandler.ListSuppliers)
	supplierRoutes.Post("/", supplierHandler.CreateSupplier)
	supplierRoutes.Get("/:id", supplierHandler.GetSupplier)
	supplierRoutes.Put("/:id", supplierHandler.UpdateSupplier)
	supplierRoutes.Delete("/:id", supplierHandler.DeleteSupplier)

	orderRoutes := protected.Group("/orders")
	orderRoutes.Get("/", supplierHandler.ListOrders)
	orderRoutes.Post("/", supplierHandler.CreateOrder)
	orderRoutes.Get("/:id", supplierHandler.GetOrder)
	orderRoutes.Put("/:id/status", supplierHandler.UpdateOrderStatus)
	orderRoutes.Delete("/:id", supplierHandler.DeleteOrder)
	log.Println("✅ Supplier/Order routes enabled")

	// ========== Printer/Receipt Routes ==========
	printerHandler := printerhandler.New(printerSvc)
	printRoutes := protected.Group("/print")
	printRoutes.Post("/receipt", printerHandler.PrintReceipt)
	printRoutes.Post("/text", printerHandler.GetTextReceipt)
	printRoutes.Post("/thermal", printerHandler.GetThermalReceipt)
	printRoutes.Post("/html", printerHandler.GetHTMLReceipt)
	printRoutes.Post("/report", printerHandler.PrintDailyReport)
	printRoutes.Get("/printers", printerHandler.GetPrinters)
	printRoutes.Post("/test", printerHandler.TestPrinter)
	printRoutes.Put("/config", printerHandler.Configure)
	printRoutes.Get("/config", printerHandler.GetConfig)
	log.Println("✅ Printer routes enabled")

	// ========== Webhook Routes (External - No Auth) ==========
	webhook := app.Group("/webhook")

	// Twilio WhatsApp
	webhook.Post("/twilio", whatsappHandler.HandleWebhook)
	webhook.Post("/twilio/status", whatsappHandler.HandleStatusCallback)
	webhook.Get("/twilio/verify", whatsappHandler.WebhookVerification)

	// M-Pesa Callbacks
	if mpesaHandler != nil {
		webhook.Post("/mpesa/stk", mpesaHandler.STKCallback)
		webhook.Post("/mpesa/b2c", mpesaHandler.B2CCallback)
		webhook.Post("/mpesa/balance", mpesaHandler.BalanceCallback)
	}

	// ========== USSD Routes ==========
	if ussdHandler != nil {
		ussdRoutes := app.Group("/api/v1/ussd")
		ussdRoutes.Post("/", ussdHandler.Handle)
		ussdRoutes.Post("/africa", ussdHandler.HandleAfricaTalking)
		ussdRoutes.Post("/callback", ussdHandler.Callback)
		log.Println("✅ USSD routes enabled")
	}

	// ========== WebSocket ==========
	websocket.Init()
	app.Get("/ws", websocket.HandleWebSocket)

	// 404 handler
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Endpoint not found",
		})
	})

	// ========== Graceful Shutdown ==========
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down gracefully...")
		app.Shutdown()
	}()

	// ========== Start Server ==========
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("🚀 Starting DukaPOS server on %s", addr)
	log.Printf("📱 WhatsApp webhook: http://your-domain/webhook/twilio")
	log.Printf("🔗 API: http://localhost:%s/api", cfg.Port)

	// Feature summary
	log.Println("📋 Features:")
	log.Printf("   • Products: ✅")
	log.Printf("   • Sales: ✅")
	if cfg.FeatureStaffAccountsEnabled {
		log.Printf("   • Staff: ✅")
	}
	if mpesaSvc != nil {
		log.Printf("   • M-Pesa: ✅")
	}
	if apiSvc != nil {
		log.Printf("   • API Keys: ✅")
		log.Printf("   • Webhooks: ✅")
	}
	if cfg.FeatureMultipleShopsEnabled {
		log.Printf("   • USSD: ✅")
		log.Printf("   • Multi-shop: ✅")
	}

	if err := app.Listen(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
