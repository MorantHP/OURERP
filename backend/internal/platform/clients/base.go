package clients

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// BaseClient 基础HTTP客户端
type BaseClient struct {
	HTTPClient *http.Client
	MaxRetries int
	RetryDelay time.Duration
	Timeout    time.Duration
}

// NewBaseClient 创建基础客户端
func NewBaseClient() *BaseClient {
	return &BaseClient{
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
			},
		},
		MaxRetries: 3,
		RetryDelay: time.Second,
		Timeout:    30 * time.Second,
	}
}

// Request 通用请求方法
func (c *BaseClient) Request(ctx context.Context, method, url string, body interface{}, headers map[string]string) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		switch v := body.(type) {
		case []byte:
			bodyReader = bytes.NewReader(v)
		case string:
			bodyReader = bytes.NewReader([]byte(v))
		default:
			jsonBody, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("序列化请求体失败: %w", err)
			}
			bodyReader = bytes.NewReader(jsonBody)
		}
	}

	var lastErr error
	for i := 0; i <= c.MaxRetries; i++ {
		if i > 0 {
			time.Sleep(c.RetryDelay * time.Duration(i))
		}

		// 如果body是[]byte或string，需要重新创建reader
		if bodyReader != nil {
			switch v := body.(type) {
			case []byte:
				bodyReader = bytes.NewReader(v)
			case string:
				bodyReader = bytes.NewReader([]byte(v))
			default:
				jsonBody, _ := json.Marshal(body)
				bodyReader = bytes.NewReader(jsonBody)
			}
		}

		req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
		if err != nil {
			return nil, err
		}

		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return respBody, nil
		}

		lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))

		// 5xx 错误重试
		if resp.StatusCode >= 500 {
			continue
		}

		// 429 限流重试
		if resp.StatusCode == 429 {
			time.Sleep(time.Second * 2)
			continue
		}

		// 其他 4xx 错误不重试
		break
	}

	return nil, lastErr
}

// Get GET请求
func (c *BaseClient) Get(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	return c.Request(ctx, http.MethodGet, url, nil, headers)
}

// Post POST请求
func (c *BaseClient) Post(ctx context.Context, url string, body interface{}, headers map[string]string) ([]byte, error) {
	if headers == nil {
		headers = make(map[string]string)
	}
	if _, ok := headers["Content-Type"]; !ok {
		headers["Content-Type"] = "application/json"
	}
	return c.Request(ctx, http.MethodPost, url, body, headers)
}

// PlatformError 平台错误
type PlatformError struct {
	Platform string `json:"platform"`
	Code     string `json:"code"`
	Message  string `json:"message"`
	SubCode  string `json:"sub_code"`
	SubMsg   string `json:"sub_msg"`
	Retry    bool   `json:"retry"`
}

func (e *PlatformError) Error() string {
	if e.SubMsg != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Platform, e.Code, e.SubMsg)
	}
	return fmt.Sprintf("[%s] %s: %s", e.Platform, e.Code, e.Message)
}

// IsRetryable 是否可重试
func (e *PlatformError) IsRetryable() bool {
	return e.Retry
}
