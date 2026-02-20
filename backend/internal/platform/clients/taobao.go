package clients

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/MorantHP/OURERP/internal/platform"
)

// TaobaoClient 淘宝/天猫API客户端
type TaobaoClient struct {
	BaseClient
	AppKey      string
	AppSecret   string
	AccessToken string
	APIURL      string
}

// NewTaobaoClient 创建淘宝客户端
func NewTaobaoClient(appKey, appSecret, accessToken string) *TaobaoClient {
	return &TaobaoClient{
		BaseClient:  *NewBaseClient(),
		AppKey:      appKey,
		AppSecret:   appSecret,
		AccessToken: accessToken,
		APIURL:      "https://eco.taobao.com/router/rest",
	}
}

// TaobaoAPIResponse 通用API响应
type TaobaoAPIResponse struct {
	ErrorResponse struct {
		Code    int    `json:"code"`
		Msg     string `json:"msg"`
		SubCode string `json:"sub_code"`
		SubMsg  string `json:"sub_msg"`
	} `json:"error_response"`
}

// TaobaoOrderResponse 订单响应
type TaobaoOrderResponse struct {
	TaobaoAPIResponse
	TradesSoldIncrementGetResponse struct {
		TotalResults int `json:"total_results"`
		Trades       struct {
			Trade []TaobaoTrade `json:"trade"`
		} `json:"trades"`
	} `json:"taobao_trades_sold_increment_get_response"`
}

// TaobaoTrade 淘宝订单
type TaobaoTrade struct {
	Tid              int64   `json:"tid"`
	Status           string  `json:"status"`
	Payment          string  `json:"payment"`
	TotalFee         string  `json:"total_fee"`
	BuyerNick        string  `json:"buyer_nick"`
	ReceiverName     string  `json:"receiver_name"`
	ReceiverMobile   string  `json:"receiver_mobile"`
	ReceiverPhone    string  `json:"receiver_phone"`
	ReceiverState    string  `json:"receiver_state"`
	ReceiverCity     string  `json:"receiver_city"`
	ReceiverDistrict string  `json:"receiver_district"`
	ReceiverAddress  string  `json:"receiver_address"`
	Created          string  `json:"created"`
	Modified         string  `json:"modified"`
	PayTime          string  `json:"pay_time"`
	ConsignTime      string  `json:"consign_time"`
	Orders           struct {
		Order []TaobaoOrder `json:"order"`
	} `json:"orders"`
}

// TaobaoOrder 订单商品
type TaobaoOrder struct {
	Oid        int64  `json:"oid"`
	SkuId      int64  `json:"sku_id"`
	OuterSkuId string `json:"outer_sku_id"`
	Title      string `json:"title"`
	Num        int    `json:"num"`
	Price      string `json:"price"`
	PicPath    string `json:"pic_path"`
}

// Call 调用淘宝API
func (c *TaobaoClient) Call(ctx context.Context, method string, params map[string]string, needSession bool) ([]byte, error) {
	allParams := map[string]string{
		"method":      method,
		"app_key":     c.AppKey,
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"format":      "json",
		"v":           "2.0",
		"sign_method": "md5",
	}

	if needSession && c.AccessToken != "" {
		allParams["session"] = c.AccessToken
	}

	// 合并业务参数
	for k, v := range params {
		allParams[k] = v
	}

	// 生成签名
	allParams["sign"] = c.generateSign(allParams)

	// 构建URL
	values := url.Values{}
	for k, v := range allParams {
		values.Set(k, v)
	}
	apiURL := c.APIURL + "?" + values.Encode()

	return c.Get(ctx, apiURL, nil)
}

