import request from '@/utils/request'

export interface Warehouse {
  id: number
  tenant_id: number
  code: string
  name: string
  address: string
  contact: string
  phone: string
  type: string
  status: number
  is_default: boolean
  created_at: string
  updated_at: string
}

export interface WarehouseListParams {
  status?: number
}

export interface CreateWarehouseParams {
  code: string
  name: string
  address?: string
  contact?: string
  phone?: string
  type?: string
}

export interface UpdateWarehouseParams {
  name?: string
  address?: string
  contact?: string
  phone?: string
  type?: string
  status?: number
  is_default?: boolean
}

export const warehouseApi = {
  list(params?: WarehouseListParams) {
    return request.get('/warehouses', { params })
  },

  get(id: number) {
    return request.get(`/warehouses/${id}`)
  },

  create(data: CreateWarehouseParams) {
    return request.post('/warehouses', data)
  },

  update(id: number, data: UpdateWarehouseParams) {
    return request.put(`/warehouses/${id}`, data)
  },

  delete(id: number) {
    return request.delete(`/warehouses/${id}`)
  },

  setDefault(id: number) {
    return request.post(`/warehouses/${id}/default`)
  }
}
