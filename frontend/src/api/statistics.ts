import request from '@/utils/request'

// 总览统计
export interface OverviewStats {
  order_count: number
  sales_amount: number
  pay_amount: number
  item_count: number
  avg_order_value: number
}

export interface GrowthStats {
  order_count_growth: number
  sales_amount_growth: number
  pay_amount_growth: number
}

export interface OverviewResponse {
  today: OverviewStats
  yesterday: OverviewStats
  this_week: OverviewStats
  this_month: OverviewStats
  growth: GrowthStats
}

// 趋势数据
export interface TrendDataPoint {
  date: string
  order_count: number
  sales_amount: number
  pay_amount: number
}

export interface TrendResponse {
  period: string
  data: TrendDataPoint[]
}

// 维度数据
export interface DimensionData {
  key: string
  label: string
  order_count: number
  sales_amount: number
  pay_amount: number
  percentage: number
}

export interface DimensionResponse {
  dimension: string
  data: DimensionData[]
}

// 订单漏斗
export interface FunnelStep {
  name: string
  count: number
  percent: number
  status: string
  status_int: number
}

export interface FunnelResponse {
  data: FunnelStep[]
}

// 热销商品
export interface TopProduct {
  sku_id: number
  sku_name: string
  quantity: number
  sales_amount: number
  order_count: number
}

export interface TopProductsResponse {
  data: TopProduct[]
}

// 查询参数
export interface StatsParams {
  start_date?: string
  end_date?: string
  platform?: string
  shop_id?: number
  category?: string
  brand?: string
}

export const statisticsApi = {
  // 总览统计
  getOverview(params?: StatsParams) {
    return request.get<OverviewResponse>('/statistics/overview', { params })
  },

  // 销售趋势
  getSalesTrend(params?: StatsParams & { period?: string }) {
    return request.get<TrendResponse>('/statistics/sales-trend', { params })
  },

  // 按平台统计
  getByPlatform(params?: StatsParams) {
    return request.get<DimensionResponse>('/statistics/by-platform', { params })
  },

  // 按店铺统计
  getByShop(params?: StatsParams) {
    return request.get<DimensionResponse>('/statistics/by-shop', { params })
  },

  // 按品类统计
  getByCategory(params?: StatsParams) {
    return request.get<DimensionResponse>('/statistics/by-category', { params })
  },

  // 按品牌统计
  getByBrand(params?: StatsParams) {
    return request.get<DimensionResponse>('/statistics/by-brand', { params })
  },

  // 订单漏斗
  getOrderFunnel(params?: StatsParams) {
    return request.get<FunnelResponse>('/statistics/order-funnel', { params })
  },

  // 热销商品
  getTopProducts(params?: StatsParams & { limit?: number }) {
    return request.get<TopProductsResponse>('/statistics/top-products', { params })
  }
}
