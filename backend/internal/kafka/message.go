package kafka

import (
	"time"

	"github.com/google/uuid"
)

// 消息类型常量
const (
	MessageTypeOrderCreate  = "order_create"
	MessageTypeOrderUpdate  = "order_update"
	MessageTypeOrderCancel  = "order_cancel"
	MessageTypeOrderRefund  = "order_refund"
	MessageTypeOrderShip    = "order_ship"
	MessageTypeOrderConfirm = "order_confirm"
)

// 平台常量
const (
	PlatformTaobao   = "taobao"
	PlatformTmall    = "tmall"
	PlatformJD       = "jd"
	PlatformDouyin   = "douyin"
	PlatformKuaishou = "kuaishou"
)

// OrderMessage 统一订单消息格式
type OrderMessage struct {
	MessageID   string      `json:"message_id"`
	MessageType string      `json:"message_type"`
	Platform    string      `json:"platform"`
	Timestamp   time.Time   `json:"timestamp"`
	Version     string      `json:"version"`
	ShopID      int64       `json:"shop_id,omitempty"`
	Data        *OrderData  `json:"data"`
}

// OrderData 订单数据
type OrderData struct {
	PlatformOrderID string         `json:"platform_order_id"`
	ParentOrderID   string         `json:"parent_order_id,omitempty"`
	OrderStatus     string         `json:"order_status"`
	OrderStatusDesc string         `json:"order_status_desc,omitempty"`
	TotalAmount     float64        `json:"total_amount"`
	PayAmount       float64        `json:"pay_amount"`
	DiscountAmount  float64        `json:"discount_amount,omitempty"`
	PostFee         float64        `json:"post_fee,omitempty"`
	BuyerInfo       *BuyerInfo     `json:"buyer_info,omitempty"`
	ReceiverInfo    *ReceiverInfo  `json:"receiver_info,omitempty"`
	Items           []*OrderItem   `json:"items"`
	PaymentInfo     *PaymentInfo   `json:"payment_info,omitempty"`
	LogisticsInfo   *LogisticsInfo `json:"logistics_info,omitempty"`
	ExtendData      map[string]interface{} `json:"extend_data,omitempty"`
	CreatedAt       time.Time      `json:"created_at,omitempty"`
	UpdatedAt       time.Time      `json:"updated_at,omitempty"`
}

// BuyerInfo 买家信息
type BuyerInfo struct {
	BuyerID      string `json:"buyer_id"`
	BuyerNick    string `json:"buyer_nick"`
	BuyerPhone   string `json:"buyer_phone,omitempty"`
	BuyerEmail   string `json:"buyer_email,omitempty"`
	VipLevel     int    `json:"vip_level,omitempty"`
	IsNewBuyer   bool   `json:"is_new_buyer,omitempty"`
}

// ReceiverInfo 收货人信息
type ReceiverInfo struct {
	ReceiverName     string `json:"receiver_name"`
	ReceiverPhone    string `json:"receiver_phone"`
	ReceiverMobile   string `json:"receiver_mobile,omitempty"`
	ReceiverProvince string `json:"receiver_province"`
	ReceiverCity     string `json:"receiver_city"`
	ReceiverDistrict string `json:"receiver_district"`
	ReceiverAddress  string `json:"receiver_address"`
	ReceiverZip      string `json:"receiver_zip,omitempty"`
}

// OrderItem 订单商品
type OrderItem struct {
	SKUID          string  `json:"sku_id"`
	SKUName        string  `json:"sku_name"`
	SKUImage       string  `json:"sku_image,omitempty"`
	SKUSpec        string  `json:"sku_spec,omitempty"`
	Quantity       int     `json:"quantity"`
	Price          float64 `json:"price"`
	TotalAmount    float64 `json:"total_amount"`
	DiscountAmount float64 `json:"discount_amount,omitempty"`
	ProductID      string  `json:"product_id,omitempty"`
	ProductName    string  `json:"product_name,omitempty"`
}

// PaymentInfo 支付信息
type PaymentInfo struct {
	PayTime     time.Time `json:"pay_time,omitempty"`
	PayType     string    `json:"pay_type,omitempty"`
	PayTradeNo  string    `json:"pay_trade_no,omitempty"`
	PayStatus   string    `json:"pay_status,omitempty"`
}

// LogisticsInfo 物流信息
type LogisticsInfo struct {
	LogisticsCompany string    `json:"logistics_company,omitempty"`
	LogisticsNo      string    `json:"logistics_no,omitempty"`
	ShipTime         time.Time `json:"ship_time,omitempty"`
	ReceiverTime     time.Time `json:"receiver_time,omitempty"`
}

// NewOrderMessage 创建新的订单消息
func NewOrderMessage(messageType, platform string, data *OrderData) *OrderMessage {
	return &OrderMessage{
		MessageID:   uuid.New().String(),
		MessageType: messageType,
		Platform:    platform,
		Timestamp:   time.Now(),
		Version:     "1.0",
		Data:        data,
	}
}

// NewOrderCreateMessage 创建新订单消息
func NewOrderCreateMessage(platform string, data *OrderData) *OrderMessage {
	return NewOrderMessage(MessageTypeOrderCreate, platform, data)
}

// NewOrderUpdateMessage 创建订单更新消息
func NewOrderUpdateMessage(platform string, data *OrderData) *OrderMessage {
	return NewOrderMessage(MessageTypeOrderUpdate, platform, data)
}

// NewOrderCancelMessage 创建订单取消消息
func NewOrderCancelMessage(platform string, data *OrderData) *OrderMessage {
	return NewOrderMessage(MessageTypeOrderCancel, platform, data)
}
