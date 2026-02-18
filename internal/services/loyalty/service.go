package loyalty

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"github.com/C9b3rD3vi1/DukaPOS/internal/repository"
	"gorm.io/gorm"
)

type Service struct {
	customerRepo *repository.CustomerRepository
	saleRepo     *repository.SaleRepository
	db           *gorm.DB
}

func NewService(customerRepo *repository.CustomerRepository, saleRepo *repository.SaleRepository, db *gorm.DB) *Service {
	return &Service{
		customerRepo: customerRepo,
		saleRepo:     saleRepo,
		db:           db,
	}
}

type EarnPointsResult struct {
	CustomerID   uint   `json:"customer_id"`
	PointsEarned int    `json:"points_earned"`
	TotalPoints  int    `json:"total_points"`
	NewTier      string `json:"new_tier"`
	TierUpgraded bool   `json:"tier_upgraded"`
}

func (s *Service) EarnPointsOnSale(saleID uint) (*EarnPointsResult, error) {
	sale, err := s.saleRepo.GetByID(saleID)
	if err != nil {
		return nil, fmt.Errorf("sale not found: %w", err)
	}

	if sale.CustomerID == nil {
		return nil, nil
	}

	customer, err := s.customerRepo.GetByID(*sale.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}

	result, err := s.AddPoints(customer.ID, sale.ShopID, sale.TotalAmount, &saleID, "Purchase")
	if err != nil {
		return nil, err
	}

	if customer.ShouldUpgradeTier() {
		oldTier := string(customer.Tier)
		customer.UpgradeTier()
		if err := s.customerRepo.Update(customer); err == nil {
			result.TierUpgraded = true
			s.createBonusTransaction(customer, oldTier, string(customer.Tier))
		}
	}

	return result, nil
}

func (s *Service) AddPoints(customerID, shopID uint, amount float64, saleID *uint, description string) (*EarnPointsResult, error) {
	customer, err := s.customerRepo.GetByID(customerID)
	if err != nil {
		return nil, err
	}

	pointsRate := s.getPointsRate(customer.Tier)
	pointsEarned := int(amount * pointsRate)

	if pointsEarned < 1 {
		pointsEarned = 1
	}

	pointsBefore := customer.LoyaltyPoints
	customer.LoyaltyPoints += pointsEarned
	customer.PointsEarned += pointsEarned
	customer.TotalSpent += amount
	customer.TotalPurchases++

	now := time.Now()
	customer.LastPurchaseAt = &now

	if err := s.customerRepo.Update(customer); err != nil {
		return nil, err
	}

	reference := s.generateReference("EP")
	transaction := &models.LoyaltyTransaction{
		CustomerID:   customerID,
		ShopID:       shopID,
		SaleID:       saleID,
		Type:         models.LoyaltyEarned,
		Points:       pointsEarned,
		PointsBefore: pointsBefore,
		PointsAfter:  customer.LoyaltyPoints,
		Amount:       amount,
		Description:  description,
		Reference:    reference,
		ExpiresAt:    s.getPointsExpiry(),
	}

	if err := s.db.Create(transaction).Error; err != nil {
		return nil, err
	}

	result := &EarnPointsResult{
		CustomerID:   customerID,
		PointsEarned: pointsEarned,
		TotalPoints:  customer.LoyaltyPoints,
		NewTier:      string(customer.Tier),
	}

	return result, nil
}

func (s *Service) RedeemPoints(customerID uint, points int, description string) (*models.LoyaltyRedemption, error) {
	if points <= 0 {
		return nil, fmt.Errorf("points must be greater than 0")
	}

	customer, err := s.customerRepo.GetByID(customerID)
	if err != nil {
		return nil, err
	}

	validPoints := s.getValidPoints(customerID)
	if validPoints < points {
		return nil, fmt.Errorf("insufficient points: have %d, need %d", validPoints, points)
	}

	pointsBefore := customer.LoyaltyPoints
	customer.LoyaltyPoints -= points
	customer.PointsRedeemed += points

	if err := s.customerRepo.Update(customer); err != nil {
		return nil, err
	}

	discountAmount := float64(points) * 0.5

	reference := s.generateReference("RP")
	transaction := &models.LoyaltyTransaction{
		CustomerID:   customerID,
		ShopID:       customer.ShopID,
		Type:         models.LoyaltyRedeemed,
		Points:       -points,
		PointsBefore: pointsBefore,
		PointsAfter:  customer.LoyaltyPoints,
		Amount:       discountAmount,
		Description:  description,
		Reference:    reference,
	}

	if err := s.db.Create(transaction).Error; err != nil {
		return nil, err
	}

	now := time.Now()
	redemption := &models.LoyaltyRedemption{
		CustomerID:     customerID,
		ShopID:         customer.ShopID,
		PointsUsed:     points,
		DiscountAmount: discountAmount,
		Status:         "completed",
		RedeemedAt:     &now,
	}

	if err := s.db.Create(redemption).Error; err != nil {
		return nil, err
	}

	return redemption, nil
}

func (s *Service) GetCustomerPoints(customerID uint) (*models.Customer, error) {
	customer, err := s.customerRepo.GetByID(customerID)
	if err != nil {
		return nil, err
	}

	customer.LoyaltyPoints = s.getValidPoints(customerID)
	return customer, nil
}

func (s *Service) getValidPoints(customerID uint) int {
	var total int
	s.db.Model(&models.LoyaltyTransaction{}).
		Where("customer_id = ? AND points > 0 AND (expires_at IS NULL OR expires_at > ?)", customerID, time.Now()).
		Select("COALESCE(SUM(points), 0)").
		Scan(&total)

	var redeemed int
	s.db.Model(&models.LoyaltyTransaction{}).
		Where("customer_id = ? AND points < 0", customerID).
		Select("COALESCE(SUM(ABS(points)), 0)").
		Scan(&redeemed)

	return total - redeemed
}

