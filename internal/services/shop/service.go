package shop

import (
	"errors"
	"fmt"

	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"github.com/C9b3rD3vi1/DukaPOS/internal/repository"
	"gorm.io/gorm"
)

var (
	ErrShopNotFound      = errors.New("shop not found")
	ErrShopExists       = errors.New("shop already exists")
	ErrMaxShopsReached  = errors.New("maximum shops reached for your plan")
	ErrNotShopOwner     = errors.New("not the shop owner")
)

// Service handles multiple shop operations
type Service struct {
	shopRepo    *repository.ShopRepository
	productRepo *repository.ProductRepository
	saleRepo    *repository.SaleRepository
}

// New creates a new shop service
func New(
	shopRepo *repository.ShopRepository,
	productRepo *repository.ProductRepository,
	saleRepo *repository.SaleRepository,
) *Service {
	return &Service{
		shopRepo:    shopRepo,
		productRepo: productRepo,
		saleRepo:    saleRepo,
	}
}

// PlanLimits defines shop limits per plan
var PlanLimits = map[models.PlanType]int{
	models.PlanFree:     1,
	models.PlanPro:     5,
	models.PlanBusiness: 50,
}

// CanAddShop checks if shop can add more shops
func (s *Service) CanAddShop(shop *models.Shop) (bool, error) {
	limit := PlanLimits[shop.Plan]
	
	// Count existing shops for this owner (by phone as proxy for now)
	// In future, we might want to link shops to an owner account
	if limit <= 1 {
		return false, ErrMaxShopsReached
	}
	
	return true, nil
}

// CreateShop creates a new shop for a shop owner
func (s *Service) CreateShop(ownerID uint, name, phone, address string) (*models.Shop, error) {
	owner, err := s.shopRepo.GetByID(ownerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrShopNotFound
		}
		return nil, err
	}

	// Check if can add more shops
	canAdd, err := s.CanAddShop(owner)
	if err != nil {
		return nil, err
	}
	if !canAdd {
		return nil, ErrMaxShopsReached
	}

	// Check if shop with phone already exists
	existing, _ := s.shopRepo.GetByPhone(phone)
	if existing != nil {
		return nil, ErrShopExists
	}

	newShop := &models.Shop{
		Name:       name,
		Phone:      phone,
		OwnerName:  owner.OwnerName,
		Address:    address,
		Plan:       owner.Plan, // Inherit plan from owner
		IsActive:   true,
	}

	if err := s.shopRepo.Create(newShop); err != nil {
		return nil, err
	}

	return newShop, nil
}

// GetShop gets a shop by ID
func (s *Service) GetShop(id uint) (*models.Shop, error) {
	return s.shopRepo.GetByID(id)
}

// GetShopsByOwner gets all shops for an owner
func (s *Service) GetShopsByOwner(ownerID uint) ([]models.Shop, error) {
	// For now, we only have the owner shop
	// In future, we'd have an owner_shops table
	shop, err := s.shopRepo.GetByID(ownerID)
	if err != nil {
		return nil, err
	}
	return []models.Shop{*shop}, nil
}

// SwitchShop switches between shops
func (s *Service) SwitchShop(shopID, userID uint) (*models.Shop, error) {
	shop, err := s.shopRepo.GetByID(shopID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrShopNotFound
		}
		return nil, err
	}

	// Verify user owns this shop
	// For now, just check if shop exists
	// In future, verify ownership

	return shop, nil
}

// UpgradePlan upgrades a shop's plan
func (s *Service) UpgradePlan(shopID uint, newPlan models.PlanType) (*models.Shop, error) {
	shop, err := s.shopRepo.GetByID(shopID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrShopNotFound
		}
		return nil, err
	}

	shop.Plan = newPlan
	if err := s.shopRepo.Update(shop); err != nil {
		return nil, err
	}

	return shop, nil
}

// GetPlanInfo returns plan information
func (s *Service) GetPlanInfo(plan models.PlanType) map[string]interface{} {
	limits := map[models.PlanType]map[string]interface{}{
		models.PlanFree: {
			"name":         "Free",
			"price":        0,
			"shops":        1,
			"products":     50,
			"staff":        0,
			"mpesa":        false,
			"analytics":    false,
		},
		models.PlanPro: {
			"name":         "Pro",
			"price":        500,
			"shops":        5,
			"products":     -1, // unlimited
			"staff":        3,
			"mpesa":        true,
			"analytics":    true,
		},
		models.PlanBusiness: {
			"name":         "Business",
			"price":        1500,
			"shops":        50,
			"products":     -1,
			"staff":        -1,
			"mpesa":        true,
			"analytics":    true,
		},
	}

	if info, ok := limits[plan]; ok {
		return info
	}
	return limits[models.PlanFree]
}

// FormatShopList formats shop list for WhatsApp
func (s *Service) FormatShopList(shops []models.Shop) string {
	if len(shops) == 0 {
		return "No shops found."
	}

	var msg = "🏪 YOUR SHOPS:\n\n"
	for i, shop := range shops {
		status := "✅"
		if !shop.IsActive {
			status = "❌"
		}
		msg += fmt.Sprintf("%d. %s %s\n   📱 %s\n\n", i+1, status, shop.Name, shop.Phone)
	}
	return msg
}

// FormatPlanInfo formats plan info for WhatsApp
func (s *Service) FormatPlanInfo(shop *models.Shop) string {
	info := s.GetPlanInfo(shop.Plan)
	return fmt.Sprintf(`💎 PLAN: %s

🛒 Shops: %d/%d
📦 Products: %d
👥 Staff: %d
💰 M-Pesa: %s
📊 Analytics: %s

Upgrade: upgrade`,
		info["name"],
		1, info["shops"].(int),
		50, info["staff"].(int),
		yesNo(info["mpesa"].(bool)),
		yesNo(info["analytics"].(bool)),
	)
}

func yesNo(b bool) string {
	if b {
		return "✅"
	}
	return "❌"
}
