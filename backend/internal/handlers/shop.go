package handlers

import (
	"net/http"
	"strconv"

	"github.com/MorantHP/OURERP/internal/middleware"
	"github.com/MorantHP/OURERP/internal/models"
	"github.com/MorantHP/OURERP/internal/platform"
	"github.com/MorantHP/OURERP/internal/repository"
	"github.com/gin-gonic/gin"
)

type ShopHandler struct {
	shopRepo *repository.ShopRepository
}

func NewShopHandler(shopRepo *repository.ShopRepository) *ShopHandler {
	return &ShopHandler{shopRepo: shopRepo}
}

// List 店铺列表
// GET /api/v1/shops
func (h *ShopHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	platform := c.Query("platform")
	statusStr := c.Query("status")

	var status *int
	if statusStr != "" {
		s, _ := strconv.Atoi(statusStr)
		status = &s
	}

	shops, total, err := h.shopRepo.List(page, size, platform, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list": shops,
		"pagination": gin.H{
			"page":  page,
			"size": size,
			"total": total,
		},
	})
}

// Get 店铺详情
// GET /api/v1/shops/:id
func (h *ShopHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	shop, err := h.shopRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "店铺不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"shop": shop})
}

// Create 创建店铺
// POST /api/v1/shops
func (h *ShopHandler) Create(c *gin.Context) {
	tenantID := middleware.GetTenantIDFromGin(c)
	if tenantID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req struct {
		Name           string `json:"name" binding:"required"`
		Platform       string `json:"platform" binding:"required"`
		PlatformShopID string `json:"platform_shop_id"`
		AppKey         string `json:"app_key"`
		AppSecret      string `json:"app_secret"`
		APIURL         string `json:"api_url"`
		WebhookURL     string `json:"webhook_url"`
		SyncInterval   int    `json:"sync_interval"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证平台是否支持
	if _, ok := platform.GetPlatformConfig(platform.PlatformType(req.Platform)); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的平台类型"})
		return
	}

	if req.SyncInterval == 0 {
		req.SyncInterval = 30
	}

	shop := &models.Shop{
		TenantID:       tenantID,
		Name:           req.Name,
		Platform:       req.Platform,
		PlatformShopID: req.PlatformShopID,
		AppKey:         req.AppKey,
		AppSecret:      req.AppSecret,
		APIURL:         req.APIURL,
		WebhookURL:     req.WebhookURL,
		SyncInterval:   req.SyncInterval,
		Status:         models.ShopStatusEnabled,
	}

	if err := h.shopRepo.Create(shop); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "创建成功",
		"shop":    shop,
	})
}

// Update 更新店铺
// PUT /api/v1/shops/:id
func (h *ShopHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	shop, err := h.shopRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "店铺不存在"})
		return
	}

	var req struct {
		Name           string `json:"name"`
		PlatformShopID string `json:"platform_shop_id"`
		AppKey         string `json:"app_key"`
		AppSecret      string `json:"app_secret"`
		APIURL         string `json:"api_url"`
		WebhookURL     string `json:"webhook_url"`
		SyncInterval   int    `json:"sync_interval"`
		Status         *int   `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != "" {
		shop.Name = req.Name
	}
	if req.PlatformShopID != "" {
		shop.PlatformShopID = req.PlatformShopID
	}
	if req.AppKey != "" {
		shop.AppKey = req.AppKey
	}
	if req.AppSecret != "" {
		shop.AppSecret = req.AppSecret
	}
	if req.APIURL != "" {
		shop.APIURL = req.APIURL
	}
	if req.WebhookURL != "" {
		shop.WebhookURL = req.WebhookURL
	}
	if req.SyncInterval > 0 {
		shop.SyncInterval = req.SyncInterval
	}
	if req.Status != nil {
		shop.Status = *req.Status
	}

	if err := h.shopRepo.Update(shop); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
		"shop":    shop,
	})
}

// Delete 删除店铺
// DELETE /api/v1/shops/:id
func (h *ShopHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if err := h.shopRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// TriggerSync 手动触发同步
// POST /api/v1/shops/:id/sync
func (h *ShopHandler) TriggerSync(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	shop, err := h.shopRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "店铺不存在"})
		return
	}

	if shop.Status != models.ShopStatusEnabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "店铺已禁用"})
		return
	}

	// TODO: 触发同步任务（需要集成SyncService）
	// 这里应该调用SyncService来执行同步

	c.JSON(http.StatusOK, gin.H{
		"message": "同步任务已启动",
		"shop_id": id,
	})
}

// GetAuthURL 获取授权URL
// GET /api/v1/shops/:id/auth-url
func (h *ShopHandler) GetAuthURL(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	shop, err := h.shopRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "店铺不存在"})
		return
	}

	// 根据平台获取授权URL
	// TODO: 集成OAuth服务

	c.JSON(http.StatusOK, gin.H{
		"auth_url":   "",
		"shop_id":    id,
		"platform":   shop.Platform,
	})
}
