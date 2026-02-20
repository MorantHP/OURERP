package services

import (
	"context"
	"time"

	"github.com/MorantHP/OURERP/internal/models"
	"gorm.io/gorm"
)

// StatsRequest 统计请求参数
type StatsRequest struct {
	TenantID  int64
	StartDate *time.Time
	EndDate   *time.Time
	Platform  string
	ShopID    int64
	Category  string
	Brand     string
}

// OverviewStats 总览统计
type OverviewStats struct {
	OrderCount    int64   `json:"order_count"`
	SalesAmount   float64 `json:"sales_amount"`
	PayAmount     float64 `json:"pay_amount"`
	ItemCount     int64   `json:"item_count"`
	AvgOrderValue float64 `json:"avg_order_value"`
}

// OverviewResponse 总览响应
type OverviewResponse struct {
	Today     OverviewStats `json:"today"`
	Yesterday OverviewStats `json:"yesterday"`
	ThisWeek  OverviewStats `json:"this_week"`
	ThisMonth OverviewStats `json:"this_month"`
	Growth    GrowthStats   `json:"growth"`
}

// GrowthStats 增长率
type GrowthStats struct {
	OrderCountGrowth  float64 `json:"order_count_growth"`
	SalesAmountGrowth float64 `json:"sales_amount_growth"`
	PayAmountGrowth   float64 `json:"pay_amount_growth"`
}

// TrendDataPoint 趋势数据点
type TrendDataPoint struct {
	Date        string  `json:"date"`
	OrderCount  int64   `json:"order_count"`
	SalesAmount float64 `json:"sales_amount"`
	PayAmount   float64 `json:"pay_amount"`
}

// TrendResponse 趋势响应
type TrendResponse struct {
	Period string           `json:"period"`
	Data   []TrendDataPoint `json:"data"`
}

// DimensionData 维度数据
type DimensionData struct {
	Key         string  `json:"key"`
	Label       string  `json:"label"`
	OrderCount  int64   `json:"order_count"`
	SalesAmount float64 `json:"sales_amount"`
	PayAmount   float64 `json:"pay_amount"`
	Percentage  float64 `json:"percentage"`
}

// DimensionResponse 维度响应
type DimensionResponse struct {
	Dimension string          `json:"dimension"`
	Data      []DimensionData `json:"data"`
}

// FunnelStep 漏斗步骤
type FunnelStep struct {
	Name      string  `json:"name"`
	Count     int64   `json:"count"`
	Percent   float64 `json:"percent"`
	Status    string  `json:"status"`
	StatusInt int     `json:"status_int"`
}

// FunnelResponse 漏斗响应
type FunnelResponse struct {
	Data []FunnelStep `json:"data"`
}

// TopProduct 热销商品
type TopProduct struct {
	SkuID       int64   `json:"sku_id"`
	SkuName     string  `json:"sku_name"`
	Quantity    int64   `json:"quantity"`
	SalesAmount float64 `json:"sales_amount"`
	OrderCount  int64   `json:"order_count"`
}

// TopProductsResponse 热销商品响应
type TopProductsResponse struct {
	Data []TopProduct `json:"data"`
}

// PlatformLabels 平台名称映射
var PlatformLabels = map[string]string{
	"taobao":       "淘宝",
	"tmall":        "天猫",
	"douyin":       "抖音",
	"kuaishou":     "快手",
	"wechat_video": "微信视频号",
	"custom":       "其他平台",
}

// StatisticsService 统计服务
type StatisticsService struct {
	db *gorm.DB
}

// NewStatisticsService 创建统计服务
func NewStatisticsService(db *gorm.DB) *StatisticsService {
	return &StatisticsService{db: db}
}

// GetOverview 获取总览统计
func (s *StatisticsService) GetOverview(ctx context.Context, req *StatsRequest) (*OverviewResponse, error) {
	now := time.Now()

	// 今日
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.Add(24 * time.Hour)

	// 昨日
	yesterdayStart := todayStart.Add(-24 * time.Hour)
	yesterdayEnd := todayStart

	// 本周
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	weekStart := todayStart.Add(time.Duration(-(weekday - 1)) * 24 * time.Hour)

	// 本月
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	resp := &OverviewResponse{
		Today:     s.getStats(req, &todayStart, &todayEnd),
		Yesterday: s.getStats(req, &yesterdayStart, &yesterdayEnd),
		ThisWeek:  s.getStats(req, &weekStart, nil),
		ThisMonth: s.getStats(req, &monthStart, nil),
	}

	// 计算增长率
	if resp.Yesterday.OrderCount > 0 {
		resp.Growth.OrderCountGrowth = float64(resp.Today.OrderCount-resp.Yesterday.OrderCount) / float64(resp.Yesterday.OrderCount)
	}
	if resp.Yesterday.SalesAmount > 0 {
		resp.Growth.SalesAmountGrowth = (resp.Today.SalesAmount - resp.Yesterday.SalesAmount) / resp.Yesterday.SalesAmount
	}
	if resp.Yesterday.PayAmount > 0 {
		resp.Growth.PayAmountGrowth = (resp.Today.PayAmount - resp.Yesterday.PayAmount) / resp.Yesterday.PayAmount
	}

	return resp, nil
}

