package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan *OutgoingMessage
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
}

type Client struct {
	conn    *websocket.Conn
	shopID  uint
	userID  uint
	isAdmin bool
}

type OutgoingMessage struct {
	ShopIDs []uint
	Message Message
}

type Message struct {
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp int64       `json:"timestamp"`
}

type SubscribeMessage struct {
	Type    string `json:"type"`
	Payload struct {
		ShopID uint `json:"shop_id"`
	} `json:"payload"`
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *OutgoingMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			log.Printf("WebSocket client connected for shop %d. Total clients: %d", client.shopID, len(h.clients))

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.conn.Close()
			}
			h.mutex.Unlock()
			log.Printf("WebSocket client disconnected. Total clients: %d", len(h.clients))

		case outgoing := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				if len(outgoing.ShopIDs) == 0 || containsShop(outgoing.ShopIDs, client.shopID) {
					err := client.conn.WriteJSON(outgoing.Message)
					if err != nil {
						client.conn.Close()
					}
				}
			}
			h.mutex.RUnlock()
		}
	}
}

func containsShop(shopIDs []uint, shopID uint) bool {
	for _, id := range shopIDs {
		if id == shopID {
			return true
		}
	}
	return false
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) Broadcast(msg Message) {
	h.broadcast <- &OutgoingMessage{
		ShopIDs: []uint{},
		Message: msg,
	}
}

func (h *Hub) SendToShop(shopID uint, msg Message) {
	h.broadcast <- &OutgoingMessage{
		ShopIDs: []uint{shopID},
		Message: msg,
	}
}

func (h *Hub) SendToShops(shopIDs []uint, msg Message) {
	h.broadcast <- &OutgoingMessage{
		ShopIDs: shopIDs,
		Message: msg,
	}
}

func (h *Hub) GetShopClients(shopID uint) int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	count := 0
	for client := range h.clients {
		if client.shopID == shopID {
			count++
		}
	}
	return count
}

func (h *Hub) ClientCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.clients)
}

var defaultHub *Hub

func Init() {
	defaultHub = NewHub()
	go defaultHub.Run()
	log.Println("WebSocket hub initialized")
}

func GetHub() *Hub {
	return defaultHub
}

func NotifyNewSale(shopID uint, productName string, amount float64, quantity int) {
	if defaultHub == nil {
		return
	}
	defaultHub.SendToShop(shopID, Message{
		Type: "new_sale",
		Payload: map[string]interface{}{
			"product":   productName,
			"amount":    amount,
			"quantity":  quantity,
			"timestamp": time.Now().Unix(),
		},
		Timestamp: time.Now().Unix(),
	})
	log.Printf("WebSocket: Notified shop %d of new sale: %s - KES %.2f", shopID, productName, amount)
}

func NotifyLowStock(shopID uint, productName string, currentStock int, threshold int) {
	if defaultHub == nil {
		return
	}
	defaultHub.SendToShop(shopID, Message{
		Type: "low_stock",
		Payload: map[string]interface{}{
			"product":       productName,
			"current_stock": currentStock,
			"threshold":     threshold,
			"is_critical":   currentStock <= threshold/2,
			"timestamp":     time.Now().Unix(),
		},
		Timestamp: time.Now().Unix(),
	})
	log.Printf("WebSocket: Notified shop %d of low stock: %s - %d remaining", shopID, productName, currentStock)
}

func NotifyPaymentReceived(shopID uint, amount float64, phone string, method string) {
	if defaultHub == nil {
		return
	}
	defaultHub.SendToShop(shopID, Message{
		Type: "payment_received",
		Payload: map[string]interface{}{
			"amount":    amount,
			"phone":     phone,
			"method":    method,
			"timestamp": time.Now().Unix(),
		},
		Timestamp: time.Now().Unix(),
	})
	log.Printf("WebSocket: Notified shop %d of payment: KES %.2f via %s", shopID, amount, method)
}

func NotifyOrderUpdate(shopID uint, orderID uint, status string, items []string) {
	if defaultHub == nil {
		return
	}
	defaultHub.SendToShop(shopID, Message{
		Type: "order_update",
		Payload: map[string]interface{}{
			"order_id":  orderID,
			"status":    status,
			"items":     items,
			"timestamp": time.Now().Unix(),
		},
		Timestamp: time.Now().Unix(),
	})
}

func NotifyStockSync(shopID uint, productID uint, quantityChange int, newStock int) {
	if defaultHub == nil {
		return
	}
	defaultHub.SendToShop(shopID, Message{
		Type: "stock_sync",
		Payload: map[string]interface{}{
			"product_id":      productID,
			"quantity_change": quantityChange,
			"new_stock":       newStock,
			"timestamp":       time.Now().Unix(),
		},
		Timestamp: time.Now().Unix(),
	})
}

func NotifyLoyaltyPoints(shopID uint, customerName string, points int, action string) {
	if defaultHub == nil {
		return
	}
	defaultHub.SendToShop(shopID, Message{
		Type: "loyalty_update",
		Payload: map[string]interface{}{
			"customer":  customerName,
			"points":    points,
			"action":    action,
			"timestamp": time.Now().Unix(),
		},
		Timestamp: time.Now().Unix(),
	})
}

func NotifyStaffAction(shopID uint, staffName string, action string, details string) {
	if defaultHub == nil {
		return
	}
	defaultHub.SendToShop(shopID, Message{
		Type: "staff_action",
		Payload: map[string]interface{}{
			"staff":     staffName,
			"action":    action,
			"details":   details,
			"timestamp": time.Now().Unix(),
		},
		Timestamp: time.Now().Unix(),
	})
}

func HandleWebSocket(c *fiber.Ctx) error {
	if !websocket.IsWebSocketUpgrade(c) {
		return c.Status(http.StatusUpgradeRequired).JSON(fiber.Map{
			"error": "WebSocket upgrade required",
		})
	}

	shopID := c.Query("shop_id")
	_ = c.Query("token")

	if shopID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "shop_id is required",
		})
	}

	wsHandler := websocket.New(func(conn *websocket.Conn) {
		var client *Client

		if defaultHub != nil {
			client = &Client{
				conn:    conn,
				shopID:  parseUint(shopID),
				userID:  0,
				isAdmin: false,
			}
			defaultHub.Register(client)
			defer defaultHub.Unregister(client)

			conn.WriteJSON(Message{
				Type:      "connected",
				Payload:   map[string]interface{}{"status": "connected", "shop_id": client.shopID},
				Timestamp: time.Now().Unix(),
			})
		}

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}

			var message SubscribeMessage
			if err := json.Unmarshal(msg, &message); err != nil {
				log.Printf("Failed to parse WebSocket message: %v", err)
				continue
			}

			switch message.Type {
			case "ping":
				conn.WriteJSON(Message{
					Type:      "pong",
					Timestamp: time.Now().Unix(),
				})
			case "subscribe":
				if client != nil && message.Payload.ShopID > 0 {
					client.shopID = message.Payload.ShopID
					log.Printf("Client subscribed to shop %d", client.shopID)
					conn.WriteJSON(Message{
						Type: "subscribed",
						Payload: map[string]interface{}{
							"shop_id": client.shopID,
						},
						Timestamp: time.Now().Unix(),
					})
				}
			case "heartbeat":
				conn.WriteJSON(Message{
					Type:      "heartbeat",
					Timestamp: time.Now().Unix(),
				})
			}
		}
	})

	return wsHandler(c)
}

func parseUint(s string) uint {
	var n uint
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + uint(c-'0')
		}
	}
	return n
}
