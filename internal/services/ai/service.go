package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/C9b3rD3vi1/DukaPOS/internal/models"
	"github.com/C9b3rD3vi1/DukaPOS/internal/repository"
)

type SalesData struct {
	Date     time.Time
	Quantity int
	Revenue  float64
}

type PredictionService struct {
	productRepo         *repository.ProductRepository
	saleRepo            *repository.SaleRepository
	summaryRepo         *repository.DailySummaryRepository
	minDataDays         int
	confidenceThreshold float64
	openAIAPIKey        string
	httpClient          *http.Client
}

func NewPredictionService(
	productRepo *repository.ProductRepository,
	saleRepo *repository.SaleRepository,
	summaryRepo *repository.DailySummaryRepository,
) *PredictionService {
	return &PredictionService{
		productRepo:         productRepo,
		saleRepo:            saleRepo,
		summaryRepo:         summaryRepo,
		minDataDays:         7,
		confidenceThreshold: 0.6,
		httpClient:          &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *PredictionService) SetOpenAIKey(apiKey string) {
	s.openAIAPIKey = apiKey
}

func (s *PredictionService) IsOpenAIConfigured() bool {
	return s.openAIAPIKey != ""
}

func (s *PredictionService) GetOpenAIInsights(ctx context.Context, shopID uint, prompt string) (string, error) {
	if !s.IsOpenAIConfigured() {
		return "", fmt.Errorf("OpenAI not configured")
	}

	sales, err := s.saleRepo.GetByShopID(shopID, 1000)
	if err != nil {
		return "", err
	}

	products, err := s.productRepo.GetByShopID(shopID)
	if err != nil {
		products = []models.Product{}
	}

	summary := fmt.Sprintf("Shop has %d products and %d recent sales. ", len(products), len(sales))

	reqBody := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{"role": "system", "content": "You are a business analyst for a Kenyan shop. Provide concise, actionable insights."},
			{"role": "user", "content": prompt + "\n\n" + summary},
		},
		"max_tokens": 200,
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+s.openAIAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
		if msg, ok := choices[0].(map[string]interface{})["message"]; ok {
			if content, ok := msg.(map[string]interface{})["content"]; ok {
				return content.(string), nil
			}
		}
	}

	return "", fmt.Errorf("no response from OpenAI")
}

type ProductPrediction struct {
	ProductID          uint      `json:"product_id"`
	ProductName        string    `json:"product_name"`
	CurrentStock       int       `json:"current_stock"`
	AvgDailySales      float64   `json:"avg_daily_sales"`
	DaysUntilStockout  int       `json:"days_until_stockout"`
	RecommendedOrder   int       `json:"recommended_order"`
	Confidence         float64   `json:"confidence"`
	Trend              string    `json:"trend"`
	TrendPercentage    float64   `json:"trend_percentage"`
	SeasonalMultiplier float64   `json:"seasonal_multiplier"`
	Priority           string    `json:"priority"`
	LastUpdated        time.Time `json:"last_updated"`
}

type ShopPrediction struct {
	ShopID        uint                `json:"shop_id"`
	TotalProducts int                 `json:"total_products"`
	Predictions   []ProductPrediction `json:"predictions"`
	UrgentCount   int                 `json:"urgent_count"`
	WarningCount  int                 `json:"warning_count"`
	HealthyCount  int                 `json:"healthy_count"`
	GeneratedAt   time.Time           `json:"generated_at"`
}

type SalesAnalytics struct {
	TotalRevenue     float64          `json:"total_revenue"`
	TotalProfit      float64          `json:"total_profit"`
	TransactionCount int              `json:"transaction_count"`
	AvgTransaction   float64          `json:"avg_transaction"`
	TopProducts      []ProductSummary `json:"top_products"`
	PeakHour         int              `json:"peak_hour"`
	PeakDay          string           `json:"peak_day"`
}

type ProductSummary struct {
	Name       string  `json:"name"`
	Quantity   int     `json:"quantity"`
	Revenue    float64 `json:"revenue"`
	Profit     float64 `json:"profit"`
	Percentage float64 `json:"percentage"`
}

