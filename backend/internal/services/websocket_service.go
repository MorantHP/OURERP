package services

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocket消息类型
const (
	MsgTypeOrderNew       = "order_new"
	MsgTypeOrderUpdate    = "order_update"
	MsgTypeInventoryAlert = "inventory_alert"
	MsgTypeSyncStatus     = "sync_status"
	MsgTypeNotification   = "notification"
	MsgTypeHeartbeat      = "heartbeat"
)

// WebSocketMessage WebSocket消息结构
type WebSocketMessage struct {
	Type      string      `json:"type"`
	TenantID  int64       `json:"tenant_id"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// Client WebSocket客户端
type Client struct {
	ID       string
	TenantID int64
	UserID   int64
	Conn     *websocket.Conn
	Send     chan *WebSocketMessage
	Hub      *WebSocketHub
}

// WebSocketHub WebSocket连接管理中心
type WebSocketHub struct {
	clients    map[string]*Client
	tenantMap  map[int64]map[string]*Client // 按租户分组
	register   chan *Client
	unregister chan *Client
	broadcast  chan *WebSocketMessage
	mu         sync.RWMutex
}

// NewWebSocketHub 创建WebSocket Hub
func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[string]*Client),
		tenantMap:  make(map[int64]map[string]*Client),
		register:   make(chan *Client, 256),
		unregister: make(chan *Client, 256),
		broadcast:  make(chan *WebSocketMessage, 1024),
	}
}

// Run 启动Hub
func (h *WebSocketHub) Run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)

		case <-ticker.C:
			h.heartbeat()
		}
	}
}

// registerClient 注册客户端
func (h *WebSocketHub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client.ID] = client

	// 添加到租户组
	if h.tenantMap[client.TenantID] == nil {
		h.tenantMap[client.TenantID] = make(map[string]*Client)
	}
	h.tenantMap[client.TenantID][client.ID] = client

	log.Printf("WebSocket客户端已连接: %s (租户: %d, 用户: %d), 当前连接数: %d",
		client.ID, client.TenantID, client.UserID, len(h.clients))
}

// RegisterClient 注册客户端（公开方法）
func (h *WebSocketHub) RegisterClient(client *Client) {
	h.register <- client
}

// UnregisterClient 注销客户端（公开方法）
func (h *WebSocketHub) UnregisterClient(client *Client) {
	h.unregister <- client
}

// unregisterClient 注销客户端
func (h *WebSocketHub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client.ID]; ok {
		delete(h.clients, client.ID)

		// 从租户组移除
		if tenantClients, ok := h.tenantMap[client.TenantID]; ok {
			delete(tenantClients, client.ID)
			if len(tenantClients) == 0 {
				delete(h.tenantMap, client.TenantID)
			}
		}

		close(client.Send)
		log.Printf("WebSocket客户端已断开: %s, 当前连接数: %d", client.ID, len(h.clients))
	}
}

// broadcastMessage 广播消息
func (h *WebSocketHub) broadcastMessage(message *WebSocketMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// 如果指定了租户，只发送给该租户的客户端
	if message.TenantID > 0 {
		if tenantClients, ok := h.tenantMap[message.TenantID]; ok {
			for _, client := range tenantClients {
				select {
				case client.Send <- message:
				default:
					// 发送通道已满，关闭连接
					close(client.Send)
					delete(tenantClients, client.ID)
				}
			}
		}
		return
	}

	// 广播给所有客户端
	for _, client := range h.clients {
		select {
		case client.Send <- message:
		default:
			// 发送通道已满，跳过
		}
	}
}

// heartbeat 心跳检测
func (h *WebSocketHub) heartbeat() {
	h.mu.RLock()
	defer h.mu.RUnlock()

	msg := &WebSocketMessage{
		Type:      MsgTypeHeartbeat,
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"ping": "pong"},
	}

	for _, client := range h.clients {
		select {
		case client.Send <- msg:
		default:
			// 连接可能已断开
		}
	}
}

// BroadcastToTenant 向指定租户广播消息
func (h *WebSocketHub) BroadcastToTenant(tenantID int64, msgType string, data interface{}) {
	message := &WebSocketMessage{
		Type:      msgType,
		TenantID:  tenantID,
		Timestamp: time.Now(),
		Data:      data,
	}
	h.broadcast <- message
}

// BroadcastToAll 向所有客户端广播消息
func (h *WebSocketHub) BroadcastToAll(msgType string, data interface{}) {
	message := &WebSocketMessage{
		Type:      msgType,
		Timestamp: time.Now(),
		Data:      data,
	}
	h.broadcast <- message
}

// GetClientCount 获取连接数
func (h *WebSocketHub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetTenantClientCount 获取租户连接数
func (h *WebSocketHub) GetTenantClientCount(tenantID int64) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if tenantClients, ok := h.tenantMap[tenantID]; ok {
		return len(tenantClients)
	}
	return 0
}

// ReadPump 读取消息（客户端->服务端）
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		// 处理客户端消息
		var msg WebSocketMessage
		if err := json.Unmarshal(message, &msg); err == nil {
			// 处理心跳响应
			if msg.Type == MsgTypeHeartbeat {
				c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			}
		}
	}
}

// WritePump 写入消息（服务端->客户端）
func (c *Client) WritePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// 全局WebSocket Hub
var wsHub *WebSocketHub
var wsHubOnce sync.Once

// GetWebSocketHub 获取全局WebSocket Hub
func GetWebSocketHub() *WebSocketHub {
	wsHubOnce.Do(func() {
		wsHub = NewWebSocketHub()
		go wsHub.Run()
	})
	return wsHub
}
