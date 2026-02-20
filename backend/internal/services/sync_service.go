package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/MorantHP/OURERP/internal/models"
	"github.com/MorantHP/OURERP/internal/platform"
	"github.com/MorantHP/OURERP/internal/platform/clients"
	"github.com/MorantHP/OURERP/internal/repository"
)

// SyncResult 同步结果
type SyncResult struct {
	ShopID       int64
	ShopName     string
	Platform     string
	TotalSynced  int
	TotalFailed  int
	ErrorMessage string
	SyncedAt     time.Time
	Duration     time.Duration
}

// SyncService 同步服务
type SyncService struct {
	orderRepo *repository.OrderRepository
	shopRepo  *repository.ShopRepository
}

// NewSyncService 创建同步服务
func NewSyncService(orderRepo *repository.OrderRepository, shopRepo *repository.ShopRepository) *SyncService {
	return &SyncService{
		orderRepo: orderRepo,
		shopRepo:  shopRepo,
	}
}

// SyncShopOrders 同步店铺订单（根据平台自动选择客户端）
func (s *SyncService) SyncShopOrders(ctx context.Context, shopID int64, startTime, endTime time.Time) (*SyncResult, error) {
	shop, err := s.shopRepo.FindByID(shopID)
	if err != nil {
		return nil, fmt.Errorf("店铺不存在: %v", err)
	}

	start := time.Now()
	result := &SyncResult{
		ShopID:   shopID,
		ShopName: shop.Name,
		Platform: shop.Platform,
		SyncedAt: time.Now(),
	}

	// 检查Token是否过期
	if s.isTokenExpired(shop) {
		result.ErrorMessage = "Token已过期，请重新授权"
		return result, fmt.Errorf("token expired")
	}

	// 根据平台获取订单
	orders, err := s.fetchPlatformOrders(ctx, shop, startTime, endTime)
	if err != nil {
		result.ErrorMessage = err.Error()
		return result, err
	}

	// 保存订单
	for _, order := range orders {
		if err := s.orderRepo.Upsert(order); err != nil {
			log.Printf("保存订单失败 [%s]: %v", order.PlatformOrderID, err)
			result.TotalFailed++
			continue
		}
		result.TotalSynced++
	}

	result.Duration = time.Since(start)

	// 更新店铺最后同步时间
	now := time.Now()
	shop.LastSyncAt = &now
	if err := s.shopRepo.Update(shop); err != nil {
		log.Printf("更新店铺同步时间失败: %v", err)
	}

	return result, nil
}

// fetchPlatformOrders 根据平台获取订单
func (s *SyncService) fetchPlatformOrders(ctx context.Context, shop *models.Shop, startTime, endTime time.Time) ([]*models.Order, error) {
	switch shop.Platform {
	case string(platform.PlatformTaobao), string(platform.PlatformTmall):
		return s.fetchTaobaoOrders(ctx, shop, startTime, endTime)
	case string(platform.PlatformDouyin):
		return s.fetchDouyinOrders(ctx, shop, startTime, endTime)
	case string(platform.PlatformKuaishou):
		return s.fetchKuaishouOrders(ctx, shop, startTime, endTime)
	case string(platform.PlatformWechatVideo):
		return s.fetchWechatVideoOrders(ctx, shop, startTime, endTime)
	case string(platform.PlatformCustom):
		return s.fetchCustomPlatformOrders(ctx, shop, startTime, endTime)
	default:
		return nil, fmt.Errorf("不支持的平台: %s", shop.Platform)
	}
}

// fetchTaobaoOrders 获取淘宝/天猫订单
func (s *SyncService) fetchTaobaoOrders(ctx context.Context, shop *models.Shop, startTime, endTime time.Time) ([]*models.Order, error) {
	client := clients.NewTaobaoClient(shop.AppKey, shop.AppSecret, shop.AccessToken)

	var allOrders []*models.Order
	pageNo := 1

	for {
		platformOrders, err := client.FetchOrders(ctx, startTime, endTime, pageNo)
		if err != nil {
			return nil, err
		}

		for i := range platformOrders {
			order := s.convertTaobaoOrder(&platformOrders[i], shop.ID)
			allOrders = append(allOrders, order)
		}

		// 淘宝API没有返回total，这里简单地判断返回数量
		if len(platformOrders) < 100 {
			break
		}
		pageNo++
	}

	return allOrders, nil
}