// generateSign 生成淘宝API签名
func (c *TaobaoClient) generateSign(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf strings.Builder
	buf.WriteString(c.AppSecret)
	for _, k := range keys {
		buf.WriteString(k)
		buf.WriteString(params[k])
	}
	buf.WriteString(c.AppSecret)

	hash := md5.Sum([]byte(buf.String()))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

// FetchOrders 获取订单列表
func (c *TaobaoClient) FetchOrders(ctx context.Context, startTime, endTime time.Time, pageNo int) ([]platform.PlatformOrder, error) {
	params := map[string]string{
		"fields":         "tid,status,payment,total_fee,buyer_nick,receiver_name,receiver_mobile,receiver_phone,receiver_state,receiver_city,receiver_district,receiver_address,created,modified,pay_time,consign_time,orders",
		"start_modified": startTime.Format("2006-01-02 15:04:05"),
		"end_modified":   endTime.Format("2006-01-02 15:04:05"),
		"page_no":        strconv.Itoa(pageNo),
		"page_size":      "100",
	}

	respBody, err := c.Call(ctx, "taobao.trades.sold.increment.get", params, true)
	if err != nil {
		return nil, err
	}

	var resp TaobaoOrderResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if resp.ErrorResponse.Code != 0 {
		return nil, &PlatformError{
			Platform: "taobao",
			Code:     strconv.Itoa(resp.ErrorResponse.Code),
			Message:  resp.ErrorResponse.Msg,
			SubCode:  resp.ErrorResponse.SubCode,
			SubMsg:   resp.ErrorResponse.SubMsg,
			Retry:    resp.ErrorResponse.Code >= 50,
		}
	}

	// 转换为统一订单格式
	orders := make([]platform.PlatformOrder, 0, len(resp.TradesSoldIncrementGetResponse.Trades.Trade))
	for _, trade := range resp.TradesSoldIncrementGetResponse.Trades.Trade {
		orders = append(orders, c.convertOrder(trade))
	}

	return orders, nil
}

// convertOrder 转换订单格式
func (c *TaobaoClient) convertOrder(trade TaobaoTrade) platform.PlatformOrder {
	order := platform.PlatformOrder{
		PlatformOrderID:  strconv.FormatInt(trade.Tid, 10),
		Platform:         platform.PlatformTaobao,
		Status:           trade.Status,
		TotalAmount:      parseFloat(trade.TotalFee),
		PayAmount:        parseFloat(trade.Payment),
		BuyerNick:        trade.BuyerNick,
		ReceiverName:     trade.ReceiverName,
		ReceiverPhone:    trade.ReceiverMobile,
		ReceiverProvince: trade.ReceiverState,
		ReceiverCity:     trade.ReceiverCity,
		ReceiverDistrict: trade.ReceiverDistrict,
		ReceiverAddress:  trade.ReceiverAddress,
		Items:            make([]platform.PlatformOrderItem, 0),
	}

	// 解析时间
	if t, err := time.ParseInLocation("2006-01-02 15:04:05", trade.Created, time.Local); err == nil {
		order.CreatedAt = t
	}
	if t, err := time.ParseInLocation("2006-01-02 15:04:05", trade.PayTime, time.Local); err == nil {
		order.PaidAt = &t
	}
	if t, err := time.ParseInLocation("2006-01-02 15:04:05", trade.ConsignTime, time.Local); err == nil {
		order.ShippedAt = &t
	}

	// 转换商品明细
	for _, item := range trade.Orders.Order {
		order.Items = append(order.Items, platform.PlatformOrderItem{
			SkuID:      strconv.FormatInt(item.SkuId, 10),
			SkuName:    item.Title,
			SkuImage:   item.PicPath,
			Quantity:   item.Num,
			Price:      parseFloat(item.Price),
			OuterSkuID: item.OuterSkuId,
		})
	}

	return order
}

// AuthURL 获取授权URL
func (c *TaobaoClient) AuthURL(redirectURI, state string) string {
	return fmt.Sprintf(
		"https://oauth.taobao.com/authorize?response_type=code&client_id=%s&redirect_uri=%s&state=%s",
		c.AppKey, url.QueryEscape(redirectURI), state,
	)
}

// ExchangeToken 用授权码换取Token
func (c *TaobaoClient) ExchangeToken(ctx context.Context, code, redirectURI string) (*platform.AuthToken, error) {
	params := map[string]string{
		"grant_type":   "authorization_code",
		"code":         code,
		"redirect_uri": redirectURI,
	}

	respBody, err := c.Call(ctx, "taobao.top.auth.token.create", params, false)
	if err != nil {
		return nil, err
	}

	var resp struct {
		TopAuthTokenCreateResponse struct {
			AuthResult struct {
				AccessToken      string `json:"access_token"`
				RefreshToken     string `json:"refresh_token"`
				ExpiresIn        int64  `json:"expires_in"`
				RefreshTokenExpiresIn int64 `json:"refresh_token_expires_in"`
			} `json:"auth_result"`
		} `json:"top_auth_token_create_response"`
	}

	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析Token响应失败: %w", err)
	}

	return &platform.AuthToken{
		AccessToken:  resp.TopAuthTokenCreateResponse.AuthResult.AccessToken,
		RefreshToken: resp.TopAuthTokenCreateResponse.AuthResult.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(resp.TopAuthTokenCreateResponse.AuthResult.ExpiresIn) * time.Second),
	}, nil
}

// RefreshToken 刷新Token
func (c *TaobaoClient) RefreshToken(ctx context.Context, refreshToken string) (*platform.AuthToken, error) {
	params := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
	}

	respBody, err := c.Call(ctx, "taobao.top.auth.token.refresh", params, false)
	if err != nil {
		return nil, err
	}

	var resp struct {
		TopAuthTokenRefreshResponse struct {
			AuthResult struct {
				AccessToken      string `json:"access_token"`
				RefreshToken     string `json:"refresh_token"`
				ExpiresIn        int64  `json:"expires_in"`
			} `json:"auth_result"`
		} `json:"top_auth_token_refresh_response"`
	}

	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析Token响应失败: %w", err)
	}

	return &platform.AuthToken{
		AccessToken:  resp.TopAuthTokenRefreshResponse.AuthResult.AccessToken,
		RefreshToken: resp.TopAuthTokenRefreshResponse.AuthResult.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(resp.TopAuthTokenRefreshResponse.AuthResult.ExpiresIn) * time.Second),
	}, nil
}

// parseFloat 解析浮点数
func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}
