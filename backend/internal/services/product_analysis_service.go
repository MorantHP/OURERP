package services

import (
	"time"

	"github.com/MorantHP/OURERP/internal/repository"
)

type ProductAnalysisService struct {
	repo *repository.DatacenterRepository
}

func NewProductAnalysisService(repo *repository.DatacenterRepository) *ProductAnalysisService {
	return &ProductAnalysisService{repo: repo}
}

// ProductTurnover 商品动销率分析
type ProductTurnover struct {
	ProductID      int64   `json:"product_id"`
	ProductName    string  `json:"product_name"`
	SalesQuantity  int     `json:"sales_quantity"`
	StockQuantity  int     `json:"stock_quantity"`
	TurnoverRate   float64 `json:"turnover_rate"`
	Status         string  `json:"status"` // high/medium/low/stagnant
}

// InventoryLevel 库存水位分析
type InventoryLevel struct {
	ProductID    int64   `json:"product_id"`
	ProductName  string  `json:"product_name"`
	Quantity     int     `json:"quantity"`
	MinQuantity  int     `json:"min_quantity"`
	MaxQuantity  int     `json:"max_quantity"`
	StockLevel   string  `json:"stock_level"` // out_of_stock/low/normal/high
	DaysOfStock  int     `json:"days_of_stock"`
	Suggestion   string  `json:"suggestion"`
}

// PurchaseStrategy 进货策略建议
type PurchaseStrategy struct {
	ProductID       int64   `json:"product_id"`
	ProductName     string  `json:"product_name"`
	CurrentStock    int     `json:"current_stock"`
	AvgDailySales   float64 `json:"avg_daily_sales"`
	SuggestedQty    int     `json:"suggested_qty"`
	Priority        string  `json:"priority"` // urgent/high/medium/low
	EstimatedDays   int     `json:"estimated_days"` // 预计可售天数
	SafetyStock     int     `json:"safety_stock"`
}

// ProductProfit 商品毛利分析
type ProductProfit struct {
	ProductID     int64   `json:"product_id"`
	ProductName   string  `json:"product_name"`
	SalesCount    int     `json:"sales_count"`
	SalesAmount   float64 `json:"sales_amount"`
	ProductCost   float64 `json:"product_cost"`
	GrossProfit   float64 `json:"gross_profit"`
	ProfitRate    float64 `json:"profit_rate"`
}

// ProductPerformance 商品表现分析
type ProductPerformance struct {
	ProductID      int64   `json:"product_id"`
	ProductName    string  `json:"product_name"`
	Category       string  `json:"category"`
	Brand          string  `json:"brand"`
	SalesRank      int     `json:"sales_rank"`
	ProfitRank     int     `json:"profit_rank"`
	TurnoverRank   int     `json:"turnover_rank"`
	Score          float64 `json:"score"` // 综合得分
}

// GetTurnoverRate 获取商品动销率
func (s *ProductAnalysisService) GetTurnoverRate(tenantID int64, startDate, endDate time.Time, limit int) ([]ProductTurnover, error) {
	data, err := s.repo.GetProductTurnoverRate(tenantID, startDate, endDate, limit)
	if err != nil {
		return nil, err
	}

	result := make([]ProductTurnover, 0)
	for _, item := range data {
		turnover := ProductTurnover{}
		if v, ok := item["product_id"].(int64); ok {
			turnover.ProductID = v
		}
		if v, ok := item["product_name"].(string); ok {
			turnover.ProductName = v
		}
		if v, ok := item["sales_quantity"].(int64); ok {
			turnover.SalesQuantity = int(v)
		}
		if v, ok := item["stock_quantity"].(int64); ok {
			turnover.StockQuantity = int(v)
		}
		if v, ok := item["turnover_rate"].(float64); ok {
			turnover.TurnoverRate = v
			// 判断动销状态
			if v >= 80 {
				turnover.Status = "high"
			} else if v >= 40 {
				turnover.Status = "medium"
			} else if v > 0 {
				turnover.Status = "low"
			} else {
				turnover.Status = "stagnant"
			}
		}
		result = append(result, turnover)
	}

	return result, nil
}

// GetInventoryLevel 获取库存水位分析
func (s *ProductAnalysisService) GetInventoryLevel(tenantID int64) ([]InventoryLevel, error) {
	data, err := s.repo.GetInventoryLevel(tenantID)
	if err != nil {
		return nil, err
	}

	result := make([]InventoryLevel, 0)
	for _, item := range data {
		level := InventoryLevel{}
		if v, ok := item["product_id"].(int64); ok {
			level.ProductID = v
		}
		if v, ok := item["product_name"].(string); ok {
			level.ProductName = v
		}
		if v, ok := item["quantity"].(int64); ok {
			level.Quantity = int(v)
		}
		if v, ok := item["min_quantity"].(int64); ok {
			level.MinQuantity = int(v)
		}
		if v, ok := item["max_quantity"].(int64); ok {
			level.MaxQuantity = int(v)
		}
		if v, ok := item["stock_level"].(string); ok {
			level.StockLevel = v
			// 生成建议
			switch v {
			case "out_of_stock":
				level.Suggestion = "紧急补货"
			case "low":
				level.Suggestion = "建议补货"
			case "high":
				level.Suggestion = "库存过高，暂停采购"
			default:
				level.Suggestion = "库存正常"
			}
		}
		result = append(result, level)
	}

	return result, nil
}

