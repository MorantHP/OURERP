package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/MorantHP/OURERP/internal/models"
	"github.com/MorantHP/OURERP/internal/repository"
)

// ERPOrderHandler ERP订单处理器
type ERPOrderHandler struct {
	orderRepo   *repository.OrderRepository
	productRepo *repository.ProductRepository
	shopRepo    *repository.ShopRepository
}

// NewERPOrderHandler 创建ERP订单处理器
func NewERPOrderHandler(
	orderRepo *repository.OrderRepository,
	productRepo *repository.ProductRepository,
	shopRepo *repository.ShopRepository,
) *ERPOrderHandler {
	return &ERPOrderHandler{
		orderRepo:   orderRepo,
		productRepo: productRepo,
		shopRepo:    shopRepo,
	}
}

// HandleOrderCreate 处理新订单
func (h *ERPOrderHandler) HandleOrderCreate(ctx context.Context, msg *OrderMessage) error {
	log.Printf("[ERPOrderHandler] 处理新订单 - 平台: %s, 订单号: %s",
		msg.Platform, msg.Data.PlatformOrderID)

	// 转换为ERP订单模型
	order := h.convertToERPOrder(msg)

	// 保存到数据库（Kafka订单自带租户ID，不需要从上下文获取）
	if err := h.orderRepo.Create(order); err != nil {
		log.Printf("[ERPOrderHandler] 保存订单失败: %v", err)
		return fmt.Errorf("保存订单失败: %w", err)
	}
	log.Printf("[ERPOrderHandler] 订单已入库 - 订单号: %s, 平台: %s", order.OrderNo, order.Platform)

	// 更新库存（如果有SKU对应关系）
	h.updateInventory(ctx, msg)

	// 发送WebSocket通知
	h.notifyNewOrder(msg)

	log.Printf("[ERPOrderHandler] 订单已保存 - ID: %d, 订单号: %s",
		order.ID, order.OrderNo)

	return nil
}

// HandleOrderUpdate 处理订单更新
func (h *ERPOrderHandler) HandleOrderUpdate(ctx context.Context, msg *OrderMessage) error {
	log.Printf("[ERPOrderHandler] 处理订单更新 - 平台: %s, 订单号: %s",
		msg.Platform, msg.Data.PlatformOrderID)

	// 查找现有订单
	existingOrder, err := h.orderRepo.FindByPlatformOrderIDWithContext(
		ctx, msg.Data.PlatformOrderID, msg.Platform)
	if err != nil {
		return fmt.Errorf("订单不存在: %w", err)
	}

	// 更新订单状态
	if err := h.orderRepo.UpdateStatusWithContext(
		ctx, existingOrder.OrderNo, h.mapOrderStatus(msg.Data.OrderStatus)); err != nil {
		return fmt.Errorf("更新订单状态失败: %w", err)
	}

	log.Printf("[ERPOrderHandler] 订单已更新 - 订单号: %s, 新状态: %s",
		existingOrder.OrderNo, msg.Data.OrderStatus)

	return nil
}

// HandleOrderCancel 处理订单取消
func (h *ERPOrderHandler) HandleOrderCancel(ctx context.Context, msg *OrderMessage) error {
	log.Printf("[ERPOrderHandler] 处理订单取消 - 平台: %s, 订单号: %s",
		msg.Platform, msg.Data.PlatformOrderID)

	// 查找现有订单
	existingOrder, err := h.orderRepo.FindByPlatformOrderIDWithContext(
		ctx, msg.Data.PlatformOrderID, msg.Platform)
	if err != nil {
		return fmt.Errorf("订单不存在: %w", err)
	}

	// 更新为已取消状态
	if err := h.orderRepo.UpdateStatusWithContext(
		ctx, existingOrder.OrderNo, models.OrderStatusCancelled); err != nil {
		return fmt.Errorf("取消订单失败: %w", err)
	}

	// 恢复库存
	h.restoreInventory(ctx, msg)

	log.Printf("[ERPOrderHandler] 订单已取消 - 订单号: %s", existingOrder.OrderNo)

	return nil
}

// HandleOrderRefund 处理订单退款
func (h *ERPOrderHandler) HandleOrderRefund(ctx context.Context, msg *OrderMessage) error {
	log.Printf("[ERPOrderHandler] 处理订单退款 - 平台: %s, 订单号: %s",
		msg.Platform, msg.Data.PlatformOrderID)

	// 查找现有订单
	existingOrder, err := h.orderRepo.FindByPlatformOrderIDWithContext(
		ctx, msg.Data.PlatformOrderID, msg.Platform)
	if err != nil {
		return fmt.Errorf("订单不存在: %w", err)
	}

	// 更新为退款状态（使用已取消状态）
	if err := h.orderRepo.UpdateStatusWithContext(
		ctx, existingOrder.OrderNo, models.OrderStatusCancelled); err != nil {
		return fmt.Errorf("更新退款状态失败: %w", err)
	}

	log.Printf("[ERPOrderHandler] 订单退款处理完成 - 订单号: %s", existingOrder.OrderNo)

	return nil
}

