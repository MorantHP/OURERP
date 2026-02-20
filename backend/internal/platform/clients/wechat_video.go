package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MorantHP/OURERP/internal/platform"
)

// WechatVideoClient 微信视频号小店API客户端
type WechatVideoClient struct {
	BaseClient
	AppID       string
	AppSecret   string
	AccessToken string
	APIURL      string
}

// NewWechatVideoClient 创建微信视频号客户端
func NewWechatVideoClient(appID, appSecret, accessToken string) *WechatVideoClient {
	return &WechatVideoClient{
		BaseClient:  *NewBaseClient(),
		AppID:       appID,
		AppSecret:   appSecret,
		AccessToken: accessToken,
		APIURL:      "https://api.weixin.qq.com",
	}
}

// WechatVideoOrderResponse 订单响应
type WechatVideoOrderResponse struct {
	Errcode   int    `json:"errcode"`
	Errmsg    string `json:"errmsg"`
	TotalNum  int    `json:"total_num"`
	NextKey   string `json:"next_key"`
	OrderIdList []string `json:"order_id_list"`
}

// WechatVideoOrderDetailResponse 订单详情响应
type WechatVideoOrderDetailResponse struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
	Order   WechatVideoOrder `json:"order"`
}

// WechatVideoOrder 微信视频号订单
type WechatVideoOrder struct {
	OrderId         string `json:"order_id"`
	OrderStatus     int    `json:"order_status"`
	TotalPrice      int64  `json:"total_price"` // 分
	PaymentPrice    int64  `json:"payment_price"`
	BuyerInfo       struct {
		NickName string `json:"nickname"`
	} `json:"buyer_info"`
	ReceiverName    string `json:"receiver_name"`
	ReceiverPhone   string `json:"receiver_phone"`
	ReceiverAddress string `json:"receiver_address"`
	CreateTime      int64  `json:"create_time"`
	PayTime         int64  `json:"pay_time"`
	ShipTime        int64  `json:"ship_time"`
	ProductList     []WechatVideoProduct `json:"product_list"`
}

// WechatVideoProduct 商品
type WechatVideoProduct struct {
	ProductId   string `json:"product_id"`
	SkuId       string `json:"sku_id"`
	OutSkuId    string `json:"out_sku_id"`
	Title       string `json:"title"`
	HeadImg     string `json:"head_img"`
	Count       int    `json:"count"`
	Price       int64  `json:"price"`
}

// Call 调用微信API
func (c *WechatVideoClient) Call(ctx context.Context, method string, params interface{}) ([]byte, error) {
	apiURL := fmt.Sprintf("%s%s?access_token=%s", c.APIURL, method, c.AccessToken)

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	return c.Post(ctx, apiURL, params, headers)
}

// GetAccessToken 获取access_token
func (c *WechatVideoClient) GetAccessToken(ctx context.Context) (string, error) {
	apiURL := fmt.Sprintf("%s/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		c.APIURL, c.AppID, c.AppSecret)

	respBody, err := c.Get(ctx, apiURL, nil)
	if err != nil {
		return "", err
	}

	var resp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		Errcode     int    `json:"errcode"`
		Errmsg      string `json:"errmsg"`
	}

	if err := json.Unmarshal(respBody, &resp); err != nil {
		return "", err
	}

	if resp.Errcode != 0 {
		return "", &PlatformError{
			Platform: "wechat_video",
			Code:     fmt.Sprintf("%d", resp.Errcode),
			Message:  resp.Errmsg,
			Retry:    resp.Errcode >= 5000,
		}
	}

	return resp.AccessToken, nil
}

// FetchOrders 获取订单ID列表
func (c *WechatVideoClient) FetchOrders(ctx context.Context, startTime, endTime time.Time, nextKey string) ([]string, string, error) {
	params := map[string]interface{}{
		"begin_time": startTime.Unix(),
		"end_time":   endTime.Unix(),
		"page_size":  100,
	}
	if nextKey != "" {
		params["next_key"] = nextKey
	}

	respBody, err := c.Call(ctx, "/channels/ec/order/list/get", params)
	if err != nil {
		return nil, "", err
	}

	var resp WechatVideoOrderResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, "", fmt.Errorf("解析响应失败: %w", err)
	}

	if resp.Errcode != 0 {
		return nil, "", &PlatformError{
			Platform: "wechat_video",
			Code:     fmt.Sprintf("%d", resp.Errcode),
			Message:  resp.Errmsg,
			Retry:    resp.Errcode >= 5000,
		}
	}

	return resp.OrderIdList, resp.NextKey, nil
}

