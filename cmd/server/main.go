package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/C9b3rD3vi1/DukaPOS/internal/config"
	"github.com/C9b3rD3vi1/DukaPOS/internal/database"
	"github.com/C9b3rD3vi1/DukaPOS/internal/handlers"
	aihandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/ai"
	apihandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/api"
	auditloghandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/auditlog"
	billinghandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/billing"
	currencyhandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/currency"
	docshandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/docs"
	emailhandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/email"
	exporthandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/export"
	jobscheduler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/jobscheduler"
	loyaltyhandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/loyalty"
	mpesahandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/mpesa"
	printerhandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/printer"
	pushhandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/push"
	qrhandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/qr"
	smshandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/sms"
	staffhandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/staff"
	supplierhandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/supplier"
	twofactorhandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/twofactor"
	ussdhandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/ussd"
	webhookhandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/webhook"
	"github.com/C9b3rD3vi1/DukaPOS/internal/middleware"
	"github.com/C9b3rD3vi1/DukaPOS/internal/repository"
	"github.com/C9b3rD3vi1/DukaPOS/internal/routes"
	"github.com/C9b3rD3vi1/DukaPOS/internal/services"
	ai "github.com/C9b3rD3vi1/DukaPOS/internal/services/ai"
	apiservice "github.com/C9b3rD3vi1/DukaPOS/internal/services/api"
	cacheservice "github.com/C9b3rD3vi1/DukaPOS/internal/services/cache"
	currencyservice "github.com/C9b3rD3vi1/DukaPOS/internal/services/currency"
	email "github.com/C9b3rD3vi1/DukaPOS/internal/services/email"
	encryption "github.com/C9b3rD3vi1/DukaPOS/internal/services/encryption"
	mpesaservice "github.com/C9b3rD3vi1/DukaPOS/internal/services/mpesa"
	printerservice "github.com/C9b3rD3vi1/DukaPOS/internal/services/printer"
	qrservice "github.com/C9b3rD3vi1/DukaPOS/internal/services/qr"
	smsservice "github.com/C9b3rD3vi1/DukaPOS/internal/services/sms"
	twofactorservice "github.com/C9b3rD3vi1/DukaPOS/internal/services/twofactor"
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
	reportHandler := handlers.NewReportHandlerWithCache(saleRepo, productRepo, summaryRepo, cacheSvc)
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
	routes.RegisterScheduledTasks(routes.SchedulerConfig{
		ShopRepo:     shopRepo,
		SaleRepo:     saleRepo,
		ProductRepo:  productRepo,
		SendWhatsApp: whatsappHandler.SendWhatsAppMessage,
	})

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

	// Serve React frontend (PWA)
	webHandler := handlers.NewWebHandler(shopRepo, productRepo, saleRepo)

	if cfg.FeatureWebDashboardEnabled {
		// Serve the React frontend built with Vite
		app.Static("/", "./dukapos-frontend/dist")

		// Initialize web handler for API fallback
		webHandler := handlers.NewWebHandler(shopRepo, productRepo, saleRepo)
		webHandler.SetAdditionalRepos(summaryRepo, customerRepo, staffRepo)

		// Legacy template routes (for backward compatibility)
		web := app.Group("")

		// Landing page
		web.Get("/", func(c *fiber.Ctx) error {
			return c.SendFile("./dukapos-frontend/dist/index.html")
		})

		// Login page
		web.Get("/login", func(c *fiber.Ctx) error {
			return c.SendFile("./dukapos-frontend/dist/index.html")
		})

		// Register page
		web.Get("/register", func(c *fiber.Ctx) error {
			return c.SendFile("./dukapos-frontend/dist/index.html")
		})

		// Dashboard (React app handles routing)
		web.Get("/dashboard/*", func(c *fiber.Ctx) error {
			return c.SendFile("./dukapos-frontend/dist/index.html")
		})

		// Admin routes
		web.Get("/admin/*", func(c *fiber.Ctx) error {
			return c.SendFile("./dukapos-frontend/dist/index.html")
		})

		log.Println("✅ React frontend enabled at /")
	}

	// Dashboard API routes - use JWT auth like protected routes
	webAPI := app.Group("/api/v1")
	webAPI.Use(middleware.JWT(authService))
	webAPI.Get("/shop/dashboard-json/:shop_id", webHandler.DashboardJSON)
	webAPI.Get("/shop/dashboard/:shop_id", webHandler.Dashboard)
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

	// ========== Create additional handlers for routes ==========
	adminHandler := handlers.NewAdminHandler()
	billingHandler := billinghandler.NewHandler(db, cfg)
	planHandler := middleware.NewPlanInfoHandler()
	customerHandler := handlers.NewCustomerHandler(customerRepo, shopRepo)
	var loyaltyHandler *loyaltyhandler.Handler
	var supplierHandler *supplierhandler.Handler
	var printerHandler *printerhandler.Handler

	if cfg.FeatureMultipleShopsEnabled {
		loyaltyHandler = loyaltyhandler.NewHandler(customerRepo, saleRepo, db)
		supplierHandler = supplierhandler.New(supplierRepo, orderRepo, productRepo)
	}

	if printerSvc != nil {
		printerHandler = printerhandler.New(printerSvc)
	}

	// Protected routes
	protected := api.Group("/v1")
	protected.Use(middleware.JWT(authService))

	// ========== Initialize Additional Handlers ==========
	// Currency Handler
	_ = currencyservice.NewService(db, cfg)
	currencyHandler := currencyhandler.NewHandler(db, cfg)
	log.Println("✅ Currency handler initialized")

	// White Label Handler (using new handler)
	whitelabelHandler := handlers.NewWhiteLabelHandler(db)
	log.Println("✅ White Label handler initialized")

	// Scheduled Report Handler
	scheduledReportHandler := handlers.NewScheduledReportHandler(db)
	log.Println("✅ Scheduled Report handler initialized")

	// Staff Role Handler
	staffRoleHandler := handlers.NewStaffRoleHandler(db)
	log.Println("✅ Staff Role handler initialized")

	// Audit Log Handler
	auditLogHandler := auditloghandler.NewAuditLogHandler(auditRepo)
	log.Println("✅ Audit Log handler initialized")

	// TwoFactor Handler
	twoFactorSvc := twofactorservice.NewService()
	twoFactorHandler := twofactorhandler.NewTwoFactorHandler(twoFactorSvc)
	log.Println("✅ TwoFactor handler initialized")

	// Push Notification Handler
	pushHandler := pushhandler.NewPushNotificationHandler(db)
	log.Println("✅ Push Notification handler initialized")

	// Plan routes
	api.Get("/plans", planHandler.GetAllPlans)

	// ========== Register All Routes ==========
	routes.RegisterAllRoutes(routes.RouteConfig{
		App:                         app,
		AuthService:                 authService,
		AuthHandler:                 authHandler,
		ShopHandler:                 shopHandler,
		ProductHandler:              productHandler,
		SaleHandler:                 saleHandler,
		ReportHandler:               reportHandler,
		ExportHandler:               exportHandler,
		StaffHandler:                staffHandler,
		WebhookHandler:              webhookHandler,
		CustomerHandler:             loyaltyHandler,
		CustHandler:                 customerHandler,
		SupplierHandler:             supplierHandler,
		MpesaHandler:                mpesaHandler,
		SMSHandler:                  smsHandler,
		EmailHandler:                emailHandler,
		AIHandler:                   aiHandler,
		PrinterHandler:              printerHandler,
		QRHandler:                   qrHandler,
		BillingHandler:              billingHandler,
		AdminHandler:                adminHandler,
		APIKeyHandler:               apiKeyHandler,
		WebHandler:                  webHandler,
		PlanInfoHandler:             planHandler,
		CurrencyHandler:             currencyHandler,
		WhiteLabelHandler:           whitelabelHandler,
		ScheduledReportHandler:      scheduledReportHandler,
		StaffRoleHandler:            staffRoleHandler,
		FeatureStaffAccountsEnabled: cfg.FeatureStaffAccountsEnabled,
		FeatureMpesaEnabled:         cfg.FeatureMpesaEnabled,
		FeatureAnalyticsEnabled:     cfg.FeatureAnalyticsEnabled,
		FeatureMultipleShopsEnabled: cfg.FeatureMultipleShopsEnabled,
		FeatureWebDashboardEnabled:  cfg.FeatureWebDashboardEnabled,
		CustomerRepo:                customerRepo,
		SaleRepo:                    saleRepo,
		DB:                          db,
	})

	// ========== Register Additional Handlers ==========
	// Currency routes
	if currencyHandler != nil {
		currencyHandler.RegisterRoutes(protected)
	}

	// White Label routes (handled in routes.go via RouteConfig)

	// Audit Log routes
	if auditLogHandler != nil {
		auditLogHandler.RegisterRoutes(protected)
	}

	// TwoFactor routes
	if twoFactorHandler != nil {
		twoFactorHandler.RegisterRoutes(protected)
	}

	// Push Notification routes
	if pushHandler != nil {
		pushHandler.RegisterRoutes(protected)
	}

	// Job Scheduler routes
	jobSchedulerHandler := jobscheduler.NewJobSchedulerHandler(routes.GetJobScheduler())
	if jobSchedulerHandler != nil {
		jobSchedulerHandler.RegisterRoutes(protected)
		log.Println("✅ Job Scheduler handler initialized")
	}

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
	app.Get("/ws/*", websocket.HandleWebSocket)

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
	log.Printf("📱 WhatsApp webhook: http://dukapos.simuxtech.com/webhook/twilio")
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
