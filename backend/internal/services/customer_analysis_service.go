package services

import (
	"time"

	"github.com/MorantHP/OURERP/internal/repository"
)

type CustomerAnalysisService struct {
	repo *repository.DatacenterRepository
}

func NewCustomerAnalysisService(repo *repository.DatacenterRepository) *CustomerAnalysisService {
	return &CustomerAnalysisService{repo: repo}
}

// CustomerAnalysisResult 客户分析结果
type CustomerAnalysisResult struct {
	// 客户统计
	TotalCustomers    int     `json:"total_customers"`
	NewCustomers      int     `json:"new_customers"`
	ActiveCustomers   int     `json:"active_customers"`
	ReturnCustomers   int     `json:"return_customers"`

	// 比率
	RepurchaseRate    float64 `json:"repurchase_rate"`
	ActivationRate    float64 `json:"activation_rate"`

	// 价值统计
	AvgOrderValue     float64 `json:"avg_order_value"`
	AvgCustomerValue  float64 `json:"avg_customer_value"`
	TotalRevenue      float64 `json:"total_revenue"`

	// 分层统计
	VipCount          int     `json:"vip_count"`
	NormalCount       int     `json:"normal_count"`
	NewCount          int     `json:"new_count"`
}

// CustomerValueDistribution 客户价值分布
type CustomerValueDistribution struct {
	ValueLevel     string  `json:"value_level"`
	CustomerCount  int     `json:"customer_count"`
	TotalAmount    float64 `json:"total_amount"`
	Percentage     float64 `json:"percentage"`
}

// GeographyDistribution 地域分布
type GeographyDistribution struct {
	Province      string  `json:"province"`
	City          string  `json:"city"`
	OrderCount    int     `json:"order_count"`
	OrderAmount   float64 `json:"order_amount"`
	CustomerCount int     `json:"customer_count"`
	Percentage    float64 `json:"percentage"`
}

// CustomerTrend 客户趋势
type CustomerTrend struct {
	Date           string  `json:"date"`
	NewCustomers   int     `json:"new_customers"`
	ActiveCustomers int    `json:"active_customers"`
	RepurchaseRate float64 `json:"repurchase_rate"`
}

