package models

import (
	"time"
)

// FinanceRecord 收支记录
type FinanceRecord struct {
	ID          int64      `json:"id" gorm:"primaryKey"`
	TenantID    int64      `json:"tenant_id" gorm:"index;not null"`
	Type        string     `json:"type" gorm:"size:20;not null"`            // income/expense
	Category    string     `json:"category" gorm:"size:50"`                 // 科目分类
	Amount      float64    `json:"amount" gorm:"type:decimal(12,2);not null"`
	Currency    string     `json:"currency" gorm:"size:10;default:'CNY'"`
	ShopID      *int64     `json:"shop_id"`       // 关联店铺
	OrderID     *int64     `json:"order_id"`      // 关联订单
	RecordDate  time.Time  `json:"record_date" gorm:"index"`
	Description string     `json:"description" gorm:"size:500"`
	VoucherNo   string     `json:"voucher_no" gorm:"size:50"`              // 凭证号
	Source      string     `json:"source" gorm:"size:20;default:'manual'"` // manual/sync
	Status      int        `json:"status"`                                  // 0-待审核 1-已审核 2-已取消
	ApprovedBy  *int64     `json:"approved_by"`
	ApprovedAt  *time.Time `json:"approved_at"`
	CreatedBy   int64      `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// TableName specifies the table name for FinanceRecord
func (FinanceRecord) TableName() string {
	return "finance_records"
}

// PlatformBill 平台账单
type PlatformBill struct {
	ID               int64      `json:"id" gorm:"primaryKey"`
	TenantID         int64      `json:"tenant_id" gorm:"index;not null"`
	ShopID           int64      `json:"shop_id" gorm:"index;not null"`
	Shop             *Shop      `json:"shop" gorm:"foreignKey:ShopID"`
	BillNo           string     `json:"bill_no" gorm:"size:50;not null"`
	BillPeriod       string     `json:"bill_period" gorm:"size:20"`         // 2024-01
	Platform         string     `json:"platform" gorm:"size:20"`
	OrderAmount      float64    `json:"order_amount" gorm:"type:decimal(12,2)"`
	RefundAmount     float64    `json:"refund_amount" gorm:"type:decimal(12,2)"`
	Commission       float64    `json:"commission" gorm:"type:decimal(12,2)"`
	ServiceFee       float64    `json:"service_fee" gorm:"type:decimal(12,2)"`
	LogisticsFee     float64    `json:"logistics_fee" gorm:"type:decimal(12,2)"`
	PromotionFee     float64    `json:"promotion_fee" gorm:"type:decimal(12,2)"`
	OtherFee         float64    `json:"other_fee" gorm:"type:decimal(12,2)"`
	SettlementAmount float64    `json:"settlement_amount" gorm:"type:decimal(12,2)"`
	BillDate         time.Time  `json:"bill_date"`
	ReconciledAmount float64    `json:"reconciled_amount" gorm:"type:decimal(12,2)"`
	Status           int        `json:"status"`         // 0-待对账 1-部分对账 2-已对账
	ReconciledAt     *time.Time `json:"reconciled_at"`
	ReconciledBy     *int64     `json:"reconciled_by"`
	SyncStatus       int        `json:"sync_status"`    // 0-未同步 1-同步中 2-已同步 3-同步失败
	SyncedAt         *time.Time `json:"synced_at"`
	RawData          string     `json:"raw_data" gorm:"type:text"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// TableName specifies the table name for PlatformBill
func (PlatformBill) TableName() string {
	return "platform_bills"
}

// PlatformBillDetail 账单明细
type PlatformBillDetail struct {
	ID               int64     `json:"id" gorm:"primaryKey"`
	TenantID         int64     `json:"tenant_id" gorm:"index;not null"`
	BillID           int64     `json:"bill_id" gorm:"index;not null"`
	Bill             *PlatformBill `json:"bill" gorm:"foreignKey:BillID"`
	ShopID           int64     `json:"shop_id" gorm:"index;not null"`
	OrderNo          string    `json:"order_no" gorm:"size:50;index"`
	OrderID          *int64    `json:"order_id"`
	ItemAmount       float64   `json:"item_amount" gorm:"type:decimal(12,2)"`
	ShippingFee      float64   `json:"shipping_fee" gorm:"type:decimal(12,2)"`
	DiscountAmount   float64   `json:"discount_amount" gorm:"type:decimal(12,2)"`
	RefundAmount     float64   `json:"refund_amount" gorm:"type:decimal(12,2)"`
	Commission       float64   `json:"commission" gorm:"type:decimal(12,2)"`
	ServiceFee       float64   `json:"service_fee" gorm:"type:decimal(12,2)"`
	SettlementAmount float64   `json:"settlement_amount" gorm:"type:decimal(12,2)"`
	TransactionTime  time.Time `json:"transaction_time"`
	Status           int       `json:"status"`              // 0-待对账 1-已对账
	ReconciledAt     *time.Time `json:"reconciled_at"`
	CreatedAt        time.Time `json:"created_at"`
}

// TableName specifies the table name for PlatformBillDetail
func (PlatformBillDetail) TableName() string {
	return "platform_bill_details"
}

// Supplier 供应商
type Supplier struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	TenantID    int64     `json:"tenant_id" gorm:"index;not null"`
	Code        string    `json:"code" gorm:"size:20"`
	Name        string    `json:"name" gorm:"size:100;not null"`
	Contact     string    `json:"contact" gorm:"size:50"`
	Phone       string    `json:"phone" gorm:"size:20"`
	Email       string    `json:"email" gorm:"size:50"`
	Address     string    `json:"address" gorm:"size:200"`
	BankName    string    `json:"bank_name" gorm:"size:50"`
	BankAccount string    `json:"bank_account" gorm:"size:30"`
	TaxNo       string    `json:"tax_no" gorm:"size:30"`
	CreditLimit float64   `json:"credit_limit" gorm:"type:decimal(12,2)"`
	Balance     float64   `json:"balance" gorm:"type:decimal(12,2)"`
	Status      int       `json:"status"`             // 1-启用 0-停用
	Remark      string    `json:"remark" gorm:"size:500"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName specifies the table name for Supplier
func (Supplier) TableName() string {
	return "suppliers"
}

// PurchaseSettlement 采购结算单
type PurchaseSettlement struct {
	ID             int64      `json:"id" gorm:"primaryKey"`
	TenantID       int64      `json:"tenant_id" gorm:"index;not null"`
	SettlementNo   string     `json:"settlement_no" gorm:"size:30;uniqueIndex:idx_tenant_settlement"`
	SupplierID     int64      `json:"supplier_id" gorm:"index;not null"`
	Supplier       *Supplier  `json:"supplier" gorm:"foreignKey:SupplierID"`
	TotalAmount    float64    `json:"total_amount" gorm:"type:decimal(12,2)"`
	PaidAmount     float64    `json:"paid_amount" gorm:"type:decimal(12,2)"`
	DiscountAmount float64    `json:"discount_amount" gorm:"type:decimal(12,2)"`
	AdjustAmount   float64    `json:"adjust_amount" gorm:"type:decimal(12,2)"`
	RealAmount     float64    `json:"real_amount" gorm:"type:decimal(12,2)"`
	SettlementDate time.Time  `json:"settlement_date"`
	DueDate        *time.Time `json:"due_date"`
	Status         int        `json:"status"`           // 0-待付款 1-部分付款 2-已付款 3-已取消
	PaymentMethod  string     `json:"payment_method" gorm:"size:20"`
	Remark         string     `json:"remark" gorm:"size:500"`
	ApprovedBy     *int64     `json:"approved_by"`
	ApprovedAt     *time.Time `json:"approved_at"`
	CreatedBy      int64      `json:"created_by"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// TableName specifies the table name for PurchaseSettlement
func (PurchaseSettlement) TableName() string {
	return "purchase_settlements"
}

// PurchaseSettlementDetail 采购结算明细
type PurchaseSettlementDetail struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	TenantID     int64     `json:"tenant_id" gorm:"index;not null"`
	SettlementID int64     `json:"settlement_id" gorm:"index;not null"`
	PurchaseID   int64     `json:"purchase_id" gorm:"index"`
	ProductID    *int64    `json:"product_id"`
	ProductSku   string    `json:"product_sku" gorm:"size:50"`
	ProductName  string    `json:"product_name" gorm:"size:200"`
	Quantity     int       `json:"quantity"`
	UnitPrice    float64   `json:"unit_price" gorm:"type:decimal(12,2)"`
	Amount       float64   `json:"amount" gorm:"type:decimal(12,2)"`
	Remark       string    `json:"remark" gorm:"size:200"`
	CreatedAt    time.Time `json:"created_at"`
}

