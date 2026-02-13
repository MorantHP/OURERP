package services

import (
	"fmt"
	"time"

	"github.com/MorantHP/OURERP/internal/models"
	"github.com/MorantHP/OURERP/internal/pkg/platform"
	"github.com/MorantHP/OURERP/internal/repository"
)

type SyncService struct {
	orderRepo     *repository.OrderRepository
	shopRepo      *repository.ShopRepository
}

func NewSyncService(orderRepo *repository.OrderRepository, shopRepo *repository.ShopRepository) *SyncService {
	return &SyncService{
		orderRepo:     orderRepo,
		shopRepo:      shopRepo,
	}
}

// SyncTaobaoOrders 同步淘宝订单
func (s *SyncService) SyncTaobaoOrders(shopID int64, startTime, endTime time.Time) (int, error) {
	shop, err := s.shopRepo.FindByID(shopID)
	if err != nil {
		return 0, fmt.Errorf("店铺不存在: %v", err)
	}

	client := platform.NewTaobaoClient(shop.AppKey, shop.AppSecret, shop.AccessToken)
	
	totalSynced := 0
	pageNo := 1
	
	for {
		orders, total, err := client.FetchOrders(startTime, endTime, pageNo)
		if err != nil {
			return totalSynced, err
		}

		for _, tbOrder := range orders {
			// 转换为内部订单模型
			order := s.convertTaobaoOrder(tbOrder, shopID)
			
			// 保存或更新订单
			if err := s.orderRepo.Upsert(order); err != nil {
				continue // 跳过失败的订单
			}
			totalSynced++
		}

		if pageNo*100 >= total {
			break
		}
		pageNo++
	}

	return totalSynced, nil
}

func (s *SyncService) convertTaobaoOrder(tbOrder platform.TaobaoOrder, shopID int64) *models.Order {
	// 状态映射
	statusMap := map[string]int{
		"WAIT_BUYER_PAY":    100, // 待付款
		"WAIT_SELLER_SEND":  200, // 待发货（已付款）
		"SELLER_CONSIGNED":  300, // 已发货
		"WAIT_BUYER_CONFIRM":400, // 已发货待确认
		"TRADE_FINISHED":    600, // 已完成
		"TRADE_CLOSED":      999, // 已关闭
	}
	
	status := statusMap[tbOrder.Status]
	if status == 0 {
		status = 100
	}

	// 解析金额
	payment := 0.0
	fmt.Sscanf(tbOrder.Payment, "%f", &payment)

	order := &models.Order{
		Platform:        "taobao",
		PlatformOrderID: tbOrder.Tid,
		ShopID:          shopID,
		Status:          status,
		PayAmount:       payment,
		BuyerNick:       tbOrder.BuyerNick,
		ReceiverName:    tbOrder.ReceiverName,
		ReceiverPhone:   tbOrder.ReceiverMobile,
		ReceiverAddress: fmt.Sprintf("%s%s%s%s", 
			tbOrder.ReceiverState, 
			tbOrder.ReceiverCity, 
			tbOrder.ReceiverDistrict, 
			tbOrder.ReceiverAddress),
		Items: make([]models.OrderItem, len(tbOrder.Orders.Order)),
	}

	// 解析时间
	if tbOrder.Created != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", tbOrder.Created); err == nil {
			order.CreatedAt = t
		}
	}

	// 转换商品明细
	for i, item := range tbOrder.Orders.Order {
		price := 0.0
		fmt.Sscanf(item.Price, "%f", &price)
		
		order.Items[i] = models.OrderItem{
			SkuID:    0, // 需要从映射表查询
			SkuName:  item.Title,
			Quantity: item.Num,
			Price:    price,
		}
	}

	return order
}