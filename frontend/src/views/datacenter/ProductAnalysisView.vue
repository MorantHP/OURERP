<template>
  <div class="product-analysis-view">
    <!-- 标签页 -->
    <el-tabs v-model="activeTab">
      <el-tab-pane label="库存水位" name="inventory">
        <!-- 库存汇总 -->
        <el-row :gutter="20">
          <el-col :span="6">
            <el-card shadow="hover">
              <el-statistic title="总商品数" :value="summary?.total_products || 0" />
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card shadow="hover">
              <el-statistic title="缺货商品" :value="summary?.out_of_stock_count || 0">
                <template #suffix>
                  <span class="text-danger">({{ summary?.out_of_stock_count || 0 }})</span>
                </template>
              </el-statistic>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card shadow="hover">
              <el-statistic title="低库存商品" :value="summary?.low_stock_count || 0">
                <template #suffix>
                  <span class="text-warning">({{ summary?.low_stock_count || 0 }})</span>
                </template>
              </el-statistic>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card shadow="hover">
              <el-statistic title="库存总数量" :value="summary?.total_quantity || 0" />
            </el-card>
          </el-col>
        </el-row>

        <!-- 库存水位表格 -->
        <el-card style="margin-top: 20px">
          <template #header>
            <div class="card-header">
              <span>库存水位明细</span>
              <el-input
                v-model="searchKeyword"
                placeholder="搜索商品名称"
                style="width: 200px"
                clearable
                @clear="filterLevels"
                @keyup.enter="filterLevels"
              >
                <template #append>
                  <el-button @click="filterLevels">搜索</el-button>
                </template>
              </el-input>
            </div>
          </template>
          <el-table :data="filteredLevels" v-loading="levelsLoading" stripe>
            <el-table-column prop="product_name" label="商品名称" min-width="200" />
            <el-table-column prop="quantity" label="当前库存" width="120" align="right" />
            <el-table-column prop="min_quantity" label="最低库存" width="120" align="right" />
            <el-table-column prop="max_quantity" label="最高库存" width="120" align="right" />
            <el-table-column label="库存状态" width="120">
              <template #default="{ row }">
                <el-tag :type="getStockLevelType(row.stock_level)">
                  {{ getStockLevelLabel(row.stock_level) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="days_of_stock" label="可售天数" width="100" align="right" />
            <el-table-column prop="suggestion" label="建议" min-width="150" />
          </el-table>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="动销分析" name="turnover">
        <!-- 日期筛选 -->
        <el-card>
          <el-form :inline="true">
            <el-form-item label="日期范围">
              <el-date-picker
                v-model="turnoverDateRange"
                type="daterange"
                range-separator="-"
                start-placeholder="开始日期"
                end-placeholder="结束日期"
                value-format="YYYY-MM-DD"
                style="width: 260px"
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="fetchTurnover">分析</el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <!-- 动销率表格 -->
        <el-card style="margin-top: 20px">
          <template #header>
            <span>商品动销率</span>
          </template>
          <el-table :data="turnoverData" v-loading="turnoverLoading" stripe>
            <el-table-column prop="product_name" label="商品名称" min-width="200" />
            <el-table-column prop="sales_quantity" label="销售数量" width="120" align="right" />
            <el-table-column prop="stock_quantity" label="库存数量" width="120" align="right" />
            <el-table-column label="动销率" width="150" align="right">
              <template #default="{ row }">
                <el-progress
                  :percentage="Math.min(row.turnover_rate, 100)"
                  :color="getTurnoverColor(row.turnover_rate)"
                  :format="() => row.turnover_rate?.toFixed(1) + '%'"
                />
              </template>
            </el-table-column>
            <el-table-column label="动销状态" width="100">
              <template #default="{ row }">
                <el-tag :type="getTurnoverType(row.status)">
                  {{ getTurnoverLabel(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="进货建议" name="purchase">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>进货策略建议</span>
              <el-button type="primary" @click="fetchPurchaseStrategy">刷新</el-button>
            </div>
          </template>
          <el-table :data="purchaseStrategies" v-loading="purchaseLoading" stripe>
            <el-table-column prop="product_name" label="商品名称" min-width="200" />
            <el-table-column prop="current_stock" label="当前库存" width="100" align="right" />
            <el-table-column prop="avg_daily_sales" label="日均销量" width="100" align="right">
              <template #default="{ row }">{{ row.avg_daily_sales?.toFixed(1) || 0 }}</template>
            </el-table-column>
            <el-table-column prop="estimated_days" label="可售天数" width="100" align="right" />
            <el-table-column prop="suggested_qty" label="建议采购量" width="120" align="right">
              <template #default="{ row }">
                <span :class="{ 'text-danger': row.suggested_qty > 0 }">
                  {{ row.suggested_qty }}
                </span>
              </template>
            </el-table-column>
            <el-table-column label="优先级" width="100">
              <template #default="{ row }">
                <el-tag :type="getPriorityType(row.priority)">
                  {{ getPriorityLabel(row.priority) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="safety_stock" label="安全库存" width="100" align="right" />
          </el-table>
        </el-card>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import {
  productAnalysisApi,
  type InventoryLevel,
  type ProductTurnover,
  type PurchaseStrategy,
  type InventorySummary
} from '@/api/datacenter'

const activeTab = ref('inventory')

// 库存水位
const levels = ref<InventoryLevel[]>([])
const filteredLevels = ref<InventoryLevel[]>([])
const levelsLoading = ref(false)
const searchKeyword = ref('')
const summary = ref<InventorySummary | null>(null)

// 动销分析
const turnoverDateRange = ref<string[]>([])
const turnoverData = ref<ProductTurnover[]>([])
const turnoverLoading = ref(false)

// 进货建议
const purchaseStrategies = ref<PurchaseStrategy[]>([])
const purchaseLoading = ref(false)

// 获取库存水位
const fetchLevels = async () => {
  levelsLoading.value = true
  try {
    const res = await productAnalysisApi.getInventoryLevel() as any
    levels.value = res.levels || []
    filteredLevels.value = levels.value
  } catch (error) {
    ElMessage.error('获取库存水位失败')
  } finally {
    levelsLoading.value = false
  }
}

// 获取库存汇总
const fetchSummary = async () => {
  try {
    const res = await productAnalysisApi.getInventorySummary() as any
    summary.value = res.summary
  } catch (error) {
    console.error('获取库存汇总失败', error)
  }
}

// 筛选库存水位
const filterLevels = () => {
  if (!searchKeyword.value) {
    filteredLevels.value = levels.value
  } else {
    filteredLevels.value = levels.value.filter(item =>
      item.product_name?.toLowerCase().includes(searchKeyword.value.toLowerCase())
    )
  }
}

// 获取动销数据
const fetchTurnover = async () => {
  turnoverLoading.value = true
  try {
    const [start, end] = turnoverDateRange.value || []
    const res = await productAnalysisApi.getTurnover(start, end, 100) as any
    turnoverData.value = res.turnover || []
  } catch (error) {
    ElMessage.error('获取动销数据失败')
  } finally {
    turnoverLoading.value = false
  }
}

// 获取进货策略
const fetchPurchaseStrategy = async () => {
  purchaseLoading.value = true
  try {
    const res = await productAnalysisApi.getPurchaseStrategy(30) as any
    purchaseStrategies.value = res.strategies || []
  } catch (error) {
    ElMessage.error('获取进货策略失败')
  } finally {
    purchaseLoading.value = false
  }
}

// 获取库存状态类型
const getStockLevelType = (level: string) => {
  switch (level) {
    case 'out_of_stock': return 'danger'
    case 'low': return 'warning'
    case 'normal': return 'success'
    case 'high': return 'info'
    default: return 'info'
  }
}

// 获取库存状态标签
const getStockLevelLabel = (level: string) => {
  switch (level) {
    case 'out_of_stock': return '缺货'
    case 'low': return '低库存'
    case 'normal': return '正常'
    case 'high': return '高库存'
    default: return level
  }
}

// 获取动销率颜色
const getTurnoverColor = (rate: number) => {
  if (rate >= 80) return '#67c23a'
  if (rate >= 40) return '#409eff'
  if (rate > 0) return '#e6a23c'
  return '#f56c6c'
}

// 获取动销状态类型
const getTurnoverType = (status: string) => {
  switch (status) {
    case 'high': return 'success'
    case 'medium': return 'primary'
    case 'low': return 'warning'
    case 'stagnant': return 'danger'
    default: return 'info'
  }
}

// 获取动销状态标签
const getTurnoverLabel = (status: string) => {
  switch (status) {
    case 'high': return '高动销'
    case 'medium': return '中动销'
    case 'low': return '低动销'
    case 'stagnant': return '滞销'
    default: return status
  }
}

// 获取优先级类型
const getPriorityType = (priority: string) => {
  switch (priority) {
    case 'urgent': return 'danger'
    case 'high': return 'warning'
    case 'medium': return 'primary'
    case 'low': return 'info'
    default: return 'info'
  }
}

// 获取优先级标签
const getPriorityLabel = (priority: string) => {
  switch (priority) {
    case 'urgent': return '紧急'
    case 'high': return '高'
    case 'medium': return '中'
    case 'low': return '低'
    default: return priority
  }
}

onMounted(() => {
  // 设置默认日期
  const end = new Date()
  const start = new Date()
  start.setDate(start.getDate() - 30)
  turnoverDateRange.value = [
    start.toISOString().split('T')[0],
    end.toISOString().split('T')[0]
  ]

  fetchLevels()
  fetchSummary()
  fetchTurnover()
  fetchPurchaseStrategy()
})
</script>

<style scoped>
.product-analysis-view {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.text-danger {
  color: #f56c6c;
}

.text-warning {
  color: #e6a23c;
}
</style>
