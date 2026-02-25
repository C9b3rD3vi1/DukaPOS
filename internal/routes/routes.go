package routes

import (
	"github.com/gofiber/fiber/v2"

	"gorm.io/gorm"

	aihandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/ai"
	apihandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/api"
	billinghandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/billing"
	currencyhandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/currency"
	emailhandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/email"
	exporthandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/export"
	loyaltyhandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/loyalty"
	mpesahandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/mpesa"
	printerhandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/printer"
	qrhandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/qr"
	smshandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/sms"
	staffhandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/staff"
	supplierhandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/supplier"
	webhookhandler "github.com/C9b3rD3vi1/DukaPOS/internal/handlers/webhook"

	"github.com/C9b3rD3vi1/DukaPOS/internal/handlers"
	"github.com/C9b3rD3vi1/DukaPOS/internal/middleware"
	"github.com/C9b3rD3vi1/DukaPOS/internal/repository"
	"github.com/C9b3rD3vi1/DukaPOS/internal/services"
)

type RouteConfig struct {
	App                         *fiber.App
	AuthService                 *services.AuthService
	AuthHandler                 *handlers.AuthHandler
	ShopHandler                 *handlers.ShopHandler
	ProductHandler              *handlers.ProductHandler
	SaleHandler                 *handlers.SaleHandler
	ReportHandler               *handlers.ReportHandler
	ExportHandler               *exporthandler.ExportHandler
	StaffHandler                *staffhandler.Handler
	WebhookHandler              *webhookhandler.Handler
	CustomerHandler             *loyaltyhandler.Handler
	CustHandler                 *handlers.CustomerHandler
	SupplierHandler             *supplierhandler.Handler
	MpesaHandler                *mpesahandler.Handler
	SMSHandler                  *smshandler.Handler
	EmailHandler                *emailhandler.Handler
	AIHandler                   *aihandler.Handler
	PrinterHandler              *printerhandler.Handler
	QRHandler                   *qrhandler.QRHandler
	BillingHandler              *billinghandler.Handler
	AdminHandler                *handlers.AdminHandler
	APIKeyHandler               *apihandler.APIKeyHandler
	WebHandler                  *handlers.WebHandler
	PlanInfoHandler             *middleware.PlanInfoHandler
	ScheduledReportHandler      *handlers.ScheduledReportHandler
	StaffRoleHandler            *handlers.StaffRoleHandler
	WhiteLabelHandler           *handlers.WhiteLabelHandler
	CurrencyHandler             *currencyhandler.Handler
	FeatureStaffAccountsEnabled bool
	FeatureMpesaEnabled         bool
	FeatureAnalyticsEnabled     bool
	FeatureMultipleShopsEnabled bool
	FeatureWebDashboardEnabled  bool
	CustomerRepo                *repository.CustomerRepository
	SaleRepo                    *repository.SaleRepository
	DB                          *gorm.DB
}

