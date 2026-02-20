package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"

	"github.com/MorantHP/OURERP/internal/models"
	"github.com/MorantHP/OURERP/internal/platform"
	"github.com/MorantHP/OURERP/internal/platform/clients"
	"github.com/MorantHP/OURERP/internal/repository"
	"github.com/gin-gonic/gin"
)

type OAuthHandler struct {
	shopRepo *repository.ShopRepository
	baseURL  string
}

func NewOAuthHandler(shopRepo *repository.ShopRepository, baseURL string) *OAuthHandler {
	return &OAuthHandler{
		shopRepo: shopRepo,
		baseURL:  baseURL,
	}
}

// state存储（生产环境应使用Redis）
var stateStore = make(map[string]int64)

// GetAuthURL 获取授权URL
// GET /api/v1/oauth/auth-url?shop_id=123
func (h *OAuthHandler) GetAuthURL(c *gin.Context) {
	shopIDStr := c.Query("shop_id")
	if shopIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少shop_id参数"})
		return
	}

	shopID, err := strconv.ParseInt(shopIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的shop_id"})
		return
	}

	shop, err := h.shopRepo.FindByID(shopID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "店铺不存在"})
		return
	}

	// 生成state（防止CSRF）
	state := generateState()
	stateStore[state] = shopID

	// 构建回调URL
	redirectURI := fmt.Sprintf("%s/api/v1/oauth/callback", h.baseURL)

	// 获取授权URL
	authURL := getAuthURL(shop.Platform, shop.AppKey, redirectURI, state)

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
		"state":    state,
	})
}

// Callback OAuth回调
// GET /api/v1/oauth/callback?code=xxx&state=xxx
func (h *OAuthHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		c.Redirect(http.StatusFound, "/shops?error=invalid_params")
		return
	}

	// 验证state
	shopID, ok := stateStore[state]
	if !ok {
		c.Redirect(http.StatusFound, "/shops?error=invalid_state")
		return
	}
	delete(stateStore, state)

	// 获取店铺信息
	shop, err := h.shopRepo.FindByID(shopID)
	if err != nil {
		c.Redirect(http.StatusFound, "/shops?error=shop_not_found")
		return
	}

	// 构建回调URL
	redirectURI := fmt.Sprintf("%s/api/v1/oauth/callback", h.baseURL)

	// 用授权码换取Token
	token, err := exchangeToken(c.Request.Context(), shop, code, redirectURI)
	if err != nil {
		c.Redirect(http.StatusFound, fmt.Sprintf("/shops?error=token_failed&msg=%s", err.Error()))
		return
	}

	// 更新店铺Token
	shop.AccessToken = token.AccessToken
	shop.RefreshToken = token.RefreshToken
	shop.TokenExpiresAt = &token.ExpiresAt
	shop.Status = models.ShopStatusEnabled

	if err := h.shopRepo.Update(shop); err != nil {
		c.Redirect(http.StatusFound, "/shops?error=save_failed")
		return
	}

	// 授权成功，重定向到前端
	c.Redirect(http.StatusFound, "/shops?success=authorized")
}

// RefreshToken 刷新Token
// POST /api/v1/oauth/refresh?shop_id=123
func (h *OAuthHandler) RefreshToken(c *gin.Context) {
	shopIDStr := c.Query("shop_id")
	if shopIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少shop_id参数"})
		return
	}

	shopID, err := strconv.ParseInt(shopIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的shop_id"})
		return
	}

	shop, err := h.shopRepo.FindByID(shopID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "店铺不存在"})
		return
	}

	if shop.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无刷新令牌，请重新授权"})
		return
	}

	// 刷新Token
	token, err := refreshToken(c.Request.Context(), shop)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("刷新失败: %s", err.Error())})
		return
	}

	// 更新店铺Token
	if err := h.shopRepo.UpdateToken(shopID, token.AccessToken, token.RefreshToken, token.ExpiresAt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "刷新成功",
		"expires_at": token.ExpiresAt,
	})
}

// getAuthURL 根据平台获取授权URL
func getAuthURL(platformType, appKey, redirectURI, state string) string {
	switch platformType {
	case string(platform.PlatformTaobao), string(platform.PlatformTmall):
		return fmt.Sprintf(
			"https://oauth.taobao.com/authorize?response_type=code&client_id=%s&redirect_uri=%s&state=%s",
			appKey, redirectURI, state,
		)
	case string(platform.PlatformDouyin):
		return fmt.Sprintf(
			"https://developer.toutiao.com/api/oauth/connect/?app_id=%s&scope=user_info,order&redirect_uri=%s&state=%s",
			appKey, redirectURI, state,
		)
	case string(platform.PlatformKuaishou):
		return fmt.Sprintf(
			"https://s.kwaixiaodian.com/oauth/authorize?app_id=%s&scope=user_info,order&redirect_uri=%s&state=%s",
			appKey, redirectURI, state,
		)
	case string(platform.PlatformWechatVideo):
		return fmt.Sprintf(
			"https://open.weixin.qq.com/connect/oauth2/authorize?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_base&state=%s#wechat_redirect",
			appKey, redirectURI, state,
		)
	default:
		return ""
	}
}

// exchangeToken 用授权码换取Token
func exchangeToken(ctx context.Context, shop *models.Shop, code, redirectURI string) (*platform.AuthToken, error) {
	switch shop.Platform {
	case string(platform.PlatformTaobao), string(platform.PlatformTmall):
		client := clients.NewTaobaoClient(shop.AppKey, shop.AppSecret, "")
		return client.ExchangeToken(ctx, code, redirectURI)
	case string(platform.PlatformDouyin):
		client := clients.NewDouyinClient(shop.AppKey, shop.AppSecret, "", shop.PlatformShopID)
		return client.ExchangeToken(ctx, code)
	case string(platform.PlatformKuaishou):
		client := clients.NewKuaishouClient(shop.AppKey, shop.AppSecret, "", shop.PlatformShopID)
		return client.ExchangeToken(ctx, code)
	case string(platform.PlatformWechatVideo):
		client := clients.NewWechatVideoClient(shop.AppKey, shop.AppSecret, "")
		return client.ExchangeToken(ctx, code)
	default:
		return nil, fmt.Errorf("不支持的平台: %s", shop.Platform)
	}
}

// refreshToken 刷新Token
func refreshToken(ctx context.Context, shop *models.Shop) (*platform.AuthToken, error) {
	switch shop.Platform {
	case string(platform.PlatformTaobao), string(platform.PlatformTmall):
		client := clients.NewTaobaoClient(shop.AppKey, shop.AppSecret, "")
		return client.RefreshToken(ctx, shop.RefreshToken)
	case string(platform.PlatformDouyin):
		client := clients.NewDouyinClient(shop.AppKey, shop.AppSecret, "", shop.PlatformShopID)
		return client.RefreshToken(ctx, shop.RefreshToken)
	case string(platform.PlatformKuaishou):
		client := clients.NewKuaishouClient(shop.AppKey, shop.AppSecret, "", shop.PlatformShopID)
		return client.RefreshToken(ctx, shop.RefreshToken)
	case string(platform.PlatformWechatVideo):
		client := clients.NewWechatVideoClient(shop.AppKey, shop.AppSecret, "")
		return client.RefreshToken(ctx, shop.RefreshToken)
	default:
		return nil, fmt.Errorf("不支持的平台: %s", shop.Platform)
	}
}

// generateState 生成随机state
func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