// convertToERPOrder 转换为ERP订单模型
func (h *ERPOrderHandler) convertToERPOrder(msg *OrderMessage) *models.Order {
	order := &models.Order{
		TenantID:        msg.ShopID, // 使用店铺ID作为租户ID（简化）
		OrderNo:         models.GenerateOrderNo(),
		Platform:        msg.Platform,
		PlatformOrderID: msg.Data.PlatformOrderID,
		ShopID:          msg.ShopID,
		Status:          h.mapOrderStatus(msg.Data.OrderStatus),
		TotalAmount:     msg.Data.TotalAmount,
		PayAmount:       msg.Data.PayAmount,
	}

	// 买家信息
	if msg.Data.BuyerInfo != nil {
		order.BuyerNick = msg.Data.BuyerInfo.BuyerNick
	}

	// 收货人信息
	if msg.Data.ReceiverInfo != nil {
		order.ReceiverName = msg.Data.ReceiverInfo.ReceiverName
		order.ReceiverPhone = msg.Data.ReceiverInfo.ReceiverPhone
		order.ReceiverAddress = fmt.Sprintf("%s%s%s%s",
			msg.Data.ReceiverInfo.ReceiverProvince,
			msg.Data.ReceiverInfo.ReceiverCity,
			msg.Data.ReceiverInfo.ReceiverDistrict,
			msg.Data.ReceiverInfo.ReceiverAddress,
		)
	}

	// 订单商品
	if len(msg.Data.Items) > 0 {
		order.Items = make([]models.OrderItem, len(msg.Data.Items))
		for i, item := range msg.Data.Items {
			order.Items[i] = models.OrderItem{
				TenantID: msg.ShopID,
				SkuID:    0, // 需要根据SKU ID查询
				SkuName:  item.SKUName,
				Quantity: item.Quantity,
				Price:    item.Price,
			}
		}
	}

	// 支付信息
	if msg.Data.PaymentInfo != nil {
		order.PaidAt = &msg.Data.PaymentInfo.PayTime
	}

	// 物流信息
	if msg.Data.LogisticsInfo != nil {
		order.LogisticsCompany = msg.Data.LogisticsInfo.LogisticsCompany
		order.LogisticsNo = msg.Data.LogisticsInfo.LogisticsNo
		if !msg.Data.LogisticsInfo.ShipTime.IsZero() {
			order.ShippedAt = &msg.Data.LogisticsInfo.ShipTime
		}
	}

	// 时间信息
	if !msg.Data.CreatedAt.IsZero() {
		order.CreatedAt = msg.Data.CreatedAt
	}
	if !msg.Data.UpdatedAt.IsZero() {
		order.UpdatedAt = msg.Data.UpdatedAt
	}

	return order
}

// mapOrderStatus 映射订单状态
func (h *ERPOrderHandler) mapOrderStatus(platformStatus string) int {
	// 通用状态映射
	statusMap := map[string]int{
		// 淘宝/天猫
		"WAIT_BUYER_PAY":     models.OrderStatusPendingPayment,
		"WAIT_SELLER_SEND":   models.OrderStatusPendingShip,
		"WAIT_BUYER_CONFIRM": models.OrderStatusShipped,
		"SELLER_CONSIGNED":   models.OrderStatusShipped,
		"TRADE_FINISHED":     models.OrderStatusCompleted,
		"TRADE_CLOSED":       models.OrderStatusCancelled,

		// 京东
		"WAIT_PAY":    models.OrderStatusPendingPayment,
		"WAIT_DELIVER": models.OrderStatusPendingShip,
		"DELIVERED":   models.OrderStatusShipped,
		"FINISHED":    models.OrderStatusCompleted,
		"CANCEL":      models.OrderStatusCancelled,

		// 抖音/快手 (数字状态)
		"10": models.OrderStatusPendingPayment,
		"20": models.OrderStatusPendingShip,
		"30": models.OrderStatusShipped,
		"40": models.OrderStatusCompleted,
		"50": models.OrderStatusCancelled,
	}

	if status, ok := statusMap[platformStatus]; ok {
		return status
	}
	return models.OrderStatusPendingPayment
}

// updateInventory 更新库存
func (h *ERPOrderHandler) updateInventory(ctx context.Context, msg *OrderMessage) {
	// TODO: 实现库存扣减逻辑
	// 1. 根据SKU ID查找对应的商品
	// 2. 扣减库存数量
}

// restoreInventory 恢复库存
func (h *ERPOrderHandler) restoreInventory(ctx context.Context, msg *OrderMessage) {
	// TODO: 实现库存恢复逻辑
}

// notifyNewOrder 通知新订单
func (h *ERPOrderHandler) notifyNewOrder(msg *OrderMessage) {
	// 发送WebSocket通知
	// TODO: 集成WebSocket服务
	notification := map[string]interface{}{
		"type":          "new_order",
		"platform":      msg.Platform,
		"order_id":      msg.Data.PlatformOrderID,
		"total_amount":  msg.Data.TotalAmount,
		"buyer_nick":    msg.Data.BuyerInfo.BuyerNick,
		"item_count":    len(msg.Data.Items),
		"created_at":    time.Now(),
	}

	// 记录日志
	notificationJSON, _ := json.Marshal(notification)
	log.Printf("[ERPOrderHandler] 新订单通知: %s", string(notificationJSON))
}
