package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/MorantHP/OURERP/internal/cache"
	"github.com/MorantHP/OURERP/internal/models"
	"github.com/MorantHP/OURERP/internal/repository"
	"gorm.io/gorm"
)

// ApiSyncService API 方式的订单同步服务
// 用于通过 HTTP API 直接接收外部系统推送的订单数据
type ApiSyncService struct {
	db             *gorm.DB
	orderRepo      *repository.OrderRepository
	shopRepo       *repository.ShopRepository
	productRepo    *repository.ProductRepository
	cacheService   cache.CacheService
}

// NewApiSyncService 创建 API 同步服务
func NewApiSyncService(
	db *gorm.DB,
	orderRepo *repository.OrderRepository,
	shopRepo *repository.ShopRepository,
	productRepo *repository.ProductRepository,
	cacheService cache.CacheService,
) *ApiSyncService {
	return &ApiSyncService{
		db:           db,
		orderRepo:    orderRepo,
		shopRepo:     shopRepo,
		productRepo:  productRepo,
		cacheService: cacheService,
	}
}

// ExternalOrder 外部订单格式（兼容淘宝、京东、抖音等平台）
type ExternalOrder struct {
	// 平台信息
	Platform       string `json:"platform" binding:"required"`        // 平台: taobao, jd, douyin, pdd
	PlatformOrderID string `json:"platform_order_id" binding:"required"` // 平台订单号
	ShopID         int64  `json:"shop_id"`                             // 店铺ID（可选，会根据平台匹配）

	// 订单状态
	Status         string `json:"status"`          // 订单状态
	OrderTime      string `json:"order_time"`      // 下单时间
	PayTime        string `json:"pay_time"`        // 支付时间

	// 金额
	TotalAmount    float64 `json:"total_amount"`    // 订单总金额
	PayAmount      float64 `json:"pay_amount"`      // 实付金额

	// 买家信息
	BuyerNick      string `json:"buyer_nick"`       // 买家昵称
	BuyerMessage   string `json:"buyer_message"`    // 买家留言

	// 收货信息
	ReceiverName   string `json:"receiver_name"`
	ReceiverPhone  string `json:"receiver_phone"`
	ReceiverAddress string `json:"receiver_address"`
	ReceiverState  string `json:"receiver_state"`  // 省
	ReceiverCity   string `json:"receiver_city"`    // 市
	ReceiverDistrict string `json:"receiver_district"` // 区
	ReceiverStreet string `json:"receiver_street"`  // 街道

	// 物流信息
	LogisticsCompany string `json:"logistics_company"` // 物流公司
	LogisticsNo      string `json:"logistics_no"`       // 物流单号

	// 订单明细
	Items []ExternalOrderItem `json:"items" binding:"required"`

	// 租户信息（用于多租户场景）
	TenantCode string `json:"tenant_code"` // 租户编码（可选，用于指定租户）
}

// ExternalOrderItem 外部订单明细
type ExternalOrderItem struct {
	SkuID          string  `json:"sku_id"`           // 商品SKU编码（外部系统）
	SkuName        string  `json:"sku_name"`         // 商品名称
	Quantity       int     `json:"quantity"`          // 数量
	Price          float64 `json:"price"`             // 单价
	TotalAmount    float64 `json:"total_amount"`     // 小计
	OuterSkuID     string  `json:"outer_sku_id"`     // 外部SKU ID
	OuterSkuName   string  `json:"outer_sku_name"`   // 外部SKU名称
	ProductName    string  `json:"product_name"`     // 商品名称（用于匹配）
	ProductBarcode string  `json:"product_barcode"`  // 商品条码（用于匹配）
}

// SyncOrderRequest 批量同步订单请求
type SyncOrderRequest struct {
	Orders []ExternalOrder `json:"orders" binding:"required"`
	Source string          `json:"source"` // 数据来源标识
}

// SyncOrderResponse 同步响应
type SyncOrderResponse struct {
	Success      bool     `json:"success"`
	Total        int      `json:"total"`
	Created      int      `json:"created"`
	Updated      int      `json:"updated"`
	Failed       int      `json:"failed"`
	Errors       []string `json:"errors,omitempty"`
	ProcessTime  int64    `json:"process_time"` // 处理耗时（毫秒）
}

