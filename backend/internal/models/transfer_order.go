package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// TransferOrder 调拨单
type TransferOrder struct {
	ID              int64           `json:"id" gorm:"primaryKey"`
	TenantID        int64           `json:"tenant_id" gorm:"index;not null"`
	OrderNo         string          `json:"order_no" gorm:"uniqueIndex;size:50;not null"`
	FromWarehouseID int64           `json:"from_warehouse_id" gorm:"index"`
	ToWarehouseID   int64           `json:"to_warehouse_id" gorm:"index"`
	Status          int             `json:"status" gorm:"default:1"` // 1-待出库 2-在途 3-已入库 4-已取消
	TotalQty        int             `json:"total_qty"`
	Remark          string          `json:"remark" gorm:"size:500"`
	OperatorID      int64           `json:"operator_id"`
	ShippedAt       *time.Time      `json:"shipped_at"`
	ReceivedAt      *time.Time      `json:"received_at"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	DeletedAt       gorm.DeletedAt  `json:"-" gorm:"index"`
	Items           []TransferItem  `json:"items" gorm:"foreignKey:OrderID"`
}

// TableName 指定表名
func (TransferOrder) TableName() string {
	return "transfer_orders"
}

// TransferItem 调拨单明细
type TransferItem struct {
	ID        int64  `json:"id" gorm:"primaryKey"`
	OrderID   int64  `json:"order_id" gorm:"index;not null"`
	ProductID int64  `json:"product_id" gorm:"index"`
	Quantity  int    `json:"quantity"`
	Remark    string `json:"remark" gorm:"size:200"`
}

// TableName 指定表名
func (TransferItem) TableName() string {
	return "transfer_items"
}

// TransferStatus 调拨状态
const (
	TransferStatusPending   = 1 // 待出库
	TransferStatusInTransit = 2 // 在途
	TransferStatusCompleted = 3 // 已入库
	TransferStatusCancelled = 4 // 已取消
)

// GenerateTransferOrderNo 生成调拨单号
func GenerateTransferOrderNo() string {
	return fmt.Sprintf("TF%s", time.Now().Format("20060102150405"))
}

// CreateTransferRequest 创建调拨单请求
type CreateTransferRequest struct {
	FromWarehouseID int64               `json:"from_warehouse_id" binding:"required"`
	ToWarehouseID   int64               `json:"to_warehouse_id" binding:"required"`
	Remark          string              `json:"remark"`
	Items           []TransferItemInput `json:"items" binding:"required,min=1"`
}

// TransferItemInput 调拨明细输入
type TransferItemInput struct {
	ProductID int64  `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
	Remark    string `json:"remark"`
}

// TransferOrderWithDetails 带详情的调拨单
type TransferOrderWithDetails struct {
	TransferOrder
	FromWarehouse *Warehouse        `json:"from_warehouse" gorm:"foreignKey:FromWarehouseID"`
	ToWarehouse   *Warehouse        `json:"to_warehouse" gorm:"foreignKey:ToWarehouseID"`
	Items         []TransferItemExt `json:"items"`
}

// TransferItemExt 扩展的调拨明细
type TransferItemExt struct {
	TransferItem
	Product *Product `json:"product" gorm:"foreignKey:ProductID"`
}
