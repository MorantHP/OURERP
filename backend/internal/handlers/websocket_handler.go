package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/MorantHP/OURERP/internal/middleware"
	"github.com/MorantHP/OURERP/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// 开发环境允许所有来源
		// 生产环境应该检查Origin
		return true
	},
}

// WebSocketHandler WebSocket处理器
type WebSocketHandler struct {
	hub *services.WebSocketHub
}

// NewWebSocketHandler 创建WebSocket处理器
func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{
		hub: services.GetWebSocketHub(),
	}
}

// HandleWebSocket 处理WebSocket连接
// GET /api/v1/ws
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// 获取用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	// 获取租户ID
	tenantID := middleware.GetTenantIDFromGin(c)

	// 升级HTTP连接为WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	// 创建客户端
	clientID := fmt.Sprintf("%d-%d-%d", userID.(int64), tenantID, time.Now().UnixNano())
	client := &services.Client{
		ID:       clientID,
		TenantID: tenantID,
		UserID:   userID.(int64),
		Conn:     conn,
		Send:     make(chan *services.WebSocketMessage, 256),
		Hub:      h.hub,
	}

	// 注册客户端
	h.hub.RegisterClient(client)

	// 发送欢迎消息
	client.Send <- &services.WebSocketMessage{
		Type:      "connected",
		TenantID:  tenantID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"client_id":    clientID,
			"message":      "WebSocket连接成功",
			"tenant_id":    tenantID,
			"connected_at": time.Now(),
		},
	}

	// 启动读写协程
	go client.WritePump()
	go client.ReadPump()
}

// GetStats 获取WebSocket统计信息
// GET /api/v1/ws/stats
func (h *WebSocketHandler) GetStats(c *gin.Context) {
	tenantID := middleware.GetTenantIDFromGin(c)

	stats := map[string]interface{}{
		"total_connections": h.hub.GetClientCount(),
	}

	if tenantID > 0 {
		stats["tenant_connections"] = h.hub.GetTenantClientCount(tenantID)
	}

	c.JSON(http.StatusOK, stats)
}
