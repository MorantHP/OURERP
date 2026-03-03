package services

import (
	"context"

	"github.com/MorantHP/OURERP/backend/internal/cache"
	"github.com/MorantHP/OURERP/backend/internal/models"
	"github.com/MorantHP/OURERP/backend/internal/pkg/errors"
	"github.com/MorantHP/OURERP/backend/internal/repository"
)

// OrderService 订单服务
type OrderService struct {
	orderRepo     *repository.OrderRepository
	inventoryRepo *repository.InventoryRepository
	cacheDecorator *CacheDecorator
}

// NewOrderService 创建订单服务
func NewOrderService(
	orderRepo *repository.OrderRepository,
	inventoryRepo *repository.InventoryRepository,
	cacheService cache.CacheService,
) *OrderService {
	return &OrderService{
		orderRepo:      orderRepo,
		inventoryRepo:  inventoryRepo,
		cacheDecorator: NewCacheDecorator(cacheService, "order"),
	}
}

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(ctx context.Context, req *models.CreateOrderRequest) (*models.Order, error) {
	tenantID := repository.GetTenantIDFromContext(ctx)

	order := &models.Order{
		OrderNo:         models.GenerateOrderNo(),
		Platform:        req.Platform,
		PlatformOrderID: req.PlatformOrderID,
		ShopID:          req.ShopID,
		Status:          models.OrderStatusPendingPayment,
		TotalAmount:     req.TotalAmount,
		PayAmount:       req.PayAmount,
		BuyerNick:       req.BuyerNick,
		ReceiverName:    req.ReceiverName,
		ReceiverPhone:   req.ReceiverPhone,
		ReceiverAddress: req.ReceiverAddress,
		Items:           make([]models.OrderItem, len(req.Items)),
	}

	for i, item := range req.Items {
		order.Items[i] = models.OrderItem{
			SkuID:    item.SkuID,
			SkuName:  item.SkuName,
			Quantity: item.Quantity,
			Price:    item.Price,
		}
	}

	if err := s.orderRepo.CreateWithContext(ctx, order); err != nil {
		return nil, errors.WrapInternal(err, "创建订单失败")
	}

	// 使缓存失效
	_ = s.cacheDecorator.InvalidateOrderCache(ctx, tenantID)

	return order, nil
}

// GetOrder 获取订单详情
func (s *OrderService) GetOrder(ctx context.Context, orderNo string) (*models.Order, error) {
	order, err := s.orderRepo.FindByOrderNoWithContext(ctx, orderNo)
	if err != nil {
		return nil, errors.ErrOrderNotFound
	}
	return order, nil
}

// ListOrders 查询订单列表
func (s *OrderService) ListOrders(ctx context.Context, page, size int, status, platform, keyword string) ([]models.Order, int64, error) {
	// 限制最大分页大小
	if size > 100 {
		size = 100
	}
	return s.orderRepo.ListWithContext(ctx, page, size, status, platform, keyword)
}

// AuditOrder 审核订单
func (s *OrderService) AuditOrder(ctx context.Context, orderNo string) error {
	if err := s.orderRepo.UpdateStatusWithContext(ctx, orderNo, models.OrderStatusPendingShip); err != nil {
		return errors.WrapInternal(err, "审核订单失败")
	}

	tenantID := repository.GetTenantIDFromContext(ctx)
	_ = s.cacheDecorator.InvalidateOrderCache(ctx, tenantID)

	return nil
}

// ShipOrder 发货
func (s *OrderService) ShipOrder(ctx context.Context, orderNo string, logisticsCompany, logisticsNo string) error {
	// 获取订单
	order, err := s.orderRepo.FindByOrderNoWithContext(ctx, orderNo)
	if err != nil {
		return errors.ErrOrderNotFound
	}

	// 检查订单状态
	if order.Status != models.OrderStatusPendingShip {
		return errors.NewAppError("INVALID_ORDER_STATUS", "订单状态不允许发货", 400, nil)
	}

	// 更新订单状态
	if err := s.orderRepo.ShipWithContext(ctx, orderNo, logisticsCompany, logisticsNo); err != nil {
		return errors.WrapInternal(err, "发货失败")
	}

	// 扣减库存（默认从第一个仓库扣减，实际应该根据业务逻辑确定）
	tenantID := repository.GetTenantIDFromContext(ctx)
	for _, item := range order.Items {
		// 这里简化处理，实际应该根据商品找到对应仓库
		// 扣减库存应该在 InventoryService 中处理
	}

	// 使缓存失效
	_ = s.cacheDecorator.InvalidateOrderCache(ctx, tenantID)

	return nil
}

// CancelOrder 取消订单
func (s *OrderService) CancelOrder(ctx context.Context, orderNo string, reason string) error {
	order, err := s.orderRepo.FindByOrderNoWithContext(ctx, orderNo)
	if err != nil {
		return errors.ErrOrderNotFound
	}

	// 检查是否可以取消
	if order.Status == models.OrderStatusShipped || order.Status == models.OrderStatusCompleted {
		return errors.NewAppError("INVALID_ORDER_STATUS", "订单已发货或已完成，无法取消", 400, nil)
	}

	// 更新订单状态
	if err := s.orderRepo.UpdateStatusWithContext(ctx, orderNo, models.OrderStatusCanceled); err != nil {
		return errors.WrapInternal(err, "取消订单失败")
	}

	tenantID := repository.GetTenantIDFromContext(ctx)
	_ = s.cacheDecorator.InvalidateOrderCache(ctx, tenantID)

	return nil
}

// GetOrderStatistics 获取订单统计
func (s *OrderService) GetOrderStatistics(ctx context.Context) (*OrderStatistics, error) {
	tenantID := repository.GetTenantIDFromContext(ctx)

	var stats OrderStatistics
	err := s.cacheDecorator.GetOrSet(ctx, cache.BuildKey(cache.CacheKeyStatistics, tenantID), &stats, cache.TTLShort, func() (interface{}, error) {
		// 这里应该调用 Repository 获取统计数据
		// 简化处理
		return &OrderStatistics{}, nil
	})

	if err != nil {
		return nil, err
	}

	return &stats, nil
}

// OrderStatistics 订单统计
type OrderStatistics struct {
	PendingPayment int64 `json:"pending_payment"`
	PendingShip    int64 `json:"pending_ship"`
	Shipped        int64 `json:"shipped"`
	Completed      int64 `json:"completed"`
	Canceled       int64 `json:"canceled"`
	TotalAmount    float64 `json:"total_amount"`
}
