<template>
  <div class="customer-analysis-view">
    <!-- 日期筛选 -->
    <el-card class="filter-card">
      <el-form :inline="true">
        <el-form-item label="日期范围">
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="-"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            style="width: 260px"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchData">分析</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- KPI 卡片 -->
    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="6">
        <el-card shadow="hover">
          <el-statistic title="总客户数" :value="analysis?.total_customers || 0" />
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <el-statistic title="新增客户" :value="analysis?.new_customers || 0" />
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <el-statistic title="活跃客户" :value="analysis?.active_customers || 0" />
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <el-statistic title="复购率" :value="analysis?.repurchase_rate || 0" :precision="2" suffix="%" />
        </el-card>
      </el-col>
    </el-row>

    <!-- 图表区域 -->
    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>客户价值分布</span>
          </template>
          <div ref="valueChartRef" style="height: 300px"></div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>地域分布</span>
          </template>
          <div ref="geoChartRef" style="height: 300px"></div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 地域明细 -->
    <el-card style="margin-top: 20px">
      <template #header>
        <span>地域销售明细</span>
      </template>
      <el-table :data="geoDistribution" v-loading="geoLoading" stripe>
        <el-table-column prop="province" label="省份" width="150" />
        <el-table-column prop="city" label="城市" width="150" />
        <el-table-column prop="order_count" label="订单数" width="120" align="right" />
        <el-table-column label="销售额" width="150" align="right">
          <template #default="{ row }">¥{{ row.order_amount?.toFixed(2) || '0.00' }}</template>
        </el-table-column>
        <el-table-column prop="customer_count" label="客户数" width="120" align="right" />
        <el-table-column label="占比" width="120" align="right">
          <template #default="{ row }">{{ row.percentage?.toFixed(2) || '0.00' }}%</template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import * as echarts from 'echarts'
import {
  customerAnalysisApi,
  type CustomerAnalysis,
  type CustomerValueDistribution,
  type GeographyDistribution
} from '@/api/datacenter'

const dateRange = ref<string[]>([])
const analysis = ref<CustomerAnalysis | null>(null)
const valueDistribution = ref<CustomerValueDistribution[]>([])
const geoDistribution = ref<GeographyDistribution[]>([])
const geoLoading = ref(false)

const valueChartRef = ref<HTMLElement>()
const geoChartRef = ref<HTMLElement>()
let valueChart: echarts.ECharts | null = null
let geoChart: echarts.ECharts | null = null

// 获取客户分析
const fetchAnalysis = async () => {
  try {
    const [start, end] = dateRange.value || []
    const res = await customerAnalysisApi.getAnalysis(start, end) as any
    analysis.value = res.analysis
  } catch (error) {
    ElMessage.error('获取客户分析失败')
  }
}

// 获取客户价值分布
const fetchValueDistribution = async () => {
  try {
    const res = await customerAnalysisApi.getValueDistribution() as any
    valueDistribution.value = res.distribution || []
    updateValueChart()
  } catch (error) {
    ElMessage.error('获取价值分布失败')
  }
}

// 获取地域分布
const fetchGeoDistribution = async () => {
  geoLoading.value = true
  try {
    const [start, end] = dateRange.value || []
    const res = await customerAnalysisApi.getGeography(start, end) as any
    geoDistribution.value = res.distribution || []
    updateGeoChart()
  } catch (error) {
    ElMessage.error('获取地域分布失败')
  } finally {
    geoLoading.value = false
  }
}

// 获取所有数据
const fetchData = async () => {
  await Promise.all([fetchAnalysis(), fetchValueDistribution(), fetchGeoDistribution()])
}

// 初始化价值分布图表
const initValueChart = () => {
  if (!valueChartRef.value) return
  valueChart = echarts.init(valueChartRef.value)
}

// 更新价值分布图表
const updateValueChart = () => {
  if (!valueChart || !valueDistribution.value.length) return

  const data = valueDistribution.value.map(item => {
    let name = item.value_level
    switch (item.value_level) {
      case 'high_value':
        name = '高价值(≥1万)'
        break
      case 'medium_value':
        name = '中价值(≥1千)'
        break
      case 'low_value':
        name = '低价值(<1千)'
        break
      case 'no_purchase':
        name = '未购买'
        break
    }
    return { value: item.customer_count, name }
  })

  const option = {
    tooltip: {
      trigger: 'item',
      formatter: '{b}: {c}人 ({d}%)'
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
          show: true,
          formatter: '{b}\n{c}人'
        },
        data
      }
    ]
  }
  valueChart.setOption(option)
}

// 初始化地域分布图表
const initGeoChart = () => {
  if (!geoChartRef.value) return
  geoChart = echarts.init(geoChartRef.value)
}

// 更新地域分布图表
const updateGeoChart = () => {
  if (!geoChart || !geoDistribution.value.length) return

  const data = geoDistribution.value.slice(0, 10).map(item => ({
    value: item.order_amount,
    name: item.province
  }))

  const option = {
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      formatter: (params: any) => {
        const item = params[0]
        return `${item.name}<br/>销售额: ¥${item.value?.toFixed(2)}`
      }
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'value',
      axisLabel: {
        formatter: '¥{value}'
      }
    },
    yAxis: {
      type: 'category',
      data: data.map(d => d.name).reverse()
    },
    series: [
      {
        type: 'bar',
        data: data.map(d => d.value).reverse(),
        itemStyle: {
          color: '#409eff',
          borderRadius: [0, 4, 4, 0]
        }
      }
    ]
  }
  geoChart.setOption(option)
}

// 窗口大小变化
const handleResize = () => {
  valueChart?.resize()
  geoChart?.resize()
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

  fetchData()
  initValueChart()
  initGeoChart()

  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  valueChart?.dispose()
  geoChart?.dispose()
})
</script>

<style scoped>
.customer-analysis-view {
  padding: 20px;
}

.filter-card {
  margin-bottom: 0;
}
</style>
