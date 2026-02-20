package services

import (
	"time"

	"github.com/MorantHP/OURERP/internal/repository"
)

type CompareAnalysisService struct {
	repo *repository.DatacenterRepository
}

func NewCompareAnalysisService(repo *repository.DatacenterRepository) *CompareAnalysisService {
	return &CompareAnalysisService{repo: repo}
}

// PeriodCompare 期间对比结果
type PeriodCompare struct {
	MetricType      string  `json:"metric_type"`
	CurrentValue    float64 `json:"current_value"`
	CompareValue    float64 `json:"compare_value"`
	ChangeValue     float64 `json:"change_value"`
	ChangeRate      float64 `json:"change_rate"`
	CurrentPeriod   string  `json:"current_period"`
	ComparePeriod   string  `json:"compare_period"`
}

// ShopCompare 店铺对比
type ShopCompare struct {
	ShopID          int64   `json:"shop_id"`
	ShopName        string  `json:"shop_name"`
	OrderCount      int     `json:"order_count"`
	OrderAmount     float64 `json:"order_amount"`
	ProfitAmount    float64 `json:"profit_amount"`
	ProfitRate      float64 `json:"profit_rate"`
	AvgOrderValue   float64 `json:"avg_order_value"`
	Percentage      float64 `json:"percentage"`
}

// PlatformCompare 平台对比
type PlatformCompare struct {
	Platform        string  `json:"platform"`
	ShopCount       int     `json:"shop_count"`
	OrderCount      int     `json:"order_count"`
	OrderAmount     float64 `json:"order_amount"`
	ProfitAmount    float64 `json:"profit_amount"`
	ProfitRate      float64 `json:"profit_rate"`
	Percentage      float64 `json:"percentage"`
}

// TargetAchievement 目标达成
type TargetAchievement struct {
	TargetName    string  `json:"target_name"`
	TargetValue   float64 `json:"target_value"`
	ActualValue   float64 `json:"actual_value"`
	AchieveRate   float64 `json:"achieve_rate"`
	Status        string  `json:"status"` // exceeded/achieved/progress/failed
	Remaining     float64 `json:"remaining"`
	DaysRemaining int     `json:"days_remaining"`
}

// PeriodRange 期间范围
type PeriodRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// PeriodCompareResult 期间对比结果
type PeriodCompareResult struct {
	CurrentPeriod  PeriodRange          `json:"current_period"`
	ComparePeriod  PeriodRange          `json:"compare_period"`
	Metrics        []PeriodCompare      `json:"metrics"`
	OrderTrend     []TrendData          `json:"order_trend"`
	AmountTrend    []TrendData          `json:"amount_trend"`
}

// TrendData 趋势数据
type TrendData struct {
	Date   string  `json:"date"`
	Current float64 `json:"current"`
	Compare float64 `json:"compare"`
}

// PeriodCompare 期间对比(同比/环比)
func (s *CompareAnalysisService) PeriodCompare(tenantID int64, currentStart, currentEnd, compareStart, compareEnd time.Time) (*PeriodCompareResult, error) {
	data, err := s.repo.GetPeriodCompare(tenantID, currentStart, currentEnd, compareStart, compareEnd)
	if err != nil {
		return nil, err
	}

	result := &PeriodCompareResult{
		CurrentPeriod: PeriodRange{Start: currentStart, End: currentEnd},
		ComparePeriod: PeriodRange{Start: compareStart, End: compareEnd},
		Metrics:       make([]PeriodCompare, 0),
	}

	// 订单数对比
	orderCompare := PeriodCompare{
		MetricType:    "orders",
		CurrentPeriod: currentStart.Format("2006-01-02") + " ~ " + currentEnd.Format("2006-01-02"),
		ComparePeriod: compareStart.Format("2006-01-02") + " ~ " + compareEnd.Format("2006-01-02"),
	}
	if v, ok := data["current_orders"].(int64); ok {
		orderCompare.CurrentValue = float64(v)
	}
	if v, ok := data["compare_orders"].(int64); ok {
		orderCompare.CompareValue = float64(v)
	}
	orderCompare.ChangeValue = orderCompare.CurrentValue - orderCompare.CompareValue
	if orderCompare.CompareValue > 0 {
		orderCompare.ChangeRate = orderCompare.ChangeValue * 100 / orderCompare.CompareValue
	}
	result.Metrics = append(result.Metrics, orderCompare)

	// 销售额对比
	amountCompare := PeriodCompare{
		MetricType:    "amount",
		CurrentPeriod: currentStart.Format("2006-01-02") + " ~ " + currentEnd.Format("2006-01-02"),
		ComparePeriod: compareStart.Format("2006-01-02") + " ~ " + compareEnd.Format("2006-01-02"),
	}
	if v, ok := data["current_amount"].(float64); ok {
		amountCompare.CurrentValue = v
	}
	if v, ok := data["compare_amount"].(float64); ok {
		amountCompare.CompareValue = v
	}
	amountCompare.ChangeValue = amountCompare.CurrentValue - amountCompare.CompareValue
	if amountCompare.CompareValue > 0 {
		amountCompare.ChangeRate = amountCompare.ChangeValue * 100 / amountCompare.CompareValue
	}
	result.Metrics = append(result.Metrics, amountCompare)

	return result, nil
}

