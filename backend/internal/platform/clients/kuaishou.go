package clients

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/MorantHP/OURERP/internal/platform"
)

// KuaishouClient 快手电商API客户端
type KuaishouClient struct {
	BaseClient
	AppID       string
	AppSecret   string
	AccessToken string
	ShopID      string
	APIURL      string
}

// NewKuaishouClient 创建快手客户端
func NewKuaishouClient(appID, appSecret, accessToken, shopID string) *KuaishouClient {
	return &KuaishouClient{
		BaseClient:  *NewBaseClient(),
		AppID:       appID,
		AppSecret:   appSecret,
		AccessToken: accessToken,
		ShopID:      shopID,
		APIURL:      "https://open.kwaixiaodian.com/api",
	}
}

// KuaishouOrderResponse 订单响应
type KuaishouOrderResponse struct {
	Result     int    `json:"result"`
	ErrorMsg   string `json:"error_msg"`
	Data       struct {
		TotalCount   int              `json:"total_count"`
		OrderList    []KuaishouOrder  `json:"order_list"`
		HasMore      bool             `json:"has_more"`
	} `json:"data"`
}

// KuaishouOrder 快手订单
type KuaishouOrder struct {
	OrderId         string `json:"order_id"`
	OrderStatus     int    `json:"order_status"`
	TotalPrice      int64  `json:"total_price"` // 分
	ActualPay       int64  `json:"actual_pay"`
	BuyerName       string `json:"buyer_name"`
	ReceiverName    string `json:"receiver_name"`
	ReceiverPhone   string `json:"receiver_phone"`
	ReceiverAddress string `json:"receiver_address"`
	CreateTime      int64  `json:"create_time"`
	PayTime         int64  `json:"pay_time"`
	ShipTime        int64  `json:"ship_time"`
	ProductList     []KuaishouProduct `json:"product_list"`
}

// KuaishouProduct 商品
type KuaishouProduct struct {
	SkuId       string `json:"sku_id"`
	OuterSkuId  string `json:"outer_sku_id"`
	Title       string `json:"title"`
	CoverUrl    string `json:"cover_url"`
	Num         int    `json:"num"`
	Price       int64  `json:"price"`
}

// Call 调用快手API
func (c *KuaishouClient) Call(ctx context.Context, method string, params map[string]interface{}) ([]byte, error) {
	timestamp := time.Now().UnixMilli()

	params["app_id"] = c.AppID
	params["timestamp"] = timestamp
	params["version"] = "1"

	// 生成签名
	sign := c.generateSign(params)
	params["sign"] = sign

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	if c.AccessToken != "" {
		headers["Access-Token"] = c.AccessToken
	}

	apiURL := c.APIURL + method
	return c.Post(ctx, apiURL, params, headers)
}

// generateSign 快手签名算法
func (c *KuaishouClient) generateSign(params map[string]interface{}) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf strings.Builder
	buf.WriteString(c.AppSecret)
	for _, k := range keys {
		buf.WriteString(k)
		buf.WriteString(fmt.Sprintf("%v", params[k]))
	}
	buf.WriteString(c.AppSecret)

	hash := md5.Sum([]byte(buf.String()))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

// FetchOrders 获取订单列表
func (c *KuaishouClient) FetchOrders(ctx context.Context, startTime, endTime time.Time, page int) ([]platform.PlatformOrder, error) {
	params := map[string]interface{}{
		"shop_id":       c.ShopID,
		"start_time":    startTime.Unix(),
		"end_time":      endTime.Unix(),
		"page_number":   page,
		"page_size":     100,
		"order_status":  0, // 0-全部
	}

	respBody, err := c.Call(ctx, "/open/seller/order/list", params)
	if err != nil {
		return nil, err
	}

	var resp KuaishouOrderResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if resp.Result != 1 {
		return nil, &PlatformError{
			Platform: "kuaishou",
			Code:     fmt.Sprintf("%d", resp.Result),
			Message:  resp.ErrorMsg,
			Retry:    resp.Result >= 5000,
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
func (c *KuaishouClient) convertOrder(o KuaishouOrder) platform.PlatformOrder {
	order := platform.PlatformOrder{
		PlatformOrderID: o.OrderId,
		Platform:        platform.PlatformKuaishou,
		Status:          fmt.Sprintf("%d", o.OrderStatus),
		TotalAmount:     float64(o.TotalPrice) / 100,
		PayAmount:       float64(o.ActualPay) / 100,
		BuyerNick:       o.BuyerName,
		ReceiverName:    o.ReceiverName,
		ReceiverPhone:   o.ReceiverPhone,
		ReceiverAddress: o.ReceiverAddress,
		CreatedAt:       time.Unix(o.CreateTime, 0),
		Items:           make([]platform.PlatformOrderItem, 0),
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
			SkuImage:   p.CoverUrl,
			Quantity:   p.Num,
			Price:      float64(p.Price) / 100,
			OuterSkuID: p.OuterSkuId,
		})
	}

	return order
}

// AuthURL 获取授权URL
func (c *KuaishouClient) AuthURL(redirectURI, state string) string {
	return fmt.Sprintf(
		"https://s.kwaixiaodian.com/oauth/authorize?app_id=%s&scope=user_info,order&redirect_uri=%s&state=%s",
		c.AppID, url.QueryEscape(redirectURI), state,
	)
}

// ExchangeToken 用授权码换取Token
func (c *KuaishouClient) ExchangeToken(ctx context.Context, code string) (*platform.AuthToken, error) {
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
func (c *KuaishouClient) RefreshToken(ctx context.Context, refreshToken string) (*platform.AuthToken, error) {
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
