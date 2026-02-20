import request from '@/utils/request'

// =============== 类型定义 ===============

// 实时概览
export interface RealtimeOverview {
  today_order_count: number
  today_order_amount: number
  today_paid_amount: number
  today_refund_count: number
  today_refund_amount: number
  today_new_customers: number
  pending_orders: number
  shipped_orders: number
  completed_orders: number
  cancelled_orders: number
  low_stock_items: number
  out_of_stock_items: number
  unhandled_alerts: number
  avg_order_value: number
  snapshot_time: string
}

// 库存状态
export interface InventoryStatus {
  total_products: number
  normal_products: number
  low_stock_products: number
  out_of_stock_products: number
  stock_value: number
  level_distribution: LevelDistribution[]
}

export interface LevelDistribution {
  level: string
  count: number
}

// 小时趋势
export interface HourlyTrend {
  hour: number
  order_count: number
  order_amount: number
}

// 客户分析
export interface CustomerAnalysis {
  total_customers: number
  new_customers: number
  active_customers: number
  return_customers: number
  repurchase_rate: number
  activation_rate: number
  avg_order_value: number
  avg_customer_value: number
  total_revenue: number
  vip_count: number
  normal_count: number
  new_count: number
}

// 客户价值分布
export interface CustomerValueDistribution {
  value_level: string
  customer_count: number
  total_amount: number
  percentage: number
}

// 地域分布
export interface GeographyDistribution {
  province: string
  city: string
  order_count: number
  order_amount: number
  customer_count: number
  percentage: number
}

// 商品动销率
export interface ProductTurnover {
  product_id: number
  product_name: string
  sales_quantity: number
  stock_quantity: number
  turnover_rate: number
  status: string // high/medium/low/stagnant
}

// 库存水位
export interface InventoryLevel {
  product_id: number
  product_name: string
  quantity: number
  min_quantity: number
  max_quantity: number
  stock_level: string // out_of_stock/low/normal/high
  days_of_stock: number
  suggestion: string
}

// 进货策略
export interface PurchaseStrategy {
  product_id: number
  product_name: string
  current_stock: number
  avg_daily_sales: number
  suggested_qty: number
  priority: string // urgent/high/medium/low
  estimated_days: number
  safety_stock: number
}

// 库存汇总
export interface InventorySummary {
  total_products: number
  total_quantity: number
  total_value: number
  out_of_stock_count: number
  low_stock_count: number
  normal_stock_count: number
  high_stock_count: number
  avg_turnover_days: number
}

// 期间对比
export interface PeriodCompare {
  metric_type: string
  current_value: number
  compare_value: number
  change_value: number
  change_rate: number
  current_period: string
  compare_period: string
}

export interface PeriodCompareResult {
  current_period: { start: string; end: string }
  compare_period: { start: string; end: string }
  metrics: PeriodCompare[]
  order_trend: TrendData[]
  amount_trend: TrendData[]
}

export interface TrendData {
  date: string
  current: number
  compare: number
}

// 店铺对比
export interface ShopCompare {
  shop_id: number
  shop_name: string
  order_count: number
  order_amount: number
  profit_amount: number
  profit_rate: number
  avg_order_value: number
  percentage: number
}

// 平台对比
export interface PlatformCompare {
  platform: string
  shop_count: number
  order_count: number
  order_amount: number
  profit_amount: number
  profit_rate: number
  percentage: number
}

// 预警规则
export interface AlertRule {
  id: number
  tenant_id: number
  name: string
  type: string
  condition: string
  threshold: number
  threshold_min: number
  notify_type: string
  notify_target: string
  level: string
  status: number
  description: string
  created_by: number
  created_at: string
  updated_at: string
}

// 预警记录
export interface AlertRecord {
  id: number
  tenant_id: number
  rule_id: number
  rule?: AlertRule
  title: string
  content: string
  level: string
  source_type: string
  source_id: number
  status: number // 0-未处理 1-已处理 2-已忽略
  handled_by?: number
  handled_at?: string
  handle_note: string
  created_at: string
}

// 预警汇总
export interface AlertSummary {
  total_alerts: number
  unhandled_alerts: number
  critical_count: number
  warning_count: number
  info_count: number
  today_alerts: number
}

// 预警类型
export interface AlertType {
  type: string
  name: string
  description: string
}

// 预警级别
export interface NotifyLevel {
  level: string
  name: string
  color: string
  description: string
}

// =============== API 函数 ===============

// 实时监控
export const realtimeApi = {
  // 获取实时概览
  getOverview: () => request.get<RealtimeOverview>('/datacenter/realtime/overview'),

  // 获取实时库存状态
  getInventory: () => request.get<InventoryStatus>('/datacenter/realtime/inventory'),

  // 获取小时趋势
  getHourlyTrend: () => request.get<HourlyTrend[]>('/datacenter/realtime/hourly-trend')
}

