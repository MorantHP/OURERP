<template>
  <div class="realtime-monitor-view">
    <!-- KPI 卡片 -->
    <el-row :gutter="20" class="kpi-row">
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card">
          <div class="kpi-content">
            <div class="kpi-icon" style="background: #409eff">
              <el-icon><Document /></el-icon>
            </div>
            <div class="kpi-info">
              <div class="kpi-value">{{ overview?.today_order_count || 0 }}</div>
              <div class="kpi-label">今日订单</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card">
          <div class="kpi-content">
            <div class="kpi-icon" style="background: #67c23a">
              <el-icon><Money /></el-icon>
            </div>
            <div class="kpi-info">
              <div class="kpi-value">¥{{ formatNumber(overview?.today_order_amount || 0) }}</div>
              <div class="kpi-label">今日销售额</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card">
          <div class="kpi-content">
            <div class="kpi-icon" style="background: #e6a23c">
              <el-icon><Wallet /></el-icon>
            </div>
            <div class="kpi-info">
              <div class="kpi-value">¥{{ formatNumber(overview?.today_paid_amount || 0) }}</div>
              <div class="kpi-label">今日实收</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card">
          <div class="kpi-content">
            <div class="kpi-icon" style="background: #f56c6c">
              <el-icon><User /></el-icon>
            </div>
            <div class="kpi-info">
              <div class="kpi-value">{{ overview?.today_new_customers || 0 }}</div>
              <div class="kpi-label">新客户</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 第二行 KPI -->
    <el-row :gutter="20" class="kpi-row">
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card small">
          <div class="kpi-content">
            <div class="kpi-info">
              <div class="kpi-label">待处理订单</div>
              <div class="kpi-value small warning">{{ overview?.pending_orders || 0 }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card small">
          <div class="kpi-content">
            <div class="kpi-info">
              <div class="kpi-label">今日退款</div>
              <div class="kpi-value small danger">¥{{ formatNumber(overview?.today_refund_amount || 0) }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card small">
          <div class="kpi-content">
            <div class="kpi-info">
              <div class="kpi-label">库存预警</div>
              <div class="kpi-value small" :class="overview?.low_stock_items > 0 ? 'danger' : ''">
                {{ overview?.low_stock_items || 0 }}
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="kpi-card small">
          <div class="kpi-content">
            <div class="kpi-info">
              <div class="kpi-label">待处理预警</div>
              <div class="kpi-value small" :class="overview?.unhandled_alerts > 0 ? 'danger' : ''">
                {{ overview?.unhandled_alerts || 0 }}
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 图表区域 -->
    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="16">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>今日销售趋势</span>
              <el-button type="primary" size="small" @click="refreshData" :loading="loading">
                <el-icon><Refresh /></el-icon>
                刷新
              </el-button>
            </div>
          </template>
          <div ref="trendChartRef" style="height: 300px"></div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card>
          <template #header>
            <span>库存状态</span>
          </template>
          <div ref="inventoryChartRef" style="height: 300px"></div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { Document, Money, Wallet, User, Refresh } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import * as echarts from 'echarts'
import { realtimeApi, type RealtimeOverview, type InventoryStatus } from '@/api/datacenter'

const loading = ref(false)
const overview = ref<RealtimeOverview | null>(null)
const inventoryStatus = ref<InventoryStatus | null>(null)
const trendChartRef = ref<HTMLElement>()
const inventoryChartRef = ref<HTMLElement>()
let trendChart: echarts.ECharts | null = null
let inventoryChart: echarts.ECharts | null = null
let refreshTimer: number | null = null

// 格式化数字
const formatNumber = (num: number) => {
  if (num >= 10000) {
    return (num / 10000).toFixed(2) + '万'
  }
  return num.toFixed(2)
}

// 获取实时数据
const fetchOverview = async () => {
  try {
    const res = await realtimeApi.getOverview() as any
    overview.value = res.overview
  } catch (error) {
    console.error('获取实时数据失败', error)
  }
}

// 获取库存状态
const fetchInventoryStatus = async () => {
  try {
    const res = await realtimeApi.getInventory() as any
    inventoryStatus.value = res.status
    updateInventoryChart()
  } catch (error) {
    console.error('获取库存状态失败', error)
  }
}

// 刷新数据
const refreshData = async () => {
  loading.value = true
  try {
    await Promise.all([fetchOverview(), fetchInventoryStatus()])
    ElMessage.success('数据已刷新')
  } finally {
    loading.value = false
  }
}

// 初始化趋势图表
const initTrendChart = () => {
  if (!trendChartRef.value) return

  trendChart = echarts.init(trendChartRef.value)
  const option = {
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'cross' }
    },
    legend: {
      data: ['订单数', '销售额']
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: Array.from({ length: 24 }, (_, i) => `${i}:00`)
    },
    yAxis: [
      {
        type: 'value',
        name: '订单数',
        position: 'left'
      },
      {
        type: 'value',
        name: '销售额',
        position: 'right'
      }
    ],
    series: [
      {
        name: '订单数',
        type: 'line',
        smooth: true,
        data: Array(24).fill(0),
        areaStyle: { opacity: 0.3 }
      },
      {
        name: '销售额',
        type: 'line',
        smooth: true,
        yAxisIndex: 1,
        data: Array(24).fill(0),
        areaStyle: { opacity: 0.3 }
      }
    ]
  }
  trendChart.setOption(option)
}

