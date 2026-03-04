package handlers

import (
	"time"

	"github.com/MorantHP/OURERP/internal/pkg/response"
	"github.com/MorantHP/OURERP/internal/services"
	"github.com/gin-gonic/gin"
)

// ApiSyncHandler API 同步处理器
type ApiSyncHandler struct {
	syncService *services.ApiSyncService
}

// NewApiSyncHandler 创建 API 同步处理器
func NewApiSyncHandler(syncService *services.ApiSyncService) *ApiSyncHandler {
	return &ApiSyncHandler{
		syncService: syncService,
	}
}

// SyncOrdersRequest 同步订单请求（添加租户信息）
type SyncOrdersRequest struct {
	TenantCode string                   `json:"tenant_code"` // 租户编码（可选，用于指定租户）
	Orders     []services.ExternalOrder `json:"orders" binding:"required"`
	Source     string                   `json:"source"` // 数据来源标识
}

// SyncOrders 批量同步订单
// POST /api/v1/sync/orders
func (h *ApiSyncHandler) SyncOrders(c *gin.Context) {
	var req SyncOrdersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 获取租户ID
	tenantID := GetTenantIDFromGin(c)
	if tenantID == 0 {
		response.Unauthorized(c, "请先选择账套")
		return
	}

	// 构建同步请求
	syncReq := &services.SyncOrderRequest{
		Orders: req.Orders,
		Source: req.Source,
	}

	// 执行同步
	result, err := h.syncService.SyncOrders(c.Request.Context(), tenantID, syncReq)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// GetSyncStatistics 获取同步统计
// GET /api/v1/sync/statistics
func (h *ApiSyncHandler) GetSyncStatistics(c *gin.Context) {
	tenantID := GetTenantIDFromGin(c)
	if tenantID == 0 {
		response.Unauthorized(c, "请先选择账套")
		return
	}

	// 获取时间范围（默认最近7天）
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	// 解析时间参数
	var startTime, endTime time.Time
	var err error

	if startTimeStr != "" {
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			response.BadRequest(c, "开始时间格式错误")
			return
		}
	} else {
		// 默认7天前
		startTime = time.Now().AddDate(0, 0, -7)
	}

	if endTimeStr != "" {
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			response.BadRequest(c, "结束时间格式错误")
			return
		}
	} else {
		endTime = time.Now()
	}

	result, err := h.syncService.GetSyncStatistics(c.Request.Context(), tenantID, startTime, endTime)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}
