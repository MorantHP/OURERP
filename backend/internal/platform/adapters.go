package platform

import (
	"fmt"
	"math/rand"
	"time"
)

// BaseAdapter 基础适配器
type BaseAdapter struct {
	config PlatformConfig
}

func (a *BaseAdapter) GetConfig() PlatformConfig {
	return a.config
}

// 各平台适配器

type TaobaoAdapter struct{ BaseAdapter }
type JDAdapter struct{ BaseAdapter }
type DouyinAdapter struct{ BaseAdapter }
type KuaishouAdapter struct{ BaseAdapter }
type XiaohongshuAdapter struct{ BaseAdapter }
type VipAdapter struct{ BaseAdapter }
type WechatAdapter struct{ BaseAdapter }
type OfflineAdapter struct{ BaseAdapter }
type CustomAdapter struct{ BaseAdapter }

// 工厂函数
func NewAdapter(platform PlatformType) (PlatformAdapter, error) {
	config, ok := GetPlatformConfig(platform)
	if !ok {
		return nil, fmt.Errorf("不支持的平台: %s", platform)
	}

	base := BaseAdapter{config: config}

	switch platform {
	case PlatformTaobao:
		return &TaobaoAdapter{base}, nil
	case PlatformJD:
		return &JDAdapter{base}, nil
	case PlatformDouyin:
		return &DouyinAdapter{base}, nil
	case PlatformKuaishou:
		return &KuaishouAdapter{base}, nil
	case PlatformXiaohongshu:
		return &XiaohongshuAdapter{base}, nil
	case PlatformVip:
		return &VipAdapter{base}, nil
	case PlatformWechat:
		return &WechatAdapter{base}, nil
	case PlatformOffline:
		return &OfflineAdapter{base}, nil
	case PlatformCustom:
		return &CustomAdapter{base}, nil
	default:
		return nil, fmt.Errorf("未知平台: %s", platform)
	}
}

// 模拟实现：生成模拟订单数据
func generateMockOrders(platform PlatformType, count int) []PlatformOrder {
	orders := make([]PlatformOrder, count)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	statuses := []string{"WAIT_PAY", "WAIT_SEND", "SEND", "SUCCESS", "CLOSE"}

	for i := 0; i < count; i++ {
		status := statuses[r.Intn(len(statuses))]
		createdAt := time.Now().Add(-time.Duration(r.Intn(168)) * time.Hour)

		order := PlatformOrder{
			PlatformOrderID:  fmt.Sprintf("%s%d%d", platform, time.Now().Unix(), r.Intn(10000)),
			Platform:         platform,
			Status:           status,
			TotalAmount:      float64(r.Intn(10000)) + float64(r.Intn(100))/100,
			PayAmount:        0,
			BuyerNick:        fmt.Sprintf("买家%d", r.Intn(10000)),
			BuyerPhone:       fmt.Sprintf("138%08d", r.Intn(100000000)),
			ReceiverName:     fmt.Sprintf("收件人%d", r.Intn(100)),
			ReceiverPhone:    fmt.Sprintf("139%08d", r.Intn(100000000)),
			ReceiverProvince: "浙江省",
			ReceiverCity:     "杭州市",
			ReceiverDistrict: "西湖区",
			ReceiverAddress:  fmt.Sprintf("文三路%d号", r.Intn(500)),
			CreatedAt:        createdAt,
			Items: []PlatformOrderItem{
				{
					SkuID:    fmt.Sprintf("SKU%d", r.Intn(10000)),
					SkuName:  "测试商品",
					Quantity: r.Intn(3) + 1,
					Price:    float64(r.Intn(1000)),
				},
			},
		}

		if status != "WAIT_PAY" && status != "CLOSE" {
			order.PayAmount = order.TotalAmount
			order.PaidAt = &createdAt
		}

		orders[i] = order
	}

	return orders
}

// 各平台适配器实现（模拟）

func (a *TaobaoAdapter) FetchOrders(startTime, endTime time.Time, page int) ([]PlatformOrder, error) {
	return generateMockOrders(PlatformTaobao, 10), nil
}

func (a *TaobaoAdapter) GetOrderDetail(orderID string) (*PlatformOrder, error) {
	orders, _ := a.FetchOrders(time.Now(), time.Now(), 1)
	if len(orders) > 0 {
		return &orders[0], nil
	}
	return nil, fmt.Errorf("订单不存在")
}

