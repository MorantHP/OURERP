<template>
  <div class="monthly-settlement-view">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>月度财务结算</span>
          <el-button type="primary" @click="showGenerateDialog">生成结算</el-button>
        </div>
      </template>

      <!-- 数据表格 -->
      <el-table :data="settlements" v-loading="loading" stripe>
        <el-table-column prop="period" label="结算期间" width="120" />
        <el-table-column label="销售额" width="120" align="right">
          <template #default="{ row }">¥{{ row.total_sales?.toFixed(2) || '0.00' }}</template>
        </el-table-column>
        <el-table-column label="退款" width="120" align="right">
          <template #default="{ row }">¥{{ row.total_refund?.toFixed(2) || '0.00' }}</template>
        </el-table-column>
        <el-table-column label="净销售" width="120" align="right">
          <template #default="{ row }">¥{{ row.net_sales?.toFixed(2) || '0.00' }}</template>
        </el-table-column>
        <el-table-column label="商品成本" width="120" align="right">
          <template #default="{ row }">¥{{ row.product_cost?.toFixed(2) || '0.00' }}</template>
        </el-table-column>
        <el-table-column label="佣金" width="100" align="right">
          <template #default="{ row }">¥{{ row.commission?.toFixed(2) || '0.00' }}</template>
        </el-table-column>
        <el-table-column label="总成本" width="120" align="right">
          <template #default="{ row }">¥{{ row.total_cost?.toFixed(2) || '0.00' }}</template>
        </el-table-column>
        <el-table-column label="毛利" width="120" align="right">
          <template #default="{ row }">
            <span :class="row.gross_profit >= 0 ? 'text-success' : 'text-danger'">
              ¥{{ row.gross_profit?.toFixed(2) || '0.00' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="利润率" width="100" align="right">
          <template #default="{ row }">
            <span :class="row.profit_rate >= 0 ? 'text-success' : 'text-danger'">
              {{ row.profit_rate?.toFixed(2) || '0.00' }}%
            </span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'warning'">
              {{ row.status === 1 ? '已结算' : '待结算' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 0"
              type="success"
              size="small"
              @click="handleConfirm(row)"
            >
              确认
            </el-button>
            <el-button type="primary" size="small" @click="showDetailDialog(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @size-change="fetchSettlements"
        @current-change="fetchSettlements"
        style="margin-top: 20px; justify-content: flex-end"
      />
    </el-card>

    <!-- 生成结算对话框 -->
    <el-dialog v-model="generateDialogVisible" title="生成月度结算" width="400px">
      <el-form :model="generateForm" label-width="100px">
        <el-form-item label="结算期间">
          <el-date-picker
            v-model="generateForm.period"
            type="month"
            placeholder="选择月份"
            value-format="YYYY-MM"
            style="width: 100%"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="generateDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleGenerate" :loading="generating">生成</el-button>
      </template>
    </el-dialog>

    <!-- 详情对话框 -->
    <el-dialog v-model="detailDialogVisible" title="结算详情" width="700px">
      <el-descriptions :column="3" border v-if="currentSettlement">
        <el-descriptions-item label="结算期间">{{ currentSettlement.period }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="currentSettlement.status === 1 ? 'success' : 'warning'">
            {{ currentSettlement.status === 1 ? '已结算' : '待结算' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="结算时间">{{ currentSettlement.settled_at || '-' }}</el-descriptions-item>
      </el-descriptions>

      <el-divider content-position="left">收入</el-divider>
      <el-row :gutter="20" v-if="currentSettlement">
        <el-col :span="8">
          <el-statistic title="销售额" :value="currentSettlement.total_sales" :precision="2" prefix="¥" />
        </el-col>
        <el-col :span="8">
          <el-statistic title="退款" :value="currentSettlement.total_refund" :precision="2" prefix="¥" />
        </el-col>
        <el-col :span="8">
          <el-statistic title="净销售" :value="currentSettlement.net_sales" :precision="2" prefix="¥" />
        </el-col>
      </el-row>

      <el-divider content-position="left">成本</el-divider>
      <el-row :gutter="20" v-if="currentSettlement">
        <el-col :span="6">
          <el-statistic title="商品成本" :value="currentSettlement.product_cost" :precision="2" prefix="¥" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="物流成本" :value="currentSettlement.shipping_cost" :precision="2" prefix="¥" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="佣金" :value="currentSettlement.commission" :precision="2" prefix="¥" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="总成本" :value="currentSettlement.total_cost" :precision="2" prefix="¥" />
        </el-col>
      </el-row>

      <el-divider content-position="left">利润</el-divider>
      <el-row :gutter="20" v-if="currentSettlement">
        <el-col :span="12">
          <el-statistic title="毛利润" :value="currentSettlement.gross_profit" :precision="2" prefix="¥" />
        </el-col>
        <el-col :span="12">
          <el-statistic title="利润率" :value="currentSettlement.profit_rate" :precision="2" suffix="%" />
        </el-col>
      </el-row>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { financialSettlementApi, type FinancialSettlement } from '@/api/finance'

const loading = ref(false)
const settlements = ref<FinancialSettlement[]>([])
const generateDialogVisible = ref(false)
const detailDialogVisible = ref(false)
const generating = ref(false)
const currentSettlement = ref<FinancialSettlement | null>(null)

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const generateForm = reactive({
  period: ''
})

// 获取结算列表
const fetchSettlements = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize
    }

    const res = await financialSettlementApi.listMonthly(params) as any
    settlements.value = res.settlements || []
    pagination.total = res.total || 0
  } catch (error) {
    ElMessage.error('获取结算列表失败')
  } finally {
    loading.value = false
  }
}

// 显示生成对话框
const showGenerateDialog = () => {
  // 默认上个月
  const now = new Date()
  now.setMonth(now.getMonth() - 1)
  generateForm.period = now.toISOString().slice(0, 7)
  generateDialogVisible.value = true
}

// 生成结算
const handleGenerate = async () => {
  if (!generateForm.period) {
    ElMessage.warning('请选择结算期间')
    return
  }

  generating.value = true
  try {
    await financialSettlementApi.generateMonthly(generateForm.period)
    ElMessage.success('生成成功')
    generateDialogVisible.value = false
    fetchSettlements()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '生成失败')
  } finally {
    generating.value = false
  }
}

// 确认结算
const handleConfirm = async (settlement: FinancialSettlement) => {
  try {
    await ElMessageBox.confirm('确定要确认该结算吗？确认后将无法修改。', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await financialSettlementApi.confirmMonthly(settlement.period)
    ElMessage.success('确认成功')
    fetchSettlements()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '确认失败')
    }
  }
}

// 显示详情对话框
const showDetailDialog = (settlement: FinancialSettlement) => {
  currentSettlement.value = settlement
  detailDialogVisible.value = true
}

onMounted(() => {
  fetchSettlements()
})
</script>

<style scoped>
.monthly-settlement-view {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.text-success {
  color: #67c23a;
}

.text-danger {
  color: #f56c6c;
}
</style>
