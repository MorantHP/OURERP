package handlers

import (
	"net/http"
	"strconv"

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

// ListOrders 订单列表
func (h *OrderHandler) ListOrders(c *gin.Context) {
	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	
	// 查询条件
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

// GetOrder 订单详情
func (h *OrderHandler) GetOrder(c *gin.Context) {
	id := c.Param("id")
	order, err := h.orderRepo.FindByOrderNo(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"order": order})
}

// CreateOrder 创建订单
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	order := &models.Order{
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

// AuditOrder 审核订单
func (h *OrderHandler) AuditOrder(c *gin.Context) {
	id := c.Param("id")
	
	if err := h.orderRepo.UpdateStatus(id, models.OrderStatusPendingShip); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "审核失败"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "审核成功"})
}

// ShipOrder 发货
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
type CreateOrderRequest struct {
	Platform        string                `json:"platform" binding:"required"`
	PlatformOrderID string                `json:"platform_order_id"`
	ShopID          int64                 `json:"shop_id"`
	TotalAmount     float64               `json:"total_amount" binding:"required"`
	PayAmount       float64               `json:"pay_amount"`
	BuyerNick       string                `json:"buyer_nick"`
	ReceiverName    string                `json:"receiver_name" binding:"required"`
	ReceiverPhone   string                `json:"receiver_phone"`
	ReceiverAddress string                `json:"receiver_address"`
	Items           []CreateOrderItemRequest `json:"items" binding:"required,min=1"`
}

type CreateOrderItemRequest struct {
	SkuID    int64   `json:"sku_id" binding:"required"`
	SkuName  string  `json:"sku_name" binding:"required"`
	Quantity int     `json:"quantity" binding:"required,min=1"`
	Price    float64 `json:"price" binding:"required"`
}