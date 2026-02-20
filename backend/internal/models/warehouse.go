package models

import (
	"time"

	"gorm.io/gorm"
)

// Warehouse 仓库
type Warehouse struct {
	ID        int64          `json:"id" gorm:"primaryKey"`
	TenantID  int64          `json:"tenant_id" gorm:"index;not null"`
	Code      string         `json:"code" gorm:"uniqueIndex:idx_wh_tenant,tenant_id;size:50;not null"`
	Name      string         `json:"name" gorm:"size:100;not null"`
	Address   string         `json:"address" gorm:"size:200"`
	Contact   string         `json:"contact" gorm:"size:50"` // 联系人
	Phone     string         `json:"phone" gorm:"size:20"`
	Type      string         `json:"type" gorm:"size:20;default:'normal'"` // normal-普通, bonded-保税
	Status    int            `json:"status" gorm:"default:1"`
	IsDefault bool           `json:"is_default" gorm:"default:false"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (Warehouse) TableName() string {
	return "warehouses"
}

// WarehouseType 仓库类型
const (
	WarehouseTypeNormal  = "normal"  // 普通仓库
	WarehouseTypeBonded = "bonded"   // 保税仓
)

// WarehouseStatus 仓库状态
const (
	WarehouseStatusDisabled = 0
	WarehouseStatusEnabled  = 1
)

// CreateWarehouseRequest 创建仓库请求
type CreateWarehouseRequest struct {
	Code    string `json:"code" binding:"required,min=2,max=50"`
	Name    string `json:"name" binding:"required,min=2,max=100"`
	Address string `json:"address"`
	Contact string `json:"contact"`
	Phone   string `json:"phone"`
	Type    string `json:"type"`
}

// UpdateWarehouseRequest 更新仓库请求
type UpdateWarehouseRequest struct {
	Name      string `json:"name"`
	Address   string `json:"address"`
	Contact   string `json:"contact"`
	Phone     string `json:"phone"`
	Type      string `json:"type"`
	Status    *int   `json:"status"`
	IsDefault *bool  `json:"is_default"`
}

// WarehouseWithStats 带统计的仓库信息
type WarehouseWithStats struct {
	Warehouse
	ProductCount int   `json:"product_count"` // 商品种类数
	TotalQty     int64 `json:"total_qty"`     // 总库存数量
	TotalValue   float64 `json:"total_value"` // 总库存价值
}