// GetCustomerAnalysis 获取客户分析
func (s *CustomerAnalysisService) GetCustomerAnalysis(tenantID int64, startDate, endDate time.Time) (*CustomerAnalysisResult, error) {
	result := &CustomerAnalysisResult{}

	// 获取复购率数据
	repurchaseData, err := s.repo.GetRepurchaseRate(tenantID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	if v, ok := repurchaseData["total_customers"].(int64); ok {
		result.TotalCustomers = int(v)
	}
	if v, ok := repurchaseData["repurchase_customers"].(int64); ok {
		result.ReturnCustomers = int(v)
	}
	if v, ok := repurchaseData["repurchase_rate"].(float64); ok {
		result.RepurchaseRate = v
	}

	// 活跃客户数(期间内下过单的客户)
	result.ActiveCustomers = result.TotalCustomers

	// 计算激活率
	if result.TotalCustomers > 0 {
		result.ActivationRate = float64(result.ActiveCustomers) * 100 / float64(result.TotalCustomers)
	}

	return result, nil
}

// GetValueDistribution 获取客户价值分布
func (s *CustomerAnalysisService) GetValueDistribution(tenantID int64) ([]CustomerValueDistribution, error) {
	data, err := s.repo.GetCustomerValueDistribution(tenantID)
	if err != nil {
		return nil, err
	}

	var totalCustomers int
	for _, item := range data {
		if v, ok := item["customer_count"].(int64); ok {
			totalCustomers += int(v)
		}
	}

	result := make([]CustomerValueDistribution, 0)
	for _, item := range data {
		dist := CustomerValueDistribution{}
		if v, ok := item["value_level"].(string); ok {
			dist.ValueLevel = v
		}
		if v, ok := item["customer_count"].(int64); ok {
			dist.CustomerCount = int(v)
			if totalCustomers > 0 {
				dist.Percentage = float64(v) * 100 / float64(totalCustomers)
			}
		}
		if v, ok := item["total_amount"].(float64); ok {
			dist.TotalAmount = v
		}
		result = append(result, dist)
	}

	return result, nil
}

// GetGeographyDistribution 获取地域分布
func (s *CustomerAnalysisService) GetGeographyDistribution(tenantID int64, startDate, endDate time.Time) ([]GeographyDistribution, error) {
	data, err := s.repo.GetSalesByRegion(tenantID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	var totalAmount float64
	for _, item := range data {
		if v, ok := item["order_amount"].(float64); ok {
			totalAmount += v
		}
	}

	result := make([]GeographyDistribution, 0)
	for _, item := range data {
		dist := GeographyDistribution{}
		if v, ok := item["province"].(string); ok {
			dist.Province = v
		}
		if v, ok := item["order_count"].(int64); ok {
			dist.OrderCount = int(v)
		}
		if v, ok := item["order_amount"].(float64); ok {
			dist.OrderAmount = v
			if totalAmount > 0 {
				dist.Percentage = v * 100 / totalAmount
			}
		}
		result = append(result, dist)
	}

	return result, nil
}

// GetCityDistribution 获取城市分布(按省份)
func (s *CustomerAnalysisService) GetCityDistribution(tenantID int64, province string, startDate, endDate time.Time) ([]GeographyDistribution, error) {
	data, err := s.repo.GetSalesByCity(tenantID, province, startDate, endDate)
	if err != nil {
		return nil, err
	}

	var totalAmount float64
	for _, item := range data {
		if v, ok := item["order_amount"].(float64); ok {
			totalAmount += v
		}
	}

	result := make([]GeographyDistribution, 0)
	for _, item := range data {
		dist := GeographyDistribution{}
		if v, ok := item["city"].(string); ok {
			dist.City = v
			dist.Province = province
		}
		if v, ok := item["order_count"].(int64); ok {
			dist.OrderCount = int(v)
		}
		if v, ok := item["order_amount"].(float64); ok {
			dist.OrderAmount = v
			if totalAmount > 0 {
				dist.Percentage = v * 100 / totalAmount
			}
		}
		result = append(result, dist)
	}

	return result, nil
}

// GetRepurchaseAnalysis 获取复购分析
func (s *CustomerAnalysisService) GetRepurchaseAnalysis(tenantID int64, startDate, endDate time.Time) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 复购率
	repurchaseData, err := s.repo.GetRepurchaseRate(tenantID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	result["repurchase_rate"] = repurchaseData["repurchase_rate"]
	result["total_customers"] = repurchaseData["total_customers"]
	result["repurchase_customers"] = repurchaseData["repurchase_customers"]

	// 计算平均购买次数
	// 这里简化处理，实际应该从订单表统计

	return result, nil
}

// CustomerSegment 客户分群
type CustomerSegment struct {
	Segment       string  `json:"segment"`
	Criteria      string  `json:"criteria"`
	CustomerCount int     `json:"customer_count"`
	TotalAmount   float64 `json:"total_amount"`
	AvgAmount     float64 `json:"avg_amount"`
}

// GetCustomerSegments 获取客户分群
func (s *CustomerAnalysisService) GetCustomerSegments(tenantID int64) ([]CustomerSegment, error) {
	// 基于RFM模型的客户分群
	segments := []CustomerSegment{
		{Segment: "高价值客户", Criteria: "消费>=10000且最近30天有购买"},
		{Segment: "主力客户", Criteria: "消费>=1000且最近60天有购买"},
		{Segment: "新客户", Criteria: "最近30天首次购买"},
		{Segment: "沉睡客户", Criteria: "超过90天未购买"},
		{Segment: "流失客户", Criteria: "超过180天未购买"},
	}

	// 这里简化返回分群定义，实际应该从数据库计算
	return segments, nil
}

// CustomerLifecycle 客户生命周期
type CustomerLifecycle struct {
	Stage         string  `json:"stage"`
	CustomerCount int     `json:"customer_count"`
	Percentage    float64 `json:"percentage"`
	StageValue    float64 `json:"stage_value"` // 该阶段客户总价值
}

// GetCustomerLifecycle 获取客户生命周期分布
func (s *CustomerAnalysisService) GetCustomerLifecycle(tenantID int64) ([]CustomerLifecycle, error) {
	now := time.Now()

	// 计算各阶段客户
	lifecycles := []CustomerLifecycle{
		{Stage: "引入期", StageValue: 0},   // 新客户，首次购买
		{Stage: "成长期", StageValue: 0},   // 2-3次购买
		{Stage: "成熟期", StageValue: 0},   // 4次以上购买
		{Stage: "衰退期", StageValue: 0},   // 60-90天未购买
		{Stage: "流失期", StageValue: 0},   // 90天以上未购买
	}

	_ = now // 避免未使用警告

	// 实际应该从数据库统计，这里简化处理
	return lifecycles, nil
}
