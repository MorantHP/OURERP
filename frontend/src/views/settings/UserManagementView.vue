<template>
  <div class="user-management">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>用户管理</span>
          <el-tag type="danger" v-if="!userStore.userInfo?.is_root">需要 Root 权限</el-tag>
        </div>
      </template>

      <!-- 提示信息 -->
      <el-alert
        v-if="!userStore.userInfo?.is_root"
        title="权限不足"
        type="warning"
        description="只有 Root 用户才能管理用户"
        :closable="false"
        show-icon
        style="margin-bottom: 20px"
      />

      <!-- 用户列表 -->
      <el-table :data="users" v-loading="loading" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="email" label="邮箱" width="200" />
        <el-table-column prop="name" label="姓名" width="120" />
        <el-table-column prop="phone" label="电话" width="130" />

        <el-table-column label="身份" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.is_root" type="danger">Root</el-tag>
            <el-tag v-else-if="row.is_approved" type="success">已审核</el-tag>
            <el-tag v-else type="warning">待审核</el-tag>
          </template>
        </el-table-column>

        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? '正常' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="created_at" label="注册时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>

        <el-table-column label="操作" fixed="right" width="280">
          <template #default="{ row }">
            <!-- Root 用户不能被操作 -->
            <span v-if="row.is_root" class="text-gray">不可操作</span>

            <template v-else>
              <!-- 审核按钮 -->
              <el-button
                v-if="!row.is_approved"
                type="success"
                size="small"
                @click="handleApprove(row, true)"
              >
                通过审核
              </el-button>
              <el-button
                v-if="!row.is_approved"
                type="danger"
                size="small"
                plain
                @click="handleApprove(row, false)"
              >
                拒绝
              </el-button>

              <!-- 状态切换 -->
              <el-button
                v-if="row.is_approved"
                :type="row.status === 1 ? 'warning' : 'success'"
                size="small"
                @click="handleToggleStatus(row)"
              >
                {{ row.status === 1 ? '禁用' : '启用' }}
              </el-button>

              <!-- 删除 -->
              <el-button
                type="danger"
                size="small"
                plain
                @click="handleDelete(row)"
              >
                删除
              </el-button>
            </template>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { userApi, type User } from '@/api/user'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const users = ref<User[]>([])
const loading = ref(false)

const formatDate = (date: string) => {
  if (!date) return '-'
  return new Date(date).toLocaleString('zh-CN')
}

const loadUsers = async () => {
  if (!userStore.userInfo?.is_root) {
    return
  }

  loading.value = true
  try {
    const res = await userApi.list()
    users.value = res.users || []
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '加载用户列表失败')
  } finally {
    loading.value = false
  }
}

const handleApprove = async (user: User, approved: boolean) => {
  const action = approved ? '通过审核' : '拒绝'

  try {
    await ElMessageBox.confirm(
      `确定要${action}用户 "${user.name}" 吗？`,
      '确认操作',
      { type: 'warning' }
    )

    await userApi.approve(user.id, approved)
    ElMessage.success(approved ? '审核通过' : '已拒绝')
    loadUsers()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '操作失败')
    }
  }
}

const handleToggleStatus = async (user: User) => {
  const newStatus = user.status === 1 ? 0 : 1
  const action = newStatus === 0 ? '禁用' : '启用'

  try {
    await ElMessageBox.confirm(
      `确定要${action}用户 "${user.name}" 吗？`,
      '确认操作',
      { type: 'warning' }
    )

    await userApi.setStatus(user.id, newStatus)
    ElMessage.success(`已${action}`)
    loadUsers()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '操作失败')
    }
  }
}

const handleDelete = async (user: User) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除用户 "${user.name}" 吗？此操作不可恢复！`,
      '确认删除',
      { type: 'error' }
    )

    await userApi.delete(user.id)
    ElMessage.success('删除成功')
    loadUsers()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '删除失败')
    }
  }
}

onMounted(() => {
  loadUsers()
})
</script>

<style scoped>
.user-management {
  padding: 20px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.text-gray {
  color: #909399;
  font-size: 12px;
}
</style>