// SyncOrders 批量同步订单
func (s *ApiSyncService) SyncOrders(ctx context.Context, tenantID int64, req *SyncOrderRequest) (*SyncOrderResponse, error) {
	startTime := time.Now()

	resp := &SyncOrderResponse{
		Total:   len(req.Orders),
		Success: true,
		Errors:  make([]string, 0),
	}

	for _, order := range req.Orders {
		// 处理单个订单
		orderID, err := s.syncSingleOrder(ctx, tenantID, &order)
		if err != nil {
			resp.Failed++
			resp.Errors = append(resp.Errors, fmt.Sprintf("订单 %s 同步失败: %v", order.PlatformOrderID, err))
		} else if orderID > 0 {
			resp.Created++
		} else {
			resp.Updated++
		}
	}

	// 如果有失败，标记为部分成功
	if resp.Failed > 0 {
		resp.Success = false
	}

	resp.ProcessTime = time.Since(startTime).Milliseconds()

	return resp, nil
}

// syncSingleOrder 同步单个订单
// 返回值: >0 表示新建订单ID, 0 表示更新订单, <0 表示失败
func (s *ApiSyncService) syncSingleOrder(ctx context.Context, tenantID int64, order *ExternalOrder) (int64, error) {
	// 1. 查找或创建店铺
	shopID, err := s.getOrCreateShop(ctx, tenantID, order)
	if err != nil {
		return -1, fmt.Errorf("获取店铺失败: %w", err)
	}

	// 2. 检查订单是否已存在
	var existingOrder models.Order
	err = s.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ? AND platform_order_id = ?", tenantID, order.Platform, order.PlatformOrderID).
		First(&existingOrder).Error

	if err == nil {
		// 订单已存在，更新订单信息
		return 0, s.updateExistingOrder(ctx, &existingOrder, order)
	} else if err != gorm.ErrRecordNotFound {
		return -1, fmt.Errorf("查询订单失败: %w", err)
	}

	// 3. 创建新订单
	newOrder := s.convertToOrder(tenantID, shopID, order)
	if err := s.db.WithContext(ctx).Create(newOrder).Error; err != nil {
		return -1, fmt.Errorf("创建订单失败: %w", err)
	}

	// 4. 使缓存失效（如果缓存服务可用）
	if s.cacheService != nil {
		_ = s.cacheService.Delete(ctx, cache.BuildKey(cache.CacheKeyOrders, tenantID))
	}

	return newOrder.ID, nil
}

// getOrCreateShop 获取或创建店铺
func (s *ApiSyncService) getOrCreateShop(ctx context.Context, tenantID int64, order *ExternalOrder) (int64, error) {
	// 如果指定了 shopID，直接返回
	if order.ShopID > 0 {
		return order.ShopID, nil
	}

	// 尝试通过平台查找默认店铺
	var shop models.Shop
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ?", tenantID, order.Platform).
		First(&shop).Error

	if err == nil {
		return shop.ID, nil
	}

	// 未找到店铺，创建一个默认店铺
	shop = models.Shop{
		TenantID:       tenantID,
		Name:           s.getPlatformDisplayName(order.Platform),
		Platform:       order.Platform,
		PlatformShopID: fmt.Sprintf("auto_%s", order.Platform),
		Status:         1,
		SyncInterval:   30,
	}

	if err := s.db.WithContext(ctx).Create(&shop).Error; err != nil {
		return 0, err
	}

	return shop.ID, nil
}

// getPlatformDisplayName 获取平台显示名称
func (s *ApiSyncService) getPlatformDisplayName(platform string) string {
	names := map[string]string{
		"taobao":   "淘宝店铺",
		"tmall":    "天猫店铺",
		"jd":       "京东店铺",
		"douyin":   "抖音店铺",
		"kuaishou": "快手店铺",
		"pdd":      "拼多多店铺",
	}
	if name, ok := names[platform]; ok {
		return name
	}
	return platform + "店铺"
}

