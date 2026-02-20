<template>
  <div class="dashboard">
    <!-- 筛选栏 -->
    <el-card class="filter-card">
      <el-form :inline="true" :model="filterForm" class="filter-form">
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
            style="width: 240px"
          />
        </el-form-item>
        <el-form-item label="平台">
          <el-select v-model="filterForm.platform" placeholder="全部平台" clearable style="width: 120px">
            <el-option v-for="p in platforms" :key="p.code" :label="p.name" :value="p.code" />
          </el-select>
        </el-form-item>
        <el-form-item label="店铺">
          <el-select v-model="filterForm.shop_id" placeholder="全部店铺" clearable style="width: 150px">
            <el-option v-for="s in shops" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 指标卡片 -->
    <div class="stats-cards" v-loading="loading">
      <el-card class="stat-card">
        <div class="stat-content">
          <div class="stat-value">{{ formatNumber(overview?.today?.order_count || 0) }}</div>
          <div class="stat-label">今日订单</div>
          <div class="stat-growth" :class="growthClass(overview?.growth?.order_count_growth)">
            {{ formatGrowth(overview?.growth?.order_count_growth) }}
          </div>
        </div>
        <el-icon class="stat-icon"><Document /></el-icon>
      </el-card>
      <el-card class="stat-card">
        <div class="stat-content">
          <div class="stat-value">{{ formatMoney(overview?.today?.sales_amount || 0) }}</div>
          <div class="stat-label">今日销售额</div>
          <div class="stat-growth" :class="growthClass(overview?.growth?.sales_amount_growth)">
            {{ formatGrowth(overview?.growth?.sales_amount_growth) }}
          </div>
        </div>
        <el-icon class="stat-icon"><Money /></el-icon>
      </el-card>
      <el-card class="stat-card">
        <div class="stat-content">
          <div class="stat-value">{{ formatMoney(overview?.today?.pay_amount || 0) }}</div>
          <div class="stat-label">今日实付</div>
          <div class="stat-growth" :class="growthClass(overview?.growth?.pay_amount_growth)">
            {{ formatGrowth(overview?.growth?.pay_amount_growth) }}
          </div>
        </div>
        <el-icon class="stat-icon"><Wallet /></el-icon>
      </el-card>
      <el-card class="stat-card">
        <div class="stat-content">
          <div class="stat-value">{{ formatMoney(overview?.today?.avg_order_value || 0) }}</div>
          <div class="stat-label">客单价</div>
        </div>
        <el-icon class="stat-icon"><TrendCharts /></el-icon>
      </el-card>
    </div>

    <!-- 销售趋势图 -->
    <el-card class="chart-card">
      <template #header>
        <div class="card-header">
          <span>销售趋势</span>
          <el-radio-group v-model="trendPeriod" size="small" @change="fetchTrend">
            <el-radio-button label="day">近30天</el-radio-button>
            <el-radio-button label="week">近7天</el-radio-button>
            <el-radio-button label="month">近1月</el-radio-button>
          </el-radio-group>
        </div>
      </template>
      <v-chart class="chart" :option="trendOption" autoresize />
    </el-card>

    <!-- 图表行 -->
    <el-row :gutter="20">
      <!-- 平台销售占比 -->
      <el-col :span="12">
        <el-card class="chart-card">
          <template #header>
            <span>平台销售占比</span>
          </template>
          <v-chart class="chart pie-chart" :option="platformOption" autoresize />
        </el-card>
      </el-col>
      <!-- 店铺销售排行 -->
      <el-col :span="12">
        <el-card class="chart-card">
          <template #header>
            <span>店铺销售排行</span>
          </template>
          <v-chart class="chart" :option="shopOption" autoresize />
        </el-card>
      </el-col>
    </el-row>

    <!-- 第二行图表 -->
    <el-row :gutter="20">
      <!-- 品类销售分析 -->
      <el-col :span="12">
        <el-card class="chart-card">
          <template #header>
            <span>品类销售分析</span>
          </template>
          <v-chart class="chart" :option="categoryOption" autoresize />
        </el-card>
      </el-col>
      <!-- 品牌销售排行 -->
      <el-col :span="12">
        <el-card class="chart-card">
          <template #header>
            <span>品牌销售排行</span>
          </template>
          <el-table :data="brandData" max-height="300" stripe>
            <el-table-column prop="label" label="品牌" />
            <el-table-column prop="sales_amount" label="销售额" align="right">
              <template #default="{ row }">
                {{ formatMoney(row.sales_amount) }}
              </template>
            </el-table-column>
            <el-table-column prop="percentage" label="占比" align="right" width="100">
              <template #default="{ row }">
                {{ row.percentage?.toFixed(1) }}%
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <!-- 第三行 -->
    <el-row :gutter="20">
      <!-- 订单漏斗 -->
      <el-col :span="12">
        <el-card class="chart-card">
          <template #header>
            <span>订单漏斗</span>
          </template>
          <v-chart class="chart" :option="funnelOption" autoresize />
        </el-card>
      </el-col>
      <!-- 热销商品TOP10 -->
      <el-col :span="12">
        <el-card class="chart-card">
          <template #header>
            <span>热销商品 TOP10</span>
          </template>
          <el-table :data="topProducts" max-height="300" stripe>
            <el-table-column type="index" label="#" width="40" />
            <el-table-column prop="sku_name" label="商品名称" show-overflow-tooltip />
            <el-table-column prop="quantity" label="销量" align="right" width="80" />
            <el-table-column prop="sales_amount" label="销售额" align="right">
              <template #default="{ row }">
                {{ formatMoney(row.sales_amount) }}
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, PieChart, BarChart, FunnelChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent
} from 'echarts/components'
import VChart from 'vue-echarts'
import { Document, Money, Wallet, TrendCharts } from '@element-plus/icons-vue'
import { statisticsApi, type OverviewResponse, type TrendResponse, type DimensionResponse, type FunnelResponse, type TopProduct } from '@/api/statistics'
import { shopApi, platformApi, type Shop, type Platform } from '@/api/shop'

