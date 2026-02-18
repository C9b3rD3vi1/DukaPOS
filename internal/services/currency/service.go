package currency

import (
	"errors"
	"sync"
	"time"

	"github.com/C9b3rD3vi1/DukaPOS/internal/config"
	"gorm.io/gorm"
)

type Service struct {
	db            *gorm.DB
	config        *config.Config
	exchangeRates map[string]float64
	lastUpdated   time.Time
	mu            sync.RWMutex
}

type Currency struct {
	ID        uint    `gorm:"primaryKey"`
	Code      string  `gorm:"size:3;uniqueIndex"`
	Name      string  `gorm:"size:50"`
	Symbol    string  `gorm:"size:5"`
	Rate      float64 `json:"rate"` // Rate to KES
	IsDefault bool    `gorm:"default:false"`
	IsActive  bool    `gorm:"default:true"`
	UpdatedAt time.Time
}

type ExchangeRate struct {
	ID           uint    `gorm:"primaryKey"`
	FromCurrency string  `gorm:"size:3;index"`
	ToCurrency   string  `gorm:"size:3;index"`
	Rate         float64 `json:"rate"`
	Source       string  `gorm:"size:50"`
	FetchedAt    time.Time
}

func NewService(db *gorm.DB, cfg *config.Config) *Service {
	svc := &Service{
		db:            db,
		config:        cfg,
		exchangeRates: make(map[string]float64),
	}

	svc.initDefaultCurrencies()
	return svc
}

func (s *Service) initDefaultCurrencies() {
	defaults := []Currency{
		{Code: "KES", Name: "Kenyan Shilling", Symbol: "KSh", Rate: 1.0, IsDefault: true},
		{Code: "USD", Name: "US Dollar", Symbol: "$", Rate: 0.0087},
		{Code: "EUR", Name: "Euro", Symbol: "€", Rate: 0.0080},
		{Code: "GBP", Name: "British Pound", Symbol: "£", Rate: 0.0069},
		{Code: "UGX", Name: "Ugandan Shilling", Symbol: "USh", Rate: 32.0},
		{Code: "TZS", Name: "Tanzanian Shilling", Symbol: "TSh", Rate: 20.0},
		{Code: "RWF", Name: "Rwandan Franc", Symbol: "FRw", Rate: 8.5},
	}

	for _, c := range defaults {
		var existing Currency
		if err := s.db.Where("code = ?", c.Code).First(&existing).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			s.db.Create(&c)
		}
		s.exchangeRates[c.Code] = c.Rate
	}
	s.lastUpdated = time.Now()
}

func (s *Service) Convert(amount float64, from, to string) (float64, error) {
	s.mu.RLock()
	fromRate, ok1 := s.exchangeRates[from]
	toRate, ok2 := s.exchangeRates[to]
	s.mu.RUnlock()

	if !ok1 {
		return 0, CurrencyError("unknown currency: " + from)
	}
	if !ok2 {
		return 0, CurrencyError("unknown currency: " + to)
	}

	kesAmount := amount * fromRate
	return kesAmount / toRate, nil
}

func (s *Service) Format(amount float64, currency string) string {
	s.mu.RLock()
	c, ok := s.exchangeRates[currency]
	s.mu.RUnlock()

	if !ok {
		currency = "KES"
		c = 1
	}

	symbol := currency
	switch currency {
	case "KES":
		symbol = "KSh"
	case "USD", "EUR", "GBP":
		symbol = "$"
	case "UGX", "TZS", "RWF":
		symbol = currency
	}

	return symbol + formatNumber(amount/c)
}

func (s *Service) GetCurrency(code string) (*Currency, error) {
	var c Currency
	if err := s.db.Where("code = ?", code).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Service) ListCurrencies() ([]Currency, error) {
	var currencies []Currency
	err := s.db.Where("is_active = ?", true).Find(&currencies).Error
	return currencies, err
}

func (s *Service) SetDefault(code string) error {
	s.db.Model(&Currency{}).Where("1=1").Update("is_default", false)
	return s.db.Model(&Currency{}).Where("code = ?", code).Update("is_default", true).Error
}

func formatNumber(n float64) string {
	return ""
}

type CurrencyError string

func (e CurrencyError) Error() string { return string(e) }
