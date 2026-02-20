<template>
  <div class="role-management">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>角色管理</span>
          <el-button type="primary" @click="showCreateDialog">新建角色</el-button>
        </div>
      </template>

      <!-- 角色列表 -->
      <el-table :data="roles" v-loading="loading" stripe>
        <el-table-column prop="name" label="角色名称" width="120" />
        <el-table-column label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="row.is_system ? 'info' : 'success'">
              {{ row.is_system ? '系统预设' : '自定义' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="200" />
        <el-table-column label="权限数量" width="100" align="center">
          <template #default="{ row }">
            {{ row.permissions?.length || 0 }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="showEditDialog(row)">
              查看
            </el-button>
            <el-button
              type="danger"
              size="small"
              @click="handleDelete(row)"
              :disabled="row.is_system"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 创建/编辑角色对话框 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑角色' : '新建角色'" width="700px">
      <el-form :model="roleForm" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="角色代码" prop="code" v-if="!isEdit">
          <el-input v-model="roleForm.code" placeholder="请输入角色代码（英文）" />
        </el-form-item>
        <el-form-item label="角色名称" prop="name">
          <el-input v-model="roleForm.name" placeholder="请输入角色名称" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="roleForm.description" type="textarea" :rows="2" placeholder="请输入角色描述" />
        </el-form-item>
        <el-form-item label="权限设置">
          <el-tabs v-model="activeTab">
            <el-tab-pane
              v-for="group in permissionGroups"
              :key="group.resource"
              :label="group.label"
              :name="group.resource"
            >
              <el-checkbox-group v-model="roleForm.permission_codes">
                <div v-for="perm in group.permissions" :key="perm.code" class="permission-item">
                  <el-checkbox :label="perm.code" :disabled="currentRole?.is_system">
                    {{ perm.name }}
                    <span class="permission-desc">{{ perm.description }}</span>
                  </el-checkbox>
                </div>
              </el-checkbox-group>
            </el-tab-pane>
          </el-tabs>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          @click="handleSave"
          :loading="saving"
          :disabled="currentRole?.is_system && isEdit"
        >
          保存
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { permissionApi, type Role, type Permission } from '@/api/permission'

const loading = ref(false)
const roles = ref<Role[]>([])
const permissions = ref<Permission[]>([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const saving = ref(false)
const currentRole = ref<Role | null>(null)
const activeTab = ref('order')
const formRef = ref<FormInstance>()

const roleForm = reactive({
  code: '',
  name: '',
  description: '',
  permission_codes: [] as string[]
})

const rules: FormRules = {
  code: [
    { required: true, message: '请输入角色代码', trigger: 'blur' },
    { pattern: /^[a-z_]+$/, message: '角色代码只能包含小写字母和下划线', trigger: 'blur' }
  ],
  name: [
    { required: true, message: '请输入角色名称', trigger: 'blur' }
  ]
}

// 权限分组
const permissionGroups = computed(() => {
  const groups: Record<string, { resource: string; label: string; permissions: Permission[] }> = {
    order: { resource: 'order', label: '订单权限', permissions: [] },
    product: { resource: 'product', label: '商品权限', permissions: [] },
    inventory: { resource: 'inventory', label: '库存权限', permissions: [] },
    warehouse: { resource: 'warehouse', label: '仓库权限', permissions: [] },
    shop: { resource: 'shop', label: '店铺权限', permissions: [] },
    user: { resource: 'user', label: '用户权限', permissions: [] },
    finance: { resource: 'finance', label: '财务权限', permissions: [] },
    report: { resource: 'report', label: '报表权限', permissions: [] },
    role: { resource: 'role', label: '角色权限', permissions: [] },
    permission: { resource: 'permission', label: '授权权限', permissions: [] },
    system: { resource: 'system', label: '系统权限', permissions: [] },
  }

  const labels: Record<string, string> = {
    order: '订单权限',
    product: '商品权限',
    inventory: '库存权限',
    warehouse: '仓库权限',
    shop: '店铺权限',
    user: '用户权限',
    finance: '财务权限',
    report: '报表权限',
    role: '角色权限',
    permission: '授权权限',
    system: '系统权限',
  }

  permissions.value.forEach(perm => {
    if (groups[perm.resource]) {
      groups[perm.resource].permissions.push(perm)
    }
  })

  return Object.values(groups).filter(g => g.permissions.length > 0)
})

// 获取角色列表
const fetchRoles = async () => {
  loading.value = true
  try {
    const res = await permissionApi.getRoles() as any
    roles.value = res.roles || []
  } catch (error) {
    ElMessage.error('获取角色列表失败')
  } finally {
    loading.value = false
  }
}

// 获取权限列表
const fetchPermissions = async () => {
  try {
    const res = await permissionApi.getPermissions() as any
    permissions.value = res.permissions || []
  } catch (error) {
    console.error('Failed to fetch permissions:', error)
  }
}

// 显示创建对话框
const showCreateDialog = () => {
  isEdit.value = false
  currentRole.value = null
  roleForm.code = ''
  roleForm.name = ''
  roleForm.description = ''
  roleForm.permission_codes = []
  activeTab.value = 'order'
  dialogVisible.value = true
}

// 显示编辑对话框
const showEditDialog = (role: Role) => {
  isEdit.value = true
  currentRole.value = role
  roleForm.code = role.code
  roleForm.name = role.name
  roleForm.description = role.description || ''
  roleForm.permission_codes = role.permissions?.map(p => p.code) || []
  activeTab.value = 'order'
  dialogVisible.value = true
}

// 保存角色
const handleSave = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    saving.value = true
    try {
      if (isEdit.value && currentRole.value) {
        await permissionApi.updateRole(currentRole.value.id, {
          name: roleForm.name,
          description: roleForm.description,
          permission_codes: roleForm.permission_codes
        })
        ElMessage.success('更新成功')
      } else {
        await permissionApi.createRole({
          code: roleForm.code,
          name: roleForm.name,
          description: roleForm.description,
          permission_codes: roleForm.permission_codes
        })
        ElMessage.success('创建成功')
      }

      dialogVisible.value = false
      fetchRoles()
    } catch (error: any) {
      ElMessage.error(error.response?.data?.error || '保存失败')
    } finally {
      saving.value = false
    }
  })
}

// 删除角色
const handleDelete = async (role: Role) => {
  if (role.is_system) {
    ElMessage.warning('系统预设角色不能删除')
    return
  }

  try {
    await ElMessageBox.confirm('确定要删除该角色吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    await permissionApi.deleteRole(role.id)
    ElMessage.success('删除成功')
    fetchRoles()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '删除失败')
    }
  }
}

onMounted(() => {
  fetchRoles()
  fetchPermissions()
})
</script>

<style scoped>
.role-management {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.permission-item {
  margin-bottom: 10px;
}

.permission-desc {
  color: #909399;
  font-size: 12px;
  margin-left: 10px;
}
</style>
