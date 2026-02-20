<template>
  <div class="order-cost-view">
    <!-- 利润分析卡片 -->
    <el-card class="analysis-card">
      <template #header>
        <div class="card-header">
          <span>利润分析</span>
          <div>
            <el-date-picker
              v-model="dateRange"
              type="daterange"
              range-separator="-"
              start-placeholder="开始日期"
              end-placeholder="结束日期"
              value-format="YYYY-MM-DD"
              style="width: 260px; margin-right: 10px"
            />
            <el-button type="primary" @click="fetchAnalysis">分析</el-button>
          </div>
        </div>
      </template>

      <el-row :gutter="20" v-loading="analysisLoading">
        <el-col :span="6">
          <el-statistic title="销售额" :value="analysis.total_sales" :precision="2" prefix="¥" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="退款金额" :value="analysis.total_refund" :precision="2" prefix="¥" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="实际销售" :value="analysis.net_sales" :precision="2" prefix="¥" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="订单数" :value="analysis.order_count" />
        </el-col>
      </el-row>

      <el-divider />

      <el-row :gutter="20">
        <el-col :span="6">
          <el-statistic title="商品成本" :value="analysis.product_cost" :precision="2" prefix="¥" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="物流成本" :value="analysis.shipping_cost" :precision="2" prefix="¥" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="平台佣金" :value="analysis.commission" :precision="2" prefix="¥" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="总成本" :value="analysis.total_cost" :precision="2" prefix="¥" />
        </el-col>
      </el-row>

      <el-divider />

      <el-row :gutter="20">
        <el-col :span="12">
          <el-statistic title="毛利润" :value="analysis.gross_profit" :precision="2" prefix="¥">
            <template #suffix>
              <span :class="analysis.gross_profit >= 0 ? 'text-success' : 'text-danger'">
                ({{ analysis.profit_rate?.toFixed(2) }}%)
              </span>
            </template>
          </el-statistic>
        </el-col>
        <el-col :span="12">
          <el-progress
            :percentage="Math.min(Math.abs(analysis.profit_rate), 100)"
            :color="analysis.profit_rate >= 0 ? '#67c23a' : '#f56c6c'"
            :stroke-width="20"
            :format="() => `利润率: ${analysis.profit_rate?.toFixed(2)}%`"
          />
        </el-col>
      </el-row>
    </el-card>

    <!-- 订单成本列表 -->
    <el-card style="margin-top: 20px">
      <template #header>
        <div class="card-header">
          <span>订单成本明细</span>
        </div>
      </template>

      <!-- 搜索栏 -->
      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="订单号">
          <el-input v-model="searchForm.order_no" placeholder="请输入订单号" clearable />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 120px">
            <el-option label="待核算" :value="0" />
            <el-option label="已核算" :value="1" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchCosts">查询</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>

      <!-- 数据表格 -->
      <el-table :data="costs" v-loading="loading" stripe>
        <el-table-column prop="order_no" label="订单号" width="180" />
        <el-table-column label="销售金额" width="120" align="right">
          <template #default="{ row }">¥{{ row.sale_amount?.toFixed(2) || '0.00' }}</template>
        </el-table-column>
        <el-table-column label="退款金额" width="120" align="right">
          <template #default="{ row }">¥{{ row.refund_amount?.toFixed(2) || '0.00' }}</template>
        </el-table-column>
        <el-table-column label="实际销售" width="120" align="right">
          <template #default="{ row }">¥{{ row.real_sale_amount?.toFixed(2) || '0.00' }}</template>
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
        <el-table-column label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'warning'">
              {{ row.status === 1 ? '已核算' : '待核算' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 0"
              type="primary"
              size="small"
              @click="handleCalculate(row)"
            >
              核算
            </el-button>
            <span v-else class="text-muted">-</span>
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
        @size-change="fetchCosts"
        @current-change="fetchCosts"
        style="margin-top: 20px; justify-content: flex-end"
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { orderCostApi, type OrderCost, type ProfitAnalysis } from '@/api/finance'

const loading = ref(false)
const analysisLoading = ref(false)
const costs = ref<OrderCost[]>([])
const dateRange = ref<string[]>([])

const searchForm = reactive({
  order_no: '',
  status: undefined as number | undefined
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const analysis = ref<ProfitAnalysis>({
  total_sales: 0,
  total_refund: 0,
  net_sales: 0,
  product_cost: 0,
  shipping_cost: 0,
  commission: 0,
  service_fee: 0,
  promotion_fee: 0,
  other_fee: 0,
  total_cost: 0,
  gross_profit: 0,
  profit_rate: 0,
  order_count: 0
})

// 获取利润分析
const fetchAnalysis = async () => {
  analysisLoading.value = true
  try {
    const params: any = {}
    if (dateRange.value && dateRange.value.length === 2) {
      params.start_date = dateRange.value[0]
      params.end_date = dateRange.value[1]
    }

    const res = await orderCostApi.getProfitAnalysis(params) as any
    analysis.value = res.analysis || analysis.value
  } catch (error) {
    ElMessage.error('获取利润分析失败')
  } finally {
    analysisLoading.value = false
  }
}

// 获取成本列表
const fetchCosts = async () => {
  loading.value = true
  try {
    const params: any = {
      page: pagination.page,
      page_size: pagination.pageSize
    }
    if (searchForm.order_no) params.order_no = searchForm.order_no
    if (searchForm.status !== undefined) params.status = searchForm.status

    const res = await orderCostApi.list(params) as any
    costs.value = res.costs || []
    pagination.total = res.total || 0
  } catch (error) {
    ElMessage.error('获取成本列表失败')
  } finally {
    loading.value = false
  }
}

// 重置搜索
const resetSearch = () => {
  searchForm.order_no = ''
  searchForm.status = undefined
  fetchCosts()
}

// 核算订单成本
const handleCalculate = async (cost: OrderCost) => {
  try {
    await orderCostApi.calculate(cost.order_id)
    ElMessage.success('核算成功')
    fetchCosts()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.error || '核算失败')
  }
}

onMounted(() => {
  // 设置默认日期范围为最近30天
  const end = new Date()
  const start = new Date()
  start.setDate(start.getDate() - 30)
  dateRange.value = [
    start.toISOString().split('T')[0],
    end.toISOString().split('T')[0]
  ]

  fetchAnalysis()
  fetchCosts()
})
</script>

<style scoped>
.order-cost-view {
  padding: 20px;
}

.analysis-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.search-form {
  margin-bottom: 20px;
}

.text-success {
  color: #67c23a;
}

.text-danger {
  color: #f56c6c;
}

.text-muted {
  color: #909399;
}
</style>
