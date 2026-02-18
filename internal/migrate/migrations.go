package migrate

import (
	"fmt"
	"time"

	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"gorm.io/gorm"
)

type Migration struct {
	Version int
	Name    string
	Run     func(*gorm.DB) error
}

var migrations []Migration

func RegisterMigration(version int, name string, run func(*gorm.DB) error) {
	migrations = append(migrations, Migration{Version: version, Name: name, Run: run})
}

func InitMigrations() {
	RegisterMigration(1, "initial_schema", func(db *gorm.DB) error {
		return db.AutoMigrate(
			&models.Shop{}, &models.Product{}, &models.Sale{},
			&models.DailySummary{}, &models.Staff{}, &models.Customer{},
			&models.LoyaltyTransaction{}, &models.Supplier{}, &models.Order{},
		)
	})

	RegisterMigration(2, "add_mpesa", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.MpesaPayment{}, &models.MpesaTransaction{})
	})

	RegisterMigration(3, "add_webhooks_api", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.Webhook{}, &models.APIKey{})
	})
}

func RunMigrations(db *gorm.DB) error {
	InitMigrations()

	for _, m := range migrations {
		var count int64
		db.Model(&MigrationRecord{}).Where("version = ?", m.Version).Count(&count)
		if count > 0 {
			continue
		}

		fmt.Printf("Running migration %d: %s\n", m.Version, m.Name)
		if err := m.Run(db); err != nil {
			return fmt.Errorf("migration %d failed: %w", m.Version, err)
		}

		db.Create(&MigrationRecord{Version: m.Version, Name: m.Name, ExecutedAt: time.Now()})
		fmt.Printf("Migration %d done\n", m.Version)
	}
	return nil
}

type MigrationRecord struct {
	ID         uint `gorm:"primaryKey"`
	Version    int  `gorm:"uniqueIndex"`
	Name       string
	ExecutedAt time.Time
}

func (MigrationRecord) TableName() string { return "schema_migrations" }
