import request from '@/utils/request'

export interface Tenant {
  id: number
  code: string
  name: string
  platform: string
  description: string
  logo: string
  status: number
  settings: Record<string, any>
  owner_id: number
  created_at: string
  updated_at: string
  role?: string // 用户在该租户中的角色
}

export interface CreateTenantParams {
  code: string
  name: string
  platform?: string
  description?: string
}

export interface UpdateTenantParams {
  name?: string
  description?: string
  logo?: string
  status?: number
  settings?: Record<string, any>
}

export interface AddUserParams {
  user_id: number
  role: 'owner' | 'admin' | 'member'
}

export interface SwitchTenantParams {
  tenant_id: number
}

export const tenantApi = {
  // 获取所有租户列表（管理员）
  list(params?: { page?: number; size?: number; platform?: string; status?: number }) {
    return request.get('/tenants', { params })
  },

  // 获取当前用户可访问的租户列表
  getMyTenants() {
    return request.get('/tenants/my')
  },

  // 获取租户详情
  get(id: number) {
    return request.get(`/tenants/${id}`)
  },

  // 创建租户
  create(data: CreateTenantParams) {
    return request.post('/tenants', data)
  },

  // 更新租户
  update(id: number, data: UpdateTenantParams) {
    return request.put(`/tenants/${id}`, data)
  },

  // 删除租户
  delete(id: number) {
    return request.delete(`/tenants/${id}`)
  },

  // 添加用户到租户
  addUser(tenantId: number, data: AddUserParams) {
    return request.post(`/tenants/${tenantId}/users`, data)
  },

  // 从租户移除用户
  removeUser(tenantId: number, userId: number) {
    return request.delete(`/tenants/${tenantId}/users/${userId}`)
  },

  // 更新用户角色
  updateUserRole(tenantId: number, userId: number, role: string) {
    return request.put(`/tenants/${tenantId}/users/${userId}/role`, { role })
  },

  // 切换当前租户
  switchTenant(data: SwitchTenantParams) {
    return request.post('/tenants/switch', data)
  }
}
