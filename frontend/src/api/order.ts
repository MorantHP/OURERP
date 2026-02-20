import request from '@/utils/request'

export interface Order {
  id: number
  order_no: string
  platform: string
  status: number
  total_amount: number
  pay_amount: number
  buyer_nick: string
  receiver_name: string
  receiver_phone: string
  created_at: string
  items?: OrderItem[]
}

export interface OrderItem {
  id: number
  sku_name: string
  quantity: number
  price: number
}

export interface OrderListParams {
  page?: number
  size?: number
  status?: number
  platform?: string
  keyword?: string
}

export interface CreateOrderItem {
  sku_id: number
  sku_name: string
  quantity: number
  price: number
}

export interface CreateOrderParams {
  platform: string
  platform_order_id?: string
  shop_id?: number
  total_amount: number
  pay_amount: number
  buyer_nick?: string
  receiver_name?: string
  receiver_phone?: string
  receiver_address?: string
  items: CreateOrderItem[]
}

export interface OrderListResponse {
  list: Order[]
  pagination: {
    total: number
    page: number
    size: number
    total_pages: number
  }
}

export const orderApi = {
  getList(params: OrderListParams): Promise<OrderListResponse> {
    return request.get('/orders', { params })
  },

  getDetail(id: string): Promise<{ order: Order }> {
    return request.get(`/orders/${id}`)
  },

  create(data: CreateOrderParams): Promise<{ order: Order }> {
    return request.post('/orders', data)
  },

  audit(id: string): Promise<{ message: string }> {
    return request.post(`/orders/${id}/audit`)
  },

  ship(id: string, logistics: { logistics_company: string; logistics_no: string }): Promise<{ message: string }> {
    return request.post(`/orders/${id}/ship`, logistics)
  }
}