package staff

import (
	"errors"
	"fmt"

	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"github.com/C9b3rD3vi1/DukaPOS/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrStaffNotFound  = errors.New("staff not found")
	ErrStaffExists    = errors.New("staff already exists")
	ErrInvalidPin     = errors.New("invalid PIN")
	ErrStaffInactive  = errors.New("staff account is inactive")
)

// Service handles staff management
type Service struct {
	staffRepo *repository.StaffRepository
	shopRepo  *repository.ShopRepository
}

// New creates a new staff service
func New(staffRepo *repository.StaffRepository, shopRepo *repository.ShopRepository) *Service {
	return &Service{
		staffRepo: staffRepo,
		shopRepo:  shopRepo,
	}
}

// Create creates a new staff member
func (s *Service) Create(shopID uint, name, phone, role, pin string) (*models.Staff, error) {
	// Check if shop exists
	_, err := s.shopRepo.GetByID(shopID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("shop not found")
		}
		return nil, err
	}

	// Check if staff with same phone exists
	existing, _ := s.staffRepo.GetByPhone(shopID, phone)
	if existing != nil {
		return nil, ErrStaffExists
	}

	// Hash PIN
	hashedPin, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	staff := &models.Staff{
		ShopID:   shopID,
		Name:     name,
		Phone:    phone,
		Role:     role,
		Pin:      string(hashedPin),
		IsActive: true,
	}

	if err := s.staffRepo.Create(staff); err != nil {
		return nil, err
	}

	return staff, nil
}

// GetByID gets a staff member by ID
func (s *Service) GetByID(id uint) (*models.Staff, error) {
	return s.staffRepo.GetByID(id)
}

// GetByPhone gets a staff member by phone
func (s *Service) GetByPhone(shopID uint, phone string) (*models.Staff, error) {
	return s.staffRepo.GetByPhone(shopID, phone)
}

// GetByShop gets all staff for a shop
func (s *Service) GetByShop(shopID uint) ([]models.Staff, error) {
	return s.staffRepo.GetByShopID(shopID)
}

// VerifyPin verifies a staff member's PIN
func (s *Service) VerifyPin(shopID uint, phone, pin string) (*models.Staff, error) {
	staff, err := s.staffRepo.GetByPhone(shopID, phone)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStaffNotFound
		}
		return nil, err
	}

	if !staff.IsActive {
		return nil, ErrStaffInactive
	}

	err = bcrypt.CompareHashAndPassword([]byte(staff.Pin), []byte(pin))
	if err != nil {
		return nil, ErrInvalidPin
	}

	return staff, nil
}

// Update updates a staff member
func (s *Service) Update(id uint, name, phone, role string, isActive bool) (*models.Staff, error) {
	staff, err := s.staffRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStaffNotFound
		}
		return nil, err
	}

	if name != "" {
		staff.Name = name
	}
	if phone != "" {
		staff.Phone = phone
	}
	if role != "" {
		staff.Role = role
	}
	staff.IsActive = isActive

	if err := s.staffRepo.Update(staff); err != nil {
		return nil, err
	}

	return staff, nil
}

// UpdatePin updates a staff member's PIN
func (s *Service) UpdatePin(id uint, newPin string) error {
	staff, err := s.staffRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrStaffNotFound
		}
		return err
	}

	hashedPin, err := bcrypt.GenerateFromPassword([]byte(newPin), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	staff.Pin = string(hashedPin)
	return s.staffRepo.Update(staff)
}

// Delete deletes a staff member
func (s *Service) Delete(id uint) error {
	staff, err := s.staffRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrStaffNotFound
		}
		return err
	}

	return s.staffRepo.Delete(staff.ID)
}

// FormatStaffList formats staff list for WhatsApp
func (s *Service) FormatStaffList(staff []models.Staff) string {
	if len(staff) == 0 {
		return "No staff members yet.\nAdd: staff add [name] [phone] [role]"
	}

	var msg string = "👥 STAFF LIST:\n\n"
	for i, st := range staff {
		status := "✅"
		if !st.IsActive {
			status = "❌"
		}
		msg += fmt.Sprintf("%d. %s %s\n   📱 %s\n   💼 %s\n\n", 
			i+1, status, st.Name, st.Phone, st.Role)
	}
	return msg
}

// Roles returns available staff roles
func Roles() []string {
	return []string{
		"manager",
		"cashier",
		"stock clerk",
		"assistant",
	}
}

// IsValidRole checks if a role is valid
func IsValidRole(role string) bool {
	for _, r := range Roles() {
		if r == role {
			return true
		}
	}
	return false
}