// fetchDouyinOrders 获取抖音订单
func (s *SyncService) fetchDouyinOrders(ctx context.Context, shop *models.Shop, startTime, endTime time.Time) ([]*models.Order, error) {
	client := clients.NewDouyinClient(shop.AppKey, shop.AppSecret, shop.AccessToken, shop.PlatformShopID)

	var allOrders []*models.Order
	page := 1

	for {
		platformOrders, err := client.FetchOrders(ctx, startTime, endTime, page)
		if err != nil {
			return nil, err
		}

		for i := range platformOrders {
			order := s.convertDouyinOrder(&platformOrders[i], shop.ID)
			allOrders = append(allOrders, order)
		}

		if len(platformOrders) < 100 {
			break
		}
		page++
	}

	return allOrders, nil
}

// fetchKuaishouOrders 获取快手订单
func (s *SyncService) fetchKuaishouOrders(ctx context.Context, shop *models.Shop, startTime, endTime time.Time) ([]*models.Order, error) {
	client := clients.NewKuaishouClient(shop.AppKey, shop.AppSecret, shop.AccessToken, shop.PlatformShopID)

	var allOrders []*models.Order
	page := 1

	for {
		platformOrders, err := client.FetchOrders(ctx, startTime, endTime, page)
		if err != nil {
			return nil, err
		}

		for i := range platformOrders {
			order := s.convertKuaishouOrder(&platformOrders[i], shop.ID)
			allOrders = append(allOrders, order)
		}

		if len(platformOrders) < 100 {
			break
		}
		page++
	}

	return allOrders, nil
}

// fetchWechatVideoOrders 获取微信视频号订单
func (s *SyncService) fetchWechatVideoOrders(ctx context.Context, shop *models.Shop, startTime, endTime time.Time) ([]*models.Order, error) {
	client := clients.NewWechatVideoClient(shop.AppKey, shop.AppSecret, shop.AccessToken)

	var allOrders []*models.Order
	nextKey := ""

	for {
		orderIDs, newNextKey, err := client.FetchOrders(ctx, startTime, endTime, nextKey)
		if err != nil {
			return nil, err
		}

		// 获取订单详情
		for _, orderID := range orderIDs {
			po, err := client.GetOrderDetail(ctx, orderID)
			if err != nil {
				log.Printf("获取微信视频号订单详情失败 [%s]: %v", orderID, err)
				continue
			}
			order := s.convertWechatVideoOrder(po, shop.ID)
			allOrders = append(allOrders, order)
		}

		if newNextKey == "" || len(orderIDs) < 100 {
			break
		}
		nextKey = newNextKey
	}

	return allOrders, nil
}

// fetchCustomPlatformOrders 获取自定义平台订单
func (s *SyncService) fetchCustomPlatformOrders(ctx context.Context, shop *models.Shop, startTime, endTime time.Time) ([]*models.Order, error) {
	client := clients.NewCustomClient(shop.Platform, shop.APIURL, shop.AppKey, shop.AppSecret, "")

	var allOrders []*models.Order
	page := 1

	for {
		platformOrders, err := client.FetchOrders(ctx, startTime, endTime, page)
		if err != nil {
			return nil, err
		}

		for i := range platformOrders {
			order := s.convertCustomOrder(&platformOrders[i], shop.ID, shop.Platform)
			allOrders = append(allOrders, order)
		}

		if len(platformOrders) < 100 {
			break
		}
		page++
	}

	return allOrders, nil
}

// convertTaobaoOrder 转换淘宝订单
func (s *SyncService) convertTaobaoOrder(po *platform.PlatformOrder, shopID int64) *models.Order {
	return s.convertPlatformOrder(po, shopID, "taobao", s.mapTaobaoStatus, true)
}

// convertDouyinOrder 转换抖音订单
func (s *SyncService) convertDouyinOrder(po *platform.PlatformOrder, shopID int64) *models.Order {
	return s.convertPlatformOrder(po, shopID, "douyin", s.mapDouyinStatus, true)
}

// convertKuaishouOrder 转换快手订单
func (s *SyncService) convertKuaishouOrder(po *platform.PlatformOrder, shopID int64) *models.Order {
	return s.convertPlatformOrder(po, shopID, "kuaishou", s.mapKuaishouStatus, false)
}

// convertWechatVideoOrder 转换微信视频号订单
func (s *SyncService) convertWechatVideoOrder(po *platform.PlatformOrder, shopID int64) *models.Order {
	return s.convertPlatformOrder(po, shopID, "wechat_video", s.mapWechatVideoStatus, false)
}

// convertCustomOrder 转换自定义平台订单
func (s *SyncService) convertCustomOrder(po *platform.PlatformOrder, shopID int64, platformName string) *models.Order {
	return s.convertPlatformOrder(po, shopID, platformName, s.parseCustomStatusFunc, false)
}

