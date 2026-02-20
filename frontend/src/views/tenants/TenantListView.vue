<template>
  <div class="tenant-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>账套管理</span>
          <el-button type="primary" @click="showCreateDialog">创建账套</el-button>
        </div>
      </template>

      <el-table :data="tenantStore.tenants" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="code" label="编码" width="120" />
        <el-table-column prop="name" label="名称" width="150" />
        <el-table-column prop="platform" label="平台" width="120">
          <template #default="{ row }">
            <el-tag :type="getPlatformTagType(row.platform)" size="small">
              {{ getPlatformLabel(row.platform) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" />
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="role" label="我的角色" width="100">
          <template #default="{ row }">
            <el-tag :type="getRoleTagType(row.role)" size="small">
              {{ getRoleLabel(row.role) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="switchToTenant(row)">
              切换
            </el-button>
            <el-button
              v-if="row.role === 'owner' || row.role === 'admin'"
              type="warning"
              size="small"
              @click="showEditDialog(row)"
            >
              编辑
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 创建/编辑账套对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑账套' : '创建账套'"
      width="500px"
    >
      <el-form :model="formData" :rules="rules" ref="formRef" label-width="80px">
        <el-form-item label="编码" prop="code" v-if="!isEdit">
          <el-input v-model="formData.code" placeholder="请输入账套编码" />
        </el-form-item>
        <el-form-item label="名称" prop="name">
          <el-input v-model="formData.name" placeholder="请输入账套名称" />
        </el-form-item>
        <el-form-item label="平台" prop="platform" v-if="!isEdit">
          <el-select v-model="formData.platform" placeholder="请选择平台" style="width: 100%">
            <el-option label="淘宝" value="taobao" />
            <el-option label="天猫" value="tmall" />
            <el-option label="抖音" value="douyin" />
            <el-option label="快手" value="kuaishou" />
            <el-option label="微信视频号" value="wechat_video" />
            <el-option label="京东" value="jd" />
            <el-option label="小红书" value="xiaohongshu" />
            <el-option label="唯品会" value="vip" />
            <el-option label="自定义" value="custom" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="formData.description"
            type="textarea"
            :rows="3"
            placeholder="请输入描述"
          />
        </el-form-item>
        <el-form-item label="状态" prop="status" v-if="isEdit">
          <el-switch
            v-model="formData.status"
            :active-value="1"
            :inactive-value="0"
            active-text="启用"
            inactive-text="禁用"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">
          {{ isEdit ? '保存' : '创建' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useTenantStore, type Tenant } from '@/stores/tenant'
import { tenantApi } from '@/api/tenant'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'

const tenantStore = useTenantStore()
const loading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref<FormInstance>()
const editId = ref<number>(0)

const formData = reactive({
  code: '',
  name: '',
  platform: '',
  description: '',
  status: 1
})

const rules: FormRules = {
  code: [
    { required: true, message: '请输入账套编码', trigger: 'blur' },
    { min: 2, max: 50, message: '编码长度为2-50个字符', trigger: 'blur' }
  ],
  name: [
    { required: true, message: '请输入账套名称', trigger: 'blur' },
    { min: 2, max: 100, message: '名称长度为2-100个字符', trigger: 'blur' }
  ],
  platform: [{ required: true, message: '请选择平台', trigger: 'change' }]
}

// 平台标签类型映射
const platformTagTypes: Record<string, string> = {
  taobao: 'warning',
  tmall: 'danger',
  douyin: '',
  kuaishou: 'success',
  wechat_video: 'success',
  jd: 'danger',
  xiaohongshu: 'danger',
  custom: 'info'
}

// 平台标签文字映射
const platformLabels: Record<string, string> = {
  taobao: '淘宝',
  tmall: '天猫',
  douyin: '抖音',
  kuaishou: '快手',
  wechat_video: '微信视频号',
  jd: '京东',
  xiaohongshu: '小红书',
  vip: '唯品会',
  custom: '自定义'
}

// 角色标签类型
const roleTagTypes: Record<string, string> = {
  owner: 'danger',
  admin: 'warning',
  member: ''
}

// 角色标签文字
const roleLabels: Record<string, string> = {
  owner: '所有者',
  admin: '管理员',
  member: '成员'
}

const getPlatformTagType = (platform: string): string => {
  return platformTagTypes[platform] || 'info'
}

const getPlatformLabel = (platform: string): string => {
  return platformLabels[platform] || platform
}

const getRoleTagType = (role: string): string => {
  return roleTagTypes[role] || 'info'
}

const getRoleLabel = (role: string): string => {
  return roleLabels[role] || role
}

const loadData = async () => {
  loading.value = true
  try {
    await tenantStore.fetchTenants()
  } finally {
    loading.value = false
  }
}

const showCreateDialog = () => {
  isEdit.value = false
  editId.value = 0
  Object.assign(formData, {
    code: '',
    name: '',
    platform: '',
    description: '',
    status: 1
  })
  dialogVisible.value = true
}

const showEditDialog = (tenant: Tenant) => {
  isEdit.value = true
  editId.value = tenant.id
  Object.assign(formData, {
    code: tenant.code,
    name: tenant.name,
    platform: tenant.platform,
    description: tenant.description || '',
    status: tenant.status
  })
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true
    try {
      if (isEdit.value) {
        await tenantApi.update(editId.value, {
          name: formData.name,
          description: formData.description,
          status: formData.status
        })
        ElMessage.success('更新成功')
      } else {
        await tenantApi.create({
          code: formData.code,
          name: formData.name,
          platform: formData.platform,
          description: formData.description
        })
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      await loadData()
    } catch (error: any) {
      ElMessage.error(error.response?.data?.error || '操作失败')
    } finally {
      submitting.value = false
    }
  })
}

const switchToTenant = async (tenant: Tenant) => {
  try {
    await tenantStore.switchTenant(tenant.id)
    ElMessage.success(`已切换到账套: ${tenant.name}`)
  } catch (error) {
    ElMessage.error('切换失败')
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.tenant-list {
  padding: 20px;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