func (s *PredictionService) GetPredictions(shopID uint) (*ShopPrediction, error) {
	products, err := s.productRepo.GetByShopID(shopID)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	prediction := &ShopPrediction{
		ShopID:        shopID,
		TotalProducts: len(products),
		Predictions:   make([]ProductPrediction, 0),
		GeneratedAt:   time.Now(),
	}

	for _, product := range products {
		pred := s.predictProduct(product.ID, product.Name, product.CurrentStock, shopID)
		prediction.Predictions = append(prediction.Predictions, *pred)

		if pred.Priority == "urgent" {
			prediction.UrgentCount++
		} else if pred.Priority == "warning" {
			prediction.WarningCount++
		} else {
			prediction.HealthyCount++
		}
	}

	return prediction, nil
}

func (s *PredictionService) predictProduct(productID uint, productName string, currentStock int, shopID uint) *ProductPrediction {
	salesData := s.getHistoricalSales(productID, shopID)

	pred := &ProductPrediction{
		ProductID:    productID,
		ProductName:  productName,
		CurrentStock: currentStock,
		LastUpdated:  time.Now(),
	}

	if len(salesData) < s.minDataDays {
		pred.AvgDailySales = 0
		pred.DaysUntilStockout = -1
		pred.RecommendedOrder = 0
		pred.Confidence = 0
		pred.Trend = "insufficient_data"
		pred.Priority = "unknown"
		return pred
	}

	avgSales := s.calculateAverageDailySales(salesData)
	trend, trendPct := s.calculateTrend(salesData)
	confidence := s.calculateConfidence(salesData)
	seasonalMultiplier := s.getSeasonalMultiplier()

	pred.AvgDailySales = avgSales
	pred.TrendPercentage = trendPct
	pred.SeasonalMultiplier = seasonalMultiplier
	pred.Confidence = confidence

	if trend > 0.1 {
		pred.Trend = "up"
	} else if trend < -0.1 {
		pred.Trend = "down"
	} else {
		pred.Trend = "stable"
	}

	if avgSales > 0 {
		pred.DaysUntilStockout = int(float64(currentStock) / avgSales)
	} else {
		pred.DaysUntilStockout = 999
	}

	safetyBuffer := 1.5
	if pred.Trend == "down" {
		safetyBuffer = 1.2
	} else if pred.Trend == "up" {
		safetyBuffer = 1.8
	}

	pred.RecommendedOrder = int(avgSales * 7 * seasonalMultiplier * safetyBuffer)
	if pred.RecommendedOrder < 1 {
		pred.RecommendedOrder = 1
	}

	if currentStock == 0 {
		pred.Priority = "urgent"
	} else if pred.DaysUntilStockout <= 3 && confidence >= s.confidenceThreshold {
		pred.Priority = "urgent"
	} else if pred.DaysUntilStockout <= 7 {
		pred.Priority = "warning"
	} else {
		pred.Priority = "healthy"
	}

	return pred
}

func (s *PredictionService) getHistoricalSales(productID uint, shopID uint) []SalesData {
	end := time.Now()
	start := end.AddDate(0, 0, -30)

	sales, err := s.saleRepo.GetByProductAndDateRange(productID, shopID, start, end)
	if err != nil {
		return []SalesData{}
	}

	data := make([]SalesData, len(sales))
	for i, sale := range sales {
		data[i] = SalesData{
			Date:     sale.CreatedAt,
			Quantity: sale.Quantity,
			Revenue:  sale.TotalAmount,
		}
	}

	return data
}

func (s *PredictionService) calculateAverageDailySales(data []SalesData) float64 {
	if len(data) == 0 {
		return 0
	}

	daySales := make(map[string]int)
	for _, d := range data {
		day := d.Date.Format("2006-01-02")
		daySales[day] += d.Quantity
	}

	total := 0
	for _, qty := range daySales {
		total += qty
	}

	days := len(daySales)
	if days == 0 {
		return 0
	}

	return float64(total) / float64(days)
}

