package webhook

import (
	"log"
	"sync"

	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"gorm.io/gorm"
)

var (
	defaultService *Manager
	once           sync.Once
)

// Manager wraps the delivery service for global access
type Manager struct {
	db          *gorm.DB
	deliverySvc *DeliveryService
	enabled     bool
	mu          sync.RWMutex
}

// Init initializes the global webhook manager
func Init(db *gorm.DB, workers, maxRetries int) {
	once.Do(func() {
		defaultService = &Manager{
			db:      db,
			enabled: true,
		}

		if workers > 0 {
			defaultService.deliverySvc = NewDeliveryService(db, workers, maxRetries)
			log.Println("Webhook manager initialized with delivery service")
		} else {
			log.Println("Webhook manager initialized (delivery disabled)")
		}
	})
}

// GetManager returns the global webhook manager
func GetManager() *Manager {
	return defaultService
}

// TriggerSaleCreated triggers a sale.created event
func (m *Manager) TriggerSaleCreated(sale *models.Sale, product *models.Product) {
	if !m.enabled || m.deliverySvc == nil {
		return
	}

	data := map[string]interface{}{
		"id":             sale.ID,
		"shop_id":        sale.ShopID,
		"product_id":     sale.ProductID,
		"product_name":   product.Name,
		"quantity":       sale.Quantity,
		"unit_price":     sale.UnitPrice,
		"total_amount":   sale.TotalAmount,
		"profit":         sale.Profit,
		"payment_method": sale.PaymentMethod,
		"created_at":     sale.CreatedAt,
	}

	if err := m.deliverySvc.TriggerEvent(EventSaleCreated, data); err != nil {
		log.Printf("Failed to trigger sale.created event: %v", err)
	}
}

// TriggerProductLowStock triggers a product.low_stock event
func (m *Manager) TriggerProductLowStock(product *models.Product) {
	if !m.enabled || m.deliverySvc == nil {
		return
	}

	data := map[string]interface{}{
		"id":                  product.ID,
		"shop_id":             product.ShopID,
		"name":                product.Name,
		"current_stock":       product.CurrentStock,
		"low_stock_threshold": product.LowStockThreshold,
	}

	if err := m.deliverySvc.TriggerEvent(EventProductLowStock, data); err != nil {
		log.Printf("Failed to trigger product.low_stock event: %v", err)
	}
}

// TriggerPaymentReceived triggers a payment.received event
func (m *Manager) TriggerPaymentReceived(sale *models.Sale, product *models.Product, phone string) {
	if !m.enabled || m.deliverySvc == nil {
		return
	}

	data := map[string]interface{}{
		"id":            sale.ID,
		"shop_id":       sale.ShopID,
		"product_id":    sale.ProductID,
		"product_name":  product.Name,
		"amount":        sale.TotalAmount,
		"phone":         phone,
		"mpesa_receipt": sale.MpesaReceipt,
		"created_at":    sale.CreatedAt,
	}

	if err := m.deliverySvc.TriggerEvent(EventPaymentReceived, data); err != nil {
		log.Printf("Failed to trigger payment.received event: %v", err)
	}
}

// TriggerCustomerCreated triggers a customer.created event
func (m *Manager) TriggerCustomerCreated(customer *models.Customer) {
	if !m.enabled || m.deliverySvc == nil {
		return
	}

	data := map[string]interface{}{
		"id":         customer.ID,
		"shop_id":    customer.ShopID,
		"name":       customer.Name,
		"phone":      customer.Phone,
		"email":      customer.Email,
		"tier":       customer.Tier,
		"points":     customer.LoyaltyPoints,
		"created_at": customer.CreatedAt,
	}

	if err := m.deliverySvc.TriggerEvent(EventCustomerCreated, data); err != nil {
		log.Printf("Failed to trigger customer.created event: %v", err)
	}
}

// TriggerOrderCreated triggers an order.created event
func (m *Manager) TriggerOrderCreated(order *models.Order, supplier *models.Supplier) {
	if !m.enabled || m.deliverySvc == nil {
		return
	}

	data := map[string]interface{}{
		"id":            order.ID,
		"shop_id":       order.ShopID,
		"supplier_id":   order.SupplierID,
		"supplier_name": supplier.Name,
		"status":        order.Status,
		"total_amount":  order.TotalAmount,
		"created_at":    order.CreatedAt,
	}

	if err := m.deliverySvc.TriggerEvent(EventOrderCreated, data); err != nil {
		log.Printf("Failed to trigger order.created event: %v", err)
	}
}

// TriggerProductCreated triggers a product.created event
func (m *Manager) TriggerProductCreated(product *models.Product) {
	if !m.enabled || m.deliverySvc == nil {
		return
	}

	data := map[string]interface{}{
		"id":                  product.ID,
		"shop_id":             product.ShopID,
		"name":                product.Name,
		"category":            product.Category,
		"unit":                product.Unit,
		"cost_price":          product.CostPrice,
		"selling_price":       product.SellingPrice,
		"current_stock":       product.CurrentStock,
		"low_stock_threshold": product.LowStockThreshold,
		"barcode":             product.Barcode,
		"created_at":          product.CreatedAt,
	}

	if err := m.deliverySvc.TriggerEvent(EventProductCreated, data); err != nil {
		log.Printf("Failed to trigger product.created event: %v", err)
	}
}

