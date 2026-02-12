// internal/models/order.go
package models

import (
	"fmt"
	"math/rand"
	"time"
)

// 订单状态
const (
	OrderStatusPendingPayment = 100 // 待付款
	OrderStatusPendingAudit   = 200 // 待审核
	OrderStatusPendingShip    = 300 // 待发货
	OrderStatusShipped        = 400 // 已发货
	OrderStatusReceived       = 500 // 已签收
	OrderStatusCompleted      = 600 // 已完成
	OrderStatusCancelled      = 999 // 已取消
)

type Order struct {
	ID              int64       `json:"id" gorm:"primaryKey"`
	OrderNo         string      `json:"order_no" gorm:"uniqueIndex;size:32;not null"`
	Platform        string      `json:"platform" gorm:"size:20;not null"` // taobao, jd, douyin, pdd
	PlatformOrderID string      `json:"platform_order_id" gorm:"size:64"`
	ShopID          int64       `json:"shop_id" gorm:"index"`
	Status          int         `json:"status" gorm:"default:100;index"`
	TotalAmount     float64     `json:"total_amount" gorm:"type:decimal(12,2)"`
	PayAmount       float64     `json:"pay_amount" gorm:"type:decimal(12,2)"`
	BuyerNick       string      `json:"buyer_nick" gorm:"size:100"`
	ReceiverName    string      `json:"receiver_name" gorm:"size:100"`
	ReceiverPhone   string      `json:"receiver_phone" gorm:"size:20"`
	ReceiverAddress string      `json:"receiver_address" gorm:"type:text"`
	Remark          string      `json:"remark" gorm:"type:text"`
	Items           []OrderItem `json:"items" gorm:"foreignKey:OrderID"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
	PaidAt          *time.Time  `json:"paid_at"`
	ShippedAt       *time.Time  `json:"shipped_at"`
}

type OrderItem struct {
	ID        int64   `json:"id" gorm:"primaryKey"`
	OrderID   int64   `json:"order_id" gorm:"index"`
	SkuID     int64   `json:"sku_id"`
	SkuName   string  `json:"sku_name" gorm:"size:200"`
	SkuImage  string  `json:"sku_image" gorm:"size:500"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price" gorm:"type:decimal(10,2)"`
	CreatedAt time.Time
}

// 生成订单号
func GenerateOrderNo() string {
	return fmt.Sprintf("ORD%s%d", time.Now().Format("20060102"), rand.Intn(9000)+1000)
}
