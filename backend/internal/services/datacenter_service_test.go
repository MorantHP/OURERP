package services

import (
	"testing"
	"time"

	"github.com/MorantHP/OURERP/internal/models"
	"github.com/MorantHP/OURERP/internal/repository"
	_ "github.com/MorantHP/OURERP/internal/repository" // 确保包被导入
)

// 测试预警服务 - GetAlertTypes
func TestAlertService_GetAlertTypes(t *testing.T) {
	svc := NewAlertService(nil, nil, nil)

	types := svc.GetAlertTypes()

	if len(types) != 5 {
		t.Errorf("期望5种预警类型, 实际得到 %d", len(types))
	}

	expectedTypes := []string{"inventory", "sales", "order", "customer", "finance"}
	for i, et := range expectedTypes {
		if types[i].Type != et {
			t.Errorf("期望类型 %s, 实际得到 %s", et, types[i].Type)
		}
	}
}

// 测试预警服务 - GetNotifyLevels
func TestAlertService_GetNotifyLevels(t *testing.T) {
	svc := NewAlertService(nil, nil, nil)

	levels := svc.GetNotifyLevels()

	if len(levels) != 3 {
		t.Errorf("期望3种预警级别, 实际得到 %d", len(levels))
	}

	expectedLevels := []string{"info", "warning", "critical"}
	for i, el := range expectedLevels {
		if levels[i].Level != el {
			t.Errorf("期望级别 %s, 实际得到 %s", el, levels[i].Level)
		}
	}
}

// 测试实时监控服务 - 结构测试
func TestRealtimeService_Structure(t *testing.T) {
	// 测试服务可以正确初始化
	t.Run("服务初始化", func(t *testing.T) {
		// 需要一个实际的repository，这里跳过
		t.Skip("需要实际的数据库连接或mock")
	})
}

// 测试客户分析服务 - 结构测试
func TestCustomerAnalysisService_Structure(t *testing.T) {
	t.Run("服务初始化", func(t *testing.T) {
		t.Skip("需要实际的数据库连接或mock")
	})
}

// 测试商品分析服务 - 结构测试
func TestProductAnalysisService_Structure(t *testing.T) {
	t.Run("服务初始化", func(t *testing.T) {
		t.Skip("需要实际的数据库连接或mock")
	})
}

// 测试对比分析服务 - 结构测试
func TestCompareAnalysisService_Structure(t *testing.T) {
	t.Run("服务初始化", func(t *testing.T) {
		t.Skip("需要实际的数据库连接或mock")
	})
}

// =============== 模型测试 ===============