func (s *Service) getPointsRate(tier models.LoyaltyTier) float64 {
	configs := map[models.LoyaltyTier]float64{
		models.TierBronze:   1.0,
		models.TierSilver:   1.5,
		models.TierGold:     2.0,
		models.TierPlatinum: 3.0,
	}
	if rate, ok := configs[tier]; ok {
		return rate
	}
	return 1.0
}

func (s *Service) getPointsExpiry() *time.Time {
	expiry := time.Now().AddDate(1, 0, 0)
	return &expiry
}

func (s *Service) generateReference(prefix string) string {
	const digits = "0123456789"
	ref := prefix
	for i := 0; i < 8; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		ref += string(digits[n.Int64()])
	}
	return ref
}

func (s *Service) createBonusTransaction(customer *models.Customer, oldTier, newTier string) {
	bonusPoints := 0
	switch models.LoyaltyTier(newTier) {
	case models.TierSilver:
		bonusPoints = 100
	case models.TierGold:
		bonusPoints = 250
	case models.TierPlatinum:
		bonusPoints = 500
	}

	if bonusPoints > 0 {
		transaction := &models.LoyaltyTransaction{
			CustomerID:  customer.ID,
			ShopID:      customer.ShopID,
			Type:        models.LoyaltyBonus,
			Points:      bonusPoints,
			Description: fmt.Sprintf("Tier upgrade bonus: %s → %s", oldTier, newTier),
			Reference:   s.generateReference("TB"),
		}
		s.db.Create(transaction)
	}
}

func (s *Service) CreateCustomer(shopID uint, name, phone, email string) (*models.Customer, error) {
	if phone == "" && email == "" {
		return nil, fmt.Errorf("phone or email is required")
	}

	existing, _ := s.customerRepo.GetByPhone(shopID, phone)
	if existing != nil {
		return existing, nil
	}

	customer := &models.Customer{
		ShopID:        shopID,
		Name:          name,
		Phone:         phone,
		Email:         email,
		LoyaltyPoints: 100,
		Tier:          models.TierBronze,
		ReferralCode:  s.generateReferralCode(),
		IsActive:      true,
	}

	if err := s.customerRepo.Create(customer); err != nil {
		return nil, err
	}

	transaction := &models.LoyaltyTransaction{
		CustomerID:  customer.ID,
		ShopID:      shopID,
		Type:        models.LoyaltyBonus,
		Points:      100,
		Description: "Welcome bonus",
		Reference:   s.generateReference("WB"),
	}
	s.db.Create(transaction)

	return customer, nil
}

func (s *Service) generateReferralCode() string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, 6)
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		code[i] = chars[n.Int64()]
	}
	return string(code)
}

func (s *Service) GetCustomerStats(customerID uint) (map[string]interface{}, error) {
	customer, err := s.customerRepo.GetByID(customerID)
	if err != nil {
		return nil, err
	}

	validPoints := s.getValidPoints(customerID)

	var transactions []models.LoyaltyTransaction
	s.db.Where("customer_id = ?", customerID).
		Order("created_at DESC").
		Limit(10).
		Find(&transactions)

	tierConfig := models.DefaultTierConfigs[customer.Tier]

	return map[string]interface{}{
		"customer": map[string]interface{}{
			"id":              customer.ID,
			"name":            customer.Name,
			"phone":           customer.Phone,
			"email":           customer.Email,
			"tier":            customer.Tier,
			"total_spent":     customer.TotalSpent,
			"total_purchases": customer.TotalPurchases,
		},
		"points": map[string]interface{}{
			"available":         validPoints,
			"lifetime_earned":   customer.PointsEarned,
			"lifetime_redeemed": customer.PointsRedeemed,
			"points_rate":       tierConfig.PointsRate,
		},
		"tier": map[string]interface{}{
			"current":   customer.Tier,
			"next":      s.getNextTier(customer.Tier),
			"min_spend": tierConfig.MinSpend,
			"to_next":   s.getSpendToNextTier(customer.Tier, customer.TotalSpent),
			"perks":     strings.Split(tierConfig.Perks, ", "),
		},
		"recent_transactions": transactions,
	}, nil
}

func (s *Service) getNextTier(current models.LoyaltyTier) string {
	order := []models.LoyaltyTier{models.TierBronze, models.TierSilver, models.TierGold, models.TierPlatinum}
	for i, t := range order {
		if t == current && i < len(order)-1 {
			return string(order[i+1])
		}
	}
	return ""
}

func (s *Service) getSpendToNextTier(current models.LoyaltyTier, spent float64) float64 {
	next := s.getNextTier(current)
	if next == "" {
		return 0
	}
	config := models.DefaultTierConfigs[models.LoyaltyTier(next)]
	return config.MinSpend - spent
}

func (s *Service) ProcessExpiredPoints() error {
	var expired []models.LoyaltyTransaction
	s.db.Where("points > 0 AND expires_at < ?", time.Now()).Find(&expired)

	for _, tx := range expired {
		customer, err := s.customerRepo.GetByID(tx.CustomerID)
		if err != nil {
			continue
		}

		customer.LoyaltyPoints -= tx.Points
		if customer.LoyaltyPoints < 0 {
			customer.LoyaltyPoints = 0
		}

		s.customerRepo.Update(customer)

		tx.Points = -tx.Points
		tx.Type = models.LoyaltyExpired
		tx.Description = "Points expired"
		s.db.Save(tx)
	}

	return nil
}