func RegisterAllRoutes(config RouteConfig) {
	api := config.App.Group("/api")

	// Auth routes (public)
	auth := api.Group("/auth")
	auth.Post("/register", config.AuthHandler.Register)
	auth.Post("/login", config.AuthHandler.Login)
	auth.Post("/otp/send", config.AuthHandler.SendOTP)
	auth.Post("/otp/verify", config.AuthHandler.VerifyOTP)

	// Plan routes
	api.Get("/subscriptions/plans", config.PlanInfoHandler.GetAllPlans)

	// Protected routes
	protected := config.App.Group("/api/v1")
	protected.Use(middleware.JWT(config.AuthService))

	// 2FA status (protected)
	protected.Get("/auth/2fa/status", config.AuthHandler.GetTwoFactorStatus)

	// Shop routes
	protected.Get("/shop/profile", config.ShopHandler.GetProfile)
	protected.Put("/shop/profile", config.ShopHandler.UpdateProfile)
	protected.Get("/shop/dashboard", config.ShopHandler.GetDashboard)
	protected.Get("/shop/account", config.ShopHandler.GetAccount)
	protected.Get("/plan", config.PlanInfoHandler.GetPlanInfo)

	// Shops list (for shop switcher)
	protected.Get("/shops", config.ShopHandler.ListShops)

	// Product routes
	protected.Get("/products", config.ProductHandler.ListProducts)
	protected.Get("/products/:id", config.ProductHandler.GetProduct)
	protected.Post("/products", config.ProductHandler.CreateProduct)
	protected.Put("/products/:id", config.ProductHandler.UpdateProduct)
	protected.Delete("/products/:id", config.ProductHandler.DeleteProduct)
	protected.Post("/products/bulk", config.ProductHandler.BulkCreateProducts)
	protected.Get("/products/categories", config.ProductHandler.ListCategories)
	protected.Post("/products/categories", config.ProductHandler.CreateCategory)
	protected.Put("/products/categories/:id", config.ProductHandler.UpdateCategory)
	protected.Delete("/products/categories/:id", config.ProductHandler.DeleteCategory)

	// Sale routes
	protected.Get("/sales", config.SaleHandler.ListSales)
	protected.Get("/sales/:id", config.SaleHandler.GetSale)
	protected.Post("/sales", config.SaleHandler.CreateSale)

	// Report routes
	protected.Get("/reports", config.ReportHandler.GetDailyReport)
	protected.Get("/reports/daily", config.ReportHandler.GetDailyReport)
	protected.Get("/reports/weekly", config.ReportHandler.GetWeeklyReport)
	protected.Get("/reports/monthly", config.ReportHandler.GetMonthlyReport)
	protected.Get("/reports/analytics", config.ReportHandler.GetAnalytics)

	// Export routes
	protected.Get("/export/products", config.ExportHandler.ExportProducts)
	protected.Get("/export/sales", config.ExportHandler.ExportSales)
	protected.Get("/export/report", config.ExportHandler.ExportReport)
	protected.Get("/export/inventory", config.ExportHandler.ExportInventory)

	// Admin routes
	admin := protected.Group("/admin")
	admin.Get("/dashboard", config.AdminHandler.Dashboard)
	admin.Get("/accounts", config.AdminHandler.GetAccounts)
	admin.Get("/accounts/:id", config.AdminHandler.GetAccount)
	admin.Put("/accounts/:id/plan", config.AdminHandler.UpdateAccountPlan)
	admin.Put("/accounts/:id/status", config.AdminHandler.UpdateAccountStatus)
	admin.Get("/shops", config.AdminHandler.GetShops)
	admin.Get("/revenue", config.AdminHandler.GetRevenueStats)
	admin.Post("/upgrade-all", config.AdminHandler.UpgradeAllAccounts)

	// Public admin fix
	api.Post("/admin/fix", config.AdminHandler.FixAdmin)

	// Billing routes
	billing := protected.Group("/billing")
	billing.Get("/plans", config.BillingHandler.GetPlans)
	billing.Get("/current", config.BillingHandler.GetCurrentPlan)
	billing.Post("/upgrade", config.BillingHandler.UpgradePlan)
	billing.Get("/history", config.BillingHandler.GetHistory)

	// Subscription routes
	subs := protected.Group("/subscriptions")
	subs.Get("/plans", config.BillingHandler.GetPlans)
	subs.Get("/current", config.BillingHandler.GetCurrentPlan)
	subs.Post("/upgrade", config.BillingHandler.UpgradePlan)

	// Web Dashboard routes
	if config.FeatureWebDashboardEnabled {
		webAPI := config.App.Group("/api/v1")
		webAPI.Get("/shop/dashboard-json/:shop_id", config.WebHandler.DashboardJSON)
		webAPI.Get("/shop/dashboard/:shop_id", config.WebHandler.Dashboard)
		webAPI.Get("/products/categories", config.ProductHandler.ListCategories)
		webAPI.Post("/products/bulk", config.ProductHandler.BulkCreateProducts)
		webAPI.Post("/products", config.WebHandler.APIProductCreate)
		webAPI.Get("/products", config.ProductHandler.ListProducts)
		webAPI.Get("/products/:id", config.ProductHandler.GetProduct)
		webAPI.Put("/products/:id", config.WebHandler.APIProductUpdate)
		webAPI.Delete("/products/:id", config.WebHandler.APIProductDelete)
		webAPI.Get("/sales/:shop_id", config.WebHandler.APISales)
		webAPI.Post("/sales", config.WebHandler.APISaleCreate)
		webAPI.Get("/reports/:shop_id", config.WebHandler.APIReports)
	}

	// Staff Routes
	if config.FeatureStaffAccountsEnabled && config.StaffHandler != nil {
		staff := protected.Group("/staff")
		staff.Get("/", config.StaffHandler.List)
		staff.Get("/:id", config.StaffHandler.Get)
		staff.Post("/", config.StaffHandler.Create)
		staff.Put("/:id", config.StaffHandler.Update)
		staff.Delete("/:id", config.StaffHandler.Delete)
		staff.Put("/:id/pin", config.StaffHandler.UpdatePin)
	}

	// Customer/Loyalty Routes - Require Pro plan
	if config.CustomerHandler != nil {
		loyalty := protected.Group("/loyalty")
		loyalty.Use(middleware.RequireFeature(middleware.FeatureLoyalty))
		config.CustomerHandler.RegisterRoutes(loyalty)
	}

	// Customer CRUD Routes - Require Pro plan
	if config.CustHandler != nil {
		customers := protected.Group("/customers")
		customers.Use(middleware.RequireFeature(middleware.FeatureLoyalty))
		customers.Get("/", config.CustHandler.List)
		customers.Get("/:id", config.CustHandler.Get)
		customers.Post("/", config.CustHandler.Create)
		customers.Put("/:id", config.CustHandler.Update)
		customers.Delete("/:id", config.CustHandler.Delete)
	}

	// Supplier/Order Routes
	if config.SupplierHandler != nil {
		suppliers := protected.Group("/suppliers")
		suppliers.Get("/", config.SupplierHandler.ListSuppliers)
		suppliers.Post("/", config.SupplierHandler.CreateSupplier)
		suppliers.Get("/:id", config.SupplierHandler.GetSupplier)
		suppliers.Put("/:id", config.SupplierHandler.UpdateSupplier)
		suppliers.Delete("/:id", config.SupplierHandler.DeleteSupplier)

		orders := protected.Group("/orders")
		orders.Get("/", config.SupplierHandler.ListOrders)
		orders.Post("/", config.SupplierHandler.CreateOrder)
		orders.Get("/:id", config.SupplierHandler.GetOrder)
		orders.Put("/:id/status", config.SupplierHandler.UpdateOrderStatus)
		orders.Delete("/:id", config.SupplierHandler.DeleteOrder)
	}

	// M-Pesa Routes - Require Pro plan
	if config.FeatureMpesaEnabled && config.MpesaHandler != nil {
		mpesa := protected.Group("/mpesa")
		mpesa.Use(middleware.RequireFeature(middleware.FeatureMpesa))
		mpesa.Post("/stk-push", config.MpesaHandler.STKPush)
		mpesa.Get("/status/:checkoutId", config.MpesaHandler.GetStatus)
		mpesa.Get("/payments", config.MpesaHandler.ListPayments)
		mpesa.Post("/payments/:id/retry", config.MpesaHandler.RetryPayment)
		mpesa.Get("/transactions", config.MpesaHandler.GetTransactions)
		mpesa.Get("/balance", config.MpesaHandler.GetBalance)
		mpesa.Post("/b2c", config.MpesaHandler.B2CSend)
	}

	// Webhook Routes - Require Business plan
	if config.FeatureAnalyticsEnabled && config.WebhookHandler != nil {
		webhooks := protected.Group("/webhooks")
		webhooks.Use(middleware.RequireFeature(middleware.FeatureWebhooks))
		webhooks.Get("/", config.WebhookHandler.List)
		webhooks.Get("/:id", config.WebhookHandler.Get)
		webhooks.Post("/", config.WebhookHandler.Create)
		webhooks.Put("/:id", config.WebhookHandler.Update)
		webhooks.Delete("/:id", config.WebhookHandler.Delete)
		webhooks.Post("/:id/test", config.WebhookHandler.Test)
	}

	// AI Routes - Require Business plan
	if config.FeatureAnalyticsEnabled && config.AIHandler != nil {
		ai := protected.Group("/ai")
		ai.Use(middleware.RequireFeature(middleware.FeatureAI))
		ai.Get("/predictions", config.AIHandler.GetPredictions)
		ai.Get("/trends", config.AIHandler.GetTrends)
		ai.Get("/inventory-value", config.AIHandler.GetInventoryValue)
		ai.Get("/restock", config.AIHandler.GetRestockRecommendations)
		ai.Get("/analytics", config.AIHandler.GetSalesAnalytics)
		ai.Post("/forecast", config.AIHandler.GenerateForecast)
	}

	// SMS Routes
	if config.SMSHandler != nil {
		sms := protected.Group("/sms")
		sms.Post("/send", config.SMSHandler.SendSMS)
		sms.Post("/bulk", config.SMSHandler.SendBulkSMS)
		sms.Get("/balance", config.SMSHandler.GetBalance)
		sms.Get("/history", config.SMSHandler.GetHistory)
	}

	// Email Routes
	if config.EmailHandler != nil {
		email := protected.Group("/email")
		email.Post("/send", config.EmailHandler.SendEmail)
		email.Post("/welcome", config.EmailHandler.SendWelcomeEmail)
		email.Get("/history", config.EmailHandler.GetHistory)
	}

	// Printer Routes
	if config.PrinterHandler != nil {
		print := protected.Group("/print")
		print.Get("/printers", config.PrinterHandler.GetPrinters)
		print.Post("/receipt", config.PrinterHandler.PrintReceipt)
	}

	// QR Routes - Require Pro plan
	if config.QRHandler != nil {
		qr := protected.Group("/qr")
		qr.Use(middleware.RequireFeature(middleware.FeatureQRPayments))
		qr.Post("/generate", config.QRHandler.GenerateDynamicQR)
		qr.Post("/static", config.QRHandler.GenerateStaticQR)
		qr.Get("/status/:id", config.QRHandler.GetPaymentStatus)
		qr.Post("/callback", config.QRHandler.HandleCallback)
	}

	// API Keys Routes - Require Business plan
	if config.APIKeyHandler != nil {
		keys := protected.Group("/api-keys")
		keys.Use(middleware.RequireFeature(middleware.FeatureAPIAccess))
		keys.Get("/", config.APIKeyHandler.List)
		keys.Post("/", config.APIKeyHandler.Create)
		keys.Delete("/:id", config.APIKeyHandler.Revoke)
	}

	// Loyalty Routes - Require Pro plan
	if config.CustomerHandler != nil {
		loyalty := protected.Group("/loyalty")
		loyalty.Use(middleware.RequireFeature(middleware.FeatureLoyalty))
	}

	// Currency Routes - Require Business plan
	if config.CurrencyHandler != nil {
		currency := protected.Group("/currency")
		currency.Use(middleware.RequireFeature(middleware.FeatureMultiCurrency))
		config.CurrencyHandler.RegisterRoutes(protected)
	}

	// White Label Routes - Require Business plan
	if config.WhiteLabelHandler != nil {
		whitelabel := protected.Group("/shop/whitelabel")
		whitelabel.Use(middleware.RequireBusiness())
		whitelabel.Get("/:shop_id", config.WhiteLabelHandler.Get)
		whitelabel.Put("/:shop_id", config.WhiteLabelHandler.Update)
	}

	// Scheduled Reports Routes
	if config.ScheduledReportHandler != nil {
		reports := protected.Group("/reports/scheduled")
		reports.Get("/", config.ScheduledReportHandler.List)
		reports.Get("/:id", config.ScheduledReportHandler.Get)
		reports.Post("/", config.ScheduledReportHandler.Create)
		reports.Put("/:id", config.ScheduledReportHandler.Update)
		reports.Delete("/:id", config.ScheduledReportHandler.Delete)
	}

	// Staff Roles Routes
	if config.StaffRoleHandler != nil {
		roles := protected.Group("/staff/roles")
		roles.Get("/", config.StaffRoleHandler.List)
		roles.Get("/:id", config.StaffRoleHandler.Get)
		roles.Post("/", config.StaffRoleHandler.Create)
		roles.Put("/:id", config.StaffRoleHandler.Update)
		roles.Delete("/:id", config.StaffRoleHandler.Delete)
	}
}
