import request from '@/utils/request'

// ==================== 类型定义 ====================

// 收支记录
export interface FinanceRecord {
  id: number
  tenant_id: number
  type: string // income/expense
  category: string
  amount: number
  currency: string
  shop_id?: number
  order_id?: number
  record_date: string
  description: string
  voucher_no: string
  source: string // manual/sync
  status: number // 0-待审核 1-已审核 2-已取消
  approved_by?: number
  approved_at?: string
  created_by: number
  created_at: string
  updated_at: string
}

// 平台账单
export interface PlatformBill {
  id: number
  tenant_id: number
  shop_id: number
  shop?: any
  bill_no: string
  bill_period: string
  platform: string
  order_amount: number
  refund_amount: number
  commission: number
  service_fee: number
  logistics_fee: number
  promotion_fee: number
  other_fee: number
  settlement_amount: number
  bill_date: string
  reconciled_amount: number
  status: number // 0-待对账 1-部分对账 2-已对账
  reconciled_at?: string
  reconciled_by?: number
  sync_status: number
  synced_at?: string
  created_at: string
  updated_at: string
}

// 账单明细
export interface PlatformBillDetail {
  id: number
  tenant_id: number
  bill_id: number
  shop_id: number
  order_no: string
  order_id?: number
  item_amount: number
  shipping_fee: number
  discount_amount: number
  refund_amount: number
  commission: number
  service_fee: number
  settlement_amount: number
  transaction_time: string
  status: number
  reconciled_at?: string
  created_at: string
}

// 供应商
export interface Supplier {
  id: number
  tenant_id: number
  code: string
  name: string
  contact: string
  phone: string
  email: string
  address: string
  bank_name: string
  bank_account: string
  tax_no: string
  credit_limit: number
  balance: number
  status: number
  remark: string
  created_at: string
  updated_at: string
}

// 采购结算单
export interface PurchaseSettlement {
  id: number
  tenant_id: number
  settlement_no: string
  supplier_id: number
  supplier?: Supplier
  total_amount: number
  paid_amount: number
  discount_amount: number
  adjust_amount: number
  real_amount: number
  settlement_date: string
  due_date?: string
  status: number // 0-待付款 1-部分付款 2-已付款 3-已取消
  payment_method: string
  remark: string
  approved_by?: number
  approved_at?: string
  created_by: number
  created_at: string
  updated_at: string
}

// 采购付款记录
export interface PurchasePayment {
  id: number
  tenant_id: number
  settlement_id: number
  payment_no: string
  amount: number
  payment_date: string
  payment_method: string
  account_no: string
  voucher_no: string
  status: number
  approved_by?: number
  approved_at?: string
  remark: string
  created_by: number
  created_at: string
  updated_at: string
}

// 商品成本
export interface ProductCost {
  id: number
  tenant_id: number
  product_id: number
  product_sku: string
  purchase_cost: number
  shipping_cost: number
  package_cost: number
  other_cost: number
  total_cost: number
  cost_method: string // weighted/fifo/standard
  effective_date: string
  stock_qty: number
  stock_value: number
  created_at: string
  updated_at: string
}

// 订单成本
export interface OrderCost {
  id: number
  tenant_id: number
  order_id: number
  order_no: string
  product_cost: number
  shipping_cost: number
  package_cost: number
  commission: number
  service_fee: number
  promotion_fee: number
  other_fee: number
  total_cost: number
  sale_amount: number
  refund_amount: number
  real_sale_amount: number
  gross_profit: number
  profit_rate: number
  status: number
  calculated_at?: string
  created_at: string
  updated_at: string
}

// 库存成本快照
export interface InventoryCostSnapshot {
  id: number
  tenant_id: number
  warehouse_id: number
  product_id: number
  product_sku: string
  snapshot_date: string
  begin_qty: number
  in_qty: number
  out_qty: number
  end_qty: number
  begin_amount: number
  in_amount: number
  out_amount: number
  end_amount: number
  unit_cost: number
  created_at: string
}

// 财务结算
export interface FinancialSettlement {
  id: number
  tenant_id: number
  settlement_type: string // monthly/yearly
  period: string
  shop_id?: number
  total_sales: number
  total_refund: number
  net_sales: number
  other_income: number
  product_cost: number
  shipping_cost: number
  commission: number
  service_fee: number
  promotion_fee: number
  other_cost: number
  total_cost: number
  gross_profit: number
  profit_rate: number
  inventory_change: number
  status: number // 0-待结算 1-已结算 2-已取消
  settled_at?: string
  settled_by?: number
  created_at: string
  updated_at: string
}

