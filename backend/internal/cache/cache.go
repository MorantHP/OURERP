package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheService 缓存服务接口
type CacheService interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	DeletePattern(ctx context.Context, pattern string) error
	Exists(ctx context.Context, key string) (bool, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)
	Incr(ctx context.Context, key string) (int64, error)
	Decr(ctx context.Context, key string) (int64, error)
	Close() error
}

// RedisCache Redis缓存实现
type RedisCache struct {
	client *redis.Client
	prefix string
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	Prefix   string
}

// NewRedisCache 创建Redis缓存
func NewRedisCache(cfg *CacheConfig) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("Redis连接成功")

	return &RedisCache{
		client: client,
		prefix: cfg.Prefix,
	}, nil
}

// NewMemoryCache 创建内存缓存（用于测试或无Redis环境）
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		data: make(map[string]*cacheItem),
	}
}

type cacheItem struct {
	value      []byte
	expiration time.Time
}

// MemoryCache 内存缓存实现
type MemoryCache struct {
	data map[string]*cacheItem
}

func (c *MemoryCache) Get(ctx context.Context, key string, dest interface{}) error {
	item, ok := c.data[key]
	if !ok {
		return fmt.Errorf("key not found")
	}
	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		delete(c.data, key)
		return fmt.Errorf("key expired")
	}
	return json.Unmarshal(item.value, dest)
}

func (c *MemoryCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	item := &cacheItem{value: data}
	if expiration > 0 {
		item.expiration = time.Now().Add(expiration)
	}
	c.data[key] = item
	return nil
}

func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	delete(c.data, key)
	return nil
}

func (c *MemoryCache) DeletePattern(ctx context.Context, pattern string) error {
	// 简化实现
	c.data = make(map[string]*cacheItem)
	return nil
}

func (c *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	_, ok := c.data[key]
	return ok, nil
}

func (c *MemoryCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	if item, ok := c.data[key]; ok {
		item.expiration = time.Now().Add(expiration)
	}
	return nil
}

func (c *MemoryCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	if item, ok := c.data[key]; ok {
		return time.Until(item.expiration), nil
	}
	return 0, fmt.Errorf("key not found")
}

func (c *MemoryCache) Incr(ctx context.Context, key string) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

func (c *MemoryCache) Decr(ctx context.Context, key string) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

func (c *MemoryCache) Close() error {
	return nil
}

// RedisCache 实现

func (c *RedisCache) buildKey(key string) string {
	if c.prefix == "" {
		return key
	}
	return c.prefix + ":" + key
}

func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	fullKey := c.buildKey(key)
	data, err := c.client.Get(ctx, fullKey).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	fullKey := c.buildKey(key)
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, fullKey, data, expiration).Err()
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	fullKey := c.buildKey(key)
	return c.client.Del(ctx, fullKey).Err()
}

func (c *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	fullPattern := c.buildKey(pattern)
	iter := c.client.Scan(ctx, 0, fullPattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := c.buildKey(key)
	n, err := c.client.Exists(ctx, fullKey).Result()
	return n > 0, err
}

func (c *RedisCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	fullKey := c.buildKey(key)
	return c.client.Expire(ctx, fullKey, expiration).Err()
}

func (c *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	fullKey := c.buildKey(key)
	return c.client.TTL(ctx, fullKey).Result()
}

func (c *RedisCache) Incr(ctx context.Context, key string) (int64, error) {
	fullKey := c.buildKey(key)
	return c.client.Incr(ctx, fullKey).Result()
}

func (c *RedisCache) Decr(ctx context.Context, key string) (int64, error) {
	fullKey := c.buildKey(key)
	return c.client.Decr(ctx, fullKey).Result()
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}

// 缓存键生成器

// CacheKey 缓存键
type CacheKey string

const (
	// 用户相关
	CacheKeyUser       CacheKey = "user"
	CacheKeyUserTenant CacheKey = "user:tenants"

	// 租户相关
	CacheKeyTenant      CacheKey = "tenant"
	CacheKeyTenantUsers CacheKey = "tenant:users"

	// 店铺相关
	CacheKeyShop    CacheKey = "shop"
	CacheKeyShops   CacheKey = "shops"

	// 商品相关
	CacheKeyProduct      CacheKey = "product"
	CacheKeyProducts     CacheKey = "products"
	CacheKeyProductCount CacheKey = "products:count"

	// 库存相关
	CacheKeyInventory      CacheKey = "inventory"
	CacheKeyInventoryAlert CacheKey = "inventory:alert"

	// 订单相关
	CacheKeyOrder      CacheKey = "order"
	CacheKeyOrders     CacheKey = "orders"
	CacheKeyOrderCount CacheKey = "orders:count"

	// 统计相关
	CacheKeyStatistics      CacheKey = "statistics"
	CacheKeyStatisticsTrend CacheKey = "statistics:trend"

	// 数据中心
	CacheKeyRealtimeOverview CacheKey = "realtime:overview"
	CacheKeyCustomerAnalysis CacheKey = "customer:analysis"
	CacheKeyProductAnalysis  CacheKey = "product:analysis"
	CacheKeyAlertRules       CacheKey = "alert:rules"
)

// BuildKey 构建缓存键
func BuildKey(base CacheKey, parts ...interface{}) string {
	key := string(base)
	for _, part := range parts {
		key = fmt.Sprintf("%s:%v", key, part)
	}
	return key
}

// 缓存时间常量
const (
	TTLShort  = 5 * time.Minute  // 短缓存
	TTLMedium = 30 * time.Minute // 中等缓存
	TTLLong   = 2 * time.Hour    // 长缓存
	TTLDay    = 24 * time.Hour   // 一天
)