// TriggerProductUpdated triggers a product.updated event
func (m *Manager) TriggerProductUpdated(product *models.Product) {
	if !m.enabled || m.deliverySvc == nil {
		return
	}

	data := map[string]interface{}{
		"id":            product.ID,
		"shop_id":       product.ShopID,
		"name":          product.Name,
		"category":      product.Category,
		"selling_price": product.SellingPrice,
		"current_stock": product.CurrentStock,
		"updated_at":    product.UpdatedAt,
	}

	if err := m.deliverySvc.TriggerEvent(EventProductUpdated, data); err != nil {
		log.Printf("Failed to trigger product.updated event: %v", err)
	}
}

// TriggerSaleUpdated triggers a sale.updated event
func (m *Manager) TriggerSaleUpdated(sale *models.Sale, product *models.Product) {
	if !m.enabled || m.deliverySvc == nil {
		return
	}

	data := map[string]interface{}{
		"id":           sale.ID,
		"shop_id":      sale.ShopID,
		"product_id":   sale.ProductID,
		"product_name": product.Name,
		"quantity":     sale.Quantity,
		"updated_at":   sale.UpdatedAt,
	}

	if err := m.deliverySvc.TriggerEvent(EventSaleUpdated, data); err != nil {
		log.Printf("Failed to trigger sale.updated event: %v", err)
	}
}

// TriggerPaymentFailed triggers a payment.failed event
func (m *Manager) TriggerPaymentFailed(sale *models.Sale, product *models.Product, reason string) {
	if !m.enabled || m.deliverySvc == nil {
		return
	}

	data := map[string]interface{}{
		"id":           sale.ID,
		"shop_id":      sale.ShopID,
		"product_id":   sale.ProductID,
		"product_name": product.Name,
		"amount":       sale.TotalAmount,
		"reason":       reason,
		"created_at":   sale.CreatedAt,
	}

	if err := m.deliverySvc.TriggerEvent(EventPaymentFailed, data); err != nil {
		log.Printf("Failed to trigger payment.failed event: %v", err)
	}
}

// TriggerCustomerTierUpgraded triggers a customer.tier_upgraded event
func (m *Manager) TriggerCustomerTierUpgraded(customer *models.Customer, oldTier, newTier string) {
	if !m.enabled || m.deliverySvc == nil {
		return
	}

	data := map[string]interface{}{
		"id":         customer.ID,
		"shop_id":    customer.ShopID,
		"name":       customer.Name,
		"phone":      customer.Phone,
		"old_tier":   oldTier,
		"new_tier":   newTier,
		"points":     customer.LoyaltyPoints,
		"updated_at": customer.UpdatedAt,
	}

	if err := m.deliverySvc.TriggerEvent(EventCustomerTier, data); err != nil {
		log.Printf("Failed to trigger customer.tier_upgraded event: %v", err)
	}
}

// TriggerOrderFulfilled triggers an order.fulfilled event
func (m *Manager) TriggerOrderFulfilled(order *models.Order, supplier *models.Supplier) {
	if !m.enabled || m.deliverySvc == nil {
		return
	}

	data := map[string]interface{}{
		"id":            order.ID,
		"shop_id":       order.ShopID,
		"supplier_id":   order.SupplierID,
		"supplier_name": supplier.Name,
		"status":        order.Status,
		"total_amount":  order.TotalAmount,
		"fulfilled_at":  order.UpdatedAt,
	}

	if err := m.deliverySvc.TriggerEvent(EventOrderFulfilled, data); err != nil {
		log.Printf("Failed to trigger order.fulfilled event: %v", err)
	}
}

// Helper functions for global access
func TriggerSaleCreated(sale *models.Sale, product *models.Product) {
	if m := GetManager(); m != nil {
		m.TriggerSaleCreated(sale, product)
	}
}

func TriggerProductLowStock(product *models.Product) {
	if m := GetManager(); m != nil {
		m.TriggerProductLowStock(product)
	}
}

func TriggerPaymentReceived(sale *models.Sale, product *models.Product, phone string) {
	if m := GetManager(); m != nil {
		m.TriggerPaymentReceived(sale, product, phone)
	}
}

func TriggerCustomerCreated(customer *models.Customer) {
	if m := GetManager(); m != nil {
		m.TriggerCustomerCreated(customer)
	}
}

func TriggerOrderCreated(order *models.Order, supplier *models.Supplier) {
	if m := GetManager(); m != nil {
		m.TriggerOrderCreated(order, supplier)
	}
}

func TriggerProductCreated(product *models.Product) {
	if m := GetManager(); m != nil {
		m.TriggerProductCreated(product)
	}
}

func TriggerProductUpdated(product *models.Product) {
	if m := GetManager(); m != nil {
		m.TriggerProductUpdated(product)
	}
}

func TriggerSaleUpdated(sale *models.Sale, product *models.Product) {
	if m := GetManager(); m != nil {
		m.TriggerSaleUpdated(sale, product)
	}
}

func TriggerPaymentFailed(sale *models.Sale, product *models.Product, reason string) {
	if m := GetManager(); m != nil {
		m.TriggerPaymentFailed(sale, product, reason)
	}
}

func TriggerCustomerTierUpgraded(customer *models.Customer, oldTier, newTier string) {
	if m := GetManager(); m != nil {
		m.TriggerCustomerTierUpgraded(customer, oldTier, newTier)
	}
}

func TriggerOrderFulfilled(order *models.Order, supplier *models.Supplier) {
	if m := GetManager(); m != nil {
		m.TriggerOrderFulfilled(order, supplier)
	}
}