func (a *TaobaoAdapter) ShipOrder(orderID string, logistics LogisticsInfo) error {
	return nil
}

func (a *TaobaoAdapter) FetchProducts(page int) ([]PlatformProduct, error) {
	return nil, nil
}

func (a *TaobaoAdapter) UpdateStock(skuID string, quantity int) error {
	return nil
}

func (a *TaobaoAdapter) AuthURL(redirectURI string) string {
	return "https://oauth.taobao.com/authorize"
}

func (a *TaobaoAdapter) ExchangeToken(code string) (*AuthToken, error) {
	return &AuthToken{AccessToken: "mock_token"}, nil
}

func (a *TaobaoAdapter) RefreshToken(refreshToken string) (*AuthToken, error) {
	return &AuthToken{AccessToken: "mock_token"}, nil
}

// 其他平台类似实现（简化）
func (a *JDAdapter) FetchOrders(startTime, endTime time.Time, page int) ([]PlatformOrder, error) {
	return generateMockOrders(PlatformJD, 10), nil
}
func (a *JDAdapter) GetOrderDetail(orderID string) (*PlatformOrder, error)   { return nil, nil }
func (a *JDAdapter) ShipOrder(orderID string, logistics LogisticsInfo) error { return nil }
func (a *JDAdapter) FetchProducts(page int) ([]PlatformProduct, error)       { return nil, nil }
func (a *JDAdapter) UpdateStock(skuID string, quantity int) error            { return nil }
func (a *JDAdapter) AuthURL(redirectURI string) string                       { return "" }
func (a *JDAdapter) ExchangeToken(code string) (*AuthToken, error)           { return nil, nil }
func (a *JDAdapter) RefreshToken(refreshToken string) (*AuthToken, error)    { return nil, nil }

func (a *DouyinAdapter) FetchOrders(startTime, endTime time.Time, page int) ([]PlatformOrder, error) {
	return generateMockOrders(PlatformDouyin, 10), nil
}
func (a *DouyinAdapter) GetOrderDetail(orderID string) (*PlatformOrder, error)   { return nil, nil }
func (a *DouyinAdapter) ShipOrder(orderID string, logistics LogisticsInfo) error { return nil }
func (a *DouyinAdapter) FetchProducts(page int) ([]PlatformProduct, error)       { return nil, nil }
func (a *DouyinAdapter) UpdateStock(skuID string, quantity int) error            { return nil }
func (a *DouyinAdapter) AuthURL(redirectURI string) string                       { return "" }
func (a *DouyinAdapter) ExchangeToken(code string) (*AuthToken, error)           { return nil, nil }
func (a *DouyinAdapter) RefreshToken(refreshToken string) (*AuthToken, error)    { return nil, nil }

// 快手、小红书、唯品会、微信、线下、自定义平台类似实现...

func (a *KuaishouAdapter) FetchOrders(startTime, endTime time.Time, page int) ([]PlatformOrder, error) {
	return generateMockOrders(PlatformKuaishou, 10), nil
}
func (a *KuaishouAdapter) GetOrderDetail(orderID string) (*PlatformOrder, error)   { return nil, nil }
func (a *KuaishouAdapter) ShipOrder(orderID string, logistics LogisticsInfo) error { return nil }
func (a *KuaishouAdapter) FetchProducts(page int) ([]PlatformProduct, error)       { return nil, nil }
func (a *KuaishouAdapter) UpdateStock(skuID string, quantity int) error            { return nil }
func (a *KuaishouAdapter) AuthURL(redirectURI string) string                       { return "" }
func (a *KuaishouAdapter) ExchangeToken(code string) (*AuthToken, error)           { return nil, nil }
func (a *KuaishouAdapter) RefreshToken(refreshToken string) (*AuthToken, error)    { return nil, nil }