// 注册 ECharts 组件
use([
  CanvasRenderer,
  LineChart,
  PieChart,
  BarChart,
  FunnelChart,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent
])

// 筛选表单
const filterForm = reactive({
  platform: '',
  shop_id: undefined as number | undefined,
  category: '',
  brand: ''
})
const dateRange = ref<[string, string] | null>(null)

// 数据
const loading = ref(false)
const overview = ref<OverviewResponse | null>(null)
const trendData = ref<TrendResponse | null>(null)
const platformData = ref<DimensionResponse | null>(null)
const shopData = ref<DimensionResponse | null>(null)
const categoryData = ref<DimensionResponse | null>(null)
const brandData = ref<DimensionResponse['data']>([])
const funnelData = ref<FunnelResponse | null>(null)
const topProducts = ref<TopProduct[]>([])

const shops = ref<Shop[]>([])
const platforms = ref<Platform[]>([])
const trendPeriod = ref('day')

// 获取筛选参数
const getFilterParams = () => {
  const params: Record<string, any> = {}
  if (dateRange.value) {
    params.start_date = dateRange.value[0]
    params.end_date = dateRange.value[1]
  }
  if (filterForm.platform) params.platform = filterForm.platform
  if (filterForm.shop_id) params.shop_id = filterForm.shop_id
  return params
}

// 获取总览数据
const fetchOverview = async () => {
  try {
    const res = await statisticsApi.getOverview(getFilterParams())
    overview.value = res as any
  } catch (error) {
    console.error('Failed to fetch overview:', error)
  }
}

// 获取趋势数据
const fetchTrend = async () => {
  try {
    const res = await statisticsApi.getSalesTrend({
      ...getFilterParams(),
      period: trendPeriod.value
    })
    trendData.value = res as any
  } catch (error) {
    console.error('Failed to fetch trend:', error)
  }
}

// 获取平台数据
const fetchPlatform = async () => {
  try {
    const res = await statisticsApi.getByPlatform(getFilterParams())
    platformData.value = res as any
  } catch (error) {
    console.error('Failed to fetch platform:', error)
  }
}

// 获取店铺数据
const fetchShop = async () => {
  try {
    const res = await statisticsApi.getByShop(getFilterParams())
    shopData.value = res as any
  } catch (error) {
    console.error('Failed to fetch shop:', error)
  }
}

// 获取品类数据
const fetchCategory = async () => {
  try {
    const res = await statisticsApi.getByCategory(getFilterParams())
    categoryData.value = res as any
  } catch (error) {
    console.error('Failed to fetch category:', error)
  }
}

// 获取品牌数据
const fetchBrand = async () => {
  try {
    const res = await statisticsApi.getByBrand(getFilterParams())
    brandData.value = (res as any).data || []
  } catch (error) {
    console.error('Failed to fetch brand:', error)
  }
}

// 获取漏斗数据
const fetchFunnel = async () => {
  try {
    const res = await statisticsApi.getOrderFunnel(getFilterParams())
    funnelData.value = res as any
  } catch (error) {
    console.error('Failed to fetch funnel:', error)
  }
}

// 获取热销商品
const fetchTopProducts = async () => {
  try {
    const res = await statisticsApi.getTopProducts({ ...getFilterParams(), limit: 10 })
    topProducts.value = (res as any).data || []
  } catch (error) {
    console.error('Failed to fetch top products:', error)
  }
}

// 获取店铺和平台列表
const fetchFilters = async () => {
  try {
    const [shopRes, platformRes] = await Promise.all([
      shopApi.list(),
      platformApi.list()
    ])
    shops.value = (shopRes as any).list || []
    platforms.value = (platformRes as any).platforms || []
  } catch (error) {
    console.error('Failed to fetch filters:', error)
  }
}

// 获取所有数据
const fetchAll = async () => {
  loading.value = true
  try {
    await Promise.all([
      fetchOverview(),
      fetchTrend(),
      fetchPlatform(),
      fetchShop(),
      fetchCategory(),
      fetchBrand(),
      fetchFunnel(),
      fetchTopProducts()
    ])
  } finally {
    loading.value = false
  }
}

