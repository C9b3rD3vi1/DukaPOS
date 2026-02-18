package notification

import (
	"fmt"
	"strings"
	"time"

	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"github.com/C9b3rD3vi1/DukaPOS/internal/repository"
)

// Service handles notifications to shops
type Service struct {
	shopRepo      *repository.ShopRepository
	productRepo   *repository.ProductRepository
	saleRepo      *repository.SaleRepository
	summaryRepo   *repository.DailySummaryRepository
}

// New creates a new notification service
func New(
	shopRepo *repository.ShopRepository,
	productRepo *repository.ProductRepository,
	saleRepo *repository.SaleRepository,
	summaryRepo *repository.DailySummaryRepository,
) *Service {
	return &Service{
		shopRepo:    shopRepo,
		productRepo: productRepo,
		saleRepo:    saleRepo,
		summaryRepo: summaryRepo,
	}
}

// Notification represents a notification to send
type Notification struct {
	ShopID    uint
	Channel   string // "whatsapp", "sms", "email"
	Recipient string
	Subject   string
	Message   string
	Type      string // "low_stock", "daily_summary", "payment_received"
}

// CheckLowStock checks for low stock products and sends alerts
func (s *Service) CheckLowStock(shopID uint) ([]Notification, error) {
	products, err := s.productRepo.GetLowStock(shopID)
	if err != nil {
		return nil, err
	}

	if len(products) == 0 {
		return nil, nil
	}

	shop, err := s.shopRepo.GetByID(shopID)
	if err != nil {
		return nil, err
	}

	var items []string
	for _, p := range products {
		items = append(items, fmt.Sprintf("• %s: %d %s (min: %d)", 
			p.Name, p.CurrentStock, p.Unit, p.LowStockThreshold))
	}

	message := fmt.Sprintf(`⚠️ LOW STOCK ALERT

%s

%s`, shop.Name, strings.Join(items, "\n"))

	return []Notification{
		{
			ShopID:    shopID,
			Channel:   "whatsapp",
			Recipient: shop.Phone,
			Subject:   "Low Stock Alert",
			Message:   message,
			Type:      "low_stock",
		},
	}, nil
}

// SendDailySummary sends daily sales summary to shop owner
func (s *Service) SendDailySummary(shopID uint) (*Notification, error) {
	shop, err := s.shopRepo.GetByID(shopID)
	if err != nil {
		return nil, err
	}

	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	sales, err := s.saleRepo.GetByDateRange(shopID, today, tomorrow)
	if err != nil {
		return nil, err
	}

	var totalSales float64
	var totalProfit float64
	productCounts := make(map[string]int)

	for _, sale := range sales {
		totalSales += sale.TotalAmount
		totalProfit += sale.Profit
		productCounts[sale.Product.Name] += sale.Quantity
	}

	// Get top 5 products
	var topProducts []string
	for name, count := range productCounts {
		topProducts = append(topProducts, fmt.Sprintf("• %s: %d sold", name, count))
		if len(topProducts) >= 5 {
			break
		}
	}

	summary := fmt.Sprintf(`📊 DAILY SUMMARY
📅 %s

💰 Total Sales: KSh %.0f
📝 Transactions: %d
💵 Profit: KSh %.0f

Top Items:
%s

Have a great day!`, 
		today.Format("Mon, Jan 2, 2006"),
		totalSales,
		len(sales),
		totalProfit,
		strings.Join(topProducts, "\n"),
	)

	return &Notification{
		ShopID:    shopID,
		Channel:   "whatsapp",
		Recipient: shop.Phone,
		Subject:   "Daily Summary",
		Message:   summary,
		Type:      "daily_summary",
	}, nil
}

// NotifyPaymentReceived notifies shop of received payment
func (s *Service) NotifyPaymentReceived(shopID uint, amount float64, phone, receipt string) (*Notification, error) {
	shop, err := s.shopRepo.GetByID(shopID)
	if err != nil {
		return nil, err
	}

	message := fmt.Sprintf(`💰 PAYMENT RECEIVED!

Amount: KSh %.0f
From: %s
Receipt: %s

Thank you!`, amount, phone, receipt)

	return &Notification{
		ShopID:    shopID,
		Channel:   "whatsapp",
		Recipient: shop.Phone,
		Subject:   "Payment Received",
		Message:   message,
		Type:      "payment_received",
	}, nil
}

// NotifySale notifies shop of a new sale
func (s *Service) NotifySale(shopID uint, sale *models.Sale) (*Notification, error) {
	shop, err := s.shopRepo.GetByID(shopID)
	if err != nil {
		return nil, err
	}

	message := fmt.Sprintf(`✅ SALE RECORDED

%s x%d = KSh %.0f
Total Today: KSh %.0f

Thank you!`,
		sale.Product.Name,
		sale.Quantity,
		sale.TotalAmount,
		sale.TotalAmount,
	)

	return &Notification{
		ShopID:    shopID,
		Channel:   "whatsapp",
		Recipient: shop.Phone,
		Subject:   "Sale Recorded",
		Message:   message,
		Type:      "sale",
	}, nil
}

// NotifyStaffAction notifies shop owner of staff action
func (s *Service) NotifyStaffAction(shopID uint, staffName, action string) (*Notification, error) {
	shop, err := s.shopRepo.GetByID(shopID)
	if err != nil {
		return nil, err
	}

	message := fmt.Sprintf(`👤 STAFF ACTION

%s: %s

Time: %s`, staffName, action, time.Now().Format("15:04"))

	return &Notification{
		ShopID:    shopID,
		Channel:   "whatsapp",
		Recipient: shop.Phone,
		Subject:   "Staff Activity",
		Message:   message,
		Type:      "staff_action",
	}, nil
}

// FormatForWhatsApp formats notification for WhatsApp
func (n *Notification) FormatForWhatsApp() string {
	return n.Message
}

// FormatForSMS formats notification for SMS (truncated)
func (n *Notification) FormatForSMS() string {
	// SMS has 160 char limit, keep it short
	msg := n.Message
	if len(msg) > 150 {
		msg = msg[:147] + "..."
	}
	return msg
}