// convertToOrder 转换为订单模型
func (s *ApiSyncService) convertToOrder(tenantID, shopID int64, order *ExternalOrder) *models.Order {
	now := time.Now()

	// 解析支付时间
	var payTimePtr *time.Time
	if order.PayTime != "" {
		if payTime, err := time.Parse(time.RFC3339, order.PayTime); err == nil {
			payTimePtr = &payTime
		}
	}

	// 组装收货地址
	receiverAddress := order.ReceiverAddress
	if receiverAddress == "" {
		receiverAddress = fmt.Sprintf("%s %s %s %s",
			order.ReceiverState, order.ReceiverCity, order.ReceiverDistrict, order.ReceiverStreet)
	}

	// 转换订单状态
	status := s.convertOrderStatus(order.Status)

	// 转换订单明细
	items := make([]models.OrderItem, len(order.Items))
	for i, item := range order.Items {
		// 尝试将 SKU ID 转换为 int64，如果失败则使用 0
		var skuID int64
		if item.SkuID != "" {
			skuID, _ = strconv.ParseInt(item.SkuID, 10, 64)
		}

		items[i] = models.OrderItem{
			SkuID:   skuID,
			SkuName: item.SkuName,
			Quantity: item.Quantity,
			Price:   item.Price,
		}
	}

	return &models.Order{
		TenantID:         tenantID,
		OrderNo:          models.GenerateOrderNo(),
		Platform:         order.Platform,
		PlatformOrderID:  order.PlatformOrderID,
		ShopID:           shopID,
		Status:           status,
		TotalAmount:      order.TotalAmount,
		PayAmount:        order.PayAmount,
		BuyerNick:        order.BuyerNick,
		ReceiverName:     order.ReceiverName,
		ReceiverPhone:    order.ReceiverPhone,
		ReceiverAddress:  receiverAddress,
		LogisticsCompany: order.LogisticsCompany,
		LogisticsNo:      order.LogisticsNo,
		Items:            items,
		PaidAt:           payTimePtr,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// convertOrderStatus 转换订单状态
func (s *ApiSyncService) convertOrderStatus(status string) int {
	statusMap := map[string]int{
		"pending_payment": models.OrderStatusPendingPayment,
		"paid":            models.OrderStatusPendingShip,
		"pending_ship":    models.OrderStatusPendingShip,
		"shipped":         models.OrderStatusShipped,
		"completed":       models.OrderStatusCompleted,
		"cancelled":       models.OrderStatusCancelled,
	}

	if s, ok := statusMap[status]; ok {
		return s
	}
	return models.OrderStatusPendingPayment
}

// updateExistingOrder 更新已存在的订单
func (s *ApiSyncService) updateExistingOrder(ctx context.Context, existing *models.Order, order *ExternalOrder) error {
	updates := make(map[string]interface{})

	// 只在订单状态为初始状态时才更新
	if existing.Status == models.OrderStatusPendingPayment {
		if order.Status != "" {
			updates["status"] = s.convertOrderStatus(order.Status)
		}
		if order.LogisticsCompany != "" {
			updates["logistics_company"] = order.LogisticsCompany
		}
		if order.LogisticsNo != "" {
			updates["logistics_no"] = order.LogisticsNo
		}
		if order.PayTime != "" {
			if payTime, err := time.Parse(time.RFC3339, order.PayTime); err == nil {
				updates["paid_at"] = &payTime
			}
		}
	}

	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		return s.db.WithContext(ctx).Model(existing).Updates(updates).Error
	}

	return nil
}

// GetSyncStatistics 获取同步统计信息
func (s *ApiSyncService) GetSyncStatistics(ctx context.Context, tenantID int64, startTime, endTime time.Time) (*SyncStatistics, error) {
	var stats SyncStatistics

	// 统计各平台订单数量
	type PlatformCount struct {
		Platform string
		Count    int64
	}
	var platformCounts []PlatformCount

	err := s.db.WithContext(ctx).
		Table("orders").
		Select("platform, count(*) as count").
		Where("tenant_id = ? AND created_at >= ? AND created_at <= ?", tenantID, startTime, endTime).
		Group("platform").
		Scan(&platformCounts).Error

	if err != nil {
		return nil, err
	}

	stats.ByPlatform = make(map[string]int64)
	for _, pc := range platformCounts {
		stats.ByPlatform[pc.Platform] = pc.Count
	}

	// 统计总订单数和总金额
	type TotalStats struct {
		Count  int64
		Amount float64
	}
	var totalStats TotalStats

	err = s.db.WithContext(ctx).
		Table("orders").
		Select("count(*) as count, sum(pay_amount) as amount").
		Where("tenant_id = ? AND created_at >= ? AND created_at <= ?", tenantID, startTime, endTime).
		Scan(&totalStats).Error

	if err != nil {
		return nil, err
	}

	stats.TotalCount = totalStats.Count
	stats.TotalAmount = totalStats.Amount

	return &stats, nil
}

// SyncStatistics 同步统计信息
type SyncStatistics struct {
	TotalCount   int64            `json:"total_count"`
	TotalAmount  float64          `json:"total_amount"`
	ByPlatform   map[string]int64 `json:"by_platform"`
}