func (s *PredictionService) calculateTrend(data []SalesData) (float64, float64) {
	if len(data) < 2 {
		return 0, 0
	}

	daySales := make(map[string]int)
	for _, d := range data {
		day := d.Date.Format("2006-01-02")
		daySales[day] += d.Quantity
	}

	type dayData struct {
		day   string
		sales int
	}

	days := make([]dayData, 0, len(daySales))
	for day, sales := range daySales {
		days = append(days, dayData{day, sales})
	}

	for i := 0; i < len(days)-1; i++ {
		for j := i + 1; j < len(days); j++ {
			if days[j].day < days[i].day {
				days[i], days[j] = days[j], days[i]
			}
		}
	}

	n := float64(len(days))
	if n < 2 {
		return 0, 0
	}

	sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0
	for i, d := range days {
		x := float64(i)
		y := float64(d.sales)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	denominator := n*sumX2 - sumX*sumX
	if denominator == 0 {
		return 0, 0
	}

	slope := (n*sumXY - sumX*sumY) / denominator

	avgY := sumY / n
	if avgY == 0 {
		return 0, 0
	}

	percentageChange := (slope / avgY) * 100

	return slope / avgY, percentageChange
}

func (s *PredictionService) calculateConfidence(data []SalesData) float64 {
	if len(data) < 7 {
		return float64(len(data)) / 10.0
	}

	daySales := make(map[string]int)
	for _, d := range data {
		day := d.Date.Format("2006-01-02")
		daySales[day] += d.Quantity
	}

	values := make([]float64, 0, len(daySales))
	for _, qty := range daySales {
		values = append(values, float64(qty))
	}

	mean := 0.0
	for _, v := range values {
		mean += v
	}
	mean /= float64(len(values))

	if mean == 0 {
		return 0
	}

	variance := 0.0
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(values))

	stdDev := variance
	if stdDev > 0 {
		stdDev = Sqrt(stdDev)
	}

	cv := (stdDev / mean)
	if cv > 1 {
		cv = 1
	}

	confidence := 1.0 - cv
	if confidence < 0 {
		confidence = 0
	}

	return Round(confidence*100) / 100
}

func (s *PredictionService) getSeasonalMultiplier() float64 {
	month := int(time.Now().Month())
	dayOfWeek := time.Now().Weekday()

	seasonalFactors := map[int]float64{
		1: 0.85, 2: 0.95, 3: 1.0, 4: 1.05,
		5: 1.0, 6: 0.9, 7: 0.8, 8: 0.85,
		9: 1.0, 10: 1.1, 11: 1.25, 12: 1.35,
	}

	weekdayFactors := map[time.Weekday]float64{
		time.Monday:    0.9,
		time.Tuesday:   1.0,
		time.Wednesday: 1.1,
		time.Thursday:  1.1,
		time.Friday:    1.2,
		time.Saturday:  1.35,
		time.Sunday:    0.75,
	}

	monthFactor := seasonalFactors[month]
	weekFactor := weekdayFactors[dayOfWeek]

	return (monthFactor + weekFactor) / 2
}

func (s *PredictionService) GetSalesAnalytics(shopID uint, days int) (*SalesAnalytics, error) {
	end := time.Now()
	start := end.AddDate(0, 0, -days)

	sales, err := s.saleRepo.GetByDateRange(shopID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get sales: %w", err)
	}

	analytics := &SalesAnalytics{
		TopProducts: make([]ProductSummary, 0),
	}

	productMap := make(map[uint]*ProductSummary)
	hourCounts := make(map[int]int)
	dayCounts := make(map[string]int)

	for _, sale := range sales {
		analytics.TotalRevenue += sale.TotalAmount
		analytics.TotalProfit += sale.Profit
		analytics.TransactionCount++

		if p, ok := productMap[sale.ProductID]; ok {
			p.Quantity += sale.Quantity
			p.Revenue += sale.TotalAmount
			p.Profit += sale.Profit
		} else {
			productMap[sale.ProductID] = &ProductSummary{
				Name:     sale.Product.Name,
				Quantity: sale.Quantity,
				Revenue:  sale.TotalAmount,
				Profit:   sale.Profit,
			}
		}

		hour := sale.CreatedAt.Hour()
		hourCounts[hour]++

		day := sale.CreatedAt.Weekday().String()
		dayCounts[day]++
	}

	for _, p := range productMap {
		p.Percentage = (p.Revenue / analytics.TotalRevenue) * 100
		analytics.TopProducts = append(analytics.TopProducts, *p)
	}

	for i := 0; i < len(analytics.TopProducts)-1; i++ {
		for j := i + 1; j < len(analytics.TopProducts); j++ {
			if analytics.TopProducts[j].Revenue > analytics.TopProducts[i].Revenue {
				analytics.TopProducts[i], analytics.TopProducts[j] = analytics.TopProducts[j], analytics.TopProducts[i]
			}
		}
	}

	if len(analytics.TopProducts) > 10 {
		analytics.TopProducts = analytics.TopProducts[:10]
	}

	if analytics.TransactionCount > 0 {
		analytics.AvgTransaction = analytics.TotalRevenue / float64(analytics.TransactionCount)
	}

	peakHour := 0
	peakHourCount := 0
	for hour, count := range hourCounts {
		if count > peakHourCount {
			peakHour = hour
			peakHourCount = count
		}
	}
	analytics.PeakHour = peakHour

	peakDay := ""
	peakDayCount := 0
	for day, count := range dayCounts {
		if count > peakDayCount {
			peakDay = day
			peakDayCount = count
		}
	}
	analytics.PeakDay = peakDay

	return analytics, nil
}

