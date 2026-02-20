package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Stocktake 盘点单
type Stocktake struct {
	ID          int64           `json:"id" gorm:"primaryKey"`
	TenantID    int64           `json:"tenant_id" gorm:"index;not null"`
	OrderNo     string          `json:"order_no" gorm:"uniqueIndex;size:50;not null"`
	WarehouseID int64           `json:"warehouse_id" gorm:"index"`
	Status      int             `json:"status" gorm:"default:1"` // 1-盘点中 2-已完成 3-已取消
	TotalQty    int             `json:"total_qty"`               // 系统库存合计
	ActualQty   int             `json:"actual_qty"`              // 实盘数量合计
	DiffQty     int             `json:"diff_qty"`                // 差异数量合计
	Remark      string          `json:"remark" gorm:"size:500"`
	OperatorID  int64           `json:"operator_id"`
	CompletedAt *time.Time      `json:"completed_at"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `json:"-" gorm:"index"`
	Items       []StocktakeItem `json:"items" gorm:"foreignKey:StocktakeID"`
}

// TableName 指定表名
func (Stocktake) TableName() string {
	return "stocktakes"
}

// StocktakeItem 盘点单明细
type StocktakeItem struct {
	ID          int64 `json:"id" gorm:"primaryKey"`
	StocktakeID int64 `json:"stocktake_id" gorm:"index;not null"`
	ProductID   int64 `json:"product_id" gorm:"index"`
	SystemQty   int   `json:"system_qty"`   // 系统库存
	ActualQty   int   `json:"actual_qty"`   // 实盘数量
	DiffQty     int   `json:"diff_qty"`     // 差异 = 实盘 - 系统
	Remark      string `json:"remark" gorm:"size:200"`
}

// TableName 指定表名
func (StocktakeItem) TableName() string {
	return "stocktake_items"
}

// StocktakeStatus 盘点状态
const (
	StocktakeStatusInProgress = 1 // 盘点中
	StocktakeStatusCompleted  = 2 // 已完成
	StocktakeStatusCancelled  = 3 // 已取消
)

// GenerateStocktakeOrderNo 生成盘点单号
func GenerateStocktakeOrderNo() string {
	return fmt.Sprintf("ST%s", time.Now().Format("20060102150405"))
}

// CreateStocktakeRequest 创建盘点单请求
type CreateStocktakeRequest struct {
	WarehouseID int64  `json:"warehouse_id" binding:"required"`
	Remark      string `json:"remark"`
}

// UpdateStocktakeItemRequest 更新盘点明细请求
type UpdateStocktakeItemRequest struct {
	ProductID int64 `json:"product_id" binding:"required"`
	ActualQty int   `json:"actual_qty" binding:"required,min=0"`
	Remark    string `json:"remark"`
}

// CompleteStocktakeRequest 完成盘点请求
type CompleteStocktakeRequest struct {
	Items []StocktakeItemInput `json:"items" binding:"required,min=1"`
}

// StocktakeItemInput 盘点明细输入
type StocktakeItemInput struct {
	ProductID int64  `json:"product_id" binding:"required"`
	ActualQty int    `json:"actual_qty" binding:"required,min=0"`
	Remark    string `json:"remark"`
}

// StocktakeWithDetails 带详情的盘点单
type StocktakeWithDetails struct {
	Stocktake
	Warehouse *Warehouse         `json:"warehouse" gorm:"foreignKey:WarehouseID"`
	Items     []StocktakeItemExt `json:"items"`
}

// StocktakeItemExt 扩展的盘点明细
type StocktakeItemExt struct {
	StocktakeItem
	Product *Product `json:"product" gorm:"foreignKey:ProductID"`
}
