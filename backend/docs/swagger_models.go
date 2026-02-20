package docs

// ==================== 通用响应 ====================

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error string `json:"error" example:"操作失败"`
}

// SuccessResponse 成功响应
type SuccessResponse struct {
	Message string `json:"message" example:"操作成功"`
}

// ==================== 认证相关 ====================

// LoginRequest 登录请求
type LoginRequest struct {
	Email    string `json:"email" example:"user@example.com" binding:"required,email"`
	Password string `json:"password" example:"password123" binding:"required,min=6"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string    `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User  UserBasic `json:"user"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Email    string `json:"email" example:"user@example.com" binding:"required,email"`
	Password string `json:"password" example:"password123" binding:"required,min=6"`
	Name     string `json:"name" example:"张三" binding:"required,min=2,max=50"`
	Phone    string `json:"phone" example:"13800138000"`
}

// UserBasic 用户基本信息
type UserBasic struct {
	ID         int64  `json:"id" example:"1"`
	Email      string `json:"email" example:"user@example.com"`
	Name       string `json:"name" example:"张三"`
	Phone      string `json:"phone" example:"13800138000"`
	IsRoot     bool   `json:"is_root" example:"false"`
	IsApproved bool   `json:"is_approved" example:"true"`
}

// ==================== 租户相关 ====================

// TenantCreateRequest 创建租户请求
type TenantCreateRequest struct {
	Code        string `json:"code" example:"COMPANY001" binding:"required"`
	Name        string `json:"name" example:"示例公司" binding:"required"`
	Platform    string `json:"platform" example:"taobao"`
	Description string `json:"description" example:"公司描述"`
}

// TenantResponse 租户响应
type TenantResponse struct {
	ID          int64  `json:"id" example:"1"`
	Code        string `json:"code" example:"COMPANY001"`
	Name        string `json:"name" example:"示例公司"`
	Platform    string `json:"platform" example:"taobao"`
	Description string `json:"description" example:"公司描述"`
	Status      int    `json:"status" example:"1"`
	Role        string `json:"role" example:"admin"`
}

// TenantSwitchRequest 切换租户请求
type TenantSwitchRequest struct {
	TenantID int64 `json:"tenant_id" example:"1" binding:"required"`
}

// ==================== 订单相关 ====================

// OrderListResponse 订单列表响应
type OrderListResponse struct {
	List       []OrderBasic `json:"list"`
	Pagination Pagination   `json:"pagination"`
}

// OrderBasic 订单基本信息
type OrderBasic struct {
	ID               int64   `json:"id" example:"1"`
	OrderNo          string  `json:"order_no" example:"TB202401010001"`
	Platform         string  `json:"platform" example:"taobao"`
	Status           int     `json:"status" example:"100"`
	TotalAmount      float64 `json:"total_amount" example:"199.00"`
	PayAmount        float64 `json:"pay_amount" example:"199.00"`
	BuyerNick        string  `json:"buyer_nick" example:"买家昵称"`
	ReceiverName     string  `json:"receiver_name" example:"张三"`
	ReceiverPhone    string  `json:"receiver_phone" example:"13800138000"`
	ReceiverAddress  string  `json:"receiver_address" example:"北京市朝阳区"`
	LogisticsCompany string  `json:"logistics_company" example:"顺丰速运"`
	LogisticsNo      string  `json:"logistics_no" example:"SF1234567890"`
	CreatedAt        string  `json:"created_at" example:"2024-01-01T10:00:00Z"`
}

// Pagination 分页信息
type Pagination struct {
	Page  int   `json:"page" example:"1"`
	Size  int   `json:"size" example:"20"`
	Total int64 `json:"total" example:"100"`
}

// ==================== 商品相关 ====================

// ProductListResponse 商品列表响应
type ProductListResponse struct {
	List       []ProductBasic `json:"list"`
	Pagination Pagination     `json:"pagination"`
}

// ProductBasic 商品基本信息
type ProductBasic struct {
	ID        int64   `json:"id" example:"1"`
	SkuCode   string  `json:"sku_code" example:"SKU001"`
	Name      string  `json:"name" example:"iPhone 15 Pro"`
	Category  string  `json:"category" example:"手机"`
	Brand     string  `json:"brand" example:"Apple"`
	CostPrice float64 `json:"cost_price" example:"7999.00"`
	SalePrice float64 `json:"sale_price" example:"8999.00"`
	Status    int     `json:"status" example:"1"`
}

// ==================== 库存相关 ====================

// InventoryListResponse 库存列表响应
type InventoryListResponse struct {
	List       []InventoryBasic `json:"list"`
	Pagination Pagination       `json:"pagination"`
}

// InventoryBasic 库存基本信息
type InventoryBasic struct {
	ID         int64   `json:"id" example:"1"`
	ProductID  int64   `json:"product_id" example:"1"`
	ProductName string `json:"product_name" example:"iPhone 15 Pro"`
	SkuCode    string  `json:"sku_code" example:"SKU001"`
	WarehouseID int64  `json:"warehouse_id" example:"1"`
	WarehouseName string `json:"warehouse_name" example:"主仓库"`
	Quantity   int     `json:"quantity" example:"100"`
	LockedQty  int     `json:"locked_qty" example:"5"`
	AlertQty   int     `json:"alert_qty" example:"10"`
	Location   string  `json:"location" example:"A-1-10"`
}

// ==================== 数据中心相关 ====================

// RealtimeOverviewResponse 实时概览响应
type RealtimeOverviewResponse struct {
	TodayOrderCount     int     `json:"today_order_count" example:"150"`
	TodayOrderAmount    float64 `json:"today_order_amount" example:"25800.00"`
	TodayPaidAmount     float64 `json:"today_paid_amount" example:"23000.00"`
	TodayRefundCount    int     `json:"today_refund_count" example:"5"`
	TodayRefundAmount   float64 `json:"today_refund_amount" example:"800.00"`
	TodayNewCustomers   int     `json:"today_new_customers" example:"20"`
	PendingOrders       int     `json:"pending_orders" example:"30"`
	ShippedOrders       int     `json:"shipped_orders" example:"80"`
	CompletedOrders     int     `json:"completed_orders" example:"35"`
	CancelledOrders     int     `json:"cancelled_orders" example:"5"`
	LowStockItems       int     `json:"low_stock_items" example:"12"`
	OutOfStockItems     int     `json:"out_of_stock_items" example:"3"`
	UnhandledAlerts     int     `json:"unhandled_alerts" example:"8"`
	AvgOrderValue       float64 `json:"avg_order_value" example:"172.00"`
}

// CustomerAnalysisResponse 客户分析响应
type CustomerAnalysisResponse struct {
	TotalCustomers     int     `json:"total_customers" example:"1000"`
	NewCustomers       int     `json:"new_customers" example:"50"`
	ActiveCustomers    int     `json:"active_customers" example:"200"`
	ReturnCustomers    int     `json:"return_customers" example:"80"`
	RepurchaseRate     float64 `json:"repurchase_rate" example:"35.5"`
	AvgOrderValue      float64 `json:"avg_order_value" example:"185.00"`
	AvgCustomerValue   float64 `json:"avg_customer_value" example:"1250.00"`
}

// ==================== 财务相关 ====================

// FinanceRecordListResponse 财务记录列表响应
type FinanceRecordListResponse struct {
	List       []FinanceRecordBasic `json:"list"`
	Pagination Pagination           `json:"pagination"`
}

// FinanceRecordBasic 财务记录基本信息
type FinanceRecordBasic struct {
	ID          int64   `json:"id" example:"1"`
	Type        string  `json:"type" example:"income"`
	Category    string  `json:"category" example:"销售收入"`
	Amount      float64 `json:"amount" example:"1000.00"`
	Description string  `json:"description" example:"订单收入"`
	RecordDate  string  `json:"record_date" example:"2024-01-01"`
	Status      int     `json:"status" example:"1"`
}

// ==================== 预警相关 ====================

// AlertRuleListResponse 预警规则列表响应
type AlertRuleListResponse struct {
	Rules      []AlertRuleBasic `json:"rules"`
	Total      int64            `json:"total"`
}

// AlertRuleBasic 预警规则基本信息
type AlertRuleBasic struct {
	ID          int64   `json:"id" example:"1"`
	Name        string  `json:"name" example:"库存不足预警"`
	Type        string  `json:"type" example:"inventory"`
	Threshold   float64 `json:"threshold" example:"10"`
	NotifyType  string  `json:"notify_type" example:"system"`
	Level       string  `json:"level" example:"warning"`
	Status      int     `json:"status" example:"1"`
	Description string  `json:"description" example:"当库存低于10时触发"`
}

// AlertSummaryResponse 预警汇总响应
type AlertSummaryResponse struct {
	TotalAlerts    int `json:"total_alerts" example:"50"`
	UnhandledAlerts int `json:"unhandled_alerts" example:"10"`
	CriticalCount  int `json:"critical_count" example:"2"`
	WarningCount   int `json:"warning_count" example:"5"`
	InfoCount      int `json:"info_count" example:"3"`
	TodayAlerts    int `json:"today_alerts" example:"8"`
}