func (s *PredictionService) GetRestockRecommendations(shopID uint) ([]ProductPrediction, error) {
	prediction, err := s.GetPredictions(shopID)
	if err != nil {
		return nil, err
	}

	recommendations := make([]ProductPrediction, 0)
	for _, p := range prediction.Predictions {
		if p.Priority == "urgent" || p.Priority == "warning" {
			recommendations = append(recommendations, p)
		}
	}

	for i := 0; i < len(recommendations)-1; i++ {
		for j := i + 1; j < len(recommendations); j++ {
			if recommendations[j].DaysUntilStockout < recommendations[i].DaysUntilStockout {
				recommendations[i], recommendations[j] = recommendations[j], recommendations[i]
			}
		}
	}

	return recommendations, nil
}

func (s *PredictionService) GetInventoryValue(shopID uint) (map[string]float64, error) {
	products, err := s.productRepo.GetByShopID(shopID)
	if err != nil {
		return nil, err
	}

	result := map[string]float64{
		"total_cost_value":   0,
		"total_retail_value": 0,
		"potential_profit":   0,
	}

	for _, p := range products {
		result["total_cost_value"] += p.CostPrice * float64(p.CurrentStock)
		result["total_retail_value"] += p.SellingPrice * float64(p.CurrentStock)
	}

	result["potential_profit"] = result["total_retail_value"] - result["total_cost_value"]

	return result, nil
}

func Sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := 1.0
	for i := 0; i < 20; i++ {
		z = (z + x/z) / 2
	}
	return z
}

func Round(x float64) float64 {
	if x < 0 {
		return -Round(-x)
	}
	return float64(int(x + 0.5))
}

type ForecastRequest struct {
	ProductID   uint `json:"product_id"`
	Days        int  `json:"days"`
	TargetStock int  `json:"target_stock"`
}

type ForecastResult struct {
	ProductID        uint    `json:"product_id"`
	ProductName      string  `json:"product_name"`
	CurrentStock     int     `json:"current_stock"`
	ForecastedSales  int     `json:"forecasted_sales"`
	DaysRemaining    int     `json:"days_remaining"`
	RecommendedOrder int     `json:"recommended_order"`
	OrderDate        string  `json:"order_date"`
	Confidence       float64 `json:"confidence"`
}

func (s *PredictionService) GenerateForecast(ctx context.Context, shopID uint, req *ForecastRequest) (*ForecastResult, error) {
	product, err := s.productRepo.GetByID(req.ProductID)
	if err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	days := req.Days
	if days <= 0 {
		days = 7
	}

	salesData := s.getHistoricalSales(req.ProductID, shopID)
	avgDaily := s.calculateAverageDailySales(salesData)
	trend, _ := s.calculateTrend(salesData)
	seasonal := s.getSeasonalMultiplier()

	var trendFactor float64 = 1.0
	if trend > 0.1 {
		trendFactor = 1.2
	} else if trend < -0.1 {
		trendFactor = 0.8
	}

	forecastedSales := int(avgDaily * float64(days) * trendFactor * seasonal)
	daysRemaining := 0
	if avgDaily > 0 {
		daysRemaining = int(float64(product.CurrentStock) / (avgDaily * trendFactor * seasonal))
	}

	targetStock := req.TargetStock
	if targetStock <= 0 {
		targetStock = forecastedSales * 2
	}

	recommendedOrder := targetStock - product.CurrentStock
	if recommendedOrder < 0 {
		recommendedOrder = 0
	}

	orderDays := daysRemaining - 3
	if orderDays < 1 {
		orderDays = 1
	}
	orderDate := time.Now().AddDate(0, 0, orderDays).Format("2006-01-02")

	confidence := s.calculateConfidence(salesData)

	return &ForecastResult{
		ProductID:        product.ID,
		ProductName:      product.Name,
		CurrentStock:     product.CurrentStock,
		ForecastedSales:  forecastedSales,
		DaysRemaining:    daysRemaining,
		RecommendedOrder: recommendedOrder,
		OrderDate:        orderDate,
		Confidence:       confidence,
	}, nil
}
