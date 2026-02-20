package models

import (
	"time"
)

// AlertRule 预警规则
type AlertRule struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	TenantID     int64     `json:"tenant_id" gorm:"index;not null"`
	Name         string    `json:"name" gorm:"size:100;not null"`
	Type         string    `json:"type" gorm:"size:20;not null"` // inventory/sales/order/customer
	Condition    string    `json:"condition" gorm:"size:500"`    // JSON条件表达式
	Threshold    float64   `json:"threshold" gorm:"type:decimal(12,2)"`
	ThresholdMin float64   `json:"threshold_min" gorm:"type:decimal(12,2)"` // 下限阈值
	NotifyType   string    `json:"notify_type" gorm:"size:50"`              // email/sms/webhook/system
	NotifyTarget string    `json:"notify_target" gorm:"size:500"`           // 通知目标，多个用逗号分隔
	Level        string    `json:"level" gorm:"size:20;default:'warning'"`  // info/warning/critical
	Status       int       `json:"status" gorm:"default:1"`                 // 1-启用 0-停用
	Description  string    `json:"description" gorm:"size:500"`
	CreatedBy    int64     `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// 关联
	Records []AlertRecord `json:"records,omitempty" gorm:"foreignKey:RuleID"`
}

// AlertRecord 预警记录
type AlertRecord struct {
	ID         int64      `json:"id" gorm:"primaryKey"`
	TenantID   int64      `json:"tenant_id" gorm:"index;not null"`
	RuleID     int64      `json:"rule_id" gorm:"index"`
	Rule       *AlertRule `json:"rule" gorm:"foreignKey:RuleID"`
	Title      string     `json:"title" gorm:"size:200"`
	Content    string     `json:"content" gorm:"type:text"`
	Level      string     `json:"level" gorm:"size:20"` // info/warning/critical
	SourceType string     `json:"source_type" gorm:"size:50"` // 触发源类型
	SourceID   int64      `json:"source_id"`                   // 触发源ID
	Status     int        `json:"status" gorm:"default:0"`     // 0-未处理 1-已处理 2-已忽略
	HandledBy  *int64     `json:"handled_by"`
	HandledAt  *time.Time `json:"handled_at"`
	HandleNote string     `json:"handle_note" gorm:"size:500"`
	CreatedAt  time.Time  `json:"created_at"`
}

// ReportTemplate 自定义报表模板
type ReportTemplate struct {
	ID         int64     `json:"id" gorm:"primaryKey"`
	TenantID   int64     `json:"tenant_id" gorm:"index;not null"`
	Name       string    `json:"name" gorm:"size:100;not null"`
	Type       string    `json:"type" gorm:"size:20"`          // sales/inventory/customer/order/finance
	DataSource string    `json:"data_source" gorm:"size:50"`   // orders/products/customers etc.
	Columns    string    `json:"columns" gorm:"type:text"`     // JSON列定义
	Filters    string    `json:"filters" gorm:"type:text"`     // JSON筛选条件
	Sorts      string    `json:"sorts" gorm:"type:text"`       // JSON排序条件
	ChartType  string    `json:"chart_type" gorm:"size:20"`    // table/line/bar/pie/radar
	ChartConfig string    `json:"chart_config" gorm:"type:text"` // JSON图表配置
	IsPublic   int       `json:"is_public" gorm:"default:0"`   // 0-私有 1-公开
	CreatedBy  int64     `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Customer 客户信息