func (a *XiaohongshuAdapter) FetchOrders(startTime, endTime time.Time, page int) ([]PlatformOrder, error) {
	return generateMockOrders(PlatformXiaohongshu, 10), nil
}
func (a *XiaohongshuAdapter) GetOrderDetail(orderID string) (*PlatformOrder, error)   { return nil, nil }
func (a *XiaohongshuAdapter) ShipOrder(orderID string, logistics LogisticsInfo) error { return nil }
func (a *XiaohongshuAdapter) FetchProducts(page int) ([]PlatformProduct, error)       { return nil, nil }
func (a *XiaohongshuAdapter) UpdateStock(skuID string, quantity int) error            { return nil }
func (a *XiaohongshuAdapter) AuthURL(redirectURI string) string                       { return "" }
func (a *XiaohongshuAdapter) ExchangeToken(code string) (*AuthToken, error)           { return nil, nil }
func (a *XiaohongshuAdapter) RefreshToken(refreshToken string) (*AuthToken, error)    { return nil, nil }

func (a *VipAdapter) FetchOrders(startTime, endTime time.Time, page int) ([]PlatformOrder, error) {
	return generateMockOrders(PlatformVip, 10), nil
}
func (a *VipAdapter) GetOrderDetail(orderID string) (*PlatformOrder, error)   { return nil, nil }
func (a *VipAdapter) ShipOrder(orderID string, logistics LogisticsInfo) error { return nil }
func (a *VipAdapter) FetchProducts(page int) ([]PlatformProduct, error)       { return nil, nil }
func (a *VipAdapter) UpdateStock(skuID string, quantity int) error            { return nil }
func (a *VipAdapter) AuthURL(redirectURI string) string                       { return "" }
func (a *VipAdapter) ExchangeToken(code string) (*AuthToken, error)           { return nil, nil }
func (a *VipAdapter) RefreshToken(refreshToken string) (*AuthToken, error)    { return nil, nil }

func (a *WechatAdapter) FetchOrders(startTime, endTime time.Time, page int) ([]PlatformOrder, error) {
	return generateMockOrders(PlatformWechat, 10), nil
}
func (a *WechatAdapter) GetOrderDetail(orderID string) (*PlatformOrder, error)   { return nil, nil }
func (a *WechatAdapter) ShipOrder(orderID string, logistics LogisticsInfo) error { return nil }
func (a *WechatAdapter) FetchProducts(page int) ([]PlatformProduct, error)       { return nil, nil }
func (a *WechatAdapter) UpdateStock(skuID string, quantity int) error            { return nil }
func (a *WechatAdapter) AuthURL(redirectURI string) string                       { return "" }
func (a *WechatAdapter) ExchangeToken(code string) (*AuthToken, error)           { return nil, nil }
func (a *WechatAdapter) RefreshToken(refreshToken string) (*AuthToken, error)    { return nil, nil }

func (a *OfflineAdapter) FetchOrders(startTime, endTime time.Time, page int) ([]PlatformOrder, error) {
	return generateMockOrders(PlatformOffline, 5), nil
}
func (a *OfflineAdapter) GetOrderDetail(orderID string) (*PlatformOrder, error)   { return nil, nil }
func (a *OfflineAdapter) ShipOrder(orderID string, logistics LogisticsInfo) error { return nil }
func (a *OfflineAdapter) FetchProducts(page int) ([]PlatformProduct, error)       { return nil, nil }
func (a *OfflineAdapter) UpdateStock(skuID string, quantity int) error            { return nil }
func (a *OfflineAdapter) AuthURL(redirectURI string) string                       { return "" }
func (a *OfflineAdapter) ExchangeToken(code string) (*AuthToken, error)           { return &AuthToken{}, nil }
func (a *OfflineAdapter) RefreshToken(refreshToken string) (*AuthToken, error) {
	return &AuthToken{}, nil
}

func (a *CustomAdapter) FetchOrders(startTime, endTime time.Time, page int) ([]PlatformOrder, error) {
	return generateMockOrders(PlatformCustom, 10), nil
}
func (a *CustomAdapter) GetOrderDetail(orderID string) (*PlatformOrder, error)   { return nil, nil }
func (a *CustomAdapter) ShipOrder(orderID string, logistics LogisticsInfo) error { return nil }
func (a *CustomAdapter) FetchProducts(page int) ([]PlatformProduct, error)       { return nil, nil }
func (a *CustomAdapter) UpdateStock(skuID string, quantity int) error            { return nil }
func (a *CustomAdapter) AuthURL(redirectURI string) string                       { return "" }
func (a *CustomAdapter) ExchangeToken(code string) (*AuthToken, error)           { return nil, nil }
func (a *CustomAdapter) RefreshToken(refreshToken string) (*AuthToken, error)    { return nil, nil }