// 初始化库存图表
const initInventoryChart = () => {
  if (!inventoryChartRef.value) return

  inventoryChart = echarts.init(inventoryChartRef.value)
  updateInventoryChart()
}

// 更新库存图表
const updateInventoryChart = () => {
  if (!inventoryChart || !inventoryStatus.value) return

  const data = inventoryStatus.value.level_distribution || []
  const option = {
    tooltip: {
      trigger: 'item',
      formatter: '{b}: {c} ({d}%)'
    },
    legend: {
      bottom: '5%',
      left: 'center'
    },
    series: [
      {
        type: 'pie',
        radius: ['40%', '70%'],
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 10,
          borderColor: '#fff',
          borderWidth: 2
        },
        label: {
          show: false,
          position: 'center'
        },
        emphasis: {
          label: {
            show: true,
            fontSize: 16,
            fontWeight: 'bold'
          }
        },
        labelLine: {
          show: false
        },
        data: data.map((item: any) => {
          let color = '#67c23a'
          let name = '正常'
          switch (item.level) {
            case 'out_of_stock':
              color = '#f56c6c'
              name = '缺货'
              break
            case 'low':
              color = '#e6a23c'
              name = '低库存'
              break
            case 'normal':
              color = '#67c23a'
              name = '正常'
              break
            case 'high':
              color = '#409eff'
              name = '高库存'
              break
          }
          return { value: item.count, name, itemStyle: { color } }
        })
      }
    ]
  }
  inventoryChart.setOption(option)
}

// 窗口大小变化时重新调整图表
const handleResize = () => {
  trendChart?.resize()
  inventoryChart?.resize()
}

onMounted(() => {
  fetchOverview()
  fetchInventoryStatus()
  initTrendChart()
  initInventoryChart()

  window.addEventListener('resize', handleResize)

  // 自动刷新，每60秒
  refreshTimer = window.setInterval(() => {
    fetchOverview()
    fetchInventoryStatus()
  }, 60000)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
  trendChart?.dispose()
  inventoryChart?.dispose()
})
</script>

<style scoped>
.realtime-monitor-view {
  padding: 20px;
}

.kpi-row {
  margin-bottom: 20px;
}

.kpi-card {
  border-radius: 8px;
}

.kpi-card.small {
  padding: 10px 0;
}

.kpi-content {
  display: flex;
  align-items: center;
  padding: 10px;
}

.kpi-icon {
  width: 60px;
  height: 60px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  color: #fff;
  margin-right: 15px;
}

.kpi-card.small .kpi-icon {
  width: 40px;
  height: 40px;
  font-size: 20px;
}

.kpi-info {
  flex: 1;
}

.kpi-value {
  font-size: 28px;
  font-weight: bold;
  color: #303133;
}

.kpi-value.small {
  font-size: 22px;
}

.kpi-value.warning {
  color: #e6a23c;
}

.kpi-value.danger {
  color: #f56c6c;
}

.kpi-label {
  font-size: 14px;
  color: #909399;
  margin-top: 5px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
