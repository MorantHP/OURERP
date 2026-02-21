// WebSocket服务 - 实时数据推送
import { useUserStore } from '@/stores/user'
import { useTenantStore } from '@/stores/tenant'

export interface WebSocketMessage {
  type: string
  tenant_id: number
  timestamp: string
  data: any
}

export type MessageHandler = (message: WebSocketMessage) => void

class WebSocketService {
  private ws: WebSocket | null = null
  private url: string
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private reconnectDelay = 3000
  private handlers: Map<string, Set<MessageHandler>> = new Map()
  private isConnected = false
  private pingInterval: number | null = null

  constructor() {
    // 根据环境配置WebSocket地址
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = import.meta.env.VITE_WS_HOST || window.location.host
    this.url = `${protocol}//${host}/api/v1/ws`
  }

  // 连接WebSocket
  connect(): Promise<boolean> {
    return new Promise((resolve) => {
      const userStore = useUserStore()
      const tenantStore = useTenantStore()

      if (!userStore.token) {
        resolve(false)
        return
      }

      // 构建WebSocket URL，带上token和tenant_id
      const wsUrl = new URL(this.url)
      wsUrl.searchParams.set('token', userStore.token)
      if (tenantStore.currentTenantId) {
        wsUrl.searchParams.set('tenant_id', String(tenantStore.currentTenantId))
      }

      try {
        this.ws = new WebSocket(wsUrl.toString())

        this.ws.onopen = () => {
          console.log('WebSocket连接成功')
          this.isConnected = true
          this.reconnectAttempts = 0
          this.startPing()
          resolve(true)
        }

        this.ws.onmessage = (event) => {
          try {
            const message: WebSocketMessage = JSON.parse(event.data)
            this.handleMessage(message)
          } catch (e) {
            console.error('解析WebSocket消息失败:', e)
          }
        }

        this.ws.onerror = (error) => {
          console.error('WebSocket错误:', error)
        }

        this.ws.onclose = () => {
          console.log('WebSocket连接关闭')
          this.isConnected = false
          this.stopPing()
          this.attemptReconnect()
        }
      } catch (error) {
        console.error('WebSocket连接失败:', error)
        resolve(false)
      }
    })
  }

  // 断开连接
  disconnect() {
    this.stopPing()
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    this.isConnected = false
  }

  // 尝试重连
  private attemptReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.log('WebSocket重连次数已达上限')
      return
    }

    this.reconnectAttempts++
    console.log(`尝试WebSocket重连 (${this.reconnectAttempts}/${this.maxReconnectAttempts})`)

    setTimeout(() => {
      this.connect()
    }, this.reconnectDelay * this.reconnectAttempts)
  }

  // 开始心跳
  private startPing() {
    this.pingInterval = window.setInterval(() => {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        this.send({ type: 'heartbeat', timestamp: new Date().toISOString(), data: {} })
      }
    }, 30000)
  }

  // 停止心跳
  private stopPing() {
    if (this.pingInterval) {
      clearInterval(this.pingInterval)
      this.pingInterval = null
    }
  }

  // 发送消息
  send(message: Partial<WebSocketMessage>) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({
        ...message,
        timestamp: message.timestamp || new Date().toISOString()
      }))
    }
  }

  // 处理收到的消息
  private handleMessage(message: WebSocketMessage) {
    const handlers = this.handlers.get(message.type)
    if (handlers) {
      handlers.forEach(handler => handler(message))
    }

    // 触发通配符处理器
    const allHandlers = this.handlers.get('*')
    if (allHandlers) {
      allHandlers.forEach(handler => handler(message))
    }
  }

  // 订阅消息
  subscribe(type: string, handler: MessageHandler): () => void {
    if (!this.handlers.has(type)) {
      this.handlers.set(type, new Set())
    }
    this.handlers.get(type)!.add(handler)

    // 返回取消订阅函数
    return () => {
      const handlers = this.handlers.get(type)
      if (handlers) {
        handlers.delete(handler)
        if (handlers.size === 0) {
          this.handlers.delete(type)
        }
      }
    }
  }

  // 订阅新订单
  onNewOrder(handler: (data: any) => void): () => void {
    return this.subscribe('order_new', (msg) => handler(msg.data))
  }

  // 订阅订单更新
  onOrderUpdate(handler: (data: any) => void): () => void {
    return this.subscribe('order_update', (msg) => handler(msg.data))
  }

  // 订阅库存预警
  onInventoryAlert(handler: (data: any) => void): () => void {
    return this.subscribe('inventory_alert', (msg) => handler(msg.data))
  }

  // 订阅同步状态
  onSyncStatus(handler: (data: any) => void): () => void {
    return this.subscribe('sync_status', (msg) => handler(msg.data))
  }

  // 订阅通知
  onNotification(handler: (data: any) => void): () => void {
    return this.subscribe('notification', (msg) => handler(msg.data))
  }

  // 获取连接状态
  getConnectionStatus(): boolean {
    return this.isConnected
  }
}

// 导出单例
export const wsService = new WebSocketService()

export default wsService
