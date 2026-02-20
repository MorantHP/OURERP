<template>
  <div class="user-permission">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>用户权限管理</span>
        </div>
      </template>

      <!-- 用户列表 -->
      <el-table :data="users" v-loading="loading" stripe>
        <el-table-column prop="name" label="用户名" width="120" />
        <el-table-column prop="email" label="邮箱" width="200" />
        <el-table-column label="角色" width="120">
          <template #default="{ row }">
            <el-tag :type="getRoleTagType(row.role)">
              {{ row.role?.name || '未设置' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="店铺权限" width="150">
          <template #default="{ row }">
            <span v-if="row.all_shops">全部店铺</span>
            <span v-else-if="row.shop_ids?.length">{{ row.shop_ids.length }} 个店铺</span>
            <span v-else class="text-muted">未设置</span>
          </template>
        </el-table-column>
        <el-table-column label="仓库权限" width="150">
          <template #default="{ row }">
            <span v-if="row.all_warehouses">全部仓库</span>
            <span v-else-if="row.warehouse_ids?.length">{{ row.warehouse_ids.length }} 个仓库</span>
            <span v-else class="text-muted">未设置</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="showEditDialog(row)" :disabled="row.role?.code === 'owner'">
              编辑
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 编辑权限对话框 -->
    <el-dialog v-model="editDialogVisible" title="编辑用户权限" width="600px">
      <el-form :model="editForm" label-width="100px" v-loading="editLoading">
        <el-form-item label="用户">
          <span>{{ currentUser?.name }} ({{ currentUser?.email }})</span>
        </el-form-item>

        <el-form-item label="角色">
          <el-select v-model="editForm.role_id" placeholder="请选择角色" style="width: 100%">
            <el-option
              v-for="role in roles"
              :key="role.id"
              :label="role.name"
              :value="role.id"
              :disabled="role.code === 'owner'"
            >
              <span>{{ role.name }}</span>
              <span style="color: #909399; font-size: 12px; margin-left: 10px;">{{ role.description }}</span>
            </el-option>
          </el-select>
        </el-form-item>

        <el-divider>资源权限</el-divider>

        <el-form-item label="店铺权限">
          <el-radio-group v-model="editForm.shop_access_type">
            <el-radio label="all">全部店铺</el-radio>
            <el-radio label="selected">指定店铺</el-radio>
          </el-radio-group>
          <el-select
            v-if="editForm.shop_access_type === 'selected'"
            v-model="editForm.shop_ids"
            multiple
            placeholder="请选择店铺"
            style="width: 100%; margin-top: 10px;"
          >
            <el-option
              v-for="shop in shops"
              :key="shop.id"
              :label="shop.name"
              :value="shop.id"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="仓库权限">
          <el-radio-group v-model="editForm.warehouse_access_type">
            <el-radio label="all">全部仓库</el-radio>
            <el-radio label="selected">指定仓库</el-radio>
          </el-radio-group>
          <el-select
            v-if="editForm.warehouse_access_type === 'selected'"
            v-model="editForm.warehouse_ids"
            multiple
            placeholder="请选择仓库"
            style="width: 100%; margin-top: 10px;"
          >
            <el-option
              v-for="warehouse in warehouses"
              :key="warehouse.id"
              :label="warehouse.name"
              :value="warehouse.id"
            />
          </el-select>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { permissionApi, type Role, type ShopResource, type WarehouseResource } from '@/api/permission'
import { useTenantStore } from '@/stores/tenant'

interface UserWithPermission {
  id: number
  name: string
  email: string
  role: Role | null
  permissions: string[]
  shop_ids: number[]
  warehouse_ids: number[]
  all_shops: boolean
  all_warehouses: boolean
}

const loading = ref(false)
const users = ref<UserWithPermission[]>([])
const roles = ref<Role[]>([])
const shops = ref<ShopResource[]>([])
const warehouses = ref<WarehouseResource[]>([])

const editDialogVisible = ref(false)
const editLoading = ref(false)
const saving = ref(false)
const currentUser = ref<UserWithPermission | null>(null)

const editForm = reactive({
  role_id: 0,
  shop_access_type: 'all',
  shop_ids: [] as number[],
  warehouse_access_type: 'all',
  warehouse_ids: [] as number[]
})

const tenantStore = useTenantStore()

// 获取租户用户列表
const fetchUsers = async () => {
  loading.value = true
  try {
    // 获取租户用户
    const tenantUsers = await tenantStore.fetchTenantUsers() as any
    const userList = tenantUsers || []

    // 获取每个用户的权限
    const usersWithPerms: UserWithPermission[] = []
    for (const user of userList) {
      try {
        const perms = await permissionApi.getUserPermissions(user.user_id) as any
        usersWithPerms.push({
          id: user.user_id,
          name: user.user?.name || user.name || '未知',
          email: user.user?.email || user.email || '',
          role: perms.role,
          permissions: perms.permissions || [],
          shop_ids: perms.shop_ids || [],
          warehouse_ids: perms.warehouse_ids || [],
          all_shops: perms.all_shops ?? true,
          all_warehouses: perms.all_warehouses ?? true
        })
      } catch (error) {
        console.error(`Failed to get permissions for user ${user.user_id}:`, error)
      }
    }
    users.value = usersWithPerms
  } catch (error) {
    ElMessage.error('获取用户列表失败')
  } finally {
    loading.value = false
  }
}

// 获取角色列表
const fetchRoles = async () => {
  try {
    const res = await permissionApi.getRoles() as any
    roles.value = res.roles || []
  } catch (error) {
    console.error('Failed to fetch roles:', error)
  }
}

// 获取店铺列表
const fetchShops = async () => {
  try {
    const res = await permissionApi.getShops() as any
    shops.value = res.shops || []
  } catch (error) {
    console.error('Failed to fetch shops:', error)
  }
}

// 获取仓库列表
const fetchWarehouses = async () => {
  try {
    const res = await permissionApi.getWarehouses() as any
    warehouses.value = res.warehouses || []
  } catch (error) {
    console.error('Failed to fetch warehouses:', error)
  }
}

// 显示编辑对话框
const showEditDialog = async (user: UserWithPermission) => {
  currentUser.value = user
  editLoading.value = true
  editDialogVisible.value = true

  // 设置表单值
  editForm.role_id = user.role?.id || 0
  editForm.shop_access_type = user.all_shops ? 'all' : 'selected'
  editForm.shop_ids = user.shop_ids || []
  editForm.warehouse_access_type = user.all_warehouses ? 'all' : 'selected'
  editForm.warehouse_ids = user.warehouse_ids || []

  editLoading.value = false
}

// 保存权限
const handleSave = async () => {
  if (!currentUser.value) return

  saving.value = true
  try {
    // 设置角色
    if (editForm.role_id) {
      await permissionApi.setUserRole(currentUser.value.id, editForm.role_id)
    }

    // 设置资源权限
    const permissions: any[] = []

    if (editForm.shop_access_type === 'selected' && editForm.shop_ids.length > 0) {
      editForm.shop_ids.forEach(shopId => {
        permissions.push({
          resource_type: 'shop',
          resource_id: shopId,
          can_read: true,
          can_write: false,
          can_delete: false
        })
      })
    }

    if (editForm.warehouse_access_type === 'selected' && editForm.warehouse_ids.length > 0) {
      editForm.warehouse_ids.forEach(warehouseId => {
        permissions.push({
          resource_type: 'warehouse',
          resource_id: warehouseId,
          can_read: true,
          can_write: false,
          can_delete: false
        })
      })
    }

    await permissionApi.setResourcePermissions(currentUser.value.id, { permissions })

    ElMessage.success('保存成功')
    editDialogVisible.value = false
    fetchUsers()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '保存失败')
  } finally {
    saving.value = false
  }
}

// 获取角色标签类型
const getRoleTagType = (role: Role | null): string => {
  if (!role) return 'info'
  switch (role.code) {
    case 'owner': return 'danger'
    case 'admin': return 'warning'
    case 'manager': return 'primary'
    default: return 'info'
  }
}

onMounted(() => {
  fetchUsers()
  fetchRoles()
  fetchShops()
  fetchWarehouses()
})
</script>

<style scoped>
.user-permission {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.text-muted {
  color: #909399;
}
</style>