type Customer struct {
	ID            int64      `json:"id" gorm:"primaryKey"`
	TenantID      int64      `json:"tenant_id" gorm:"index;not null"`
	Code          string     `json:"code" gorm:"size:50;uniqueIndex:idx_tenant_customer_code"`
	Name          string     `json:"name" gorm:"size:100;index"`
	Phone         string     `json:"phone" gorm:"size:20;index"`
	Email         string     `json:"email" gorm:"size:100"`
	Type          string     `json:"type" gorm:"size:20;default:'b2c'"`   // b2b/b2c
	Level         string     `json:"level" gorm:"size:20;default:'normal'"` // vip/normal/new
	Source        string     `json:"source" gorm:"size:50"`              // 来源平台
	Province      string     `json:"province" gorm:"size:50;index"`
	City          string     `json:"city" gorm:"size:50;index"`
	District      string     `json:"district" gorm:"size:50"`
	Address       string     `json:"address" gorm:"size:300"`
	TotalOrders   int        `json:"total_orders"`
	TotalAmount   float64    `json:"total_amount" gorm:"type:decimal(12,2);default:0"`
	TotalPaid     float64    `json:"total_paid" gorm:"type:decimal(12,2);default:0"`
	FirstOrderAt  *time.Time `json:"first_order_at"`
	LastOrderAt   *time.Time `json:"last_order_at"`
	AvgOrderValue float64    `json:"avg_order_value" gorm:"type:decimal(12,2);default:0"`
	Tags          string     `json:"tags" gorm:"size:200"` // 标签，逗号分隔
	Remark        string     `json:"remark" gorm:"size:500"`
	Status        int        `json:"status" gorm:"default:1"` // 1-正常 0-禁用
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// RealtimeSnapshot 实时监控快照
type RealtimeSnapshot struct {
	ID              int64     `json:"id" gorm:"primaryKey"`
	TenantID        int64     `json:"tenant_id" gorm:"index;not null"`
	SnapshotTime    time.Time `json:"snapshot_time" gorm:"index"`
	OrderCount      int       `json:"order_count"`
	OrderAmount     float64   `json:"order_amount" gorm:"type:decimal(12,2)"`
	PaidAmount      float64   `json:"paid_amount" gorm:"type:decimal(12,2)"`
	RefundCount     int       `json:"refund_count"`
	RefundAmount    float64   `json:"refund_amount" gorm:"type:decimal(12,2)"`
	PendingOrders   int       `json:"pending_orders"`
	ShippedOrders   int       `json:"shipped_orders"`
	CompletedOrders int       `json:"completed_orders"`
	CancelledOrders int       `json:"cancelled_orders"`
	LowStockItems   int       `json:"low_stock_items"`
	OutOfStockItems int       `json:"out_of_stock_items"`
	NewCustomers    int       `json:"new_customers"`
	CreatedAt       time.Time `json:"created_at"`
}

// ProductAnalysis 商品分析结果缓存
type ProductAnalysis struct {
	ID             int64     `json:"id" gorm:"primaryKey"`
	TenantID       int64     `json:"tenant_id" gorm:"index;not null"`
	ProductID      int64     `json:"product_id" gorm:"index"`
	SkuID          int64     `json:"sku_id" gorm:"index"`
	AnalysisDate   time.Time `json:"analysis_date" gorm:"index"`
	PeriodType     string    `json:"period_type" gorm:"size:20"` // day/week/month
	SalesCount     int       `json:"sales_count"`
	SalesAmount    float64   `json:"sales_amount" gorm:"type:decimal(12,2)"`
	SalesQuantity  int       `json:"sales_quantity"`
	RefundCount    int       `json:"refund_count"`
	RefundAmount   float64   `json:"refund_amount" gorm:"type:decimal(12,2)"`
	RefundRate     float64   `json:"refund_rate" gorm:"type:decimal(5,2)"` // 退货率
	AvgPrice       float64   `json:"avg_price" gorm:"type:decimal(10,2)"`
	ProductCost    float64   `json:"product_cost" gorm:"type:decimal(12,2)"`
	GrossProfit    float64   `json:"gross_profit" gorm:"type:decimal(12,2)"`
	ProfitRate     float64   `json:"profit_rate" gorm:"type:decimal(5,2)"` // 毛利率
	InventoryQty   int       `json:"inventory_qty"`
	TurnoverRate   float64   `json:"turnover_rate" gorm:"type:decimal(5,2)"` // 周转率
	SalesRate      float64   `json:"sales_rate" gorm:"type:decimal(5,2)"`     // 动销率
	StockDays      int       `json:"stock_days"`                              // 库存可售天数
	CreatedAt      time.Time `json:"created_at"`
}

// CustomerAnalysis 客户分析结果缓存
type CustomerAnalysis struct {
	ID               int64     `json:"id" gorm:"primaryKey"`
	TenantID         int64     `json:"tenant_id" gorm:"index;not null"`
	AnalysisDate     time.Time `json:"analysis_date" gorm:"index"`
	PeriodType       string    `json:"period_type" gorm:"size:20"` // day/week/month
	NewCustomers     int       `json:"new_customers"`              // 新客户数
	ActiveCustomers  int       `json:"active_customers"`           // 活跃客户数
	ReturnCustomers  int       `json:"return_customers"`           // 回购客户数
	RepurchaseRate   float64   `json:"repurchase_rate" gorm:"type:decimal(5,2)"` // 复购率
	AvgOrderValue    float64   `json:"avg_order_value" gorm:"type:decimal(10,2)"`
	AvgCustomerValue float64   `json:"avg_customer_value" gorm:"type:decimal(10,2)"`
	VipCount         int       `json:"vip_count"`
	NormalCount      int       `json:"normal_count"`
	NewCount         int       `json:"new_count"`
	CreatedAt        time.Time `json:"created_at"`
}

// RegionAnalysis 地域分析结果
type RegionAnalysis struct {
	ID            int64     `json:"id" gorm:"primaryKey"`
	TenantID      int64     `json:"tenant_id" gorm:"index;not null"`
	AnalysisDate  time.Time `json:"analysis_date" gorm:"index"`
	PeriodType    string    `json:"period_type" gorm:"size:20"`
	Province      string    `json:"province" gorm:"size:50;index"`
	City          string    `json:"city" gorm:"size:50;index"`
	OrderCount    int       `json:"order_count"`
	OrderAmount   float64   `json:"order_amount" gorm:"type:decimal(12,2)"`
	CustomerCount int       `json:"customer_count"`
	ProductCount  int       `json:"product_count"` // 销售SKU数
	AvgOrderValue float64   `json:"avg_order_value" gorm:"type:decimal(10,2)"`
	CreatedAt     time.Time `json:"created_at"`
}

// CompareAnalysis 对比分析结果
type CompareAnalysis struct {
	ID            int64     `json:"id" gorm:"primaryKey"`
	TenantID      int64     `json:"tenant_id" gorm:"index;not null"`
	AnalysisDate  time.Time `json:"analysis_date" gorm:"index"`
	CompareType   string    `json:"compare_type" gorm:"size:20"` // yoy/mom/target/shop/platform
	CurrentPeriod string    `json:"current_period" gorm:"size:50"`
	ComparePeriod string    `json:"compare_period" gorm:"size:50"`
	MetricType    string    `json:"metric_type" gorm:"size:50"`  // orders/sales/profit/customers
	CurrentValue  float64   `json:"current_value" gorm:"type:decimal(12,2)"`
	CompareValue  float64   `json:"compare_value" gorm:"type:decimal(12,2)"`
	ChangeValue   float64   `json:"change_value" gorm:"type:decimal(12,2)"`
	ChangeRate    float64   `json:"change_rate" gorm:"type:decimal(5,2)"`
	TargetValue   float64   `json:"target_value" gorm:"type:decimal(12,2)"`
	AchieveRate   float64   `json:"achieve_rate" gorm:"type:decimal(5,2)"`
	ExtraData     string    `json:"extra_data" gorm:"type:text"` // JSON额外数据
	CreatedAt     time.Time `json:"created_at"`
}

// DashboardWidget 仪表盘组件配置
type DashboardWidget struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	TenantID    int64     `json:"tenant_id" gorm:"index;not null"`
	UserID      int64     `json:"user_id" gorm:"index"`
	Name        string    `json:"name" gorm:"size:100"`
	Type        string    `json:"type" gorm:"size:50"`        // kpi/chart/table/list
	DataSource  string    `json:"data_source" gorm:"size:50"` // API数据源
	Config      string    `json:"config" gorm:"type:text"`    // JSON配置
	Position    string    `json:"position" gorm:"size:50"`    // 位置信息
	Width       int       `json:"width"`                      // 宽度(格子数)
	Height      int       `json:"height"`                     // 高度(格子数)
	SortOrder   int       `json:"sort_order"`
	Status      int       `json:"status" gorm:"default:1"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 查询过滤器类型

// AlertRuleFilter 预警规则过滤器
type AlertRuleFilter struct {
	Type   string
	Status *int
}

// AlertRecordFilter 预警记录过滤器
type AlertRecordFilter struct {
	RuleID     *int64
	Level      string
	Status     *int
	SourceType string
	StartDate  *time.Time
	EndDate    *time.Time
}

// CustomerFilter 客户过滤器
type CustomerFilter struct {
	Name     string
	Phone    string
	Type     string
	Level    string
	Province string
	City     string
	Source   string
	Status   *int
	Tags     string
}

// ReportTemplateFilter 报表模板过滤器
type ReportTemplateFilter struct {
	Type       string
	DataSource string
	IsPublic   *int
	CreatedBy  *int64
}