// GetYOYCompare 同比分析(与去年同期对比)
func (s *CompareAnalysisService) GetYOYCompare(tenantID int64, startDate, endDate time.Time) (*PeriodCompareResult, error) {
	// 计算去年同期
	compareStart := startDate.AddDate(-1, 0, 0)
	compareEnd := endDate.AddDate(-1, 0, 0)

	return s.PeriodCompare(tenantID, startDate, endDate, compareStart, compareEnd)
}

// GetMOMCompare 环比分析(与上月同期对比)
func (s *CompareAnalysisService) GetMOMCompare(tenantID int64, startDate, endDate time.Time) (*PeriodCompareResult, error) {
	// 计算上一个周期
	duration := endDate.Sub(startDate)
	compareEnd := startDate.AddDate(0, 0, -1)
	compareStart := compareEnd.Add(-duration)

	return s.PeriodCompare(tenantID, startDate, endDate, compareStart, compareEnd)
}

// ShopCompareResult 店铺对比结果
type ShopCompareResult struct {
	Shops       []ShopCompare `json:"shops"`
	TotalAmount float64       `json:"total_amount"`
	TotalOrders int           `json:"total_orders"`
	BestShop    *ShopCompare  `json:"best_shop"`
	WorstShop   *ShopCompare  `json:"worst_shop"`
}

// GetShopCompare 店铺对比
func (s *CompareAnalysisService) GetShopCompare(tenantID int64, shopIDs []int64, startDate, endDate time.Time) (*ShopCompareResult, error) {
	// 这里简化处理，实际应该从订单表按店铺分组统计
	result := &ShopCompareResult{
		Shops: make([]ShopCompare, 0),
	}

	_ = tenantID
	_ = shopIDs
	_ = startDate
	_ = endDate

	return result, nil
}

// PlatformCompareResult 平台对比结果
type PlatformCompareResult struct {
	Platforms    []PlatformCompare `json:"platforms"`
	TotalAmount  float64           `json:"total_amount"`
	TotalOrders  int               `json:"total_orders"`
	BestPlatform *PlatformCompare  `json:"best_platform"`
}

// GetPlatformCompare 平台对比
func (s *CompareAnalysisService) GetPlatformCompare(tenantID int64, startDate, endDate time.Time) (*PlatformCompareResult, error) {
	// 这里简化处理，实际应该从订单表按平台分组统计
	result := &PlatformCompareResult{
		Platforms: make([]PlatformCompare, 0),
	}

	_ = tenantID
	_ = startDate
	_ = endDate

	return result, nil
}

// Target 目标定义
type Target struct {
	Name       string  `json:"name"`
	TargetType string  `json:"target_type"` // orders/amount/profit
	Value      float64 `json:"value"`
	Deadline   string  `json:"deadline"`
}

// TargetAchievementResult 目标达成结果
type TargetAchievementResult struct {
	Targets     []TargetAchievement `json:"targets"`
	OverallRate float64             `json:"overall_rate"`
	Status      string              `json:"status"`
}