// 结算账户
export interface FinanceBankAccount {
  id: number
  tenant_id: number
  account_name: string
  account_no: string
  bank_name: string
  bank_branch: string
  account_type: string
  currency: string
  balance: number
  status: number
  is_default: boolean
  remark: string
  created_at: string
  updated_at: string
}

// 利润分析
export interface ProfitAnalysis {
  total_sales: number
  total_refund: number
  net_sales: number
  product_cost: number
  shipping_cost: number
  commission: number
  service_fee: number
  promotion_fee: number
  other_fee: number
  total_cost: number
  gross_profit: number
  profit_rate: number
  order_count: number
}

// ==================== 收支记录 API ====================

export const financeRecordApi = {
  list: (params?: any) => request.get('/finance/records', { params }),
  get: (id: number) => request.get(`/finance/records/${id}`),
  create: (data: Partial<FinanceRecord>) => request.post('/finance/records', data),
  update: (id: number, data: Partial<FinanceRecord>) => request.put(`/finance/records/${id}`, data),
  delete: (id: number) => request.delete(`/finance/records/${id}`),
  approve: (id: number) => request.post(`/finance/records/${id}/approve`),
}

// ==================== 平台账单 API ====================

export const platformBillApi = {
  list: (params?: any) => request.get('/finance/bills', { params }),
  get: (id: number) => request.get(`/finance/bills/${id}`),
  create: (data: Partial<PlatformBill>) => request.post('/finance/bills', data),
  getDetails: (id: number) => request.get(`/finance/bills/${id}/details`),
  reconcileDetail: (billId: number, detailId: number, orderId: number) =>
    request.post(`/finance/bills/${billId}/details/${detailId}/reconcile`, { order_id: orderId }),
}

// ==================== 供应商 API ====================

export const supplierApi = {
  list: (params?: any) => request.get('/finance/suppliers', { params }),
  get: (id: number) => request.get(`/finance/suppliers/${id}`),
  create: (data: Partial<Supplier>) => request.post('/finance/suppliers', data),
  update: (id: number, data: Partial<Supplier>) => request.put(`/finance/suppliers/${id}`, data),
  delete: (id: number) => request.delete(`/finance/suppliers/${id}`),
}

// ==================== 采购结算 API ====================

export const purchaseSettlementApi = {
  list: (params?: any) => request.get('/finance/settlements', { params }),
  get: (id: number) => request.get(`/finance/settlements/${id}`),
  create: (data: Partial<PurchaseSettlement>) => request.post('/finance/settlements', data),
  pay: (id: number, data: Partial<PurchasePayment>) => request.post(`/finance/settlements/${id}/pay`, data),
  getPayments: (id: number) => request.get(`/finance/settlements/${id}/payments`),
}

// ==================== 商品成本 API ====================

export const productCostApi = {
  list: (params?: any) => request.get('/finance/product-costs', { params }),
  update: (id: number, data: Partial<ProductCost>) => request.put(`/finance/product-costs/${id}`, data),
  batchUpdate: (costs: Partial<ProductCost>[]) => request.post('/finance/product-costs/batch', { costs }),
}

// ==================== 订单成本 API ====================

export const orderCostApi = {
  list: (params?: any) => request.get('/finance/order-costs', { params }),
  calculate: (id: number) => request.post(`/finance/order-costs/${id}/calculate`),
  getProfitAnalysis: (params?: any) => request.get('/finance/order-costs/profit', { params }),
}

// ==================== 库存成本快照 API ====================

export const inventorySnapshotApi = {
  list: (params?: any) => request.get('/finance/inventory-snapshots', { params }),
  generate: (date: string) => request.post('/finance/inventory-snapshots/generate', { date }),
}

// ==================== 财务结算 API ====================

export const financialSettlementApi = {
  // 月度结算
  listMonthly: (params?: any) => request.get('/finance/monthly-settlements', { params }),
  generateMonthly: (period: string, shopId?: number) =>
    request.post('/finance/monthly-settlements/generate', { period, shop_id: shopId }),
  confirmMonthly: (period: string) => request.post(`/finance/monthly-settlements/${period}/confirm`),

  // 年度结算
  listYearly: (params?: any) => request.get('/finance/yearly-settlements', { params }),
  generateYearly: (year: string) => request.post('/finance/yearly-settlements/generate', { year }),
}

// ==================== 结算账户 API ====================

export const bankAccountApi = {
  list: () => request.get('/finance/bank-accounts'),
  create: (data: Partial<FinanceBankAccount>) => request.post('/finance/bank-accounts', data),
  update: (id: number, data: Partial<FinanceBankAccount>) => request.put(`/finance/bank-accounts/${id}`, data),
  delete: (id: number) => request.delete(`/finance/bank-accounts/${id}`),
}
