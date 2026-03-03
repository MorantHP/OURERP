package handlers

import (
	"strconv"

	"github.com/MorantHP/OURERP/backend/internal/models"
	"github.com/MorantHP/OURERP/backend/internal/pkg/response"
	"github.com/MorantHP/OURERP/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// OrderHandlerV2 订单处理器（重构版）
type OrderHandlerV2 struct {
	orderService *services.OrderService
}

// NewOrderHandlerV2 创建订单处理器
func NewOrderHandlerV2(orderService *services.OrderService) *OrderHandlerV2 {
	return &OrderHandlerV2{
		orderService: orderService,
	}
}

// List 订单列表
// GET /api/v1/orders
func (h *OrderHandlerV2) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	status := c.Query("status")
	platform := c.Query("platform")
	keyword := c.Query("keyword")

	ctx := c.Request.Context()
	orders, total, err := h.orderService.ListOrders(ctx, page, size, status, platform, keyword)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessPage(c, orders, page, size, total)
}

// Get 订单详情
// GET /api/v1/orders/:order_no
func (h *OrderHandlerV2) Get(c *gin.Context) {
	orderNo := c.Param("order_no")

	ctx := c.Request.Context()
	order, err := h.orderService.GetOrder(ctx, orderNo)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, order)
}

// Create 创建订单
// POST /api/v1/orders
func (h *OrderHandlerV2) Create(c *gin.Context) {
	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	ctx := c.Request.Context()
	order, err := h.orderService.CreateOrder(ctx, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, order)
}

// Audit 审核订单
// POST /api/v1/orders/:order_no/audit
func (h *OrderHandlerV2) Audit(c *gin.Context) {
	orderNo := c.Param("order_no")

	ctx := c.Request.Context()
	if err := h.orderService.AuditOrder(ctx, orderNo); err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithMessage(c, "审核成功", nil)
}

// Ship 发货
// POST /api/v1/orders/:order_no/ship
func (h *OrderHandlerV2) Ship(c *gin.Context) {
	orderNo := c.Param("order_no")

	var req struct {
		LogisticsCompany string `json:"logistics_company" binding:"required"`
		LogisticsNo      string `json:"logistics_no" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	ctx := c.Request.Context()
	if err := h.orderService.ShipOrder(ctx, orderNo, req.LogisticsCompany, req.LogisticsNo); err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithMessage(c, "发货成功", nil)
}

// Cancel 取消订单
// POST /api/v1/orders/:order_no/cancel
func (h *OrderHandlerV2) Cancel(c *gin.Context) {
	orderNo := c.Param("order_no")

	var req struct {
		Reason string `json:"reason"`
	}

	_ = c.ShouldBindJSON(&req) // reason 可选

	ctx := c.Request.Context()
	if err := h.orderService.CancelOrder(ctx, orderNo, req.Reason); err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithMessage(c, "取消成功", nil)
}

// Statistics 订单统计
// GET /api/v1/orders/statistics
func (h *OrderHandlerV2) Statistics(c *gin.Context) {
	ctx := c.Request.Context()
	stats, err := h.orderService.GetOrderStatistics(ctx)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, stats)
}
