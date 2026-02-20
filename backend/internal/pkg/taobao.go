package pkg

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"
)

type TaobaoClient struct {
	AppKey    string
	AppSecret string
	Session   string
}

type TaobaoOrder struct {
	Tid             string `json:"tid"`
	Status          string `json:"status"`
	Payment         string `json:"payment"`
	BuyerNick       string `json:"buyer_nick"`
	ReceiverName    string `json:"receiver_name"`
	ReceiverMobile  string `json:"receiver_mobile"`
	ReceiverState   string `json:"receiver_state"`
	ReceiverCity    string `json:"receiver_city"`
	ReceiverDistrict string `json:"receiver_district"`
	ReceiverAddress string `json:"receiver_address"`
	Created         string `json:"created"`
	PayTime         string `json:"pay_time"`
	Orders          struct {
		Order []struct {
			SkuID       string `json:"sku_id"`
			OuterSkuID  string `json:"outer_sku_id"`
			Title       string `json:"title"`
			Num         int    `json:"num"`
			Price       string `json:"price"`
			PicPath     string `json:"pic_path"`
		} `json:"order"`
	} `json:"orders"`
}

func NewTaobaoClient(appKey, appSecret, session string) *TaobaoClient {
	return &TaobaoClient{
		AppKey:    appKey,
		AppSecret: appSecret,
		Session:   session,
	}
}

// FetchOrders 拉取订单
func (c *TaobaoClient) FetchOrders(startTime, endTime time.Time, pageNo int) ([]TaobaoOrder, int, error) {
	params := map[string]string{
		"method":       "taobao.trades.sold.increment.get",
		"app_key":      c.AppKey,
		"timestamp":    time.Now().Format("2006-01-02 15:04:05"),
		"format":       "json",
		"v":            "2.0",
		"sign_method":  "md5",
		"session":      c.Session,
		"fields":       "tid,status,payment,buyer_nick,receiver_name,receiver_mobile,receiver_state,receiver_city,receiver_district,receiver_address,created,pay_time,orders",
		"start_modified": startTime.Format("2006-01-02 15:04:05"),
		"end_modified":   endTime.Format("2006-01-02 15:04:05"),
		"page_no":        strconv.Itoa(pageNo),
		"page_size":      "100",
	}

	params["sign"] = c.generateSign(params)

	apiURL := "https://eco.taobao.com/router/rest"
	resp, err := http.Get(apiURL + "?" + encodeParams(params))
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	var result struct {
		TradesSoldGetResponse struct {
			Trades struct {
				Trade []TaobaoOrder `json:"trade"`
			} `json:"trades"`
			TotalResults int `json:"total_results"`
		} `json:"trades_sold_get_response"`
		ErrorResponse struct {
			SubMsg string `json:"sub_msg"`
			Msg    string `json:"msg"`
		} `json:"error_response"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, 0, err
	}

	if result.ErrorResponse.Msg != "" {
		return nil, 0, fmt.Errorf("淘宝API错误: %s", result.ErrorResponse.SubMsg)
	}

	return result.TradesSoldGetResponse.Trades.Trade, 
	       result.TradesSoldGetResponse.TotalResults, nil
}

// generateSign 生成淘宝API签名
func (c *TaobaoClient) generateSign(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf []byte
	buf = append(buf, c.AppSecret...)
	for _, k := range keys {
		buf = append(buf, k...)
		buf = append(buf, params[k]...)
	}
	buf = append(buf, c.AppSecret...)

	hash := md5.Sum(buf)
	return hex.EncodeToString(hash[:])
}

func encodeParams(params map[string]string) string {
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	return values.Encode()
}