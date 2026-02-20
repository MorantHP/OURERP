package services

import (
	"time"

	"github.com/MorantHP/OURERP/internal/models"
	"github.com/MorantHP/OURERP/internal/repository"
)

type RealtimeService struct {
	repo *repository.DatacenterRepository
}

func NewRealtimeService(repo *repository.DatacenterRepository) *RealtimeService {
	return &RealtimeService{repo: repo}
}

// RealtimeOverview 实时概览数据
type RealtimeOverview struct {
	// 今日数据
	TodayOrderCount    int     `json:"today_order_count"`
	TodayOrderAmount   float64 `json:"today_order_amount"`
	TodayPaidAmount    float64 `json:"today_paid_amount"`
	TodayRefundCount   int     `json:"today_refund_count"`
	TodayRefundAmount  float64 `json:"today_refund_amount"`
	TodayNewCustomers  int     `json:"today_new_customers"`

	// 订单状态统计
	PendingOrders   int `json:"pending_orders"`
	ShippedOrders   int `json:"shipped_orders"`
	CompletedOrders int `json:"completed_orders"`
	CancelledOrders int `json:"cancelled_orders"`

	// 库存状态
	LowStockItems   int `json:"low_stock_items"`
	OutOfStockItems int `json:"out_of_stock_items"`

	// 未处理预警
	UnhandledAlerts int `json:"unhandled_alerts"`

	// 客单价
	AvgOrderValue float64 `json:"avg_order_value"`

	// 实时时间
	SnapshotTime time.Time `json:"snapshot_time"`
}

// OrderStream 订单流数据
type OrderStream struct {
	OrderNo     string    `json:"order_no"`
	ShopName    string    `json:"shop_name"`
	Amount      float64   `json:"amount"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	BuyerName   string    `json:"buyer_name"`
	Province    string    `json:"province"`
}

// InventoryStatus 库存状态
type InventoryStatus struct {
	TotalProducts   int     `json:"total_products"`
	NormalProducts  int     `json:"normal_products"`
	LowStockProducts int    `json:"low_stock_products"`
	OutOfStockProducts int  `json:"out_of_stock_products"`
	StockValue      float64 `json:"stock_value"`
	LevelDistribution []LevelDistribution `json:"level_distribution"`
}

type LevelDistribution struct {
	Level string `json:"level"`
	Count int    `json:"count"`
}

// GetOverview 获取实时概览数据
func (s *RealtimeService) GetOverview(tenantID int64) (*RealtimeOverview, error) {
	// 获取今日开始时间
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 获取实时统计
	stats, err := s.repo.GetRealtimeStats(tenantID, todayStart)
	if err != nil {
		return nil, err
	}

	overview := &RealtimeOverview{
		SnapshotTime: now,
	}

	// 今日数据
	if v, ok := stats["order_count"].(int64); ok {
		overview.TodayOrderCount = int(v)
	}
	if v, ok := stats["order_amount"].(float64); ok {
		overview.TodayOrderAmount = v
	}
	if v, ok := stats["paid_amount"].(float64); ok {
		overview.TodayPaidAmount = v
	}
	if v, ok := stats["refund_count"].(int64); ok {
		overview.TodayRefundCount = int(v)
	}
	if v, ok := stats["refund_amount"].(float64); ok {
		overview.TodayRefundAmount = v
	}
	if v, ok := stats["new_customers"].(int64); ok {
		overview.TodayNewCustomers = int(v)
	}

	// 订单状态
	if v, ok := stats["pending_orders"].(int64); ok {
		overview.PendingOrders = int(v)
	}

	// 库存预警
	if v, ok := stats["low_stock_items"].(int64); ok {
		overview.LowStockItems = int(v)
	}

	// 计算客单价
	if overview.TodayOrderCount > 0 {
		overview.AvgOrderValue = overview.TodayOrderAmount / float64(overview.TodayOrderCount)
	}

	// 获取未处理预警数
	unhandledAlerts, _ := s.repo.CountUnhandledAlerts(tenantID)
	overview.UnhandledAlerts = int(unhandledAlerts)

	return overview, nil
}

// CreateSnapshot 创建实时快照
func (s *RealtimeService) CreateSnapshot(tenantID int64) error {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	stats, err := s.repo.GetRealtimeStats(tenantID, todayStart)
	if err != nil {
		return err
	}

	snapshot := &models.RealtimeSnapshot{
		TenantID:     tenantID,
		SnapshotTime: now,
	}

	if v, ok := stats["order_count"].(int64); ok {
		snapshot.OrderCount = int(v)
	}
	if v, ok := stats["order_amount"].(float64); ok {
		snapshot.OrderAmount = v
	}
	if v, ok := stats["paid_amount"].(float64); ok {
		snapshot.PaidAmount = v
	}
	if v, ok := stats["refund_count"].(int64); ok {
		snapshot.RefundCount = int(v)
	}
	if v, ok := stats["refund_amount"].(float64); ok {
		snapshot.RefundAmount = v
	}
	if v, ok := stats["pending_orders"].(int64); ok {
		snapshot.PendingOrders = int(v)
	}
	if v, ok := stats["low_stock_items"].(int64); ok {
		snapshot.LowStockItems = int(v)
	}
	if v, ok := stats["new_customers"].(int64); ok {
		snapshot.NewCustomers = int(v)
	}

	return s.repo.CreateSnapshot(snapshot)
}

// GetInventoryStatus 获取库存状态
func (s *RealtimeService) GetInventoryStatus(tenantID int64) (*InventoryStatus, error) {
	levels, err := s.repo.GetInventoryLevel(tenantID)
	if err != nil {
		return nil, err
	}

	status := &InventoryStatus{
		LevelDistribution: make([]LevelDistribution, 0),
	}

	levelCounts := make(map[string]int)
	for _, item := range levels {
		status.TotalProducts++
		if level, ok := item["stock_level"].(string); ok {
			levelCounts[level]++
			switch level {
			case "normal":
				status.NormalProducts++
			case "low":
				status.LowStockProducts++
			case "out_of_stock":
				status.OutOfStockProducts++
			}
		}
	}

	// 构建分布数据
	for level, count := range levelCounts {
		status.LevelDistribution = append(status.LevelDistribution, LevelDistribution{
			Level: level,
			Count: count,
		})
	}

	return status, nil
}

// HourlyTrend 小时趋势
type HourlyTrend struct {
	Hour        int     `json:"hour"`
	OrderCount  int     `json:"order_count"`
	OrderAmount float64 `json:"order_amount"`
}

// GetHourlyTrend 获取今日小时趋势
func (s *RealtimeService) GetHourlyTrend(tenantID int64) ([]HourlyTrend, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 这里简化处理，实际应该从订单表按小时分组统计
	// 返回24小时的趋势数据
	trends := make([]HourlyTrend, 24)
	for i := 0; i < 24; i++ {
		trends[i] = HourlyTrend{
			Hour:        i,
			OrderCount:  0,
			OrderAmount: 0,
		}
	}

	// 从快照获取数据（如果有的话）
	snapshots, err := s.repo.GetSnapshotsByPeriod(tenantID, todayStart, now)
	if err == nil && len(snapshots) > 0 {
		for _, snap := range snapshots {
			hour := snap.SnapshotTime.Hour()
			trends[hour].OrderCount += snap.OrderCount
			trends[hour].OrderAmount += snap.OrderAmount
		}
	}

	return trends, nil
}
