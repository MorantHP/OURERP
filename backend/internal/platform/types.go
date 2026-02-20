package platform

import "time"

// PlatformType 平台类型
type PlatformType string

const (
	// 优先实现平台
	PlatformTaobao      PlatformType = "taobao"       // 淘宝
	PlatformTmall       PlatformType = "tmall"        // 天猫(与淘宝共用API)
	PlatformDouyin      PlatformType = "douyin"       // 抖音电商
	PlatformKuaishou    PlatformType = "kuaishou"     // 快手电商
	PlatformWechatVideo PlatformType = "wechat_video" // 微信视频号
	PlatformTikTok      PlatformType = "tiktok"       // TikTok小店
	PlatformJingqi      PlatformType = "jingqi"       // 京企直卖

	// 后续扩展平台
	PlatformJD          PlatformType = "jd"           // 京东
	PlatformXiaohongshu PlatformType = "xiaohongshu"  // 小红书
	PlatformVip         PlatformType = "vip"          // 唯品会
	Platform1688        PlatformType = "1688"         // 1688
	PlatformWechat      PlatformType = "wechat"       // 微信小商店

	// 通用
	PlatformOffline     PlatformType = "offline"      // 实体店铺
	PlatformCustom      PlatformType = "custom"       // 自定义第三方
)

// PlatformConfig 平台配置
type PlatformConfig struct {
	Code        PlatformType `json:"code"`
	Name        string       `json:"name"`
	Icon        string       `json:"icon"`
	Description string       `json:"description"`
	Features    []string     `json:"features"`  // 支持的功能
	AuthType    string       `json:"auth_type"` // oauth, apikey, manual
}

// PlatformAdapter 平台适配器接口
type PlatformAdapter interface {
	// 基础信息
	GetConfig() PlatformConfig

	// 订单相关
	FetchOrders(startTime, endTime time.Time, page int) ([]PlatformOrder, error)
	GetOrderDetail(orderID string) (*PlatformOrder, error)
	ShipOrder(orderID string, logistics LogisticsInfo) error

	// 商品相关
	FetchProducts(page int) ([]PlatformProduct, error)
	UpdateStock(skuID string, quantity int) error

	// 授权相关
	AuthURL(redirectURI string) string
	ExchangeToken(code string) (*AuthToken, error)
	RefreshToken(refreshToken string) (*AuthToken, error)
}

// PlatformOrder 平台订单统一格式
type PlatformOrder struct {
	PlatformOrderID  string              `json:"platform_order_id"`
	Platform         PlatformType        `json:"platform"`
	Status           string              `json:"status"`
	TotalAmount      float64             `json:"total_amount"`
	PayAmount        float64             `json:"pay_amount"`
	BuyerNick        string              `json:"buyer_nick"`
	BuyerPhone       string              `json:"buyer_phone"`
	ReceiverName     string              `json:"receiver_name"`
	ReceiverPhone    string              `json:"receiver_phone"`
	ReceiverProvince string              `json:"receiver_province"`
	ReceiverCity     string              `json:"receiver_city"`
	ReceiverDistrict string              `json:"receiver_district"`
	ReceiverAddress  string              `json:"receiver_address"`
	Remark           string              `json:"remark"`
	CreatedAt        time.Time           `json:"created_at"`
	PaidAt           *time.Time          `json:"paid_at"`
	ShippedAt        *time.Time          `json:"shipped_at"`
	Items            []PlatformOrderItem `json:"items"`
}

type PlatformOrderItem struct {
	SkuID      string  `json:"sku_id"`
	SkuName    string  `json:"sku_name"`
	SkuImage   string  `json:"sku_image"`
	Quantity   int     `json:"quantity"`
	Price      float64 `json:"price"`
	OuterSkuID string  `json:"outer_sku_id"`
}

type PlatformProduct struct {
	SkuID      string  `json:"sku_id"`
	Name       string  `json:"name"`
	Image      string  `json:"image"`
	Price      float64 `json:"price"`
	Stock      int     `json:"stock"`
	OuterSkuID string  `json:"outer_sku_id"`
}

type LogisticsInfo struct {
	Company     string `json:"company"`
	CompanyCode string `json:"company_code"`
	TrackingNo  string `json:"tracking_no"`
}

type AuthToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}