// statusMapper 状态映射函数类型
type statusMapper func(string) int

// convertPlatformOrder 通用的平台订单转换函数
func (s *SyncService) convertPlatformOrder(po *platform.PlatformOrder, shopID int64, platformName string, mapStatus statusMapper, concatAddress bool) *models.Order {
	address := po.ReceiverAddress
	if concatAddress {
		address = fmt.Sprintf("%s%s%s%s", po.ReceiverProvince, po.ReceiverCity, po.ReceiverDistrict, po.ReceiverAddress)
	}

	order := &models.Order{
		Platform:        platformName,
		PlatformOrderID: po.PlatformOrderID,
		ShopID:          shopID,
		Status:          mapStatus(po.Status),
		TotalAmount:     po.TotalAmount,
		PayAmount:       po.PayAmount,
		BuyerNick:       po.BuyerNick,
		ReceiverName:    po.ReceiverName,
		ReceiverPhone:   po.ReceiverPhone,
		ReceiverAddress: address,
		Items:           make([]models.OrderItem, len(po.Items)),
	}

	if !po.CreatedAt.IsZero() {
		order.CreatedAt = po.CreatedAt
	}
	order.PaidAt = po.PaidAt
	order.ShippedAt = po.ShippedAt

	for i, item := range po.Items {
		order.Items[i] = models.OrderItem{
			SkuID:    0,
			SkuName:  item.SkuName,
			Quantity: item.Quantity,
			Price:    item.Price,
		}
	}

	return order
}

// parseCustomStatusFunc 用于自定义平台的状态映射
func (s *SyncService) parseCustomStatusFunc(status string) int {
	return s.parseCustomStatus(status)
}

// parseCustomStatus 解析自定义平台状态
func (s *SyncService) parseCustomStatus(status string) int {
	// 尝试解析为数字
	var statusInt int
	if n, err := fmt.Sscanf(status, "%d", &statusInt); n == 1 && err == nil {
		return statusInt
	}
	return models.OrderStatusPendingPayment
}

// isTokenExpired 检查Token是否过期
func (s *SyncService) isTokenExpired(shop *models.Shop) bool {
	if shop.TokenExpiresAt == nil {
		return shop.AccessToken == ""
	}
	// 提前5分钟判断过期
	return time.Now().Add(5 * time.Minute).After(*shop.TokenExpiresAt)
}

// 状态映射函数

func (s *SyncService) mapTaobaoStatus(status string) int {
	statusMap := map[string]int{
		"WAIT_BUYER_PAY":     models.OrderStatusPendingPayment,
		"WAIT_SELLER_SEND":   models.OrderStatusPendingShip,
		"SELLER_CONSIGNED":   models.OrderStatusShipped,
		"WAIT_BUYER_CONFIRM": models.OrderStatusShipped,
		"TRADE_FINISHED":     models.OrderStatusCompleted,
		"TRADE_CLOSED":       models.OrderStatusCancelled,
	}
	if v, ok := statusMap[status]; ok {
		return v
	}
	return models.OrderStatusPendingPayment
}

func (s *SyncService) mapDouyinStatus(status string) int {
	statusMap := map[string]int{
		"10": models.OrderStatusPendingPayment,
		"20": models.OrderStatusPendingShip,
		"30": models.OrderStatusShipped,
		"40": models.OrderStatusCompleted,
		"50": models.OrderStatusCancelled,
	}
	if v, ok := statusMap[status]; ok {
		return v
	}
	return models.OrderStatusPendingPayment
}

func (s *SyncService) mapKuaishouStatus(status string) int {
	statusMap := map[string]int{
		"10": models.OrderStatusPendingPayment,
		"20": models.OrderStatusPendingShip,
		"30": models.OrderStatusShipped,
		"40": models.OrderStatusCompleted,
		"50": models.OrderStatusCancelled,
	}
	if v, ok := statusMap[status]; ok {
		return v
	}
	return models.OrderStatusPendingPayment
}

func (s *SyncService) mapWechatVideoStatus(status string) int {
	statusMap := map[string]int{
		"10":   models.OrderStatusPendingPayment,
		"20":   models.OrderStatusPendingShip,
		"30":   models.OrderStatusShipped,
		"100":  models.OrderStatusCompleted,
		"200":  models.OrderStatusCancelled,
	}
	if v, ok := statusMap[status]; ok {
		return v
	}
	return models.OrderStatusPendingPayment
}
