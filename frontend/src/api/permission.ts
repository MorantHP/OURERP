import request from '@/utils/request'

// 角色接口
export interface Role {
  id: number
  tenant_id: number
  code: string
  name: string
  description: string
  is_system: boolean
  permissions: Permission[]
  created_at: string
  updated_at: string
}

// 权限接口
export interface Permission {
  id: number
  code: string
  name: string
  resource: string
  action: string
  description: string
  created_at: string
}

// 用户资源权限
export interface UserResourcePermission {
  id: number
  tenant_id: number
  user_id: number
  resource_type: 'shop' | 'warehouse'
  resource_id: number
  can_read: boolean
  can_write: boolean
  can_delete: boolean
  created_at: string
}

// 用户权限汇总
export interface UserPermissions {
  role: Role | null
  permissions: string[]
  shop_ids: number[]
  warehouse_ids: number[]
  all_shops: boolean
  all_warehouses: boolean
  resource_permissions: UserResourcePermission[]
}

// 创建角色参数
export interface CreateRoleParams {
  code: string
  name: string
  description?: string
  permission_codes: string[]
}

// 更新角色参数
export interface UpdateRoleParams {
  name: string
  description?: string
  permission_codes: string[]
}

// 设置资源权限参数
export interface SetResourcePermissionsParams {
  permissions: Array<{
    resource_type: 'shop' | 'warehouse'
    resource_id: number
    can_read: boolean
    can_write: boolean
    can_delete: boolean
  }>
}

// 店铺简要信息
export interface ShopResource {
  id: number
  name: string
  platform: string
}

// 仓库简要信息
export interface WarehouseResource {
  id: number
  name: string
  code: string
}

export const permissionApi = {
  // ==================== 角色管理 ====================

  // 获取角色列表
  getRoles(): Promise<{ roles: Role[] }> {
    return request.get('/permissions/roles')
  },

  // 获取角色详情
  getRole(id: number): Promise<{ role: Role }> {
    return request.get(`/permissions/roles/${id}`)
  },

  // 创建自定义角色
  createRole(data: CreateRoleParams): Promise<{ role: Role }> {
    return request.post('/permissions/roles', data)
  },

  // 更新角色
  updateRole(id: number, data: UpdateRoleParams): Promise<{ message: string }> {
    return request.put(`/permissions/roles/${id}`, data)
  },

  // 删除角色
  deleteRole(id: number): Promise<{ message: string }> {
    return request.delete(`/permissions/roles/${id}`)
  },

  // ==================== 权限查询 ====================

  // 获取所有权限列表
  getPermissions(): Promise<{ permissions: Permission[] }> {
    return request.get('/permissions')
  },

  // 获取当前用户权限
  getMyPermissions(): Promise<UserPermissions> {
    return request.get('/permissions/my')
  },

  // 获取指定用户权限
  getUserPermissions(userId: number): Promise<UserPermissions> {
    return request.get(`/permissions/users/${userId}`)
  },

  // ==================== 用户授权 ====================

  // 设置用户角色
  setUserRole(userId: number, roleId: number): Promise<{ message: string }> {
    return request.put(`/permissions/users/${userId}/role`, { role_id: roleId })
  },

  // 设置用户资源权限
  setResourcePermissions(userId: number, data: SetResourcePermissionsParams): Promise<{ message: string }> {
    return request.put(`/permissions/users/${userId}/resources`, data)
  },

  // 添加用户资源权限
  addResourcePermission(userId: number, permission: Omit<UserResourcePermission, 'id' | 'tenant_id' | 'user_id' | 'created_at'>): Promise<{ permission: UserResourcePermission }> {
    return request.post(`/permissions/users/${userId}/resources`, permission)
  },

  // 移除用户资源权限
  removeResourcePermission(userId: number, permissionId: number): Promise<{ message: string }> {
    return request.delete(`/permissions/users/${userId}/resources/${permissionId}`)
  },

  // ==================== 可授权资源 ====================

  // 获取可授权的店铺列表
  getShops(): Promise<{ shops: ShopResource[] }> {
    return request.get('/permissions/resources/shops')
  },

  // 获取可授权的仓库列表
  getWarehouses(): Promise<{ warehouses: WarehouseResource[] }> {
    return request.get('/permissions/resources/warehouses')
  }
}

// 权限代码常量
export const PermissionCodes = {
  // 订单权限
  ORDER_READ: 'order:read',
  ORDER_WRITE: 'order:write',
  ORDER_DELETE: 'order:delete',
  ORDER_AUDIT: 'order:audit',
  ORDER_SHIP: 'order:ship',

  // 商品权限
  PRODUCT_READ: 'product:read',
  PRODUCT_WRITE: 'product:write',
  PRODUCT_DELETE: 'product:delete',

  // 库存权限
  INVENTORY_READ: 'inventory:read',
  INVENTORY_WRITE: 'inventory:write',

  // 仓库权限
  WAREHOUSE_READ: 'warehouse:read',
  WAREHOUSE_WRITE: 'warehouse:write',

  // 用户权限
  USER_READ: 'user:read',
  USER_WRITE: 'user:write',
  USER_DELETE: 'user:delete',

  // 店铺权限
  SHOP_READ: 'shop:read',
  SHOP_WRITE: 'shop:write',
  SHOP_DELETE: 'shop:delete',

  // 财务权限
  FINANCE_READ: 'finance:read',
  FINANCE_WRITE: 'finance:write',
  FINANCE_EXPORT: 'finance:export',

  // 报表权限
  REPORT_READ: 'report:read',
  REPORT_EXPORT: 'report:export',

  // 角色权限
  ROLE_READ: 'role:read',
  ROLE_WRITE: 'role:write',

  // 授权权限
  PERMISSION_ASSIGN: 'permission:assign',

  // 系统权限
  SYSTEM_CONFIG: 'system:config',
  SYSTEM_LOG: 'system:log',
}

// 角色代码常量
export const RoleCodes = {
  OWNER: 'owner',
  ADMIN: 'admin',
  MANAGER: 'manager',
  OPERATOR: 'operator',
  FINANCE: 'finance',
  WAREHOUSE: 'warehouse',
  CUSTOMER: 'customer',
  VIEWER: 'viewer',
}
