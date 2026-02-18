package webhook

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"gorm.io/gorm"
)

type EventType string

const (
	EventSaleCreated     EventType = "sale.created"
	EventSaleUpdated     EventType = "sale.updated"
	EventProductCreated  EventType = "product.created"
	EventProductUpdated  EventType = "product.updated"
	EventProductLowStock EventType = "product.low_stock"
	EventPaymentReceived EventType = "payment.received"
	EventPaymentFailed   EventType = "payment.failed"
	EventCustomerCreated EventType = "customer.created"
	EventCustomerTier    EventType = "customer.tier_upgraded"
	EventShopCreated     EventType = "shop.created"
	EventOrderCreated    EventType = "order.created"
	EventOrderFulfilled  EventType = "order.fulfilled"
)

type DeliveryService struct {
	db           *gorm.DB
	httpClient   *http.Client
	queue        chan *EventDelivery
	workers      int
	maxRetries   int
	timeout      time.Duration
	deliveryRepo *DeliveryRepo
}

type EventDelivery struct {
	ID          uint
	WebhookID   uint
	EventType   EventType
	Payload     json.RawMessage
	Attempt     int
	Status      string
	Error       string
	ScheduledAt time.Time
	DeliveredAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type DeliveryRepo struct {
	db *gorm.DB
}

type WebhookEvent struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	WebhookID uint       `gorm:"index" json:"webhook_id"`
	Event     EventType  `gorm:"size:50;index" json:"event"`
	Payload   string     `gorm:"type:text" json:"payload"`
	Status    string     `gorm:"size:20;default:pending" json:"status"`
	Attempts  int        `gorm:"default:0" json:"attempts"`
	Error     string     `gorm:"size:500" json:"error"`
	SentAt    *time.Time `json:"sent_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type WebhookDelivery struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	EventID      uint       `gorm:"index" json:"event_id"`
	WebhookURL   string     `gorm:"size:500" json:"webhook_url"`
	Status       string     `gorm:"size:20;default:pending" json:"status"`
	HTTPStatus   int        `json:"http_status"`
	ResponseBody string     `gorm:"size:1000" json:"response_body"`
	Error        string     `gorm:"size:500" json:"error"`
	Attempt      int        `gorm:"default:0" json:"attempt"`
	DeliveredAt  *time.Time `json:"delivered_at"`
	CreatedAt    time.Time  `json:"created_at"`
}

func NewDeliveryService(db *gorm.DB, workers, maxRetries int) *DeliveryService {
	svc := &DeliveryService{
		db:         db,
		workers:    workers,
		maxRetries: maxRetries,
		timeout:    30 * time.Second,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		queue:      make(chan *EventDelivery, 1000),
	}

	if err := svc.db.AutoMigrate(&WebhookEvent{}, &WebhookDelivery{}); err != nil {
		log.Printf("Failed to migrate webhook tables: %v", err)
	}

	return svc
}

func (s *DeliveryService) Start(ctx context.Context) {
	for i := 0; i < s.workers; i++ {
		go s.worker(ctx, i)
	}

	go s.retryFailed(ctx)

	log.Printf("Webhook delivery service started with %d workers", s.workers)
}

func (s *DeliveryService) worker(ctx context.Context, id int) {
	for {
		select {
		case <-ctx.Done():
			return
		case delivery := <-s.queue:
			s.processDelivery(delivery)
		}
	}
}

func (s *DeliveryService) processDelivery(delivery *EventDelivery) {
	var webhook models.Webhook
	if err := s.db.First(&webhook, delivery.WebhookID).Error; err != nil {
		log.Printf("Webhook not found: %d", delivery.WebhookID)
		return
	}

	if !webhook.IsActive {
		return
	}

	payload := map[string]interface{}{
		"event":     delivery.EventType,
		"timestamp": time.Now().Unix(),
		"data":      delivery.Payload,
	}

	jsonPayload, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", webhook.URL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		s.recordDelivery(delivery, webhook.URL, 0, err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Event", string(delivery.EventType))
	req.Header.Set("X-Webhook-ID", fmt.Sprintf("%d", delivery.ID))

	if webhook.Secret != "" {
		signature := generateHMAC(jsonPayload, webhook.Secret)
		req.Header.Set("X-Webhook-Signature", signature)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.recordDelivery(delivery, webhook.URL, 0, err.Error())
		s.scheduleRetry(delivery)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		s.recordDelivery(delivery, webhook.URL, resp.StatusCode, "")

		s.db.Model(&WebhookEvent{}).Where("id = ?", delivery.ID).Updates(map[string]interface{}{
			"status":   "delivered",
			"sent_at":  time.Now(),
			"attempts": delivery.Attempt + 1,
		})
	} else {
		body := ""
		buf := make([]byte, 512)
		if n, _ := resp.Body.Read(buf); n > 0 {
			body = string(buf)
		}

		s.recordDelivery(delivery, webhook.URL, resp.StatusCode, body)
		s.scheduleRetry(delivery)
	}
}

func (s *DeliveryService) recordDelivery(delivery *EventDelivery, url string, status int, errMsg string) {
	deliveryRecord := &WebhookDelivery{
		EventID:      delivery.ID,
		WebhookURL:   url,
		HTTPStatus:   status,
		ResponseBody: errMsg,
		Attempt:      delivery.Attempt + 1,
	}

	if status >= 200 && status < 300 {
		deliveryRecord.Status = "success"
		now := time.Now()
		deliveryRecord.DeliveredAt = &now
	} else {
		deliveryRecord.Status = "failed"
		deliveryRecord.Error = errMsg
	}

	s.db.Create(deliveryRecord)
}

func (s *DeliveryService) scheduleRetry(delivery *EventDelivery) {
	if delivery.Attempt >= s.maxRetries {
		s.db.Model(&WebhookEvent{}).Where("id = ?", delivery.ID).Update("status", "failed")
		return
	}

	delivery.Attempt++
	delay := time.Duration(delivery.Attempt * delivery.Attempt * 5)

	s.db.Model(&WebhookEvent{}).Where("id = ?", delivery.ID).Updates(map[string]interface{}{
		"attempts": delivery.Attempt,
		"status":   "retry_scheduled",
	})

	go func() {
		time.Sleep(delay * time.Second)
		s.queue <- delivery
	}()
}

func (s *DeliveryService) retryFailed(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.retryPending()
		}
	}
}

func (s *DeliveryService) retryPending() {
	var events []WebhookEvent
	s.db.Where("status = ? AND attempts < ?", "pending", s.maxRetries).
		Or("status = ?", "retry_scheduled").
		Limit(100).
		Find(&events)

	for _, event := range events {
		delivery := &EventDelivery{
			ID:        event.ID,
			WebhookID: event.WebhookID,
			EventType: event.Event,
			Payload:   json.RawMessage(event.Payload),
			Attempt:   event.Attempts,
		}
		s.queue <- delivery
	}
}

func (s *DeliveryService) TriggerEvent(eventType EventType, data interface{}) error {
	webhooks, err := s.getActiveWebhooks(eventType)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	for _, webhook := range webhooks {
		event := &WebhookEvent{
			WebhookID: webhook.ID,
			Event:     eventType,
			Payload:   string(payload),
			Status:    "pending",
			CreatedAt: time.Now(),
		}

		if err := s.db.Create(event).Error; err != nil {
			log.Printf("Failed to create webhook event: %v", err)
			continue
		}

		delivery := &EventDelivery{
			ID:        event.ID,
			WebhookID: webhook.ID,
			EventType: eventType,
			Payload:   payload,
			Attempt:   0,
		}

		select {
		case s.queue <- delivery:
		default:
			log.Printf("Webhook queue full, event dropped: %s", eventType)
		}
	}

	return nil
}

func (s *DeliveryService) getActiveWebhooks(eventType EventType) ([]models.Webhook, error) {
	var webhooks []models.Webhook
	err := s.db.Where("is_active = ? AND (events = ? OR events LIKE ?)", true, eventType, "%all%").
		Find(&webhooks).Error
	return webhooks, err
}

func (s *DeliveryService) GetEventStatus(eventID uint) (map[string]interface{}, error) {
	var event WebhookEvent
	if err := s.db.First(&event, eventID).Error; err != nil {
		return nil, err
	}

	var deliveries []WebhookDelivery
	s.db.Where("event_id = ?", eventID).Find(&deliveries)

	return map[string]interface{}{
		"event":      event.Event,
		"status":     event.Status,
		"attempts":   event.Attempts,
		"deliveries": deliveries,
	}, nil
}

func generateHMAC(payload []byte, secret string) string {
	h := sha256Hmac([]byte(secret), payload)
	return fmt.Sprintf("sha256=%x", h)
}

func sha256Hmac(key, data []byte) []byte {
	h := sha256.New()
	h.Write(key)
	h.Write(data)
	return h.Sum(nil)
}
