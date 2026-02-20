package models

import (
	"time"
)

// Inventory 库存
type Inventory struct {
	ID          int64       `json:"id" gorm:"primaryKey"`
	TenantID    int64       `json:"tenant_id" gorm:"index;not null"`
	ProductID   int64       `json:"product_id" gorm:"uniqueIndex:idx_inv_product_wh,warehouse_id;not null;index"`
	WarehouseID int64       `json:"warehouse_id" gorm:"uniqueIndex:idx_inv_product_wh,product_id;not null;index"`
	Quantity    int         `json:"quantity" gorm:"default:0"`    // 可用库存
	LockedQty   int         `json:"locked_qty" gorm:"default:0"`  // 锁定库存
	TotalQty    int         `json:"total_qty" gorm:"default:0"`   // 总库存 = 可用 + 锁定
	AlertQty    int         `json:"alert_qty" gorm:"default:0"`   // 库存预警数量
	Location    string      `json:"location" gorm:"size:50"`      // 库位
	BatchNo     string      `json:"batch_no" gorm:"size:50"`      // 批次号
	ExpireAt    *time.Time  `json:"expire_at"`                    // 过期日期
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// TableName 指定表名
func (Inventory) TableName() string {
	return "inventories"
}

// InventoryLog 库存流水
type InventoryLog struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	TenantID    int64     `json:"tenant_id" gorm:"index;not null"`
	ProductID   int64     `json:"product_id" gorm:"index;not null"`
	WarehouseID int64     `json:"warehouse_id" gorm:"index;not null"`
	ChangeQty   int       `json:"change_qty"`              // 变动数量(正负)
	BeforeQty   int       `json:"before_qty"`              // 变动前库存
	AfterQty    int       `json:"after_qty"`               // 变动后库存
	RefType     string    `json:"ref_type" gorm:"size:20"` // 关联类型: inbound/outbound/stocktake/transfer/adjust
	RefID       int64     `json:"ref_id"`                  // 关联单据ID
	RefNo       string    `json:"ref_no" gorm:"size:50"`   // 关联单号
	OperatorID  int64     `json:"operator_id"`             // 操作人ID
	Remark      string    `json:"remark" gorm:"size:200"`
	CreatedAt   time.Time `json:"created_at"`
}

// TableName 指定表名
func (InventoryLog) TableName() string {
	return "inventory_logs"
}

// RefType 关联类型
const (
	RefTypeInbound  = "inbound"  // 入库
	RefTypeOutbound = "outbound" // 出库
	RefTypeStocktake = "stocktake" // 盘点
	RefTypeTransfer = "transfer" // 调拨
	RefTypeAdjust   = "adjust"   // 调整
)

// InventoryWithProduct 带商品信息的库存
type InventoryWithProduct struct {
	Inventory
	Product  *Product  `json:"product" gorm:"foreignKey:ProductID"`
	Warehouse *Warehouse `json:"warehouse" gorm:"foreignKey:WarehouseID"`
}

// AdjustInventoryRequest 库存调整请求
type AdjustInventoryRequest struct {
	ProductID   int64  `json:"product_id" binding:"required"`
	WarehouseID int64  `json:"warehouse_id" binding:"required"`
	ChangeQty   int    `json:"change_qty" binding:"required"` // 正数增加，负数减少
	Remark      string `json:"remark"`
}

// UpdateInventoryRequest 更新库存请求
type UpdateInventoryRequest struct {
	AlertQty *int    `json:"alert_qty"`
	Location string  `json:"location"`
	BatchNo  string  `json:"batch_no"`
	ExpireAt *string `json:"expire_at"`
}