// 客户分析
export const customerAnalysisApi = {
  // 获取客户分析
  getAnalysis: (startDate?: string, endDate?: string) =>
    request.get<CustomerAnalysis>('/datacenter/customers/analysis', {
      params: { start_date: startDate, end_date: endDate }
    }),

  // 获取客户价值分布
  getValueDistribution: () =>
    request.get<CustomerValueDistribution[]>('/datacenter/customers/value-distribution'),

  // 获取地域分布
  getGeography: (startDate?: string, endDate?: string) =>
    request.get<GeographyDistribution[]>('/datacenter/customers/geography', {
      params: { start_date: startDate, end_date: endDate }
    }),

  // 获取城市分布
  getCity: (province: string, startDate?: string, endDate?: string) =>
    request.get<GeographyDistribution[]>('/datacenter/customers/city', {
      params: { province, start_date: startDate, end_date: endDate }
    }),

  // 获取复购分析
  getRepurchase: (startDate?: string, endDate?: string) =>
    request.get('/datacenter/customers/repurchase', {
      params: { start_date: startDate, end_date: endDate }
    })
}

// 商品分析
export const productAnalysisApi = {
  // 获取商品动销率
  getTurnover: (startDate?: string, endDate?: string, limit?: number) =>
    request.get<ProductTurnover[]>('/datacenter/products/turnover', {
      params: { start_date: startDate, end_date: endDate, limit }
    }),

  // 获取库存水位
  getInventoryLevel: () =>
    request.get<InventoryLevel[]>('/datacenter/products/inventory-level'),

  // 获取进货策略
  getPurchaseStrategy: (days?: number) =>
    request.get<PurchaseStrategy[]>('/datacenter/products/purchase-strategy', {
      params: { days }
    }),

  // 获取低库存商品
  getLowStock: () =>
    request.get<InventoryLevel[]>('/datacenter/products/low-stock'),

  // 获取库存汇总
  getInventorySummary: () =>
    request.get<InventorySummary>('/datacenter/products/inventory-summary')
}

// 对比分析
export const compareAnalysisApi = {
  // 期间对比
  periodCompare: (currentStart: string, currentEnd: string, compareStart: string, compareEnd: string) =>
    request.get<PeriodCompareResult>('/datacenter/compare/period', {
      params: {
        current_start_date: currentStart,
        current_end_date: currentEnd,
        compare_start_date: compareStart,
        compare_end_date: compareEnd
      }
    }),

  // 同比分析
  yoyCompare: (startDate: string, endDate: string) =>
    request.get<PeriodCompareResult>('/datacenter/compare/yoy', {
      params: { start_date: startDate, end_date: endDate }
    }),

  // 环比分析
  momCompare: (startDate: string, endDate: string) =>
    request.get<PeriodCompareResult>('/datacenter/compare/mom', {
      params: { start_date: startDate, end_date: endDate }
    }),

  // 店铺对比
  shopCompare: (shopIds: number[], startDate?: string, endDate?: string) =>
    request.get('/datacenter/compare/shop', {
      params: { shop_ids: shopIds.join(','), start_date: startDate, end_date: endDate }
    }),

  // 平台对比
  platformCompare: (startDate?: string, endDate?: string) =>
    request.get('/datacenter/compare/platform', {
      params: { start_date: startDate, end_date: endDate }
    })
}

// 预警管理
export const alertApi = {
  // 获取预警规则列表
  listRules: (params?: { type?: string; status?: number; page?: number; page_size?: number }) =>
    request.get<{ rules: AlertRule[]; total: number }>('/datacenter/alerts/rules', { params }),

  // 创建预警规则
  createRule: (data: Partial<AlertRule>) =>
    request.post<AlertRule>('/datacenter/alerts/rules', data),

  // 更新预警规则
  updateRule: (id: number, data: Partial<AlertRule>) =>
    request.put<AlertRule>(`/datacenter/alerts/rules/${id}`, data),

  // 删除预警规则
  deleteRule: (id: number) =>
    request.delete(`/datacenter/alerts/rules/${id}`),

  // 启用/停用预警规则
  toggleRule: (id: number, status: number) =>
    request.post(`/datacenter/alerts/rules/${id}/toggle`, null, {
      params: { status }
    }),

  // 获取预警汇总
  getSummary: () =>
    request.get<AlertSummary>('/datacenter/alerts/summary'),

  // 获取预警记录列表
  listRecords: (params?: {
    rule_id?: number
    level?: string
    status?: number
    source_type?: string
    page?: number
    page_size?: number
  }) =>
    request.get<{ records: AlertRecord[]; total: number }>('/datacenter/alerts/records', { params }),

  // 处理预警
  handleRecord: (id: number, note: string) =>
    request.post(`/datacenter/alerts/records/${id}/handle`, { note }),

  // 忽略预警
  ignoreRecord: (id: number, note: string) =>
    request.post(`/datacenter/alerts/records/${id}/ignore`, { note }),

  // 手动检查预警
  checkAlerts: () =>
    request.post<{ alerts: AlertRecord[]; new_count: number }>('/datacenter/alerts/check'),

  // 获取预警类型
  getTypes: () =>
    request.get<{ types: AlertType[] }>('/datacenter/alerts/types'),

  // 获取预警级别
  getLevels: () =>
    request.get<{ levels: NotifyLevel[] }>('/datacenter/alerts/levels')
}
