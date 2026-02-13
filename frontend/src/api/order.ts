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

export const orderApi = {
  getList(params: OrderListParams) {
    return request.get('/orders', { params })
  },
  
  getDetail(id: string) {
    return request.get(`/orders/${id}`)
  },
  
  create(data: any) {
    return request.post('/orders', data)
  },
  
  audit(id: string) {
    return request.post(`/orders/${id}/audit`)
  },
  
  ship(id: string, logistics: { logistics_company: string; logistics_no: string }) {
    return request.post(`/orders/${id}/ship`, logistics)
  }
}