package routes

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/C9b3rD3vi1/DukaPOS/internal/repository"
	"github.com/C9b3rD3vi1/DukaPOS/internal/services/job"
)

var (
	defaultJobScheduler        *job.Scheduler
	defaultJobSchedulerStarted bool
)

type SchedulerConfig struct {
	ShopRepo     *repository.ShopRepository
	SaleRepo     *repository.SaleRepository
	ProductRepo  *repository.ProductRepository
	SendWhatsApp func(phone, message string) error
}

func GetJobScheduler() *job.Scheduler {
	return defaultJobScheduler
}

func RegisterScheduledTasks(config SchedulerConfig) {
	// Initialize the advanced job defaultJobScheduler
	defaultJobScheduler = job.GetScheduler()
	defaultJobScheduler.Start()
	defaultJobSchedulerStarted = true

	// Daily report task - runs every 24 hours
	defaultJobScheduler.AddPeriodicJob("daily_reports", 24*time.Hour, func() error {
		log.Println("📊 Running daily reports task...")

		shops, _, err := config.ShopRepo.List(1000, 0)
		if err != nil {
			log.Printf("❌ Failed to get shops: %v", err)
			return err
		}

		for _, shop := range shops {
			if !shop.IsActive {
				continue
			}

			sales, err := config.SaleRepo.GetTodaySales(shop.ID)
			if err != nil {
				continue
			}

			totalSales := 0.0
			totalProfit := 0.0
			for _, s := range sales {
				totalSales += s.TotalAmount
				totalProfit += s.Profit
			}

			if len(sales) > 0 {
				reportMsg := fmt.Sprintf("📊 DAILY REPORT - %s\n\n💰 Today's Sales: KSh %.0f\n💵 Profit: KSh %.0f\n📝 Transactions: %d\n\nSent automatically by DukaPOS", shop.Name, totalSales, totalProfit, len(sales))

				if err := config.SendWhatsApp(shop.Phone, reportMsg); err != nil {
					log.Printf("❌ Failed to send daily report to shop %s: %v", shop.Name, err)
				} else {
					log.Printf("✅ Daily report sent to shop %s", shop.Name)
				}
			}
		}

		log.Println("✅ Daily reports task completed")
		return nil
	})

	// Low stock check - runs every 6 hours
	defaultJobScheduler.AddPeriodicJob("low_stock_check", 6*time.Hour, func() error {
		log.Println("⚠️ Running low stock check...")

		shops, _, err := config.ShopRepo.List(1000, 0)
		if err != nil {
			return err
		}

		for _, shop := range shops {
			if !shop.IsActive {
				continue
			}

			lowStock, err := config.ProductRepo.GetLowStock(shop.ID)
			if err != nil {
				continue
			}

			if len(lowStock) > 0 {
				var productList strings.Builder
				productList.WriteString("⚠️ LOW STOCK ALERT\n\n")
				for _, p := range lowStock {
					productList.WriteString(fmt.Sprintf("• %s: %d (min: %d)\n", p.Name, p.CurrentStock, p.LowStockThreshold))
				}
				productList.WriteString("\nAdd stock: add [name] [price] [qty]")

				if err := config.SendWhatsApp(shop.Phone, productList.String()); err != nil {
					log.Printf("❌ Failed to send low stock alert to shop %s: %v", shop.Name, err)
				} else {
					log.Printf("✅ Low stock alert sent to shop %s", shop.Name)
				}
			}
		}

		log.Println("✅ Low stock check completed")
		return nil
	})

	// Weekly report task - runs every 7 days
	defaultJobScheduler.AddPeriodicJob("weekly_reports", 7*24*time.Hour, func() error {
		log.Println("📊 Running weekly reports task...")

		shops, _, err := config.ShopRepo.List(1000, 0)
		if err != nil {
			return err
		}

		for _, shop := range shops {
			if !shop.IsActive {
				continue
			}

			end := time.Now()
			start := end.AddDate(0, 0, -7)
			sales, err := config.SaleRepo.GetByDateRange(shop.ID, start, end)
			if err != nil {
				continue
			}

			if len(sales) > 0 {
				totalSales := 0.0
				totalProfit := 0.0
				for _, s := range sales {
					totalSales += s.TotalAmount
					totalProfit += s.Profit
				}

				reportMsg := fmt.Sprintf("📊 WEEKLY REPORT\n\n💰 Weekly Sales: KSh %.0f\n💵 Profit: KSh %.0f\n📝 Transactions: %d\n\nHave a great week!", totalSales, totalProfit, len(sales))

				if err := config.SendWhatsApp(shop.Phone, reportMsg); err != nil {
					log.Printf("❌ Failed to send weekly report to shop %s: %v", shop.Name, err)
				}
			}
		}

		log.Println("✅ Weekly reports task completed")
		return nil
	})

	// Monthly report task - runs every 30 days
	defaultJobScheduler.AddPeriodicJob("monthly_reports", 30*24*time.Hour, func() error {
		log.Println("📊 Running monthly reports task...")

		shops, _, err := config.ShopRepo.List(1000, 0)
		if err != nil {
			return err
		}

		for _, shop := range shops {
			if !shop.IsActive {
				continue
			}

			end := time.Now()
			start := end.AddDate(0, -1, 0)
			sales, err := config.SaleRepo.GetByDateRange(shop.ID, start, end)
			if err != nil {
				continue
			}

			if len(sales) > 0 {
				totalSales := 0.0
				totalProfit := 0.0
				for _, s := range sales {
					totalSales += s.TotalAmount
					totalProfit += s.Profit
				}

				avgDaily := totalSales / 30

				reportMsg := fmt.Sprintf("📊 MONTHLY REPORT\n\n💰 Monthly Sales: KSh %.0f\n💵 Profit: KSh %.0f\n📝 Transactions: %d\n📈 Daily Avg: KSh %.0f\n\nGreat progress this month! 🎉", totalSales, totalProfit, len(sales), avgDaily)

				if err := config.SendWhatsApp(shop.Phone, reportMsg); err != nil {
					log.Printf("❌ Failed to send monthly report to shop %s: %v", shop.Name, err)
				}
			}
		}

		log.Println("✅ Monthly reports task completed")
		return nil
	})

	log.Println("✅ Advanced job defaultJobScheduler initialized with jobs:")
	log.Println("   - daily_reports (24h)")
	log.Println("   - low_stock_check (6h)")
	log.Println("   - weekly_reports (7d)")
	log.Println("   - monthly_reports (30d)")
}
