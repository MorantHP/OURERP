<template>
  <div class="income-expense-view">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>收支管理</span>
          <el-button type="primary" @click="showCreateDialog">新建记录</el-button>
        </div>
      </template>

      <!-- 搜索栏 -->
      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="类型">
          <el-select v-model="searchForm.type" placeholder="全部" clearable style="width: 120px">
            <el-option label="收入" value="income" />
            <el-option label="支出" value="expense" />
          </el-select>
        </el-form-item>
        <el-form-item label="日期范围">
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="-"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            style="width: 240px"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 120px">
            <el-option label="待审核" :value="0" />
            <el-option label="已审核" :value="1" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchRecords">查询</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>

      <!-- 统计卡片 -->
      <el-row :gutter="20" class="stat-cards">
        <el-col :span="8">
          <el-statistic title="总收入" :value="stats.totalIncome" :precision="2" prefix="¥" />
        </el-col>
        <el-col :span="8">
          <el-statistic title="总支出" :value="stats.totalExpense" :precision="2" prefix="¥" />
        </el-col>
        <el-col :span="8">
          <el-statistic title="净收入" :value="stats.netIncome" :precision="2" prefix="¥" />
        </el-col>
      </el-row>

      <!-- 数据表格 -->
      <el-table :data="records" v-loading="loading" stripe style="margin-top: 20px">
        <el-table-column prop="record_date" label="日期" width="120" />
        <el-table-column label="类型" width="80">
          <template #default="{ row }">
            <el-tag :type="row.type === 'income' ? 'success' : 'danger'">
              {{ row.type === 'income' ? '收入' : '支出' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="category" label="分类" width="120" />
        <el-table-column prop="description" label="描述" min-width="200" />
        <el-table-column label="金额" width="120" align="right">
          <template #default="{ row }">
            <span :class="row.type === 'income' ? 'text-success' : 'text-danger'">
              {{ row.type === 'income' ? '+' : '-' }}¥{{ row.amount.toFixed(2) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="voucher_no" label="凭证号" width="120" />
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'warning'">
              {{ row.status === 1 ? '已审核' : '待审核' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 0"
              type="success"
              size="small"
              @click="handleApprove(row)"
            >
              审核
            </el-button>
            <el-button type="primary" size="small" @click="showEditDialog(row)">编辑</el-button>
            <el-button type="danger" size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next"
        @size-change="fetchRecords"
        @current-change="fetchRecords"
        style="margin-top: 20px; justify-content: flex-end"
      />
    </el-card>

    <!-- 创建/编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑记录' : '新建记录'" width="500px">
      <el-form :model="recordForm" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="类型" prop="type">
          <el-radio-group v-model="recordForm.type">
            <el-radio label="income">收入</el-radio>
            <el-radio label="expense">支出</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="分类" prop="category">
          <el-input v-model="recordForm.category" placeholder="请输入分类" />
        </el-form-item>
        <el-form-item label="金额" prop="amount">
          <el-input-number v-model="recordForm.amount" :precision="2" :min="0" style="width: 100%" />
        </el-form-item>
        <el-form-item label="日期" prop="record_date">
          <el-date-picker v-model="recordForm.record_date" type="date" placeholder="选择日期" value-format="YYYY-MM-DD" style="width: 100%" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="recordForm.description" type="textarea" :rows="2" placeholder="请输入描述" />
        </el-form-item>
        <el-form-item label="凭证号">
          <el-input v-model="recordForm.voucher_no" placeholder="请输入凭证号" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { financeRecordApi, type FinanceRecord } from '@/api/finance'

const loading = ref(false)
const records = ref<FinanceRecord[]>([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const saving = ref(false)
const currentId = ref<number | null>(null)
const formRef = ref<FormInstance>()
const dateRange = ref<string[]>([])

const searchForm = reactive({
  type: '',
  status: undefined as number | undefined
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const recordForm = reactive({
  type: 'income',
  category: '',
  amount: 0,
  record_date: '',
  description: '',
  voucher_no: ''
})

const rules: FormRules = {
  type: [{ required: true, message: '请选择类型', trigger: 'change' }],
  category: [{ required: true, message: '请输入分类', trigger: 'blur' }],
  amount: [{ required: true, message: '请输入金额', trigger: 'blur' }],
  record_date: [{ required: true, message: '请选择日期', trigger: 'change' }]
}

// 统计数据
const stats = computed(() => {
  let totalIncome = 0
  let totalExpense = 0
  records.value.forEach(r => {
    if (r.status === 1) {
      if (r.type === 'income') {
        totalIncome += r.amount
      } else {
        totalExpense += r.amount
      }
    }
  })
  return {
    totalIncome,
    totalExpense,
    netIncome: totalIncome - totalExpense
  }
})

// 获取记录列表
const fetchRecords = async () => {
  loading.value = true
  try {
    const params: any = {
      page: pagination.page,
      page_size: pagination.pageSize
    }
    if (searchForm.type) params.type = searchForm.type
    if (searchForm.status !== undefined) params.status = searchForm.status
    if (dateRange.value && dateRange.value.length === 2) {
      params.start_date = dateRange.value[0]
      params.end_date = dateRange.value[1]
    }

    const res = await financeRecordApi.list(params) as any
    records.value = res.records || []
    pagination.total = res.total || 0
  } catch (error) {
    ElMessage.error('获取记录列表失败')
  } finally {
    loading.value = false
  }
}

// 重置搜索
const resetSearch = () => {
  searchForm.type = ''
  searchForm.status = undefined
  dateRange.value = []
  fetchRecords()
}

// 显示创建对话框
const showCreateDialog = () => {
  isEdit.value = false
  currentId.value = null
  recordForm.type = 'income'
  recordForm.category = ''
  recordForm.amount = 0
  recordForm.record_date = new Date().toISOString().split('T')[0]
  recordForm.description = ''
  recordForm.voucher_no = ''
  dialogVisible.value = true
}

// 显示编辑对话框
const showEditDialog = (record: FinanceRecord) => {
  isEdit.value = true
  currentId.value = record.id
  recordForm.type = record.type
  recordForm.category = record.category
  recordForm.amount = record.amount
  recordForm.record_date = record.record_date
  recordForm.description = record.description
  recordForm.voucher_no = record.voucher_no
  dialogVisible.value = true
}

// 保存记录
const handleSave = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    saving.value = true
    try {
      if (isEdit.value && currentId.value) {
        await financeRecordApi.update(currentId.value, recordForm)
        ElMessage.success('更新成功')
      } else {
        await financeRecordApi.create(recordForm)
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      fetchRecords()
    } catch (error: any) {
      ElMessage.error(error.response?.data?.error || '保存失败')
    } finally {
      saving.value = false
    }
  })
}

// 审核记录
const handleApprove = async (record: FinanceRecord) => {
  try {
    await ElMessageBox.confirm('确定要审核该记录吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await financeRecordApi.approve(record.id)
    ElMessage.success('审核成功')
    fetchRecords()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '审核失败')
    }
  }
}

// 删除记录
const handleDelete = async (record: FinanceRecord) => {
  try {
    await ElMessageBox.confirm('确定要删除该记录吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await financeRecordApi.delete(record.id)
    ElMessage.success('删除成功')
    fetchRecords()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '删除失败')
    }
  }
}

onMounted(() => {
  fetchRecords()
})
</script>

<style scoped>
.income-expense-view {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.search-form {
  margin-bottom: 20px;
}

.stat-cards {
  margin-top: 20px;
  text-align: center;
}

.text-success {
  color: #67c23a;
}

.text-danger {
  color: #f56c6c;
}
</style>
