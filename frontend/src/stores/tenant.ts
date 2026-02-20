import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { tenantApi, type Tenant } from '@/api/tenant'

export const useTenantStore = defineStore('tenant', () => {
  // 当前选中的租户ID
  const currentTenantId = ref<number>(0)

  // 当前租户信息
  const currentTenant = ref<Tenant | null>(null)

  // 用户可访问的租户列表
  const tenants = ref<Tenant[]>([])

  // 是否已选择租户
  const hasTenant = computed(() => currentTenantId.value > 0)

  // 当前用户在租户中的角色
  const currentRole = computed(() => currentTenant.value?.role || '')

  // 是否是管理员（owner或admin）
  const isAdmin = computed(() =>
    currentRole.value === 'owner' || currentRole.value === 'admin'
  )

  // 设置当前租户
  const setCurrentTenant = (tenant: Tenant | null) => {
    currentTenant.value = tenant
    currentTenantId.value = tenant?.id || 0
    if (tenant) {
      localStorage.setItem('tenant_id', String(tenant.id))
    } else {
      localStorage.removeItem('tenant_id')
    }
  }

  // 获取用户可访问的租户列表
  const fetchTenants = async () => {
    try {
      const res = await tenantApi.getMyTenants()
      tenants.value = res.list || []

      // 如果有当前租户ID，更新当前租户信息
      if (currentTenantId.value > 0) {
        const found = tenants.value.find(t => t.id === currentTenantId.value)
        if (found) {
          currentTenant.value = found
        }
      }

      // 如果没有当前租户但有可用租户，选择第一个
      if (!currentTenantId.value && tenants.value.length > 0) {
        await switchTenant(tenants.value[0].id)
      }

      return tenants.value
    } catch (error) {
      console.error('Failed to fetch tenants:', error)
      throw error
    }
  }

  // 切换租户
  const switchTenant = async (tenantId: number) => {
    try {
      const res = await tenantApi.switchTenant({ tenant_id: tenantId })
      // 使用后端返回的租户信息更新当前租户
      if (res.tenant) {
        setCurrentTenant(res.tenant as Tenant)
      }
      return res
    } catch (error) {
      console.error('Failed to switch tenant:', error)
      throw error
    }
  }

  // 初始化 - 从本地存储恢复租户ID
  const init = () => {
    const savedTenantId = localStorage.getItem('tenant_id')
    if (savedTenantId) {
      currentTenantId.value = parseInt(savedTenantId, 10)
    }
  }

  // 清除租户信息
  const clearTenant = () => {
    currentTenant.value = null
    currentTenantId.value = 0
    tenants.value = []
    localStorage.removeItem('tenant_id')
  }

  // 初始化
  init()

  return {
    currentTenantId,
    currentTenant,
    tenants,
    hasTenant,
    currentRole,
    isAdmin,
    setCurrentTenant,
    fetchTenants,
    switchTenant,
    clearTenant
  }
})
