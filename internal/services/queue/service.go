package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// Message represents a queued message
type Message struct {
	ID          string          `json:"id"`
	Type        string          `json:"type"` // "whatsapp", "sms", "ussd"
	Payload     json.RawMessage `json:"payload"`
	Priority    int             `json:"priority"` // Higher = more urgent
	RetryCount  int             `json:"retry_count"`
	MaxRetries  int             `json:"max_retries"`
	CreatedAt   time.Time       `json:"created_at"`
	ProcessedAt *time.Time      `json:"processed_at,omitempty"`
	Error       string          `json:"error,omitempty"`
}

// Handler processes messages from the queue
type Handler func(msg *Message) error

// Queue implements a simple message queue
type Queue struct {
	name     string
	messages []*Message
	handlers map[string]Handler
	mu       sync.RWMutex
	workers  int
	closed   bool
	ctx      context.Context
	cancel   context.CancelFunc
}

// Options for queue configuration
type Options struct {
	Name    string
	Workers int
	MaxSize int
}

// New creates a new message queue
func New(opts *Options) *Queue {
	if opts == nil {
		opts = &Options{
			Name:    "default",
			Workers: 2,
		}
	}
	if opts.Workers < 1 {
		opts.Workers = 1
	}

	ctx, cancel := context.WithCancel(context.Background())

	q := &Queue{
		name:     opts.Name,
		messages: make([]*Message, 0, 100),
		handlers: make(map[string]Handler),
		workers:  opts.Workers,
		ctx:      ctx,
		cancel:   cancel,
	}

	// Start workers
	for i := 0; i < q.workers; i++ {
		go q.worker(i)
	}

	log.Printf("📬 Queue '%s' started with %d workers", q.name, q.workers)

	return q
}

// RegisterHandler registers a message handler
func (q *Queue) RegisterHandler(msgType string, handler Handler) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.handlers[msgType] = handler
	log.Printf("📬 Registered handler for: %s", msgType)
}

// Enqueue adds a message to the queue
func (q *Queue) Enqueue(msg *Message) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return fmt.Errorf("queue is closed")
	}

	msg.ID = generateMessageID()
	msg.CreatedAt = time.Now()

	// Insert in priority order (higher priority first)
	inserted := false
	for i, m := range q.messages {
		if msg.Priority > m.Priority {
			q.messages = append(q.messages[:i], append([]*Message{msg}, q.messages[i:]...)...)
			inserted = true
			break
		}
	}
	if !inserted {
		q.messages = append(q.messages, msg)
	}

	log.Printf("📬 Enqueued message: %s (type: %s, priority: %d)", msg.ID, msg.Type, msg.Priority)
	return nil
}

// EnqueueWithPayload enqueues a message with JSON payload
func (q *Queue) EnqueueWithPayload(msgType string, payload interface{}, priority int) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	msg := &Message{
		Type:       msgType,
		Payload:    payloadBytes,
		Priority:   priority,
		MaxRetries: 3,
	}

	return q.Enqueue(msg)
}

// Dequeue removes and returns the next message
func (q *Queue) Dequeue() *Message {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.messages) == 0 {
		return nil
	}

	msg := q.messages[0]
	q.messages = q.messages[1:]
	return msg
}

// Size returns the number of messages in queue
func (q *Queue) Size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.messages)
}

// Close shuts down the queue
func (q *Queue) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.closed = true
	q.cancel()
	log.Printf("📬 Queue '%s' closed", q.name)
}

// worker processes messages from the queue
func (q *Queue) worker(id int) {
	log.Printf("📬 Worker %d started for queue '%s'", id, q.name)

	for {
		select {
		case <-q.ctx.Done():
			log.Printf("📬 Worker %d stopping", id)
			return
		default:
			msg := q.Dequeue()
			if msg == nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			q.processMessage(msg)
		}
	}
}

func (q *Queue) processMessage(msg *Message) {
	log.Printf("📬 Processing message: %s (type: %s)", msg.ID, msg.Type)

	q.mu.RLock()
	handler, exists := q.handlers[msg.Type]
	q.mu.RUnlock()

	if !exists {
		log.Printf("⚠️ No handler for message type: %s", msg.Type)
		return
	}

	err := handler(msg)
	now := time.Now()

	if err != nil {
		log.Printf("❌ Message %s failed: %v", msg.ID, err)
		msg.Error = err.Error()
		msg.RetryCount++

		if msg.RetryCount < msg.MaxRetries {
			// Re-queue with lower priority
			msg.Priority = msg.Priority / 2
			q.Enqueue(msg)
			log.Printf("📬 Message %s re-queued (retry %d/%d)", msg.ID, msg.RetryCount, msg.MaxRetries)
		} else {
			log.Printf("❌ Message %s failed after %d retries", msg.ID, msg.MaxRetries)
		}
	} else {
		msg.ProcessedAt = &now
		log.Printf("✅ Message %s processed successfully", msg.ID)
	}
}

// Stats returns queue statistics
func (q *Queue) Stats() map[string]interface{} {
	q.mu.RLock()
	defer q.mu.RUnlock()

	handlers := make([]string, 0, len(q.handlers))
	for t := range q.handlers {
		handlers = append(handlers, t)
	}

	return map[string]interface{}{
		"name":       q.name,
		"queue_size": len(q.messages),
		"workers":    q.workers,
		"registered": handlers,
		"closed":     q.closed,
	}
}

func generateMessageID() string {
	return fmt.Sprintf("msg_%d_%d", time.Now().UnixNano(), time.Now().Nanosecond()%1000)
}