// GetPurchaseStrategy 获取进货策略建议
func (s *ProductAnalysisService) GetPurchaseStrategy(tenantID int64, days int) ([]PurchaseStrategy, error) {
	// 获取库存水位
	data, err := s.repo.GetInventoryLevel(tenantID)
	if err != nil {
		return nil, err
	}

	result := make([]PurchaseStrategy, 0)
	for _, item := range data {
		strategy := PurchaseStrategy{}
		if v, ok := item["product_id"].(int64); ok {
			strategy.ProductID = v
		}
		if v, ok := item["product_name"].(string); ok {
			strategy.ProductName = v
		}

		quantity := 0
		minQuantity := 0
		maxQuantity := 0

		if v, ok := item["quantity"].(int64); ok {
			quantity = int(v)
			strategy.CurrentStock = quantity
		}
		if v, ok := item["min_quantity"].(int64); ok {
			minQuantity = int(v)
			strategy.SafetyStock = minQuantity
		}
		if v, ok := item["max_quantity"].(int64); ok {
			maxQuantity = int(v)
		}

		// 简化计算日均销量（实际应该从历史数据计算）
		// 假设库存周转天数为30天
		strategy.AvgDailySales = float64(maxQuantity) / 30.0
		if strategy.AvgDailySales < 1 {
			strategy.AvgDailySales = 1
		}

		// 计算可售天数
		if strategy.AvgDailySales > 0 {
			strategy.EstimatedDays = int(float64(quantity) / strategy.AvgDailySales)
		}

		// 计算建议采购量
		targetStock := maxQuantity
		if targetStock == 0 {
			targetStock = minQuantity * 3
		}
		strategy.SuggestedQty = targetStock - quantity
		if strategy.SuggestedQty < 0 {
			strategy.SuggestedQty = 0
		}

		// 判断优先级
		stockLevel, _ := item["stock_level"].(string)
		switch stockLevel {
		case "out_of_stock":
			strategy.Priority = "urgent"
		case "low":
			strategy.Priority = "high"
		default:
			strategy.Priority = "low"
		}

		result = append(result, strategy)
	}

	return result, nil
}

// GetLowStockProducts 获取低库存商品
func (s *ProductAnalysisService) GetLowStockProducts(tenantID int64) ([]InventoryLevel, error) {
	levels, err := s.GetInventoryLevel(tenantID)
	if err != nil {
		return nil, err
	}

	result := make([]InventoryLevel, 0)
	for _, level := range levels {
		if level.StockLevel == "low" || level.StockLevel == "out_of_stock" {
			result = append(result, level)
		}
	}

	return result, nil
}

// InventorySummary 库存汇总
type InventorySummary struct {
	TotalProducts      int     `json:"total_products"`
	TotalQuantity      int     `json:"total_quantity"`
	TotalValue         float64 `json:"total_value"`
	OutOfStockCount    int     `json:"out_of_stock_count"`
	LowStockCount      int     `json:"low_stock_count"`
	NormalStockCount   int     `json:"normal_stock_count"`
	HighStockCount     int     `json:"high_stock_count"`
	AvgTurnoverDays    float64 `json:"avg_turnover_days"`
}

// GetInventorySummary 获取库存汇总
func (s *ProductAnalysisService) GetInventorySummary(tenantID int64) (*InventorySummary, error) {
	data, err := s.repo.GetInventoryLevel(tenantID)
	if err != nil {
		return nil, err
	}

	summary := &InventorySummary{}
	for _, item := range data {
		summary.TotalProducts++
		if v, ok := item["quantity"].(int64); ok {
			summary.TotalQuantity += int(v)
		}
		if level, ok := item["stock_level"].(string); ok {
			switch level {
			case "out_of_stock":
				summary.OutOfStockCount++
			case "low":
				summary.LowStockCount++
			case "normal":
				summary.NormalStockCount++
			case "high":
				summary.HighStockCount++
			}
		}
	}

	return summary, nil
}

// SalesTrend 销售趋势
type SalesTrend struct {
	Date        string  `json:"date"`
	SalesCount  int     `json:"sales_count"`
	SalesAmount float64 `json:"sales_amount"`
	Quantity    int     `json:"quantity"`
}

// GetProductSalesTrend 获取商品销售趋势
func (s *ProductAnalysisService) GetProductSalesTrend(tenantID int64, productID int64, startDate, endDate time.Time) ([]SalesTrend, error) {
	// 这里简化处理，实际应该从订单明细按日期分组统计
	trends := make([]SalesTrend, 0)

	// 生成日期范围内的趋势数据
	current := startDate
	for current.Before(endDate) || current.Equal(endDate) {
		trends = append(trends, SalesTrend{
			Date:        current.Format("2006-01-02"),
			SalesCount:  0,
			SalesAmount: 0,
			Quantity:    0,
		})
		current = current.AddDate(0, 0, 1)
	}

	return trends, nil
}

// CategoryAnalysis 类目分析
type CategoryAnalysis struct {
	Category       string  `json:"category"`
	ProductCount   int     `json:"product_count"`
	SalesCount     int     `json:"sales_count"`
	SalesAmount    float64 `json:"sales_amount"`
	Quantity       int     `json:"quantity"`
	AvgPrice       float64 `json:"avg_price"`
	Percentage     float64 `json:"percentage"`
}

// GetCategoryAnalysis 获取类目销售分析
func (s *ProductAnalysisService) GetCategoryAnalysis(tenantID int64, startDate, endDate time.Time) ([]CategoryAnalysis, error) {
	// 这里简化处理，实际应该从数据库按类目分组统计
	result := []CategoryAnalysis{}
	_ = tenantID
	_ = startDate
	_ = endDate

	return result, nil
}