// TableName specifies the table name for PurchaseSettlementDetail
func (PurchaseSettlementDetail) TableName() string {
	return "purchase_settlement_details"
}

// PurchasePayment 采购付款记录
type PurchasePayment struct {
	ID            int64      `json:"id" gorm:"primaryKey"`
	TenantID      int64      `json:"tenant_id" gorm:"index;not null"`
	SettlementID  int64      `json:"settlement_id" gorm:"index;not null"`
	PaymentNo     string     `json:"payment_no" gorm:"size:30"`
	Amount        float64    `json:"amount" gorm:"type:decimal(12,2)"`
	PaymentDate   time.Time  `json:"payment_date"`
	PaymentMethod string     `json:"payment_method" gorm:"size:20"` // bank/alipay/wechat等
	AccountNo     string     `json:"account_no" gorm:"size:50"`
	VoucherNo     string     `json:"voucher_no" gorm:"size:50"`
	Status        int        `json:"status"`                        // 0-待审核 1-已审核 2-已取消
	ApprovedBy    *int64     `json:"approved_by"`
	ApprovedAt    *time.Time `json:"approved_at"`
	Remark        string     `json:"remark" gorm:"size:500"`
	CreatedBy     int64      `json:"created_by"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// TableName specifies the table name for PurchasePayment
func (PurchasePayment) TableName() string {
	return "purchase_payments"
}

// ProductCost 商品成本
type ProductCost struct {
	ID            int64     `json:"id" gorm:"primaryKey"`
	TenantID      int64     `json:"tenant_id" gorm:"index;not null"`
	ProductID     int64     `json:"product_id" gorm:"index;not null"`
	ProductSku    string    `json:"product_sku" gorm:"size:50"`
	PurchaseCost  float64   `json:"purchase_cost" gorm:"type:decimal(12,2)"`
	ShippingCost  float64   `json:"shipping_cost" gorm:"type:decimal(12,2)"`
	PackageCost   float64   `json:"package_cost" gorm:"type:decimal(12,2)"`
	OtherCost     float64   `json:"other_cost" gorm:"type:decimal(12,2)"`
	TotalCost     float64   `json:"total_cost" gorm:"type:decimal(12,2)"`
	CostMethod    string    `json:"cost_method" gorm:"size:20;default:'weighted'"` // weighted/fifo/standard
	EffectiveDate time.Time `json:"effective_date"`
	StockQty      int       `json:"stock_qty"`
	StockValue    float64   `json:"stock_value" gorm:"type:decimal(12,2)"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TableName specifies the table name for ProductCost
func (ProductCost) TableName() string {
	return "product_costs"
}

// OrderCost 订单成本
type OrderCost struct {
	ID             int64      `json:"id" gorm:"primaryKey"`
	TenantID       int64      `json:"tenant_id" gorm:"index;not null"`
	OrderID        int64      `json:"order_id" gorm:"uniqueIndex:idx_tenant_order_cost"`
	OrderNo        string     `json:"order_no" gorm:"size:50;index"`
	ProductCost    float64    `json:"product_cost" gorm:"type:decimal(12,2)"`
	ShippingCost   float64    `json:"shipping_cost" gorm:"type:decimal(12,2)"`
	PackageCost    float64    `json:"package_cost" gorm:"type:decimal(12,2)"`
	Commission     float64    `json:"commission" gorm:"type:decimal(12,2)"`
	ServiceFee     float64    `json:"service_fee" gorm:"type:decimal(12,2)"`
	PromotionFee   float64    `json:"promotion_fee" gorm:"type:decimal(12,2)"`
	OtherFee       float64    `json:"other_fee" gorm:"type:decimal(12,2)"`
	TotalCost      float64    `json:"total_cost" gorm:"type:decimal(12,2)"`
	SaleAmount     float64    `json:"sale_amount" gorm:"type:decimal(12,2)"`
	RefundAmount   float64    `json:"refund_amount" gorm:"type:decimal(12,2)"`
	RealSaleAmount float64    `json:"real_sale_amount" gorm:"type:decimal(12,2)"`
	GrossProfit    float64    `json:"gross_profit" gorm:"type:decimal(12,2)"`
	ProfitRate     float64    `json:"profit_rate" gorm:"type:decimal(5,2)"`
	Status         int        `json:"status"`                    // 0-待核算 1-已核算
	CalculatedAt   *time.Time `json:"calculated_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// TableName specifies the table name for OrderCost
func (OrderCost) TableName() string {
	return "order_costs"
}

// InventoryCostSnapshot 库存成本快照
type InventoryCostSnapshot struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	TenantID     int64     `json:"tenant_id" gorm:"index;not null"`
	WarehouseID  int64     `json:"warehouse_id" gorm:"index;not null"`
	ProductID    int64     `json:"product_id" gorm:"index;not null"`
	ProductSku   string    `json:"product_sku" gorm:"size:50"`
	SnapshotDate time.Time `json:"snapshot_date" gorm:"index"`
	BeginQty     int       `json:"begin_qty"`
	InQty        int       `json:"in_qty"`
	OutQty       int       `json:"out_qty"`
	EndQty       int       `json:"end_qty"`
	BeginAmount  float64   `json:"begin_amount" gorm:"type:decimal(12,2)"`
	InAmount     float64   `json:"in_amount" gorm:"type:decimal(12,2)"`
	OutAmount    float64   `json:"out_amount" gorm:"type:decimal(12,2)"`
	EndAmount    float64   `json:"end_amount" gorm:"type:decimal(12,2)"`
	UnitCost     float64   `json:"unit_cost" gorm:"type:decimal(12,4)"`
	CreatedAt    time.Time `json:"created_at"`
}

// TableName specifies the table name for InventoryCostSnapshot
func (InventoryCostSnapshot) TableName() string {
	return "inventory_cost_snapshots"
}

// FinancialSettlement 财务结算
type FinancialSettlement struct {
	ID              int64      `json:"id" gorm:"primaryKey"`
	TenantID        int64      `json:"tenant_id" gorm:"index;not null"`
	SettlementType  string     `json:"settlement_type" gorm:"size:20;not null"` // monthly/yearly
	Period          string     `json:"period" gorm:"size:10;uniqueIndex:idx_tenant_period"`
	ShopID          *int64     `json:"shop_id"`
	TotalSales      float64    `json:"total_sales" gorm:"type:decimal(12,2)"`
	TotalRefund     float64    `json:"total_refund" gorm:"type:decimal(12,2)"`
	NetSales        float64    `json:"net_sales" gorm:"type:decimal(12,2)"`
	OtherIncome     float64    `json:"other_income" gorm:"type:decimal(12,2)"`
	ProductCost     float64    `json:"product_cost" gorm:"type:decimal(12,2)"`
	ShippingCost    float64    `json:"shipping_cost" gorm:"type:decimal(12,2)"`
	Commission      float64    `json:"commission" gorm:"type:decimal(12,2)"`
	ServiceFee      float64    `json:"service_fee" gorm:"type:decimal(12,2)"`
	PromotionFee    float64    `json:"promotion_fee" gorm:"type:decimal(12,2)"`
	OtherCost       float64    `json:"other_cost" gorm:"type:decimal(12,2)"`
	TotalCost       float64    `json:"total_cost" gorm:"type:decimal(12,2)"`
	GrossProfit     float64    `json:"gross_profit" gorm:"type:decimal(12,2)"`
	ProfitRate      float64    `json:"profit_rate" gorm:"type:decimal(5,2)"`
	InventoryChange float64    `json:"inventory_change" gorm:"type:decimal(12,2)"`
	Status          int        `json:"status"`                    // 0-待结算 1-已结算 2-已取消
	SettledAt       *time.Time `json:"settled_at"`
	SettledBy       *int64     `json:"settled_by"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// TableName specifies the table name for FinancialSettlement
func (FinancialSettlement) TableName() string {
	return "financial_settlements"
}

// FinanceBankAccount 结算账户
type FinanceBankAccount struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	TenantID    int64     `json:"tenant_id" gorm:"index;not null"`
	AccountName string    `json:"account_name" gorm:"size:100;not null"`
	AccountNo   string    `json:"account_no" gorm:"size:50"`
	BankName    string    `json:"bank_name" gorm:"size:100"`
	BankBranch  string    `json:"bank_branch" gorm:"size:100"`
	AccountType string    `json:"account_type" gorm:"size:20"` // bank/alipay/wechat等
	Currency    string    `json:"currency" gorm:"size:10;default:'CNY'"`
	Balance     float64   `json:"balance" gorm:"type:decimal(12,2)"`
	Status      int       `json:"status"`                      // 1-启用 0-停用
	IsDefault   bool      `json:"is_default"`
	Remark      string    `json:"remark" gorm:"size:500"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName specifies the table name for FinanceBankAccount
func (FinanceBankAccount) TableName() string {
	return "finance_bank_accounts"
}
