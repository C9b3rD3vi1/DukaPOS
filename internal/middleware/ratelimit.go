package middleware

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// TokenRateLimiter implements a token bucket rate limiter
type TokenRateLimiter struct {
	requests map[string]*bucket
	mu       sync.RWMutex
	rate     int
	window   time.Duration
}

type bucket struct {
	tokens    int
	lastReset time.Time
}

// NewTokenRateLimiter creates a new rate limiter
func NewTokenRateLimiter(requests int, window time.Duration) *TokenRateLimiter {
	rl := &TokenRateLimiter{
		requests: make(map[string]*bucket),
		rate:     requests,
		window:   window,
	}
	
	// Cleanup old entries periodically
	go rl.cleanup()
	
	return rl
}

// RateLimiter returns Fiber middleware for rate limiting
func RateLimiter(maxRequests int, windowSeconds int) fiber.Handler {
	rl := NewTokenRateLimiter(maxRequests, time.Duration(windowSeconds)*time.Second)
	
	return func(c *fiber.Ctx) error {
		key := rl.getKey(c)
		
		if !rl.allow(key) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "rate limit exceeded",
				"retry_after": windowSeconds,
			})
		}
		
		return c.Next()
	}
}

// getKey generates a unique key for the client
func (rl *TokenRateLimiter) getKey(c *fiber.Ctx) string {
	// Use API key if present, otherwise use IP
	if apiKey := c.Get("X-API-Key"); apiKey != "" {
		return fmt.Sprintf("api:%s", apiKey)
	}
	return fmt.Sprintf("ip:%s", c.IP())
}

// allow checks if request is allowed
func (rl *TokenRateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	b, exists := rl.requests[key]
	now := time.Now()
	
	if !exists {
		rl.requests[key] = &bucket{
			tokens:    rl.rate - 1,
			lastReset: now,
		}
		return true
	}
	
	// Reset tokens if window has passed
	if now.Sub(b.lastReset) >= rl.window {
		b.tokens = rl.rate - 1
		b.lastReset = now
		return true
	}
	
	// Check if tokens available
	if b.tokens > 0 {
		b.tokens--
		return true
	}
	
	return false
}

// cleanup removes old entries
func (rl *TokenRateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, b := range rl.requests {
			if now.Sub(b.lastReset) > rl.window*2 {
				delete(rl.requests, key)
			}
		}
		rl.mu.Unlock()
	}
}

// GetRemaining returns remaining requests for a key
func (rl *TokenRateLimiter) GetRemaining(c *fiber.Ctx) int {
	key := rl.getKey(c)
	
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	
	b, exists := rl.requests[key]
	if !exists {
		return rl.rate
	}
	
	now := time.Now()
	if now.Sub(b.lastReset) >= rl.window {
		return rl.rate
	}
	
	return b.tokens
}