// 搜索和重置
const handleSearch = () => {
  fetchAll()
}

const handleReset = () => {
  dateRange.value = null
  filterForm.platform = ''
  filterForm.shop_id = undefined
  filterForm.category = ''
  filterForm.brand = ''
  fetchAll()
}

// 格式化函数
const formatNumber = (num: number) => {
  return num.toLocaleString('zh-CN')
}

const formatMoney = (num: number) => {
  return '¥' + num.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

const formatGrowth = (growth: number | undefined) => {
  if (growth === undefined || growth === null) return '-'
  const percent = (growth * 100).toFixed(1)
  return growth >= 0 ? `↑ ${percent}%` : `↓ ${Math.abs(parseFloat(percent))}%`
}

const growthClass = (growth: number | undefined) => {
  if (growth === undefined || growth === null) return ''
  return growth >= 0 ? 'positive' : 'negative'
}

// ECharts 选项
const trendOption = computed(() => ({
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
    data: trendData.value?.data?.map(d => d.date) || []
  },
  yAxis: [
    { type: 'value', name: '订单数' },
    { type: 'value', name: '销售额' }
  ],
  series: [
    {
      name: '订单数',
      type: 'line',
      data: trendData.value?.data?.map(d => d.order_count) || [],
      smooth: true,
      areaStyle: { opacity: 0.3 }
    },
    {
      name: '销售额',
      type: 'line',
      yAxisIndex: 1,
      data: trendData.value?.data?.map(d => d.sales_amount) || [],
      smooth: true,
      areaStyle: { opacity: 0.3 }
    }
  ]
}))

const platformOption = computed(() => ({
  tooltip: {
    trigger: 'item',
    formatter: '{b}: ¥{c} ({d}%)'
  },
  legend: {
    orient: 'vertical',
    left: 'left'
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
      data: (platformData.value?.data || []).map(d => ({
        value: d.sales_amount,
        name: d.label
      }))
    }
  ]
}))

const shopOption = computed(() => ({
  tooltip: {
    trigger: 'axis',
    axisPointer: { type: 'shadow' }
  },
  grid: {
    left: '3%',
    right: '4%',
    bottom: '3%',
    containLabel: true
  },
  xAxis: {
    type: 'value'
  },
  yAxis: {
    type: 'category',
    data: (shopData.value?.data || []).map(d => d.label).reverse()
  },
  series: [
    {
      type: 'bar',
      data: (shopData.value?.data || []).map(d => d.sales_amount).reverse(),
      itemStyle: {
        color: '#409EFF',
        borderRadius: [0, 4, 4, 0]
      }
    }
  ]
}))

const categoryOption = computed(() => ({
  tooltip: {
    trigger: 'axis',
    axisPointer: { type: 'shadow' }
  },
  grid: {
    left: '3%',
    right: '4%',
    bottom: '3%',
    containLabel: true
  },
  xAxis: {
    type: 'category',
    data: (categoryData.value?.data || []).map(d => d.label),
    axisLabel: {
      rotate: 30
    }
  },
  yAxis: {
    type: 'value'
  },
  series: [
    {
      type: 'bar',
      data: (categoryData.value?.data || []).map(d => d.sales_amount),
      itemStyle: {
        color: '#67C23A',
        borderRadius: [4, 4, 0, 0]
      }
    }
  ]
}))

const funnelOption = computed(() => ({
  tooltip: {
    trigger: 'item',
    formatter: '{b}: {c} ({d}%)'
  },
  series: [
    {
      type: 'funnel',
      left: '10%',
      top: 30,
      bottom: 30,
      width: '80%',
      min: 0,
      max: 100,
      minSize: '20%',
      maxSize: '100%',
      sort: 'descending',
      gap: 2,
      label: {
        show: true,
        position: 'inside'
      },
      labelLine: {
        length: 10,
        lineStyle: {
          width: 1,
          type: 'solid'
        }
      },
      itemStyle: {
        borderColor: '#fff',
        borderWidth: 1
      },
      emphasis: {
        label: {
          fontSize: 14
        }
      },
      data: (funnelData.value?.data || []).map(d => ({
        value: d.percent,
        name: d.name
      }))
    }
  ]
}))

onMounted(() => {
  fetchFilters()
  fetchAll()
})
</script>

<style scoped>
.dashboard {
  padding: 20px;
}

.filter-card {
  margin-bottom: 20px;
}

.filter-form {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.stats-cards {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  margin-bottom: 20px;
}

.stat-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 28px;
  font-weight: bold;
  color: #303133;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-top: 5px;
}

.stat-growth {
  font-size: 12px;
  margin-top: 5px;
}

.stat-growth.positive {
  color: #67C23A;
}

.stat-growth.negative {
  color: #F56C6C;
}

.stat-icon {
  font-size: 48px;
  color: #409EFF;
  opacity: 0.3;
}

.chart-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.chart {
  height: 300px;
  width: 100%;
}

.pie-chart {
  height: 280px;
}

@media (max-width: 1200px) {
  .stats-cards {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .stats-cards {
    grid-template-columns: 1fr;
  }
}
</style>
