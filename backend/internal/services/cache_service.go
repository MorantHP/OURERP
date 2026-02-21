package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MorantHP/OURERP/internal/cache"
	"github.com/MorantHP/OURERP/internal/models"
	"github.com/MorantHP/OURERP/internal/repository"
)

// CacheService 缓存服务
type CacheService struct {
	cache      cache.CacheService
	orderRepo  *repository.OrderRepository
	shopRepo   *repository.ShopRepository
	productRepo *repository.ProductRepository
}

// NewCacheService 创建缓存服务
func NewCacheService(
	cache cache.CacheService,
	orderRepo *repository.OrderRepository,
	shopRepo *repository.ShopRepository,
	productRepo *repository.ProductRepository,
) *CacheService {
	return &CacheService{
		cache:       cache,
		orderRepo:   orderRepo,
		shopRepo:    shopRepo,
		productRepo: productRepo,
	}
}

// 订单缓存操作

// GetOrderCount 获取订单数量（带缓存）
func (s *CacheService) GetOrderCount(ctx context.Context, tenantID int64) (int64, error) {
	key := cache.BuildKey(cache.CacheKeyOrderCount, tenantID)

	var count int64
	err := s.cache.Get(ctx, key, &count)
	if err == nil {
		return count, nil
	}

	// 从数据库获取
	count, err = s.orderRepo.CountByTenantID(tenantID)
	if err != nil {
		return 0, err
	}

	// 缓存结果
	_ = s.cache.Set(ctx, key, count, cache.TTLShort)
	return count, nil
}

// InvalidateOrderCount 使订单数量缓存失效
func (s *CacheService) InvalidateOrderCount(ctx context.Context, tenantID int64) error {
	key := cache.BuildKey(cache.CacheKeyOrderCount, tenantID)
	return s.cache.Delete(ctx, key)
}

// 店铺缓存操作

// GetShop 获取店铺（带缓存）
func (s *CacheService) GetShop(ctx context.Context, shopID int64) (*models.Shop, error) {
	key := cache.BuildKey(cache.CacheKeyShop, shopID)

	var shop models.Shop
	err := s.cache.Get(ctx, key, &shop)
	if err == nil {
		return &shop, nil
	}

	// 从数据库获取
	result, err := s.shopRepo.FindByID(shopID)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	_ = s.cache.Set(ctx, key, result, cache.TTLMedium)
	return result, nil
}

// InvalidateShop 使店铺缓存失效
func (s *CacheService) InvalidateShop(ctx context.Context, shopID int64) error {
	key := cache.BuildKey(cache.CacheKeyShop, shopID)
	return s.cache.Delete(ctx, key)
}

// 统计数据缓存操作

// GetStatistics 获取统计数据（带缓存）
func (s *CacheService) GetStatistics(ctx context.Context, tenantID int64, date string, dest interface{}) error {
	key := fmt.Sprintf("statistics:%d:%s", tenantID, date)
	return s.cache.Get(ctx, key, dest)
}

// SetStatistics 设置统计数据缓存
func (s *CacheService) SetStatistics(ctx context.Context, tenantID int64, date string, data interface{}) error {
	key := fmt.Sprintf("statistics:%d:%s", tenantID, date)
	return s.cache.Set(ctx, key, data, cache.TTLShort)
}

// InvalidateStatistics 使统计数据缓存失效
func (s *CacheService) InvalidateStatistics(ctx context.Context, tenantID int64) error {
	pattern := fmt.Sprintf("statistics:%d:*", tenantID)
	return s.cache.DeletePattern(ctx, pattern)
}

// 实时数据缓存

// GetRealtimeOverview 获取实时概览（带缓存，短TTL）
func (s *CacheService) GetRealtimeOverview(ctx context.Context, tenantID int64) (*RealtimeOverviewData, error) {
	key := cache.BuildKey(cache.CacheKeyRealtimeOverview, tenantID)

	var data RealtimeOverviewData
	err := s.cache.Get(ctx, key, &data)
	if err == nil {
		return &data, nil
	}

	return nil, err
}

// SetRealtimeOverview 设置实时概览缓存
func (s *CacheService) SetRealtimeOverview(ctx context.Context, tenantID int64, data *RealtimeOverviewData) error {
	key := cache.BuildKey(cache.CacheKeyRealtimeOverview, tenantID)
	// 实时数据缓存时间较短
	return s.cache.Set(ctx, key, data, time.Minute)
}

// RealtimeOverviewData 实时概览数据
type RealtimeOverviewData struct {
	OrderCount     int64   `json:"order_count"`
	OrderAmount    float64 `json:"order_amount"`
	TodayOrders    int64   `json:"today_orders"`
	TodayAmount    float64 `json:"today_amount"`
	PendingOrders  int64   `json:"pending_orders"`
	LowStockItems  int64   `json:"low_stock_items"`
}

// 预警缓存

// GetAlertRules 获取预警规则（带缓存）
func (s *CacheService) GetAlertRules(ctx context.Context, tenantID int64) ([]models.AlertRule, error) {
	key := cache.BuildKey(cache.CacheKeyAlertRules, tenantID)

	var rules []models.AlertRule
	err := s.cache.Get(ctx, key, &rules)
	if err == nil {
		return rules, nil
	}

	return nil, err
}

// SetAlertRules 设置预警规则缓存
func (s *CacheService) SetAlertRules(ctx context.Context, tenantID int64, rules []models.AlertRule) error {
	key := cache.BuildKey(cache.CacheKeyAlertRules, tenantID)
	return s.cache.Set(ctx, key, rules, cache.TTLMedium)
}

// InvalidateAlertRules 使预警规则缓存失效
func (s *CacheService) InvalidateAlertRules(ctx context.Context, tenantID int64) error {
	key := cache.BuildKey(cache.CacheKeyAlertRules, tenantID)
	return s.cache.Delete(ctx, key)
}

// 批量操作

// InvalidateTenant 使租户所有缓存失效
func (s *CacheService) InvalidateTenant(ctx context.Context, tenantID int64) error {
	patterns := []string{
		fmt.Sprintf("order*:%d:*", tenantID),
		fmt.Sprintf("shop*:%d:*", tenantID),
		fmt.Sprintf("product*:%d:*", tenantID),
		fmt.Sprintf("statistics:%d:*", tenantID),
		fmt.Sprintf("inventory*:%d:*", tenantID),
		fmt.Sprintf("realtime*:%d:*", tenantID),
		fmt.Sprintf("alert*:%d:*", tenantID),
	}

	for _, pattern := range patterns {
		if err := s.cache.DeletePattern(ctx, pattern); err != nil {
			return err
		}
	}

	return nil
}

// 预热缓存

// WarmupCache 预热缓存
func (s *CacheService) WarmupCache(ctx context.Context, tenantID int64) error {
	// 预热订单数量
	_, _ = s.GetOrderCount(ctx, tenantID)

	// 预热实时概览（会在服务层处理）

	return nil
}

// 辅助函数

// serialize 序列化数据
func serialize(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

// deserialize 反序列化数据
func deserialize(data []byte, dest interface{}) error {
	return json.Unmarshal(data, dest)
}
