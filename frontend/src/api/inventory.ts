import request from '@/utils/request'

export interface Inventory {
  id: number
  tenant_id: number
  product_id: number
  warehouse_id: number
  quantity: number
  locked_qty: number
  total_qty: number
  alert_qty: number
  location: string
  batch_no: string
  expire_at: string | null
  created_at: string
  updated_at: string
  product?: {
    id: number
    sku_code: string
    name: string
    category: string
    brand: string
    unit: string
    sale_price: number
  }
  warehouse?: {
    id: number
    name: string
    code: string
  }
}

export interface InventoryLog {
  id: number
  tenant_id: number
  product_id: number
  warehouse_id: number
  change_qty: number
  before_qty: number
  after_qty: number
  ref_type: string
  ref_id: number
  ref_no: string
  operator_id: number
  remark: string
  created_at: string
}

export interface InventoryListParams {
  page?: number
  size?: number
  warehouse_id?: number
  product_id?: number
  keyword?: string
  low_stock?: boolean
}

export interface InventoryLogParams {
  page?: number
  size?: number
  warehouse_id?: number
  product_id?: number
  ref_type?: string
}

export interface AdjustInventoryParams {
  product_id: number
  warehouse_id: number
  change_qty: number
  remark?: string
}

export interface UpdateInventoryParams {
  alert_qty?: number
  location?: string
  batch_no?: string
  expire_at?: string
}

export const inventoryApi = {
  list(params: InventoryListParams) {
    return request.get('/inventory', { params })
  },

  get(id: number) {
    return request.get(`/inventory/${id}`)
  },

  update(id: number, data: UpdateInventoryParams) {
    return request.put(`/inventory/${id}`, data)
  },

  adjust(data: AdjustInventoryParams) {
    return request.post('/inventory/adjust', data)
  },

  logs(params: InventoryLogParams) {
    return request.get('/inventory/logs', { params })
  },

  alert() {
    return request.get('/inventory/alert')
  }
}
