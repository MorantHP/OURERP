import request from '@/utils/request'

export interface Shop {
  id: number
  name: string
  platform: string
  platform_shop_id: string
  status: number
  sync_interval: number
  last_sync_at: string | null
  token_expires_at: string | null
  created_at: string
  api_url?: string
  webhook_url?: string
}

export interface Platform {
  code: string
  name: string
  icon: string
  description: string
  features: string[]
  auth_type: string
}

export const shopApi = {
  // 获取店铺列表
  list: (params?: { platform?: string; status?: number; page?: number; size?: number }) =>
    request.get('/shops', { params }),

  // 获取单个店铺
  get: (id: number) =>
    request.get(`/shops/${id}`),

  // 创建店铺
  create: (data: Partial<Shop>) =>
    request.post('/shops', data),

  // 更新店铺
  update: (id: number, data: Partial<Shop>) =>
    request.put(`/shops/${id}`, data),

  // 删除店铺
  delete: (id: number) =>
    request.delete(`/shops/${id}`),

  // 手动触发同步
  triggerSync: (shopId: number) =>
    request.post(`/shops/${shopId}/sync`),

  // 获取授权URL
  getAuthUrl: (shopId: number) =>
    request.get(`/shops/${shopId}/auth-url`),
}

export const platformApi = {
  // 获取所有支持的平台
  list: (): Promise<{ platforms: Platform[] }> =>
    request.get('/platforms'),

  // 获取单个平台
  get: (code: string): Promise<{ platform: Platform }> =>
    request.get(`/platforms/${code}`),
}

export const oauthApi = {
  // 获取授权URL
  getAuthUrl: (shopId: number): Promise<{ auth_url: string; state: string }> =>
    request.get('/oauth/auth-url', { params: { shop_id: shopId } }),

  // 刷新Token
  refreshToken: (shopId: number) =>
    request.post('/oauth/refresh', null, { params: { shop_id: shopId } }),
}
