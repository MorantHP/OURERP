<template>
  <div class="compare-analysis-view">
    <!-- 标签页 -->
    <el-tabs v-model="activeTab">
      <el-tab-pane label="同比分析" name="yoy">
        <ComparePanel
          title="同比分析(与去年同期对比)"
          :loading="yoyLoading"
          :result="yoyResult"
          @compare="handleYOY"
        />
      </el-tab-pane>

      <el-tab-pane label="环比分析" name="mom">
        <ComparePanel
          title="环比分析(与上个周期对比)"
          :loading="momLoading"
          :result="momResult"
          @compare="handleMOM"
        />
      </el-tab-pane>

      <el-tab-pane label="自定义对比" name="custom">
        <el-card>
          <template #header>
            <span>自定义期间对比</span>
          </template>
          <el-form :inline="true">
            <el-form-item label="当前期间">
              <el-date-picker
                v-model="currentPeriod"
                type="daterange"
                range-separator="-"
                start-placeholder="开始日期"
                end-placeholder="结束日期"
                value-format="YYYY-MM-DD"
                style="width: 260px"
              />
            </el-form-item>
            <el-form-item label="对比期间">
              <el-date-picker
                v-model="comparePeriod"
                type="daterange"
                range-separator="-"
                start-placeholder="开始日期"
                end-placeholder="结束日期"
                value-format="YYYY-MM-DD"
                style="width: 260px"
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="handleCustomCompare" :loading="customLoading">
                对比分析
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>

        <CompareResult
          v-if="customResult"
          :result="customResult"
          style="margin-top: 20px"
        />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, defineComponent, h } from 'vue'
import { ElMessage } from 'element-plus'
import {
  compareAnalysisApi,
  type PeriodCompareResult
} from '@/api/datacenter'

// 对比面板组件
const ComparePanel = defineComponent({
  name: 'ComparePanel',
  props: {
    title: String,
    loading: Boolean,
    result: Object as () => PeriodCompareResult | null
  },
  emits: ['compare'],
  setup(props, { emit }) {
    const dateRange = ref<string[]>([])

    const handleCompare = () => {
      if (!dateRange.value || dateRange.value.length !== 2) {
        ElMessage.warning('请选择日期范围')
        return
      }
      emit('compare', dateRange.value[0], dateRange.value[1])
    }

    return () => h('div', [
      h('div', { class: 'filter-section' }, [
        h('el-card', [
          h('el-form', { inline: true }, [
            h('el-form-item', { label: '日期范围' }, [
              h('el-date-picker', {
                modelValue: dateRange.value,
                'onUpdate:modelValue': (val: string[]) => { dateRange.value = val },
                type: 'daterange',
                rangeSeparator: '-',
                startPlaceholder: '开始日期',
                endPlaceholder: '结束日期',
                valueFormat: 'YYYY-MM-DD',
                style: 'width: 260px'
              })
            ]),
            h('el-form-item', [
              h('el-button', {
                type: 'primary',
                loading: props.loading,
                onClick: handleCompare
              }, '对比分析')
            ])
          ])
        ])
      ]),
      props.result && h(CompareResult, { result: props.result, style: 'margin-top: 20px' })
    ])
  }
})

