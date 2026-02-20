package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/MorantHP/OURERP/internal/repository"
	"github.com/MorantHP/OURERP/internal/services"
	"github.com/gin-gonic/gin"
)

type StatisticsHandler struct {
	statsService *services.StatisticsService
}

func NewStatisticsHandler(statsService *services.StatisticsService) *StatisticsHandler {
	return &StatisticsHandler{statsService: statsService}
}

// parseStatsRequest 解析统计请求参数
func (h *StatisticsHandler) parseStatsRequest(c *gin.Context) *services.StatsRequest {
	req := &services.StatsRequest{}

	// 从上下文获取租户ID
	tenantID := repository.GetTenantIDFromContext(c.Request.Context())
	req.TenantID = tenantID

	// 解析日期
	if startDate := c.Query("start_date"); startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			req.StartDate = &t
		}
	}
	if endDate := c.Query("end_date"); endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			// 设置为当天的最后一秒
			endOfDay := t.Add(24*time.Hour - time.Second)
			req.EndDate = &endOfDay
		}
	}

	// 解析筛选条件
	req.Platform = c.Query("platform")
	if shopID := c.Query("shop_id"); shopID != "" {
		if id, err := strconv.ParseInt(shopID, 10, 64); err == nil {
			req.ShopID = id
		}
	}
	req.Category = c.Query("category")
	req.Brand = c.Query("brand")

	return req
}

// Overview 总览统计
func (h *StatisticsHandler) Overview(c *gin.Context) {
	req := h.parseStatsRequest(c)

	result, err := h.statsService.GetOverview(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// SalesTrend 销售趋势
func (h *StatisticsHandler) SalesTrend(c *gin.Context) {
	req := h.parseStatsRequest(c)
	period := c.DefaultQuery("period", "day") // day, week, month

	result, err := h.statsService.GetSalesTrend(c.Request.Context(), req, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ByPlatform 按平台统计
func (h *StatisticsHandler) ByPlatform(c *gin.Context) {
	req := h.parseStatsRequest(c)

	result, err := h.statsService.GetByPlatform(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ByShop 按店铺统计
func (h *StatisticsHandler) ByShop(c *gin.Context) {
	req := h.parseStatsRequest(c)

	result, err := h.statsService.GetByShop(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ByCategory 按品类统计
func (h *StatisticsHandler) ByCategory(c *gin.Context) {
	req := h.parseStatsRequest(c)

	result, err := h.statsService.GetByCategory(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ByBrand 按品牌统计
func (h *StatisticsHandler) ByBrand(c *gin.Context) {
	req := h.parseStatsRequest(c)

	result, err := h.statsService.GetByBrand(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// OrderFunnel 订单漏斗
func (h *StatisticsHandler) OrderFunnel(c *gin.Context) {
	req := h.parseStatsRequest(c)

	result, err := h.statsService.GetOrderFunnel(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// TopProducts 热销商品
func (h *StatisticsHandler) TopProducts(c *gin.Context) {
	req := h.parseStatsRequest(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	result, err := h.statsService.GetTopProducts(c.Request.Context(), req, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