// getStats 获取指定时间范围的统计
func (s *StatisticsService) getStats(req *StatsRequest, startDate, endDate *time.Time) OverviewStats {
	var stats OverviewStats

	query := s.db.Model(&models.Order{}).Where("tenant_id = ?", req.TenantID)

	if startDate != nil {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != nil {
		query = query.Where("created_at < ?", endDate)
	}
	if req.Platform != "" {
		query = query.Where("platform = ?", req.Platform)
	}
	if req.ShopID > 0 {
		query = query.Where("shop_id = ?", req.ShopID)
	}

	// 订单数和金额
	var result struct {
		OrderCount int64
		SalesSum   float64
		PaySum     float64
	}

	query.Select("COUNT(*) as order_count, COALESCE(SUM(total_amount), 0) as sales_sum, COALESCE(SUM(pay_amount), 0) as pay_sum").
		Scan(&result)

	stats.OrderCount = result.OrderCount
	stats.SalesAmount = result.SalesSum
	stats.PayAmount = result.PaySum

	// 计算商品件数
	var itemCount int64
	s.db.Model(&models.OrderItem{}).
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.tenant_id = ?", req.TenantID).
		Scan(&itemCount)
	stats.ItemCount = itemCount

	// 计算客单价
	if stats.OrderCount > 0 {
		stats.AvgOrderValue = stats.PayAmount / float64(stats.OrderCount)
	}

	return stats
}

// GetSalesTrend 获取销售趋势
func (s *StatisticsService) GetSalesTrend(ctx context.Context, req *StatsRequest, period string) (*TrendResponse, error) {
	now := time.Now()
	var startDate time.Time
	var dateFormat string

	switch period {
	case "week":
		startDate = now.AddDate(0, 0, -7)
		dateFormat = "%Y-%m-%d"
	case "month":
		startDate = now.AddDate(0, -1, 0)
		dateFormat = "%Y-%m-%d"
	default: // day
		startDate = now.AddDate(0, 0, -30)
		dateFormat = "%Y-%m-%d"
	}

	var results []struct {
		Date        string
		OrderCount  int64
		SalesAmount float64
		PayAmount   float64
	}

	query := s.db.Model(&models.Order{}).
		Select("strftime(?, created_at) as date, COUNT(*) as order_count, SUM(total_amount) as sales_amount, SUM(pay_amount) as pay_amount", dateFormat).
		Where("tenant_id = ?", req.TenantID).
		Where("created_at >= ?", startDate).
		Group("date").
		Order("date")

	if req.Platform != "" {
		query = query.Where("platform = ?", req.Platform)
	}
	if req.ShopID > 0 {
		query = query.Where("shop_id = ?", req.ShopID)
	}

	query.Scan(&results)

	data := make([]TrendDataPoint, len(results))
	for i, r := range results {
		data[i] = TrendDataPoint{
			Date:        r.Date,
			OrderCount:  r.OrderCount,
			SalesAmount: r.SalesAmount,
			PayAmount:   r.PayAmount,
		}
	}

	return &TrendResponse{
		Period: period,
		Data:   data,
	}, nil
}

// GetByPlatform 按平台统计
func (s *StatisticsService) GetByPlatform(ctx context.Context, req *StatsRequest) (*DimensionResponse, error) {
	return s.getByDimension(ctx, req, "platform", PlatformLabels)
}

// GetByShop 按店铺统计
func (s *StatisticsService) GetByShop(ctx context.Context, req *StatsRequest) (*DimensionResponse, error) {
	// 获取店铺名称映射
	var shops []models.Shop
	s.db.Where("tenant_id = ?", req.TenantID).Find(&shops)
	shopLabels := make(map[string]string)
	for _, shop := range shops {
		shopLabels[string(rune(shop.ID))] = shop.Name
	}

	return s.getByDimension(ctx, req, "shop_id", shopLabels)
}

// GetByCategory 按品类统计
func (s *StatisticsService) GetByCategory(ctx context.Context, req *StatsRequest) (*DimensionResponse, error) {
	var results []struct {
		Category    string
		OrderCount  int64
		SalesAmount float64
		PayAmount   float64
	}

	query := s.db.Model(&models.OrderItem{}).
		Select("p.category, COUNT(DISTINCT o.id) as order_count, SUM(oi.price * oi.quantity) as sales_amount, SUM(oi.price * oi.quantity) as pay_amount").
		Joins("JOIN orders o ON o.id = order_items.order_id").
		Joins("LEFT JOIN products p ON p.sku_code = order_items.sku_name").
		Where("o.tenant_id = ?", req.TenantID)

	if req.StartDate != nil {
		query = query.Where("o.created_at >= ?", req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("o.created_at <= ?", req.EndDate)
	}
	if req.Platform != "" {
		query = query.Where("o.platform = ?", req.Platform)
	}
	if req.ShopID > 0 {
		query = query.Where("o.shop_id = ?", req.ShopID)
	}

	query.Group("p.category").Order("sales_amount DESC").Scan(&results)

	var totalSales float64
	for _, r := range results {
		totalSales += r.SalesAmount
	}

	data := make([]DimensionData, len(results))
	for i, r := range results {
		percentage := 0.0
		if totalSales > 0 {
			percentage = r.SalesAmount / totalSales * 100
		}
		label := r.Category
		if label == "" {
			label = "未分类"
		}
		data[i] = DimensionData{
			Key:         r.Category,
			Label:       label,
			OrderCount:  r.OrderCount,
			SalesAmount: r.SalesAmount,
			PayAmount:   r.PayAmount,
			Percentage:  percentage,
		}
	}

	return &DimensionResponse{
		Dimension: "category",
		Data:      data,
	}, nil
}

// GetByBrand 按品牌统计
func (s *StatisticsService) GetByBrand(ctx context.Context, req *StatsRequest) (*DimensionResponse, error) {
	var results []struct {
		Brand       string
		OrderCount  int64
		SalesAmount float64
		PayAmount   float64
	}

	query := s.db.Model(&models.OrderItem{}).
		Select("p.brand, COUNT(DISTINCT o.id) as order_count, SUM(oi.price * oi.quantity) as sales_amount, SUM(oi.price * oi.quantity) as pay_amount").
		Joins("JOIN orders o ON o.id = order_items.order_id").
		Joins("LEFT JOIN products p ON p.sku_code = order_items.sku_name").
		Where("o.tenant_id = ?", req.TenantID)

	if req.StartDate != nil {
		query = query.Where("o.created_at >= ?", req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("o.created_at <= ?", req.EndDate)
	}
	if req.Platform != "" {
		query = query.Where("o.platform = ?", req.Platform)
	}
	if req.ShopID > 0 {
		query = query.Where("o.shop_id = ?", req.ShopID)
	}

	query.Group("p.brand").Order("sales_amount DESC").Scan(&results)

	var totalSales float64
	for _, r := range results {
		totalSales += r.SalesAmount
	}

	data := make([]DimensionData, len(results))
	for i, r := range results {
		percentage := 0.0
		if totalSales > 0 {
			percentage = r.SalesAmount / totalSales * 100
		}
		label := r.Brand
		if label == "" {
			label = "其他品牌"
		}
		data[i] = DimensionData{
			Key:         r.Brand,
			Label:       label,
			OrderCount:  r.OrderCount,
			SalesAmount: r.SalesAmount,
			PayAmount:   r.PayAmount,
			Percentage:  percentage,
		}
	}

	return &DimensionResponse{
		Dimension: "brand",
		Data:      data,
	}, nil
}

// getByDimension 通用维度统计
func (s *StatisticsService) getByDimension(ctx context.Context, req *StatsRequest, field string, labels map[string]string) (*DimensionResponse, error) {
	var results []struct {
		Key         string
		OrderCount  int64
		SalesAmount float64
		PayAmount   float64
	}

	query := s.db.Model(&models.Order{}).
		Select(field + " as key, COUNT(*) as order_count, SUM(total_amount) as sales_amount, SUM(pay_amount) as pay_amount").
		Where("tenant_id = ?", req.TenantID)

	if req.StartDate != nil {
		query = query.Where("created_at >= ?", req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("created_at <= ?", req.EndDate)
	}
	if req.Platform != "" && field != "platform" {
		query = query.Where("platform = ?", req.Platform)
	}
	if req.ShopID > 0 && field != "shop_id" {
		query = query.Where("shop_id = ?", req.ShopID)
	}

	query.Group(field).Order("sales_amount DESC").Scan(&results)

	var totalSales float64
	for _, r := range results {
		totalSales += r.SalesAmount
	}

	data := make([]DimensionData, len(results))
	for i, r := range results {
		percentage := 0.0
		if totalSales > 0 {
			percentage = r.SalesAmount / totalSales * 100
		}
		label := r.Key
		if l, ok := labels[r.Key]; ok {
			label = l
		}
		if label == "" {
			label = "未知"
		}
		data[i] = DimensionData{
			Key:         r.Key,
			Label:       label,
			OrderCount:  r.OrderCount,
			SalesAmount: r.SalesAmount,
			PayAmount:   r.PayAmount,
			Percentage:  percentage,
		}
	}

	return &DimensionResponse{
		Dimension: field,
		Data:      data,
	}, nil
}

// GetOrderFunnel 获取订单漏斗
func (s *StatisticsService) GetOrderFunnel(ctx context.Context, req *StatsRequest) (*FunnelResponse, error) {
	var counts struct {
		Total      int64
		PendingPay int64
		PendingShip int64
		Shipped    int64
		Completed  int64
		Cancelled  int64
	}

	query := s.db.Model(&models.Order{}).Where("tenant_id = ?", req.TenantID)

	if req.StartDate != nil {
		query = query.Where("created_at >= ?", req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("created_at <= ?", req.EndDate)
	}
	if req.Platform != "" {
		query = query.Where("platform = ?", req.Platform)
	}
	if req.ShopID > 0 {
		query = query.Where("shop_id = ?", req.ShopID)
	}

	query.Select(`
		COUNT(*) as total,
		SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) as pending_pay,
		SUM(CASE WHEN status = 2 THEN 1 ELSE 0 END) as pending_ship,
		SUM(CASE WHEN status = 3 THEN 1 ELSE 0 END) as shipped,
		SUM(CASE WHEN status = 4 THEN 1 ELSE 0 END) as completed,
		SUM(CASE WHEN status = 5 THEN 1 ELSE 0 END) as cancelled
	`).Scan(&counts)

	total := counts.Total
	if total == 0 {
		total = 1
	}

	steps := []FunnelStep{
		{Name: "全部订单", Count: counts.Total, Status: "total", StatusInt: 0, Percent: 100},
		{Name: "待付款", Count: counts.PendingPay, Status: "pending_pay", StatusInt: 1, Percent: float64(counts.PendingPay) / float64(total) * 100},
		{Name: "待发货", Count: counts.PendingShip, Status: "pending_ship", StatusInt: 2, Percent: float64(counts.PendingShip) / float64(total) * 100},
		{Name: "已发货", Count: counts.Shipped, Status: "shipped", StatusInt: 3, Percent: float64(counts.Shipped) / float64(total) * 100},
		{Name: "已完成", Count: counts.Completed, Status: "completed", StatusInt: 4, Percent: float64(counts.Completed) / float64(total) * 100},
		{Name: "已取消", Count: counts.Cancelled, Status: "cancelled", StatusInt: 5, Percent: float64(counts.Cancelled) / float64(total) * 100},
	}

	return &FunnelResponse{Data: steps}, nil
}

// GetTopProducts 获取热销商品
func (s *StatisticsService) GetTopProducts(ctx context.Context, req *StatsRequest, limit int) (*TopProductsResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	var results []struct {
		SkuID       int64
		SkuName     string
		Quantity    int64
		SalesAmount float64
		OrderCount  int64
	}

	query := s.db.Model(&models.OrderItem{}).
		Select("sku_id, sku_name, SUM(quantity) as quantity, SUM(price * quantity) as sales_amount, COUNT(DISTINCT order_id) as order_count").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.tenant_id = ?", req.TenantID)

	if req.StartDate != nil {
		query = query.Where("orders.created_at >= ?", req.StartDate)
	}
	if req.EndDate != nil {
		query = query.Where("orders.created_at <= ?", req.EndDate)
	}
	if req.Platform != "" {
		query = query.Where("orders.platform = ?", req.Platform)
	}
	if req.ShopID > 0 {
		query = query.Where("orders.shop_id = ?", req.ShopID)
	}

	query.Group("sku_id, sku_name").Order("sales_amount DESC").Limit(limit).Scan(&results)

	data := make([]TopProduct, len(results))
	for i, r := range results {
		data[i] = TopProduct{
			SkuID:       r.SkuID,
			SkuName:     r.SkuName,
			Quantity:    r.Quantity,
			SalesAmount: r.SalesAmount,
			OrderCount:  r.OrderCount,
		}
	}

	return &TopProductsResponse{Data: data}, nil
}
