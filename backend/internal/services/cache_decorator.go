package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/MorantHP/OURERP/internal/cache"
)

// CacheDecorator 缓存装饰器
type CacheDecorator struct {
	cache  cache.CacheService
	prefix string
}

// NewCacheDecorator 创建缓存装饰器
func NewCacheDecorator(cache cache.CacheService, prefix string) *CacheDecorator {
	return &CacheDecorator{
		cache:  cache,
		prefix: prefix,
	}
}

// GetOrSet 获取或设置缓存
func (d *CacheDecorator) GetOrSet(ctx context.Context, key string, dest interface{}, ttl time.Duration, fn func() (interface{}, error)) error {
	fullKey := d.prefix + ":" + key

	// 尝试从缓存获取
	err := d.cache.Get(ctx, fullKey, dest)
	if err == nil {
		return nil
	}

	// 缓存未命中，执行函数
	data, err := fn()
	if err != nil {
		return err
	}

	// 设置缓存
	if err := d.cache.Set(ctx, fullKey, data, ttl); err != nil {
		// 缓存失败不影响业务
	}

	// 将结果复制到dest
	if bytes, err := json.Marshal(data); err == nil {
		json.Unmarshal(bytes, dest)
	}

	return nil
}

// Delete 删除缓存
func (d *CacheDecorator) Delete(ctx context.Context, key string) error {
	return d.cache.Delete(ctx, d.prefix+":"+key)
}

// DeletePattern 批量删除缓存
func (d *CacheDecorator) DeletePattern(ctx context.Context, pattern string) error {
	return d.cache.DeletePattern(ctx, d.prefix+":"+pattern)
}

// InvalidateProductCache 使商品缓存失效
func (d *CacheDecorator) InvalidateProductCache(ctx context.Context, productID int64, tenantID int64) error {
	keys := []string{
		cache.BuildKey(cache.CacheKeyProduct, tenantID, productID),
		cache.BuildKey(cache.CacheKeyProducts, tenantID),
		cache.BuildKey(cache.CacheKeyProductCount, tenantID),
	}
	for _, key := range keys {
		_ = d.cache.Delete(ctx, key)
	}
	return nil
}

// InvalidateInventoryCache 使库存缓存失效
func (d *CacheDecorator) InvalidateInventoryCache(ctx context.Context, productID int64, warehouseID int64, tenantID int64) error {
	keys := []string{
		cache.BuildKey(cache.CacheKeyInventory, tenantID, productID, warehouseID),
		cache.BuildKey(cache.CacheKeyInventoryAlert, tenantID),
	}
	for _, key := range keys {
		_ = d.cache.Delete(ctx, key)
	}
	return nil
}

// InvalidateOrderCache 使订单缓存失效
func (d *CacheDecorator) InvalidateOrderCache(ctx context.Context, tenantID int64) error {
	keys := []string{
		cache.BuildKey(cache.CacheKeyOrders, tenantID),
		cache.BuildKey(cache.CacheKeyOrderCount, tenantID),
		cache.BuildKey(cache.CacheKeyStatistics, tenantID),
	}
	for _, key := range keys {
		_ = d.cache.Delete(ctx, key)
	}
	return nil
}
