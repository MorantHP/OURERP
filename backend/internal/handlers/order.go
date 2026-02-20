package handlers

import (
	"net/http"
	"strconv"

	"github.com/MorantHP/OURERP/internal/middleware"
	"github.com/MorantHP/OURERP/internal/models"
	"github.com/MorantHP/OURERP/internal/repository"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderRepo *repository.OrderRepository
}

func NewOrderHandler(orderRepo *repository.OrderRepository) *OrderHandler {
	return &OrderHandler{orderRepo: orderRepo}
}

func (h *OrderHandler) ListOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	status := c.Query("status")
	platform := c.Query("platform")
	keyword := c.Query("keyword")

	orders, total, err := h.orderRepo.List(page, size, status, platform, keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list": orders,
		"pagination": gin.H{
			"total":       total,
			"page":        page,
			"size":        size,
			"total_pages": (total + int64(size) - 1) / int64(size),
		},
	})
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	id := c.Param("id")
	order, err := h.orderRepo.FindByOrderNo(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"order": order})
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	tenantID := middleware.GetTenantIDFromGin(c)
	if tenantID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order := &models.Order{
		TenantID:        tenantID,
		OrderNo:         models.GenerateOrderNo(),
		Platform:        req.Platform,
		PlatformOrderID: req.PlatformOrderID,
		ShopID:          req.ShopID,
		Status:          models.OrderStatusPendingPayment,
		TotalAmount:     req.TotalAmount,
		PayAmount:       req.PayAmount,
		BuyerNick:       req.BuyerNick,
		ReceiverName:    req.ReceiverName,
		ReceiverPhone:   req.ReceiverPhone,
		ReceiverAddress: req.ReceiverAddress,
		Items:           make([]models.OrderItem, len(req.Items)),
	}

	for i, item := range req.Items {
		order.Items[i] = models.OrderItem{
			TenantID: tenantID,
			SkuID:    item.SkuID,
			SkuName:  item.SkuName,
			Quantity: item.Quantity,
			Price:    item.Price,
		}
	}

	if err := h.orderRepo.Create(order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建订单失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"order": order})
}

func (h *OrderHandler) AuditOrder(c *gin.Context) {
	id := c.Param("id")

	if err := h.orderRepo.UpdateStatus(id, models.OrderStatusPendingShip); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "审核失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "审核成功"})
}

func (h *OrderHandler) ShipOrder(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		LogisticsCompany string `json:"logistics_company"`
		LogisticsNo      string `json:"logistics_no"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.orderRepo.Ship(id, req.LogisticsCompany, req.LogisticsNo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发货失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "发货成功"})
}