func TestAlertRule_Model(t *testing.T) {
	rule := models.AlertRule{
		ID:          1,
		TenantID:    1,
		Name:        "测试预警",
		Type:        "inventory",
		Threshold:   10,
		NotifyType:  "system",
		Level:       "warning",
		Status:      1,
		Description: "这是一个测试预警规则",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if rule.Name != "测试预警" {
		t.Errorf("规则名称不匹配")
	}
	if rule.Type != "inventory" {
		t.Errorf("规则类型不匹配")
	}
	if rule.Status != 1 {
		t.Errorf("规则状态不匹配")
	}
}

func TestAlertRecord_Model(t *testing.T) {
	now := time.Now()
	record := models.AlertRecord{
		ID:         1,
		TenantID:   1,
		RuleID:     1,
		Title:      "库存预警",
		Content:    "商品库存不足",
		Level:      "warning",
		SourceType: "product",
		SourceID:   100,
		Status:     0,
		CreatedAt:  now,
	}

	if record.Title != "库存预警" {
		t.Errorf("预警标题不匹配")
	}
	if record.Status != 0 {
		t.Errorf("预警状态不匹配")
	}
}

func TestCustomer_Model(t *testing.T) {
	now := time.Now()
	customer := models.Customer{
		ID:           1,
		TenantID:     1,
		Code:         "C001",
		Name:         "测试客户",
		Phone:        "13800138000",
		Type:         "b2c",
		Level:        "normal",
		TotalOrders:  10,
		TotalAmount:  1000.00,
		FirstOrderAt: &now,
		LastOrderAt:  &now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if customer.Code != "C001" {
		t.Errorf("客户编码不匹配")
	}
	if customer.TotalOrders != 10 {
		t.Errorf("订单数不匹配")
	}
	if customer.Type != "b2c" {
		t.Errorf("客户类型不匹配")
	}
}

func TestRealtimeSnapshot_Model(t *testing.T) {
	snapshot := models.RealtimeSnapshot{
		ID:              1,
		TenantID:        1,
		SnapshotTime:    time.Now(),
		OrderCount:      100,
		OrderAmount:     10000.00,
		PendingOrders:   10,
		ShippedOrders:   50,
		CompletedOrders: 40,
		LowStockItems:   5,
		NewCustomers:    20,
	}

	if snapshot.OrderCount != 100 {
		t.Errorf("订单数不匹配")
	}
	if snapshot.PendingOrders != 10 {
		t.Errorf("待处理订单数不匹配")
	}
}

func TestReportTemplate_Model(t *testing.T) {
	template := models.ReportTemplate{
		ID:         1,
		TenantID:   1,
		Name:       "销售报表",
		Type:       "sales",
		DataSource: "orders",
		Columns:    `["date", "amount", "count"]`,
		ChartType:  "line",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if template.Name != "销售报表" {
		t.Errorf("模板名称不匹配")
	}
	if template.Type != "sales" {
		t.Errorf("模板类型不匹配")
	}
}

func TestProductAnalysis_Model(t *testing.T) {
	analysis := models.ProductAnalysis{
		ID:            1,
		TenantID:      1,
		ProductID:     100,
		AnalysisDate:  time.Now(),
		PeriodType:    "day",
		SalesCount:    50,
		SalesAmount:   5000.00,
		SalesQuantity: 100,
		ProductCost:   2000.00,
		GrossProfit:   3000.00,
		ProfitRate:    60.0,
		TurnoverRate:  45.5,
	}

	if analysis.SalesCount != 50 {
		t.Errorf("销售数量不匹配")
	}
	if analysis.ProfitRate != 60.0 {
		t.Errorf("利润率不匹配")
	}
}

func TestCompareAnalysis_Model(t *testing.T) {
	analysis := models.CompareAnalysis{
		ID:            1,
		TenantID:      1,
		AnalysisDate:  time.Now(),
		CompareType:   "yoy",
		CurrentPeriod: "2025-01",
		ComparePeriod: "2024-01",
		MetricType:    "sales",
		CurrentValue:  10000.00,
		CompareValue:  8000.00,
		ChangeValue:   2000.00,
		ChangeRate:    25.0,
	}

	if analysis.CompareType != "yoy" {
		t.Errorf("对比类型不匹配")
	}
	if analysis.ChangeRate != 25.0 {
		t.Errorf("变化率不匹配")
	}
}

// =============== 过滤器测试 ===============

func TestAlertRuleFilter_Defaults(t *testing.T) {
	filter := models.AlertRuleFilter{}

	if filter.Type != "" {
		t.Errorf("默认类型应为空")
	}
	if filter.Status != nil {
		t.Errorf("默认状态应为nil")
	}
}

func TestCustomerFilter_Fields(t *testing.T) {
	status := 1
	filter := models.CustomerFilter{
		Name:     "测试",
		Phone:    "13800138000",
		Type:     "b2c",
		Level:    "vip",
		Province: "广东",
		City:     "深圳",
		Status:   &status,
	}

	if filter.Name != "测试" {
		t.Errorf("名称不匹配")
	}
	if *filter.Status != 1 {
		t.Errorf("状态不匹配")
	}
}

// =============== 基准测试 ===============

func BenchmarkAlertService_GetAlertTypes(b *testing.B) {
	svc := NewAlertService(nil, nil, nil)

	for i := 0; i < b.N; i++ {
		svc.GetAlertTypes()
	}
}

func BenchmarkAlertService_GetNotifyLevels(b *testing.B) {
	svc := NewAlertService(nil, nil, nil)

	for i := 0; i < b.N; i++ {
		svc.GetNotifyLevels()
	}
}

// 确保编译时接口实现检查
var _ = repository.NewDatacenterRepository(nil)
