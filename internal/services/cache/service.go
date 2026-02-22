package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheService struct {
	client *redis.Client
	cfg    *Config
}

type Config struct {
	URL      string
	Password string
	DB       int
}

type DailySummaryCache struct {
	TotalSales       float64            `json:"total_sales"`
	TotalProfit      float64            `json:"total_profit"`
	TransactionCount int                `json:"transaction_count"`
	AverageSale      float64            `json:"average_sale"`
	TopProducts      []TopProductCache  `json:"top_products"`
	ByPaymentMethod  map[string]float64 `json:"by_payment_method"`
	GeneratedAt      time.Time          `json:"generated_at"`
}

type TopProductCache struct {
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Revenue  float64 `json:"revenue"`
}

func NewCacheService(cfg *Config) (*CacheService, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.URL,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &CacheService{
		client: client,
		cfg:    cfg,
	}, nil
}

func (s *CacheService) Close() error {
	return s.client.Close()
}

func (s *CacheService) GetDailySummary(shopID uint, date time.Time) (*DailySummaryCache, error) {
	key := fmt.Sprintf("daily_summary:%d:%s", shopID, date.Format("2006-01-02"))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	data, err := s.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var summary DailySummaryCache
	if err := json.Unmarshal(data, &summary); err != nil {
		return nil, err
	}

	return &summary, nil
}

func (s *CacheService) SetDailySummary(shopID uint, date time.Time, summary *DailySummaryCache, ttl time.Duration) error {
	key := fmt.Sprintf("daily_summary:%d:%s", shopID, date.Format("2006-01-02"))

	summary.GeneratedAt = time.Now()

	data, err := json.Marshal(summary)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return s.client.Set(ctx, key, data, ttl).Err()
}

func (s *CacheService) InvalidateDailySummary(shopID uint, date time.Time) error {
	key := fmt.Sprintf("daily_summary:%d:%s", shopID, date.Format("2006-01-02"))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return s.client.Del(ctx, key).Err()
}

func (s *CacheService) GetProduct(shopID uint, productID uint) ([]byte, error) {
	key := fmt.Sprintf("product:%d:%d", shopID, productID)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return s.client.Get(ctx, key).Bytes()
}

func (s *CacheService) SetProduct(shopID uint, productID uint, data []byte, ttl time.Duration) error {
	key := fmt.Sprintf("product:%d:%d", shopID, productID)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return s.client.Set(ctx, key, data, ttl).Err()
}

func (s *CacheService) InvalidateProduct(shopID uint, productID uint) error {
	key := fmt.Sprintf("product:%d:%d", shopID, productID)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return s.client.Del(ctx, key).Err()
}

func (s *CacheService) GetShop(shopID uint) ([]byte, error) {
	key := fmt.Sprintf("shop:%d", shopID)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return s.client.Get(ctx, key).Bytes()
}

func (s *CacheService) SetShop(shopID uint, data []byte, ttl time.Duration) error {
	key := fmt.Sprintf("shop:%d", shopID)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return s.client.Set(ctx, key, data, ttl).Err()
}

func (s *CacheService) InvalidateShop(shopID uint) error {
	key := fmt.Sprintf("shop:%d", shopID)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return s.client.Del(ctx, key).Err()
}

func (s *CacheService) IsHealthy() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return s.client.Ping(ctx).Err() == nil
}
