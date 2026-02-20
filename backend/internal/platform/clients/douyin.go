package clients

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/MorantHP/OURERP/internal/platform"
)

// DouyinClient 抖音电商API客户端
type DouyinClient struct {
	BaseClient
	AppKey      string
	AppSecret   string
	AccessToken string
	ShopID      string
	APIURL      string
}

// NewDouyinClient 创建抖音客户端
func NewDouyinClient(appKey, appSecret, accessToken, shopID string) *DouyinClient {
	return &DouyinClient{
		BaseClient:  *NewBaseClient(),
		AppKey:      appKey,
		AppSecret:   appSecret,
		AccessToken: accessToken,
		ShopID:      shopID,
		APIURL:      "https://developer.toutiao.com/api",
	}
}

// DouyinOrderResponse 订单响应
type DouyinOrderResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Data      struct {
		Total       int            `json:"total"`
		HasMore     bool           `json:"has_more"`
		OrderList   []DouyinOrder  `json:"order_list"`
	} `json:"data"`
	Extra     struct {
		SubCode  int    `json:"sub_code"`
		SubMsg   string `json:"sub_msg"`
	} `json:"extra"`
}

// DouyinOrder 抖音订单
type DouyinOrder struct {
	OrderID        string `json:"order_id"`
	OrderStatus    int    `json:"order_status"`
	TotalAmount    int64  `json:"total_amount"` // 分
	PayAmount      int64  `json:"pay_amount"`
	BuyerName      string `json:"buyer_name"`
	BuyerPhone     string `json:"buyer_phone"`
	ReceiverName   string `json:"receiver_name"`
	ReceiverPhone  string `json:"receiver_phone"`
	ReceiverState  string `json:"receiver_state"`
	ReceiverCity   string `json:"receiver_city"`
	ReceiverDistrict string `json:"receiver_district"`
	ReceiverAddress  string `json:"receiver_address"`
	CreateTime     int64  `json:"create_time"`
	PayTime        int64  `json:"pay_time"`
	ShipTime       int64  `json:"ship_time"`
	ProductList    []DouyinProduct `json:"product_list"`
}

// DouyinProduct 商品
type DouyinProduct struct {
	ProductID   string `json:"product_id"`
	SkuID       string `json:"sku_id"`
	OuterSkuID  string `json:"outer_sku_id"`
	Name        string `json:"name"`
	Image       string `json:"image"`
	Count       int    `json:"count"`
	Price       int64  `json:"price"` // 分
}

// Call 调用抖音API
func (c *DouyinClient) Call(ctx context.Context, method string, params interface{}) ([]byte, error) {
	timestamp := time.Now().Unix()

	reqBody, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	// 生成签名
	sign := c.generateSign(method, string(reqBody), timestamp)

	apiURL := fmt.Sprintf("%s%s?app_id=%s&timestamp=%d&sign=%s",
		c.APIURL, method, c.AppKey, timestamp, sign)

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	if c.AccessToken != "" {
		headers["Access-Token"] = c.AccessToken
	}

	return c.Post(ctx, apiURL, params, headers)
}

// generateSign 生成抖音API签名
func (c *DouyinClient) generateSign(method, body string, timestamp int64) string {
	h := hmac.New(sha256.New, []byte(c.AppSecret))
	h.Write([]byte(fmt.Sprintf("%s%s%d", method, body, timestamp)))
	return hex.EncodeToString(h.Sum(nil))
}

// FetchOrders 获取订单列表
func (c *DouyinClient) FetchOrders(ctx context.Context, startTime, endTime time.Time, page int) ([]platform.PlatformOrder, error) {
	params := map[string]interface{}{
		"shop_id":       c.ShopID,
		"start_time":    startTime.Unix(),
		"end_time":      endTime.Unix(),
		"page":          page,
		"size":          100,
		"order_status":  0, // 0-全部
	}

	respBody, err := c.Call(ctx, "/order/search", params)
	if err != nil {
		return nil, err
	}

	var resp DouyinOrderResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if resp.Code != 0 {
		return nil, &PlatformError{
			Platform: "douyin",
			Code:     fmt.Sprintf("%d", resp.Code),
			Message:  resp.Message,
			SubMsg:   resp.Extra.SubMsg,
			Retry:    resp.Code >= 5000,
		}
	}

	// 转换为统一订单格式
	orders := make([]platform.PlatformOrder, 0, len(resp.Data.OrderList))
	for _, o := range resp.Data.OrderList {
		orders = append(orders, c.convertOrder(o))
	}

	return orders, nil
}

// convertOrder 转换订单格式
func (c *DouyinClient) convertOrder(o DouyinOrder) platform.PlatformOrder {
	order := platform.PlatformOrder{
		PlatformOrderID:  o.OrderID,
		Platform:         platform.PlatformDouyin,
		Status:           fmt.Sprintf("%d", o.OrderStatus),
		TotalAmount:      float64(o.TotalAmount) / 100,
		PayAmount:        float64(o.PayAmount) / 100,
		BuyerNick:        o.BuyerName,
		BuyerPhone:       o.BuyerPhone,
		ReceiverName:     o.ReceiverName,
		ReceiverPhone:    o.ReceiverPhone,
		ReceiverProvince: o.ReceiverState,
		ReceiverCity:     o.ReceiverCity,
		ReceiverDistrict: o.ReceiverDistrict,
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
			SkuID:      p.SkuID,
			SkuName:    p.Name,
			SkuImage:   p.Image,
			Quantity:   p.Count,
			Price:      float64(p.Price) / 100,
			OuterSkuID: p.OuterSkuID,
		})
	}

	return order
}

// AuthURL 获取授权URL
func (c *DouyinClient) AuthURL(redirectURI, state string) string {
	return fmt.Sprintf(
		"https://developer.toutiao.com/api/oauth/connect/?app_id=%s&scope=user_info,order&redirect_uri=%s&state=%s",
		c.AppKey, url.QueryEscape(redirectURI), state,
	)
}

// ExchangeToken 用授权码换取Token
func (c *DouyinClient) ExchangeToken(ctx context.Context, code string) (*platform.AuthToken, error) {
	params := map[string]interface{}{
		"code":       code,
		"grant_type": "authorization_code",
	}

	respBody, err := c.Call(ctx, "/oauth/token", params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
	}

	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析Token响应失败: %w", err)
	}

	return &platform.AuthToken{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second),
	}, nil
}

// RefreshToken 刷新Token
func (c *DouyinClient) RefreshToken(ctx context.Context, refreshToken string) (*platform.AuthToken, error) {
	params := map[string]interface{}{
		"refresh_token": refreshToken,
		"grant_type":    "refresh_token",
	}

	respBody, err := c.Call(ctx, "/oauth/token", params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
	}

	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析Token响应失败: %w", err)
	}

	return &platform.AuthToken{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second),
	}, nil
}
