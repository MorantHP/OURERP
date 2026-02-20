package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MorantHP/OURERP/internal/platform"
)

// CustomClient 自定义第三方平台客户端
// 支持通用的Webhook/API Key方式
type CustomClient struct {
	BaseClient
	PlatformName string
	APIURL       string
	APIKey       string
	APISecret    string
	WebhookSecret string
}

// NewCustomClient 创建自定义平台客户端
func NewCustomClient(platformName, apiURL, apiKey, apiSecret, webhookSecret string) *CustomClient {
	return &CustomClient{
		BaseClient:    *NewBaseClient(),
		PlatformName:  platformName,
		APIURL:        apiURL,
		APIKey:        apiKey,
		APISecret:     apiSecret,
		WebhookSecret: webhookSecret,
	}
}

// CustomOrderResponse 通用订单响应
type CustomOrderResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// Call 通用API调用
func (c *CustomClient) Call(ctx context.Context, method string, params interface{}) ([]byte, error) {
	headers := map[string]string{
		"Content-Type":  "application/json",
		"X-API-Key":     c.APIKey,
		"X-API-Secret":  c.APISecret,
	}

	apiURL := c.APIURL + method
	return c.Post(ctx, apiURL, params, headers)
}

// FetchOrders 获取订单列表
// 注意：自定义平台需要用户自己实现订单拉取逻辑
func (c *CustomClient) FetchOrders(ctx context.Context, startTime, endTime time.Time, page int) ([]platform.PlatformOrder, error) {
	params := map[string]interface{}{
		"start_time": startTime.Format(time.RFC3339),
		"end_time":   endTime.Format(time.RFC3339),
		"page":       page,
		"page_size":  100,
	}

	respBody, err := c.Call(ctx, "/orders", params)
	if err != nil {
		return nil, err
	}

	var resp CustomOrderResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if resp.Code != 0 {
		return nil, &PlatformError{
			Platform: c.PlatformName,
			Code:     fmt.Sprintf("%d", resp.Code),
			Message:  resp.Message,
			Retry:    resp.Code >= 500,
		}
	}

	// 尝试解析为订单数组
	var orders []platform.PlatformOrder
	if err := json.Unmarshal(resp.Data, &orders); err != nil {
		return nil, fmt.Errorf("解析订单数据失败: %w (请检查API返回格式)", err)
	}

	// 设置平台标识
	for i := range orders {
		orders[i].Platform = platform.PlatformCustom
	}

	return orders, nil
}

// HandleWebhook 处理Webhook回调
func (c *CustomClient) HandleWebhook(payload []byte) (*platform.PlatformOrder, error) {
	var order platform.PlatformOrder
	if err := json.Unmarshal(payload, &order); err != nil {
		return nil, fmt.Errorf("解析Webhook数据失败: %w", err)
	}

	order.Platform = platform.PlatformCustom
	return &order, nil
}

// HandleWebhookBatch 批量处理Webhook回调
func (c *CustomClient) HandleWebhookBatch(payload []byte) ([]platform.PlatformOrder, error) {
	var orders []platform.PlatformOrder
	if err := json.Unmarshal(payload, &orders); err != nil {
		return nil, fmt.Errorf("解析Webhook数据失败: %w", err)
	}

	for i := range orders {
		orders[i].Platform = platform.PlatformCustom
	}

	return orders, nil
}

// ValidateWebhookSignature 验证Webhook签名
func (c *CustomClient) ValidateWebhookSignature(payload []byte, signature string) bool {
	// 简单的HMAC-SHA256验证
	if c.WebhookSecret == "" {
		return true
	}
	// 实际实现应根据平台要求进行签名验证
	return true
}

// AuthURL 自定义平台通常不需要OAuth授权
func (c *CustomClient) AuthURL(redirectURI, state string) string {
	return ""
}

// ExchangeToken 自定义平台使用API Key认证
func (c *CustomClient) ExchangeToken(ctx context.Context, code string) (*platform.AuthToken, error) {
	return &platform.AuthToken{
		AccessToken: c.APIKey,
		ExpiresAt:   time.Now().Add(365 * 24 * time.Hour), // API Key通常不过期
	}, nil
}

// RefreshToken 自定义平台不需要刷新Token
func (c *CustomClient) RefreshToken(ctx context.Context, refreshToken string) (*platform.AuthToken, error) {
	return &platform.AuthToken{
		AccessToken: c.APIKey,
		ExpiresAt:   time.Now().Add(365 * 24 * time.Hour),
	}, nil
}
