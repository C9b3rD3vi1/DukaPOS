package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
)

// Event represents a webhook event
type Event struct {
	Type      string      `json:"type"`
	ShopID    uint       `json:"shop_id"`
	Timestamp time.Time  `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// Delivery represents a webhook delivery attempt
type Delivery struct {
	ID          uint      `json:"id"`
	EventID    string    `json:"event_id"`
	WebhookID  uint      `json:"webhook_id"`
	URL        string    `json:"url"`
	Status     string    `json:"status"` // pending, success, failed
	StatusCode int       `json:"status_code"`
	Response   string    `json:"response"`
	Attempts   int       `json:"attempts"`
	CreatedAt  time.Time `json:"created_at"`
	SentAt     *time.Time `json:"sent_at"`
}

// Service handles webhook delivery
type Service struct {
	webhookRepo  WebhookRepository
	deliveryRepo DeliveryRepository
	httpClient   *http.Client
	maxRetries  int
	timeout     time.Duration
}

// New creates a new webhook service
func New(webhookRepo WebhookRepository, deliveryRepo DeliveryRepository) *Service {
	return &Service{
		webhookRepo:  webhookRepo,
		deliveryRepo: deliveryRepo,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		maxRetries: 3,
		timeout:   30 * time.Second,
	}
}

// WebhookRepository interface
type WebhookRepository interface {
	GetByShopIDAndEvent(shopID uint, event string) ([]models.Webhook, error)
	GetActiveByShop(shopID uint) ([]models.Webhook, error)
}

// DeliveryRepository interface
type DeliveryRepository interface {
	Create(delivery *Delivery) error
	Update(delivery *Delivery) error
	GetByEventID(eventID string) ([]Delivery, error)
}

// Supported events
var SupportedEvents = map[string]string{
	"sale.created":       "A new sale is recorded",
	"sale.updated":        "A sale is updated",
	"product.created":    "A new product is added",
	"product.updated":    "A product is updated",
	"product.low_stock":  "Product stock is low",
	"product.out_of_stock": "Product is out of stock",
	"payment.completed":  "Payment received",
	"payment.failed":     "Payment failed",
	"shop.upgraded":     "Shop plan upgraded",
}

// SendEvent sends a webhook event to all subscribed webhooks
func (s *Service) SendEvent(shopID uint, eventType string, data interface{}) error {
	// Get all active webhooks for this shop subscribed to this event
	webhooks, err := s.webhookRepo.GetByShopIDAndEvent(shopID, eventType)
	if err != nil {
		return fmt.Errorf("failed to get webhooks: %w", err)
	}

	if len(webhooks) == 0 {
		return nil // No webhooks to send
	}

	// Create event
	event := Event{
		Type:      eventType,
		ShopID:    shopID,
		Timestamp: time.Now(),
		Data:      data,
	}

	// Send to each webhook asynchronously
	var wg sync.WaitGroup
	for _, webhook := range webhooks {
		wg.Add(1)
		go func(webhook models.Webhook) {
			defer wg.Done()
			s.deliver(webhook, event)
		}(webhook)
	}
	wg.Wait()

	return nil
}

// deliver sends a single webhook
func (s *Service) deliver(webhook models.Webhook, event Event) {
	delivery := &Delivery{
		EventID:   fmt.Sprintf("%d-%s-%d", event.ShopID, event.Type, time.Now().Unix()),
		WebhookID: webhook.ID,
		URL:       webhook.URL,
		Status:    "pending",
		Attempts:  0,
		CreatedAt: time.Now(),
	}

	// Create delivery record
	s.deliveryRepo.Create(delivery)

	// Prepare payload
	payload, err := json.Marshal(event)
	if err != nil {
		delivery.Status = "failed"
		delivery.Response = err.Error()
		s.deliveryRepo.Update(delivery)
		return
	}

	// Generate signature
	signature := generateSignature(payload, webhook.Secret)

	// Send request
	for attempt := 1; attempt <= s.maxRetries; attempt++ {
		delivery.Attempts = attempt

		req, err := http.NewRequest("POST", webhook.URL, bytes.NewBuffer(payload))
		if err != nil {
			delivery.Response = err.Error()
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Webhook-Signature", signature)
		req.Header.Set("X-Webhook-Event", event.Type)
		req.Header.Set("X-Webhook-ShopID", fmt.Sprintf("%d", event.ShopID))

		resp, err := s.httpClient.Do(req)
		if err != nil {
			delivery.Response = err.Error()
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}
		defer resp.Body.Close()

		delivery.StatusCode = resp.StatusCode

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			now := time.Now()
			delivery.SentAt = &now
			delivery.Status = "success"
			s.deliveryRepo.Update(delivery)
			return
		}

		delivery.Response = fmt.Sprintf("HTTP %d", resp.StatusCode)
		time.Sleep(time.Duration(attempt) * time.Second)
	}

	delivery.Status = "failed"
	s.deliveryRepo.Update(delivery)
}

// generateSignature creates HMAC signature for payload
func generateSignature(payload []byte, secret string) string {
	if secret == "" {
		return ""
	}

	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

// VerifySignature verifies webhook signature
func VerifySignature(payload []byte, signature, secret string) bool {
	expected := generateSignature(payload, secret)
	return hmac.Equal([]byte(expected), []byte(signature))
}

// FormatEventDescription returns human-readable event description
func FormatEventDescription(eventType string) string {
	if desc, ok := SupportedEvents[eventType]; ok {
		return desc
	}
	return "Unknown event"
}

// GetSupportedEvents returns all supported events
func GetSupportedEvents() map[string]string {
	return SupportedEvents
}

// AsyncSendEvent sends event asynchronously (non-blocking)
func (s *Service) AsyncSendEvent(shopID uint, eventType string, data interface{}) {
	go func() {
		if err := s.SendEvent(shopID, eventType, data); err != nil {
			fmt.Printf("Webhook delivery error: %v\n", err)
		}
	}()
}