// 对比结果组件
const CompareResult = defineComponent({
  name: 'CompareResult',
  props: {
    result: Object as () => PeriodCompareResult
  },
  setup(props) {
    const getChangeClass = (change: number) => {
      if (change > 0) return 'text-success'
      if (change < 0) return 'text-danger'
      return ''
    }

    const getChangeIcon = (change: number) => {
      if (change > 0) return '↑'
      if (change < 0) return '↓'
      return '→'
    }

    return () => {
      if (!props.result) return null

      return h('div', [
        h('el-row', { gutter: 20 }, [
          h('el-col', { span: 24 }, [
            h('el-card', [
              h('template', { '#header': () => '对比结果' }, []),
              h('el-row', { gutter: 20 }, props.result.metrics.map((metric: any) =>
                h('el-col', { span: 8, key: metric.metric_type }, [
                  h('el-card', { shadow: 'hover', class: 'metric-card' }, [
                    h('div', { class: 'metric-title' },
                      metric.metric_type === 'orders' ? '订单数' : '销售额'
                    ),
                    h('div', { class: 'metric-comparison' }, [
                      h('div', { class: 'metric-item' }, [
                        h('div', { class: 'metric-label' }, '当前期间'),
                        h('div', { class: 'metric-value' },
                          metric.metric_type === 'orders'
                            ? Math.round(metric.current_value)
                            : '¥' + metric.current_value?.toFixed(2)
                        )
                      ]),
                      h('div', { class: 'metric-divider' }, 'vs'),
                      h('div', { class: 'metric-item' }, [
                        h('div', { class: 'metric-label' }, '对比期间'),
                        h('div', { class: 'metric-value' },
                          metric.metric_type === 'orders'
                            ? Math.round(metric.compare_value)
                            : '¥' + metric.compare_value?.toFixed(2)
                        )
                      ])
                    ]),
                    h('div', { class: 'metric-change' }, [
                      h('span', { class: getChangeClass(metric.change_value) }, [
                        getChangeIcon(metric.change_value),
                        metric.change_value > 0 ? '+' : '',
                        metric.metric_type === 'orders'
                          ? Math.round(metric.change_value)
                          : '¥' + metric.change_value?.toFixed(2),
                        ' (',
                        metric.change_rate > 0 ? '+' : '',
                        metric.change_rate?.toFixed(2),
                        '%)'
                      ])
                    ])
                  ])
                ])
              ))
            ])
          ])
        ])
      ])
    }
  }
})

const activeTab = ref('yoy')
const dateRange = ref<string[]>([])

// 同比
const yoyLoading = ref(false)
const yoyResult = ref<PeriodCompareResult | null>(null)

// 环比
const momLoading = ref(false)
const momResult = ref<PeriodCompareResult | null>(null)

// 自定义
const currentPeriod = ref<string[]>([])
const comparePeriod = ref<string[]>([])
const customLoading = ref(false)
const customResult = ref<PeriodCompareResult | null>(null)

// 同比分析
const handleYOY = async (startDate: string, endDate: string) => {
  yoyLoading.value = true
  try {
    const res = await compareAnalysisApi.yoyCompare(startDate, endDate) as any
    yoyResult.value = res
  } catch (error) {
    ElMessage.error('同比分析失败')
  } finally {
    yoyLoading.value = false
  }
}

// 环比分析
const handleMOM = async (startDate: string, endDate: string) => {
  momLoading.value = true
  try {
    const res = await compareAnalysisApi.momCompare(startDate, endDate) as any
    momResult.value = res
  } catch (error) {
    ElMessage.error('环比分析失败')
  } finally {
    momLoading.value = false
  }
}

// 自定义对比
const handleCustomCompare = async () => {
  if (!currentPeriod.value || currentPeriod.value.length !== 2) {
    ElMessage.warning('请选择当前期间')
    return
  }
  if (!comparePeriod.value || comparePeriod.value.length !== 2) {
    ElMessage.warning('请选择对比期间')
    return
  }

  customLoading.value = true
  try {
    const res = await compareAnalysisApi.periodCompare(
      currentPeriod.value[0],
      currentPeriod.value[1],
      comparePeriod.value[0],
      comparePeriod.value[1]
    ) as any
    customResult.value = res
  } catch (error) {
    ElMessage.error('对比分析失败')
  } finally {
    customLoading.value = false
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
})
</script>

<style scoped>
.compare-analysis-view {
  padding: 20px;
}

.filter-section {
  margin-bottom: 20px;
}

.metric-card {
  text-align: center;
}

.metric-title {
  font-size: 16px;
  font-weight: bold;
  color: #303133;
  margin-bottom: 15px;
}

.metric-comparison {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 20px;
}

.metric-item {
  text-align: center;
}

.metric-label {
  font-size: 12px;
  color: #909399;
  margin-bottom: 5px;
}

.metric-value {
  font-size: 20px;
  font-weight: bold;
  color: #303133;
}

.metric-divider {
  font-size: 14px;
  color: #909399;
}

.metric-change {
  margin-top: 15px;
  font-size: 14px;
}

.text-success {
  color: #67c23a;
}

.text-danger {
  color: #f56c6c;
}
</style>
