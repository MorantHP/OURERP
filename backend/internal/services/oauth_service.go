package services

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/MorantHP/OURERP/internal/models"
	"github.com/MorantHP/OURERP/internal/platform"
	"github.com/MorantHP/OURERP/internal/repository"
)

// OAuthConfig OAuth配置
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	AuthURL      string
	TokenURL     string
	Scope        string
}

// OAuthToken OAuth令牌响应
type OAuthToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
}

// OAuthService OAuth服务
type OAuthService struct {
	shopRepo  *repository.ShopRepository
	configs   map[platform.PlatformType]*OAuthConfig
	httpClient *http.Client
	baseURL   string
}

// NewOAuthService 创建OAuth服务
func NewOAuthService(shopRepo *repository.ShopRepository, baseURL string) *OAuthService {
	// 创建HTTP客户端（跳过SSL验证，用于开发环境）
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{
		Transport: tr,
		Timeout:   30 * time.Second,
	}

	service := &OAuthService{
		shopRepo:   shopRepo,
		httpClient: httpClient,
		baseURL:    baseURL,
		configs:    make(map[platform.PlatformType]*OAuthConfig),
	}

	// 初始化各平台配置
	service.initConfigs()

	return service
}

// initConfigs 初始化各平台的OAuth配置
func (s *OAuthService) initConfigs() {
	// 淘宝/天猫
	s.configs[platform.PlatformTaobao] = &OAuthConfig{
		AuthURL:     "https://oauth.taobao.com/authorize",
		TokenURL:    "https://oauth.taobao.com/token",
		RedirectURI: s.baseURL + "/api/v1/oauth/callback",
		Scope:       "item order",
	}
	s.configs[platform.PlatformTmall] = s.configs[platform.PlatformTaobao]

	// 抖音
	s.configs[platform.PlatformDouyin] = &OAuthConfig{
		AuthURL:     "https://developer.toutiao.com/oauth/authorize",
		TokenURL:    "https://developer.toutiao.com/oauth/access_token",
		RedirectURI: s.baseURL + "/api/v1/oauth/callback",
		Scope:       "order.item",
	}

	// 快手
	s.configs[platform.PlatformKuaishou] = &OAuthConfig{
		AuthURL:     "https://open.kuaishou.com/oauth2/authorize",
		TokenURL:    "https://open.kuaishou.com/oauth2/access_token",
		RedirectURI: s.baseURL + "/api/v1/oauth/callback",
		Scope:       "order",
	}

	// 微信视频号
	s.configs[platform.PlatformWechatVideo] = &OAuthConfig{
		AuthURL:     "https://open.weixin.qq.com/connect/oauth2/authorize",
		TokenURL:    "https://api.weixin.qq.com/sns/oauth2/access_token",
		RedirectURI: s.baseURL + "/api/v1/oauth/callback",
		Scope:       "snsapi_base",
	}
}

// GetAuthURL 获取授权URL
func (s *OAuthService) GetAuthURL(shopID int64, state string) (string, error) {
	shop, err := s.shopRepo.FindByID(shopID)
	if err != nil {
		return "", fmt.Errorf("店铺不存在: %w", err)
	}

	platformType := platform.PlatformType(shop.Platform)
	config, ok := s.configs[platformType]
	if !ok {
		return "", fmt.Errorf("不支持的平台: %s", shop.Platform)
	}

	// 构建授权URL
	params := url.Values{}
	params.Set("client_id", shop.AppKey)
	params.Set("redirect_uri", config.RedirectURI)
	params.Set("response_type", "code")
	params.Set("state", state)
	if config.Scope != "" {
		params.Set("scope", config.Scope)
	}

	authURL := fmt.Sprintf("%s?%s", config.AuthURL, params.Encode())

	// 微信视频号特殊处理
	if platformType == platform.PlatformWechatVideo {
		authURL += "#wechat_redirect"
	}

	return authURL, nil
}