// GetTargetAchievement 目标达成率
func (s *CompareAnalysisService) GetTargetAchievement(tenantID int64, targets []Target, startDate, endDate time.Time) (*TargetAchievementResult, error) {
	result := &TargetAchievementResult{
		Targets: make([]TargetAchievement, 0),
	}

	// 获取期间统计数据
	data, err := s.repo.GetPeriodCompare(tenantID, startDate, endDate, startDate, endDate)
	if err != nil {
		return nil, err
	}

	var totalTarget float64
	var totalAchieved float64

	for _, target := range targets {
		achievement := TargetAchievement{
			TargetName:  target.Name,
			TargetValue: target.Value,
		}

		// 根据目标类型获取实际值
		switch target.TargetType {
		case "orders":
			if v, ok := data["current_orders"].(int64); ok {
				achievement.ActualValue = float64(v)
			}
		case "amount":
			if v, ok := data["current_amount"].(float64); ok {
				achievement.ActualValue = v
			}
		case "profit":
			// 简化处理，实际应该从利润表获取
			achievement.ActualValue = 0
		}

		// 计算达成率
		if achievement.TargetValue > 0 {
			achievement.AchieveRate = achievement.ActualValue * 100 / achievement.TargetValue
		}

		// 计算剩余
		achievement.Remaining = achievement.TargetValue - achievement.ActualValue
		if achievement.Remaining < 0 {
			achievement.Remaining = 0
		}

		// 判断状态
		if achievement.AchieveRate >= 100 {
			achievement.Status = "exceeded"
		} else if achievement.AchieveRate >= 80 {
			achievement.Status = "achieved"
		} else if achievement.AchieveRate >= 50 {
			achievement.Status = "progress"
		} else {
			achievement.Status = "failed"
		}

		result.Targets = append(result.Targets, achievement)

		totalTarget += achievement.TargetValue
		totalAchieved += achievement.ActualValue
	}

	// 计算总体达成率
	if totalTarget > 0 {
		result.OverallRate = totalAchieved * 100 / totalTarget
	}

	// 判断总体状态
	if result.OverallRate >= 100 {
		result.Status = "exceeded"
	} else if result.OverallRate >= 80 {
		result.Status = "achieved"
	} else if result.OverallRate >= 50 {
		result.Status = "progress"
	} else {
		result.Status = "failed"
	}

	return result, nil
}

// TrendCompare 趋势对比
type TrendCompare struct {
	Period      string      `json:"period"`
	CurrentData []TrendData `json:"current_data"`
	CompareData []TrendData `json:"compare_data"`
	Summary     TrendSummary `json:"summary"`
}

// TrendSummary 趋势摘要
type TrendSummary struct {
	CurrentTotal float64 `json:"current_total"`
	CompareTotal float64 `json:"compare_total"`
	ChangeRate   float64 `json:"change_rate"`
	Trend        string  `json:"trend"` // up/down/stable
}

// GetTrendCompare 趋势对比分析
func (s *CompareAnalysisService) GetTrendCompare(tenantID int64, metricType string, currentStart, currentEnd, compareStart, compareEnd time.Time) (*TrendCompare, error) {
	result := &TrendCompare{
		CurrentData: make([]TrendData, 0),
		CompareData: make([]TrendData, 0),
	}

	// 这里简化处理，实际应该按天获取趋势数据
	_ = tenantID
	_ = metricType

	result.Period = currentStart.Format("2006-01-02") + " vs " + compareStart.Format("2006-01-02")

	return result, nil
}

// MultiPeriodCompare 多期间对比
type MultiPeriodCompare struct {
	Periods []PeriodData `json:"periods"`
	Metrics []MetricTrend `json:"metrics"`
}

// PeriodData 期间数据
type PeriodData struct {
	PeriodName  string  `json:"period_name"`
	StartDate   string  `json:"start_date"`
	EndDate     string  `json:"end_date"`
	OrderCount  int     `json:"order_count"`
	OrderAmount float64 `json:"order_amount"`
}

// MetricTrend 指标趋势
type MetricTrend struct {
	MetricName string    `json:"metric_name"`
	Values     []float64 `json:"values"`
}

// GetMultiPeriodCompare 多期间对比
func (s *CompareAnalysisService) GetMultiPeriodCompare(tenantID int64, periods []PeriodRange) (*MultiPeriodCompare, error) {
	result := &MultiPeriodCompare{
		Periods: make([]PeriodData, 0),
		Metrics: make([]MetricTrend, 0),
	}

	for _, period := range periods {
		data, err := s.repo.GetPeriodCompare(tenantID, period.Start, period.End, period.Start, period.End)
		if err != nil {
			continue
		}

		periodData := PeriodData{
			PeriodName: period.Start.Format("2006-01"),
			StartDate:  period.Start.Format("2006-01-02"),
			EndDate:    period.End.Format("2006-01-02"),
		}

		if v, ok := data["current_orders"].(int64); ok {
			periodData.OrderCount = int(v)
		}
		if v, ok := data["current_amount"].(float64); ok {
			periodData.OrderAmount = v
		}

		result.Periods = append(result.Periods, periodData)
	}

	return result, nil
}
