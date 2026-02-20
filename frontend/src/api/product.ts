import request from '@/utils/request'

export interface Product {
  id: number
  tenant_id: number
  sku_code: string
  name: string
  category: string
  brand: string
  barcode: string
  image_url: string
  unit: string
  cost_price: number
  sale_price: number
  specs: Record<string, any>
  status: number
  remark: string
  created_at: string
  updated_at: string
  total_quantity?: number
}

export interface ProductListParams {
  page?: number
  size?: number
  category?: string
  brand?: string
  keyword?: string
  status?: number
}

export interface CreateProductParams {
  sku_code: string
  name: string
  category?: string
  brand?: string
  barcode?: string
  image_url?: string
  unit?: string
  cost_price?: number
  sale_price?: number
  specs?: Record<string, any>
  remark?: string
}

export interface UpdateProductParams {
  name?: string
  category?: string
  brand?: string
  barcode?: string
  image_url?: string
  unit?: string
  cost_price?: number
  sale_price?: number
  specs?: Record<string, any>
  status?: number
  remark?: string
}

export const productApi = {
  list(params: ProductListParams) {
    return request.get('/products', { params })
  },

  get(id: number) {
    return request.get(`/products/${id}`)
  },

  create(data: CreateProductParams) {
    return request.post('/products', data)
  },

  update(id: number, data: UpdateProductParams) {
    return request.put(`/products/${id}`, data)
  },

  delete(id: number) {
    return request.delete(`/products/${id}`)
  },

  getCategories() {
    return request.get('/products/categories')
  },

  getBrands() {
    return request.get('/products/brands')
  }
}
