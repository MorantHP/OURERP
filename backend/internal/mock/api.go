package mock

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/MorantHP/OURERP/internal/models"
	"github.com/gin-gonic/gin"
)

// MockTaobaoHandler 模拟淘宝API处理器
type MockTaobaoHandler struct {
	generator *OrderGenerator
	orders    []*models.Order // 内存存储
}

func NewMockTaobaoHandler() *MockTaobaoHandler {
	gen := NewOrderGenerator()
	
	// 预生成100个历史订单
	orders := gen.GenerateOrders(100, 1, "taobao")
	
	return &MockTaobaoHandler{
		generator: gen,
		orders:    orders,
	}
}

// RegisterRoutes 注册模拟API路由
func (h *MockTaobaoHandler) RegisterRoutes(r *gin.RouterGroup) {
	mock := r.Group("/mock")
	{
		// 模拟淘宝API
		mock.GET("/taobao/trades/sold/get", h.GetSoldTrades)
		mock.GET("/taobao/trade/fullinfo/get", h.GetTradeFullInfo)
		
		// 模拟数据管理
		mock.POST("/generate", h.GenerateOrders)
		mock.GET("/orders", h.ListMockOrders)
		mock.POST("/realtime/start", h.StartRealtime)
	}
}

// GetSoldTrades 模拟 taobao.trades.sold.get
func (h *MockTaobaoHandler) GetSoldTrades(c *gin.Context) {
	// 解析参数
	pageNo, _ := strconv.Atoi(c.DefaultQuery("page_no", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "40"))
	
	startTime := c.Query("start_modified")
	endTime := c.Query("end_modified")
	
	fmt.Printf("[Mock] 收到订单查询请求: page=%d, start=%s, end=%s\n", 
		pageNo, startTime, endTime)
	
	// 模拟延迟
	time.Sleep(100 * time.Millisecond)
	
	// 分页
	total := len(h.orders)
	start := (pageNo - 1) * pageSize
	end := start + pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	
	pageOrders := h.orders[start:end]
	
	// 转换为淘宝格式
	trades := make([]map[string]interface{}, len(pageOrders))
	for i, order := range pageOrders {
		trades[i] = h.convertToTaobaoFormat(order)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"trades_sold_get_response": gin.H{
			"trades": gin.H{
				"trade": trades,
			},
			"total_results": total,
			"has_next": end < total,
		},
	})
}

// GetTradeFullInfo 模拟获取订单详情
func (h *MockTaobaoHandler) GetTradeFullInfo(c *gin.Context) {
	tid := c.Query("tid")
	
	for _, order := range h.orders {
		if order.PlatformOrderID == tid {
			c.JSON(http.StatusOK, gin.H{
				"trade_fullinfo_get_response": gin.H{
					"trade": h.convertToTaobaoFormat(order),
				},
			})
			return
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"error_response": gin.H{
			"code": 15,
			"msg": "Invalid session",
			"sub_code": "isv.trade-not-exist",
			"sub_msg": "订单不存在",
		},
	})
}

// GenerateOrders 手动生成订单
func (h *MockTaobaoHandler) GenerateOrders(c *gin.Context) {
	var req struct {
		Count    int    `json:"count" binding:"required,min=1,max=1000"`
		Platform string `json:"platform" binding:"required"`
		ShopID   int64  `json:"shop_id" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	newOrders := h.generator.GenerateOrders(req.Count, req.ShopID, req.Platform)
	h.orders = append(newOrders, h.orders...)
	
	c.JSON(http.StatusOK, gin.H{
		"message": "生成成功",
		"count":   len(newOrders),
		"total":   len(h.orders),
	})
}

// ListMockOrders 查看模拟订单列表
func (h *MockTaobaoHandler) ListMockOrders(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit > len(h.orders) {
		limit = len(h.orders)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"total": len(h.orders),
		"orders": h.orders[:limit],
	})
}

// StartRealtime 开始实时生成订单
func (h *MockTaobaoHandler) StartRealtime(c *gin.Context) {
	ch := h.generator.GenerateRealtimeOrders(1)
	
	// 启动goroutine接收订单
	go func() {
		for order := range ch {
			h.orders = append([]*models.Order{order}, h.orders...)
			fmt.Printf("[Mock] 新订单: %s, 买家: %s, 金额: %.2f\n", 
				order.OrderNo, order.BuyerNick, order.TotalAmount)
		}
	}()
	
	c.JSON(http.StatusOK, gin.H{"message": "实时订单生成已启动"})
}

// convertToTaobaoFormat 转换为淘宝API格式
func (h *MockTaobaoHandler) convertToTaobaoFormat(order *models.Order) map[string]interface{} {
	orders := make([]map[string]interface{}, len(order.Items))
	for i, item := range order.Items {
		orders[i] = map[string]interface{}{
			"oid":          fmt.Sprintf("%d", i+1),
			"sku_id":       fmt.Sprintf("%d", item.SkuID),
			"outer_sku_id": fmt.Sprintf("%d", item.SkuID),
			"title":        item.SkuName,
			"num":          item.Quantity,
			"price":        fmt.Sprintf("%.2f", item.Price),
			"total_fee":    fmt.Sprintf("%.2f", item.Price*float64(item.Quantity)),
		}
	}
	
	statusMap := map[int]string{
		100: "WAIT_BUYER_PAY",
		200: "WAIT_SELLER_SEND_GOODS",
		300: "WAIT_SELLER_SEND_GOODS",
		400: "WAIT_BUYER_CONFIRM_GOODS",
		500: "TRADE_BUYER_SIGNED",
		600: "TRADE_FINISHED",
		999: "TRADE_CLOSED",
	}
	
	return map[string]interface{}{
		"tid":               order.PlatformOrderID,
		"seller_nick":       "模拟店铺",
		"buyer_nick":        order.BuyerNick,
		"title":             order.Items[0].SkuName,
		"type":              "fixed",
		"status":            statusMap[order.Status],
		"receiver_name":     order.ReceiverName,
		"receiver_state":    "浙江省",
		"receiver_city":     "杭州市",
		"receiver_district": "西湖区",
		"receiver_address":  order.ReceiverAddress,
		"receiver_mobile":   order.ReceiverPhone,
		"receiver_phone":    order.ReceiverPhone,
		"created":           order.CreatedAt.Format("2006-01-02 15:04:05"),
		"pay_time":          formatTime(order.PaidAt),
		"send_time":         formatTime(order.ShippedAt),
		"consign_time":      formatTime(order.ShippedAt),
		"end_time":          "",
		"modified":          time.Now().Format("2006-01-02 15:04:05"),
		"num":               order.Items[0].Quantity,
		"price":             fmt.Sprintf("%.2f", order.Items[0].Price),
		"payment":           fmt.Sprintf("%.2f", order.PayAmount),
		"orders": map[string]interface{}{
			"order": orders,
		},
	}
}

func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}