// GetOrderDetail 获取订单详情
func (c *WechatVideoClient) GetOrderDetail(ctx context.Context, orderID string) (*platform.PlatformOrder, error) {
	params := map[string]interface{}{
		"order_id": orderID,
	}

	respBody, err := c.Call(ctx, "/channels/ec/order/get", params)
	if err != nil {
		return nil, err
	}

	var resp WechatVideoOrderDetailResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if resp.Errcode != 0 {
		return nil, &PlatformError{
			Platform: "wechat_video",
			Code:     fmt.Sprintf("%d", resp.Errcode),
			Message:  resp.Errmsg,
			Retry:    resp.Errcode >= 5000,
		}
	}

	order := c.convertOrder(resp.Order)
	return &order, nil
}

// convertOrder 转换订单格式
func (c *WechatVideoClient) convertOrder(o WechatVideoOrder) platform.PlatformOrder {
	order := platform.PlatformOrder{
		PlatformOrderID:  o.OrderId,
		Platform:         platform.PlatformWechatVideo,
		Status:           fmt.Sprintf("%d", o.OrderStatus),
		TotalAmount:      float64(o.TotalPrice) / 100,
		PayAmount:        float64(o.PaymentPrice) / 100,
		BuyerNick:        o.BuyerInfo.NickName,
		ReceiverName:     o.ReceiverName,
		ReceiverPhone:    o.ReceiverPhone,
		ReceiverAddress:  o.ReceiverAddress,
		CreatedAt:        time.Unix(o.CreateTime, 0),
		Items:            make([]platform.PlatformOrderItem, 0),
	}

	if o.PayTime > 0 {
		t := time.Unix(o.PayTime, 0)
		order.PaidAt = &t
	}
	if o.ShipTime > 0 {
		t := time.Unix(o.ShipTime, 0)
		order.ShippedAt = &t
	}

	// 转换商品明细
	for _, p := range o.ProductList {
		order.Items = append(order.Items, platform.PlatformOrderItem{
			SkuID:      p.SkuId,
			SkuName:    p.Title,
			SkuImage:   p.HeadImg,
			Quantity:   p.Count,
			Price:      float64(p.Price) / 100,
			OuterSkuID: p.OutSkuId,
		})
	}

	return order
}

// ExchangeToken 换取Token (微信使用授权码换取access_token)
func (c *WechatVideoClient) ExchangeToken(ctx context.Context, code string) (*platform.AuthToken, error) {
	apiURL := fmt.Sprintf("%s/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		c.APIURL, c.AppID, c.AppSecret, code)

	respBody, err := c.Get(ctx, apiURL, nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
		Openid       string `json:"openid"`
		Errcode      int    `json:"errcode"`
		Errmsg       string `json:"errmsg"`
	}

	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析Token响应失败: %w", err)
	}

	if resp.Errcode != 0 {
		return nil, &PlatformError{
			Platform: "wechat_video",
			Code:     fmt.Sprintf("%d", resp.Errcode),
			Message:  resp.Errmsg,
		}
	}

	return &platform.AuthToken{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second),
	}, nil
}

// RefreshToken 刷新Token
func (c *WechatVideoClient) RefreshToken(ctx context.Context, refreshToken string) (*platform.AuthToken, error) {
	apiURL := fmt.Sprintf("%s/sns/oauth2/refresh_token?appid=%s&grant_type=refresh_token&refresh_token=%s",
		c.APIURL, c.AppID, refreshToken)

	respBody, err := c.Get(ctx, apiURL, nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
		Errcode      int    `json:"errcode"`
		Errmsg       string `json:"errmsg"`
	}

	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析Token响应失败: %w", err)
	}

	if resp.Errcode != 0 {
		return nil, &PlatformError{
			Platform: "wechat_video",
			Code:     fmt.Sprintf("%d", resp.Errcode),
			Message:  resp.Errmsg,
		}
	}

	return &platform.AuthToken{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second),
	}, nil
}
