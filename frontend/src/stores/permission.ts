import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { permissionApi, type Role, type Permission, type UserPermissions } from '@/api/permission'
import { useUserStore } from './user'

export const usePermissionStore = defineStore('permission', () => {
  // 状态
  const role = ref<Role | null>(null)
  const permissions = ref<string[]>([])
  const shopIds = ref<number[]>([])
  const warehouseIds = ref<number[]>([])
  const allShops = ref(true)
  const allWarehouses = ref(true)
  const loading = ref(false)
  const loaded = ref(false)

  // 计算属性
  const isOwner = computed(() => role.value?.code === 'owner')
  const isAdmin = computed(() => role.value?.code === 'admin' || isOwner.value)
  const canAssignPermission = computed(() => hasPermission('permission:assign'))

  // 方法
  async function fetchPermissions() {
    const userStore = useUserStore()
    if (!userStore.token) {
      return
    }

    loading.value = true
    try {
      const res = await permissionApi.getMyPermissions() as any
      role.value = res.role
      permissions.value = res.permissions || []
      shopIds.value = res.shop_ids || []
      warehouseIds.value = res.warehouse_ids || []
      allShops.value = res.all_shops ?? true
      allWarehouses.value = res.all_warehouses ?? true
      loaded.value = true
    } catch (error) {
      console.error('Failed to fetch permissions:', error)
    } finally {
      loading.value = false
    }
  }

  function hasPermission(perm: string): boolean {
    // 如果是主账号，拥有所有权限
    if (isOwner.value) {
      return true
    }
    return permissions.value.includes(perm)
  }

  function hasAnyPermission(perms: string[]): boolean {
    // 如果是主账号，拥有所有权限
    if (isOwner.value) {
      return true
    }
    return perms.some(perm => permissions.value.includes(perm))
  }

  function hasAllPermissions(perms: string[]): boolean {
    // 如果是主账号，拥有所有权限
    if (isOwner.value) {
      return true
    }
    return perms.every(perm => permissions.value.includes(perm))
  }

  function canAccessShop(shopId: number): boolean {
    // 如果是主账号或管理员，可以访问所有店铺
    if (isOwner.value || isAdmin.value) {
      return true
    }
    // 如果设置了全部访问
    if (allShops.value) {
      return true
    }
    // 检查是否在允许列表中
    return shopIds.value.includes(shopId)
  }

  function canAccessWarehouse(warehouseId: number): boolean {
    // 如果是主账号或管理员，可以访问所有仓库
    if (isOwner.value || isAdmin.value) {
      return true
    }
    // 如果设置了全部访问
    if (allWarehouses.value) {
      return true
    }
    // 检查是否在允许列表中
    return warehouseIds.value.includes(warehouseId)
  }

  function canRead(resource: string): boolean {
    return hasPermission(`${resource}:read`)
  }

  function canWrite(resource: string): boolean {
    return hasPermission(`${resource}:write`)
  }

  function canDelete(resource: string): boolean {
    return hasPermission(`${resource}:delete`)
  }

  function clear() {
    role.value = null
    permissions.value = []
    shopIds.value = []
    warehouseIds.value = []
    allShops.value = true
    allWarehouses.value = true
    loaded.value = false
  }

  return {
    // 状态
    role,
    permissions,
    shopIds,
    warehouseIds,
    allShops,
    allWarehouses,
    loading,
    loaded,
    // 计算属性
    isOwner,
    isAdmin,
    canAssignPermission,
    // 方法
    fetchPermissions,
    hasPermission,
    hasAnyPermission,
    hasAllPermissions,
    canAccessShop,
    canAccessWarehouse,
    canRead,
    canWrite,
    canDelete,
    clear,
  }
})
