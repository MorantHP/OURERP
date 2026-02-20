package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

const (
	OrderStatusPendingPayment = iota + 1
	OrderStatusPendingShip
	OrderStatusShipped
	OrderStatusCompleted
	OrderStatusCancelled
)

type Order struct {
	ID               int64          `json:"id" gorm:"primaryKey"`
	TenantID         int64          `json:"tenant_id" gorm:"index;not null"` // 租户ID
	OrderNo          string         `json:"order_no" gorm:"uniqueIndex:idx_order_tenant,tenant_id;size:50;not null"`
	Platform         string         `json:"platform" gorm:"size:20;not null"`
	PlatformOrderID  string         `json:"platform_order_id" gorm:"size:100"`
	ShopID           int64          `json:"shop_id"`
	Status           int            `json:"status" gorm:"default:1"`
	TotalAmount      float64        `json:"total_amount"`
	PayAmount        float64        `json:"pay_amount"`
	BuyerNick        string         `json:"buyer_nick" gorm:"size:100"`
	ReceiverName     string         `json:"receiver_name" gorm:"size:50"`
	ReceiverPhone    string         `json:"receiver_phone" gorm:"size:20"`
	ReceiverAddress  string         `json:"receiver_address" gorm:"size:500"`
	LogisticsCompany string         `json:"logistics_company" gorm:"size:50"`
	LogisticsNo      string         `json:"logistics_no" gorm:"size:100"`
	Items            []OrderItem    `json:"items" gorm:"foreignKey:OrderID"`
	PaidAt           *time.Time     `json:"paid_at"`
	ShippedAt        *time.Time     `json:"shipped_at"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`
}

type OrderItem struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	TenantID  int64     `json:"tenant_id" gorm:"index;not null"`
	OrderID   int64     `json:"order_id" gorm:"index;not null"`
	SkuID     int64     `json:"sku_id"`
	SkuName   string    `json:"sku_name" gorm:"size:200"`
	Quantity  int       `json:"quantity"`
	Price     float64   `json:"price"`
	CreatedAt time.Time `json:"created_at"`
}

func GenerateOrderNo() string {
	return fmt.Sprintf("ORD%d", time.Now().UnixNano())
}

type CreateOrderRequest struct {
	Platform        string                     `json:"platform" binding:"required"`
	PlatformOrderID string                     `json:"platform_order_id"`
	ShopID          int64                      `json:"shop_id"`
	TotalAmount     float64                    `json:"total_amount" binding:"required"`
	PayAmount       float64                    `json:"pay_amount"`
	BuyerNick       string                     `json:"buyer_nick"`
	ReceiverName    string                     `json:"receiver_name" binding:"required"`
	ReceiverPhone   string                     `json:"receiver_phone"`
	ReceiverAddress string                     `json:"receiver_address"`
	Items           []CreateOrderItemRequest   `json:"items" binding:"required,min=1"`
}

type CreateOrderItemRequest struct {
	SkuID    int64   `json:"sku_id" binding:"required"`
	SkuName  string  `json:"sku_name" binding:"required"`
	Quantity int     `json:"quantity" binding:"required,min=1"`
	Price    float64 `json:"price" binding:"required"`
}
