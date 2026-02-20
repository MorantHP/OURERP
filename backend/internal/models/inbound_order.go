package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// InboundOrder 入库单
type InboundOrder struct {
	ID          int64          `json:"id" gorm:"primaryKey"`
	TenantID    int64          `json:"tenant_id" gorm:"index;not null"`
	OrderNo     string         `json:"order_no" gorm:"uniqueIndex;size:50;not null"`
	WarehouseID int64          `json:"warehouse_id" gorm:"index"`
	Type        string         `json:"type" gorm:"size:20"`      // purchase-采购, return-退货, transfer-调拨入
	Status      int            `json:"status" gorm:"default:1"`  // 1-待入库 2-已入库 3-已取消
	TotalQty    int            `json:"total_qty"`
	TotalAmount float64        `json:"total_amount"`
	Supplier    string         `json:"supplier" gorm:"size:100"`
	Remark      string         `json:"remark" gorm:"size:500"`
	OperatorID  int64          `json:"operator_id"`
	CompletedAt *time.Time     `json:"completed_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	Items       []InboundItem  `json:"items" gorm:"foreignKey:OrderID"`
}

// TableName 指定表名
func (InboundOrder) TableName() string {
	return "inbound_orders"
}

// InboundItem 入库单明细
type InboundItem struct {
	ID        int64      `json:"id" gorm:"primaryKey"`
	OrderID   int64      `json:"order_id" gorm:"index;not null"`
	ProductID int64      `json:"product_id" gorm:"index"`
	Quantity  int        `json:"quantity"`
	CostPrice float64    `json:"cost_price"`
	BatchNo   string     `json:"batch_no" gorm:"size:50"`
	ExpireAt  *time.Time `json:"expire_at"`
	Remark    string     `json:"remark" gorm:"size:200"`
}

// TableName 指定表名
func (InboundItem) TableName() string {
	return "inbound_items"
}

// InboundType 入库类型
const (
	InboundTypePurchase  = "purchase"  // 采购入库
	InboundTypeReturn    = "return"    // 退货入库
	InboundTypeTransfer  = "transfer"  // 调拨入库
	InboundTypeOther     = "other"     // 其他
)

// InboundStatus 入库状态
const (
	InboundStatusPending   = 1 // 待入库
	InboundStatusCompleted = 2 // 已入库
	InboundStatusCancelled = 3 // 已取消
)

// GenerateInboundOrderNo 生成入库单号
func GenerateInboundOrderNo() string {
	return fmt.Sprintf("IN%s", time.Now().Format("20060102150405"))
}

// CreateInboundRequest 创建入库单请求
type CreateInboundRequest struct {
	WarehouseID int64               `json:"warehouse_id" binding:"required"`
	Type        string              `json:"type" binding:"required"`
	Supplier    string              `json:"supplier"`
	Remark      string              `json:"remark"`
	Items       []InboundItemInput  `json:"items" binding:"required,min=1"`
}

// InboundItemInput 入库明细输入
type InboundItemInput struct {
	ProductID int64      `json:"product_id" binding:"required"`
	Quantity  int        `json:"quantity" binding:"required,min=1"`
	CostPrice float64    `json:"cost_price"`
	BatchNo   string     `json:"batch_no"`
	ExpireAt  *time.Time `json:"expire_at"`
	Remark    string     `json:"remark"`
}

// InboundOrderWithDetails 带详情的入库单
type InboundOrderWithDetails struct {
	InboundOrder
	Warehouse *Warehouse       `json:"warehouse" gorm:"foreignKey:WarehouseID"`
	Items     []InboundItemExt `json:"items"`
}

// InboundItemExt 扩展的入库明细
type InboundItemExt struct {
	InboundItem
	Product *Product `json:"product" gorm:"foreignKey:ProductID"`
}
