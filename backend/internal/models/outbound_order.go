package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// OutboundOrder 出库单
type OutboundOrder struct {
	ID          int64           `json:"id" gorm:"primaryKey"`
	TenantID    int64           `json:"tenant_id" gorm:"index;not null"`
	OrderNo     string          `json:"order_no" gorm:"uniqueIndex;size:50;not null"`
	WarehouseID int64           `json:"warehouse_id" gorm:"index"`
	Type        string          `json:"type" gorm:"size:20"`       // sale-销售, transfer-调拨出, scrap-报废
	Status      int             `json:"status" gorm:"default:1"`   // 1-待出库 2-已出库 3-已取消
	TotalQty    int             `json:"total_qty"`
	RefType     string          `json:"ref_type" gorm:"size:20"`   // 关联类型: order-订单
	RefID       int64           `json:"ref_id"`                    // 关联订单ID
	RefNo       string          `json:"ref_no" gorm:"size:50"`     // 关联订单号
	Remark      string          `json:"remark" gorm:"size:500"`
	OperatorID  int64           `json:"operator_id"`
	CompletedAt *time.Time      `json:"completed_at"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `json:"-" gorm:"index"`
	Items       []OutboundItem  `json:"items" gorm:"foreignKey:OrderID"`
}

// TableName 指定表名
func (OutboundOrder) TableName() string {
	return "outbound_orders"
}

// OutboundItem 出库单明细
type OutboundItem struct {
	ID        int64   `json:"id" gorm:"primaryKey"`
	OrderID   int64   `json:"order_id" gorm:"index;not null"`
	ProductID int64   `json:"product_id" gorm:"index"`
	Quantity  int     `json:"quantity"`
	SalePrice float64 `json:"sale_price"`
	Remark    string  `json:"remark" gorm:"size:200"`
}

// TableName 指定表名
func (OutboundItem) TableName() string {
	return "outbound_items"
}

// OutboundType 出库类型
const (
	OutboundTypeSale     = "sale"     // 销售出库
	OutboundTypeTransfer = "transfer" // 调拨出库
	OutboundTypeScrap    = "scrap"    // 报废出库
	OutboundTypeOther    = "other"    // 其他
)

// OutboundStatus 出库状态
const (
	OutboundStatusPending   = 1 // 待出库
	OutboundStatusCompleted = 2 // 已出库
	OutboundStatusCancelled = 3 // 已取消
)

// GenerateOutboundOrderNo 生成出库单号
func GenerateOutboundOrderNo() string {
	return fmt.Sprintf("OUT%s", time.Now().Format("20060102150405"))
}

// CreateOutboundRequest 创建出库单请求
type CreateOutboundRequest struct {
	WarehouseID int64                `json:"warehouse_id" binding:"required"`
	Type        string               `json:"type" binding:"required"`
	RefType     string               `json:"ref_type"`
	RefID       int64                `json:"ref_id"`
	RefNo       string               `json:"ref_no"`
	Remark      string               `json:"remark"`
	Items       []OutboundItemInput  `json:"items" binding:"required,min=1"`
}

// OutboundItemInput 出库明细输入
type OutboundItemInput struct {
	ProductID int64   `json:"product_id" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required,min=1"`
	SalePrice float64 `json:"sale_price"`
	Remark    string  `json:"remark"`
}

// OutboundOrderWithDetails 带详情的出库单
type OutboundOrderWithDetails struct {
	OutboundOrder
	Warehouse *Warehouse        `json:"warehouse" gorm:"foreignKey:WarehouseID"`
	Items     []OutboundItemExt `json:"items"`
}

// OutboundItemExt 扩展的出库明细
type OutboundItemExt struct {
	OutboundItem
	Product *Product `json:"product" gorm:"foreignKey:ProductID"`
}
