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
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mutex      sync.RWMutex
}

type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			log.Printf("WebSocket client connected. Total clients: %d", len(h.clients))

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			h.mutex.Unlock()
			log.Printf("WebSocket client disconnected. Total clients: %d", len(h.clients))

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					client.Close()
				}
			}
			h.mutex.RUnlock()
		}
	}
}

func (h *Hub) Register(client *websocket.Conn) {
	h.register <- client
}

func (h *Hub) Unregister(client *websocket.Conn) {
	h.unregister <- client
}

func (h *Hub) Broadcast(msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal WebSocket message: %v", err)
		return
	}
	h.broadcast <- data
}

func (h *Hub) SendToShop(shopID uint, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal WebSocket message: %v", err)
		return
	}
	h.broadcast <- data
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
}

func GetHub() *Hub {
	return defaultHub
}

func NotifyNewSale(shopID uint, productName string, amount float64) {
	if defaultHub == nil {
		return
	}
	defaultHub.SendToShop(shopID, Message{
		Type: "new_sale",
		Payload: map[string]interface{}{
			"product":   productName,
			"amount":    amount,
			"timestamp": time.Now().Unix(),
		},
	})
}

func NotifyLowStock(shopID uint, productName string, currentStock int) {
	if defaultHub == nil {
		return
	}
	defaultHub.SendToShop(shopID, Message{
		Type: "low_stock",
		Payload: map[string]interface{}{
			"product":       productName,
			"current_stock": currentStock,
			"timestamp":     time.Now().Unix(),
		},
	})
}

func NotifyPaymentReceived(shopID uint, amount float64, phone string) {
	if defaultHub == nil {
		return
	}
	defaultHub.SendToShop(shopID, Message{
		Type: "payment_received",
		Payload: map[string]interface{}{
			"amount":    amount,
			"phone":     phone,
			"timestamp": time.Now().Unix(),
		},
	})
}

func HandleWebSocket(c *fiber.Ctx) error {
	if !websocket.IsWebSocketUpgrade(c) {
		return c.Status(http.StatusUpgradeRequired).JSON(fiber.Map{
			"error": "WebSocket upgrade required",
		})
	}

	wsHandler := websocket.New(func(conn *websocket.Conn) {
		if defaultHub != nil {
			defaultHub.Register(conn)
			defer defaultHub.Unregister(conn)
		}

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}

			var message Message
			if err := json.Unmarshal(msg, &message); err != nil {
				log.Printf("Failed to parse WebSocket message: %v", err)
				continue
			}

			switch message.Type {
			case "ping":
				conn.WriteJSON(Message{Type: "pong"})
			case "subscribe":
				log.Printf("Client subscribed to: %v", message.Payload)
			}
		}
	})

	return wsHandler(c)
}