// HandleCallback 处理OAuth回调
func (s *OAuthService) HandleCallback(ctx context.Context, shopID int64, code string) (*models.Shop, error) {
	shop, err := s.shopRepo.FindByID(shopID)
	if err != nil {
		return nil, fmt.Errorf("店铺不存在: %w", err)
	}

	platformType := platform.PlatformType(shop.Platform)
	config, ok := s.configs[platformType]
	if !ok {
		return nil, fmt.Errorf("不支持的平台: %s", shop.Platform)
	}

	// 请求访问令牌
	token, err := s.requestToken(config, shop.AppKey, shop.AppSecret, code)
	if err != nil {
		return nil, fmt.Errorf("获取令牌失败: %w", err)
	}

	// 更新店铺令牌信息
	now := time.Now()
	expiresAt := now.Add(time.Duration(token.ExpiresIn) * time.Second)

	shop.AccessToken = token.AccessToken
	shop.RefreshToken = token.RefreshToken
	shop.TokenExpiresAt = &expiresAt
	shop.Status = models.ShopStatusEnabled

	if err := s.shopRepo.Update(shop); err != nil {
		return nil, fmt.Errorf("更新店铺失败: %w", err)
	}

	return shop, nil
}

// RefreshToken 刷新访问令牌
func (s *OAuthService) RefreshToken(ctx context.Context, shopID int64) (*models.Shop, error) {
	shop, err := s.shopRepo.FindByID(shopID)
	if err != nil {
		return nil, fmt.Errorf("店铺不存在: %w", err)
	}

	if shop.RefreshToken == "" {
		return nil, fmt.Errorf("店铺没有刷新令牌")
	}

	platformType := platform.PlatformType(shop.Platform)
	config, ok := s.configs[platformType]
	if !ok {
		return nil, fmt.Errorf("不支持的平台: %s", shop.Platform)
	}

	// 使用刷新令牌获取新的访问令牌
	token, err := s.refreshAccessToken(config, shop.AppKey, shop.AppSecret, shop.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("刷新令牌失败: %w", err)
	}

	// 更新店铺令牌信息
	now := time.Now()
	expiresAt := now.Add(time.Duration(token.ExpiresIn) * time.Second)

	shop.AccessToken = token.AccessToken
	if token.RefreshToken != "" {
		shop.RefreshToken = token.RefreshToken
	}
	shop.TokenExpiresAt = &expiresAt

	if err := s.shopRepo.Update(shop); err != nil {
		return nil, fmt.Errorf("更新店铺失败: %w", err)
	}

	return shop, nil
}

// requestToken 请求访问令牌
func (s *OAuthService) requestToken(config *OAuthConfig, appKey, appSecret, code string) (*OAuthToken, error) {
	params := url.Values{}
	params.Set("grant_type", "authorization_code")
	params.Set("client_id", appKey)
	params.Set("client_secret", appSecret)
	params.Set("code", code)
	params.Set("redirect_uri", config.RedirectURI)

	return s.doTokenRequest(config.TokenURL, params)
}

// refreshAccessToken 刷新访问令牌
func (s *OAuthService) refreshAccessToken(config *OAuthConfig, appKey, appSecret, refreshToken string) (*OAuthToken, error) {
	params := url.Values{}
	params.Set("grant_type", "refresh_token")
	params.Set("client_id", appKey)
	params.Set("client_secret", appSecret)
	params.Set("refresh_token", refreshToken)

	return s.doTokenRequest(config.TokenURL, params)
}

// doTokenRequest 执行令牌请求
func (s *OAuthService) doTokenRequest(tokenURL string, params url.Values) (*OAuthToken, error) {
	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("令牌请求失败: %s", string(body))
	}

	var token OAuthToken
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("解析令牌响应失败: %w", err)
	}

	if token.AccessToken == "" {
		return nil, fmt.Errorf("响应中没有访问令牌: %s", string(body))
	}

	return &token, nil
}

// CheckTokenExpiry 检查令牌是否即将过期
func (s *OAuthService) CheckTokenExpiry(shop *models.Shop) bool {
	if shop.TokenExpiresAt == nil {
		return shop.AccessToken == ""
	}
	// 提前5分钟判断过期
	return time.Now().Add(5 * time.Minute).After(*shop.TokenExpiresAt)
}

// IsTokenExpired 检查令牌是否已过期
func (s *OAuthService) IsTokenExpired(shop *models.Shop) bool {
	if shop.TokenExpiresAt == nil {
		return shop.AccessToken == ""
	}
	return time.Now().After(*shop.TokenExpiresAt)
}
