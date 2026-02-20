<template>
  <div class="alert-management-view">
    <!-- 标签页 -->
    <el-tabs v-model="activeTab">
      <el-tab-pane label="预警记录" name="records">
        <!-- 预警汇总 -->
        <el-row :gutter="20">
          <el-col :span="6">
            <el-card shadow="hover">
              <el-statistic title="未处理预警" :value="summary?.unhandled_alerts || 0">
                <template #suffix>
                  <span v-if="summary?.unhandled_alerts" class="text-danger">条</span>
                  <span v-else class="text-success">条</span>
                </template>
              </el-statistic>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card shadow="hover">
              <el-statistic title="严重预警" :value="summary?.critical_count || 0" />
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card shadow="hover">
              <el-statistic title="警告" :value="summary?.warning_count || 0" />
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card shadow="hover">
              <el-statistic title="今日预警" :value="summary?.today_alerts || 0" />
            </el-card>
          </el-col>
        </el-row>

        <!-- 预警记录列表 -->
        <el-card style="margin-top: 20px">
          <template #header>
            <div class="card-header">
              <span>预警记录</span>
              <el-button type="primary" @click="checkAlerts" :loading="checking">
                <el-icon><Refresh /></el-icon>
                检查预警
              </el-button>
            </div>
          </template>

          <!-- 筛选 -->
          <el-form :inline="true" class="filter-form">
            <el-form-item label="级别">
              <el-select v-model="recordFilter.level" placeholder="全部" clearable style="width: 120px">
                <el-option label="严重" value="critical" />
                <el-option label="警告" value="warning" />
                <el-option label="信息" value="info" />
              </el-select>
            </el-form-item>
            <el-form-item label="状态">
              <el-select v-model="recordFilter.status" placeholder="全部" clearable style="width: 120px">
                <el-option label="未处理" :value="0" />
                <el-option label="已处理" :value="1" />
                <el-option label="已忽略" :value="2" />
              </el-select>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="fetchRecords">查询</el-button>
            </el-form-item>
          </el-form>

          <!-- 表格 -->
          <el-table :data="records" v-loading="recordsLoading" stripe>
            <el-table-column prop="title" label="预警标题" min-width="200" />
            <el-table-column prop="content" label="内容" min-width="300" show-overflow-tooltip />
            <el-table-column label="级别" width="100">
              <template #default="{ row }">
                <el-tag :type="getLevelType(row.level)">
                  {{ getLevelLabel(row.level) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="getStatusType(row.status)">
                  {{ getStatusLabel(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间" width="180">
              <template #default="{ row }">
                {{ formatTime(row.created_at) }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="150" fixed="right">
              <template #default="{ row }">
                <el-button
                  v-if="row.status === 0"
                  type="success"
                  size="small"
                  @click="showHandleDialog(row)"
                >
                  处理
                </el-button>
                <el-button
                  v-if="row.status === 0"
                  type="warning"
                  size="small"
                  @click="ignoreRecord(row)"
                >
                  忽略
                </el-button>
              </template>
            </el-table-column>
          </el-table>

          <!-- 分页 -->
          <el-pagination
            v-model:current-page="recordPagination.page"
            v-model:page-size="recordPagination.pageSize"
            :total="recordPagination.total"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next"
            @size-change="fetchRecords"
            @current-change="fetchRecords"
            style="margin-top: 20px; justify-content: flex-end"
          />
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="预警规则" name="rules">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>预警规则配置</span>
              <el-button type="primary" @click="showCreateRuleDialog">
                <el-icon><Plus /></el-icon>
                新建规则
              </el-button>
            </div>
          </template>

          <!-- 表格 -->
          <el-table :data="rules" v-loading="rulesLoading" stripe>
            <el-table-column prop="name" label="规则名称" min-width="150" />
            <el-table-column label="类型" width="120">
              <template #default="{ row }">
                <el-tag>{{ getTypeLabel(row.type) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="threshold" label="阈值" width="100" align="right" />
            <el-table-column label="级别" width="100">
              <template #default="{ row }">
                <el-tag :type="getLevelType(row.level)">
                  {{ getLevelLabel(row.level) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="通知方式" width="120">
              <template #default="{ row }">
                {{ getNotifyTypeLabel(row.notify_type) }}
              </template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-switch
                  :model-value="row.status === 1"
                  @change="(val: boolean) => toggleRule(row.id, val)"
                />
              </template>
            </el-table-column>
            <el-table-column label="操作" width="150" fixed="right">
              <template #default="{ row }">
                <el-button type="primary" size="small" @click="showEditRuleDialog(row)">编辑</el-button>
                <el-button type="danger" size="small" @click="deleteRule(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <!-- 处理预警对话框 -->
    <el-dialog v-model="handleDialogVisible" title="处理预警" width="400px">
      <el-form :model="handleForm" label-width="80px">
        <el-form-item label="处理说明">
          <el-input
            v-model="handleForm.note"
            type="textarea"
            :rows="3"
            placeholder="请输入处理说明"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="handleDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleRecord" :loading="handling">确认处理</el-button>
      </template>
    </el-dialog>

    <!-- 创建/编辑规则对话框 -->
    <el-dialog v-model="ruleDialogVisible" :title="isEditRule ? '编辑规则' : '新建规则'" width="500px">
      <el-form :model="ruleForm" :rules="ruleRules" ref="ruleFormRef" label-width="100px">
        <el-form-item label="规则名称" prop="name">
          <el-input v-model="ruleForm.name" placeholder="请输入规则名称" />
        </el-form-item>
        <el-form-item label="预警类型" prop="type">
          <el-select v-model="ruleForm.type" placeholder="请选择类型" style="width: 100%">
            <el-option label="库存预警" value="inventory" />
            <el-option label="销售预警" value="sales" />
            <el-option label="订单预警" value="order" />
            <el-option label="客户预警" value="customer" />
          </el-select>
        </el-form-item>
        <el-form-item label="阈值">
          <el-input-number v-model="ruleForm.threshold" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="预警级别" prop="level">
          <el-select v-model="ruleForm.level" placeholder="请选择级别" style="width: 100%">
            <el-option label="信息" value="info" />
            <el-option label="警告" value="warning" />
            <el-option label="严重" value="critical" />
          </el-select>
        </el-form-item>
        <el-form-item label="通知方式">
          <el-select v-model="ruleForm.notify_type" placeholder="请选择通知方式" style="width: 100%">
            <el-option label="系统通知" value="system" />
            <el-option label="邮件" value="email" />
            <el-option label="短信" value="sms" />
            <el-option label="Webhook" value="webhook" />
          </el-select>
        </el-form-item>
        <el-form-item label="通知目标">
          <el-input v-model="ruleForm.notify_target" placeholder="邮箱/手机号/Webhook URL" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="ruleForm.description" type="textarea" :rows="2" placeholder="规则描述" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="ruleDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveRule" :loading="savingRule">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Plus, Refresh } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import {
  alertApi,
  type AlertSummary,
  type AlertRecord,
  type AlertRule
} from '@/api/datacenter'

const activeTab = ref('records')

// 预警汇总
const summary = ref<AlertSummary | null>(null)

// 预警记录
const records = ref<AlertRecord[]>([])
const recordsLoading = ref(false)
const recordFilter = reactive({
  level: '',
  status: undefined as number | undefined
})
const recordPagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

// 预警规则
const rules = ref<AlertRule[]>([])
const rulesLoading = ref(false)

// 处理预警
const handleDialogVisible = ref(false)
const handling = ref(false)
const currentRecord = ref<AlertRecord | null>(null)
const handleForm = reactive({
  note: ''
})

// 检查预警
const checking = ref(false)

// 规则表单
const ruleDialogVisible = ref(false)
const isEditRule = ref(false)
const savingRule = ref(false)
const currentRuleId = ref<number | null>(null)
const ruleFormRef = ref<FormInstance>()
const ruleForm = reactive({
  name: '',
  type: '',
  threshold: 0,
  level: 'warning',
  notify_type: 'system',
  notify_target: '',
  description: ''
})
const ruleRules: FormRules = {
  name: [{ required: true, message: '请输入规则名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择预警类型', trigger: 'change' }],
  level: [{ required: true, message: '请选择预警级别', trigger: 'change' }]
}

// 获取预警汇总
const fetchSummary = async () => {
  try {
    const res = await alertApi.getSummary() as any
    summary.value = res.summary
  } catch (error) {
    console.error('获取预警汇总失败', error)
  }
}

// 获取预警记录
const fetchRecords = async () => {
  recordsLoading.value = true
  try {
    const params: any = {
      page: recordPagination.page,
      page_size: recordPagination.pageSize
    }
    if (recordFilter.level) params.level = recordFilter.level
    if (recordFilter.status !== undefined) params.status = recordFilter.status

    const res = await alertApi.listRecords(params) as any
    records.value = res.records || []
    recordPagination.total = res.total || 0
  } catch (error) {
    ElMessage.error('获取预警记录失败')
  } finally {
    recordsLoading.value = false
  }
}

// 获取预警规则
const fetchRules = async () => {
  rulesLoading.value = true
  try {
    const res = await alertApi.listRules() as any
    rules.value = res.rules || []
  } catch (error) {
    ElMessage.error('获取预警规则失败')
  } finally {
    rulesLoading.value = false
  }
}

// 检查预警
const checkAlerts = async () => {
  checking.value = true
  try {
    const res = await alertApi.checkAlerts() as any
    ElMessage.success(`检查完成，新增 ${res.new_count} 条预警`)
    fetchSummary()
    fetchRecords()
  } catch (error) {
    ElMessage.error('检查预警失败')
  } finally {
    checking.value = false
  }
}

// 显示处理对话框
const showHandleDialog = (record: AlertRecord) => {
  currentRecord.value = record
  handleForm.note = ''
  handleDialogVisible.value = true
}

// 处理预警
const handleRecord = async () => {
  if (!currentRecord.value) return

  handling.value = true
  try {
    await alertApi.handleRecord(currentRecord.value.id, handleForm.note)
    ElMessage.success('处理成功')
    handleDialogVisible.value = false
    fetchRecords()
    fetchSummary()
  } catch (error) {
    ElMessage.error('处理失败')
  } finally {
    handling.value = false
  }
}

// 忽略预警
const ignoreRecord = async (record: AlertRecord) => {
  try {
    await ElMessageBox.confirm('确定要忽略该预警吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await alertApi.ignoreRecord(record.id, '手动忽略')
    ElMessage.success('已忽略')
    fetchRecords()
    fetchSummary()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败')
    }
  }
}

// 显示创建规则对话框
const showCreateRuleDialog = () => {
  isEditRule.value = false
  currentRuleId.value = null
  Object.assign(ruleForm, {
    name: '',
    type: '',
    threshold: 0,
    level: 'warning',
    notify_type: 'system',
    notify_target: '',
    description: ''
  })
  ruleDialogVisible.value = true
}

// 显示编辑规则对话框
const showEditRuleDialog = (rule: AlertRule) => {
  isEditRule.value = true
  currentRuleId.value = rule.id
  Object.assign(ruleForm, {
    name: rule.name,
    type: rule.type,
    threshold: rule.threshold,
    level: rule.level,
    notify_type: rule.notify_type,
    notify_target: rule.notify_target,
    description: rule.description
  })
  ruleDialogVisible.value = true
}

// 保存规则
const saveRule = async () => {
  if (!ruleFormRef.value) return

  await ruleFormRef.value.validate(async (valid) => {
    if (!valid) return

    savingRule.value = true
    try {
      if (isEditRule.value && currentRuleId.value) {
        await alertApi.updateRule(currentRuleId.value, ruleForm)
        ElMessage.success('更新成功')
      } else {
        await alertApi.createRule(ruleForm)
        ElMessage.success('创建成功')
      }
      ruleDialogVisible.value = false
      fetchRules()
    } catch (error: any) {
      ElMessage.error(error.response?.data?.error || '保存失败')
    } finally {
      savingRule.value = false
    }
  })
}

// 删除规则
const deleteRule = async (rule: AlertRule) => {
  try {
    await ElMessageBox.confirm('确定要删除该规则吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await alertApi.deleteRule(rule.id)
    ElMessage.success('删除成功')
    fetchRules()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 切换规则状态
const toggleRule = async (id: number, enabled: boolean) => {
  try {
    await alertApi.toggleRule(id, enabled ? 1 : 0)
    ElMessage.success(enabled ? '已启用' : '已停用')
    fetchRules()
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

// 辅助函数
const getLevelType = (level: string) => {
  switch (level) {
    case 'critical': return 'danger'
    case 'warning': return 'warning'
    case 'info': return 'info'
    default: return 'info'
  }
}

const getLevelLabel = (level: string) => {
  switch (level) {
    case 'critical': return '严重'
    case 'warning': return '警告'
    case 'info': return '信息'
    default: return level
  }
}

const getStatusType = (status: number) => {
  switch (status) {
    case 0: return 'warning'
    case 1: return 'success'
    case 2: return 'info'
    default: return 'info'
  }
}

const getStatusLabel = (status: number) => {
  switch (status) {
    case 0: return '未处理'
    case 1: return '已处理'
    case 2: return '已忽略'
    default: return '未知'
  }
}

const getTypeLabel = (type: string) => {
  switch (type) {
    case 'inventory': return '库存预警'
    case 'sales': return '销售预警'
    case 'order': return '订单预警'
    case 'customer': return '客户预警'
    default: return type
  }
}

const getNotifyTypeLabel = (type: string) => {
  switch (type) {
    case 'system': return '系统通知'
    case 'email': return '邮件'
    case 'sms': return '短信'
    case 'webhook': return 'Webhook'
    default: return type
  }
}

const formatTime = (time: string) => {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

onMounted(() => {
  fetchSummary()
  fetchRecords()
  fetchRules()
})
</script>

<style scoped>
.alert-management-view {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.filter-form {
  margin-bottom: 20px;
}

.text-success {
  color: #67c23a;
}

.text-danger {
  color: #f56c6c;
}
</style>
