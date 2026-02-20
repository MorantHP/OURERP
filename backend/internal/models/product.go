package models

import (
	"time"

	"gorm.io/gorm"
)

// Product 商品
type Product struct {
	ID        int64          `json:"id" gorm:"primaryKey"`
	TenantID  int64          `json:"tenant_id" gorm:"index;not null"`
	SkuCode   string         `json:"sku_code" gorm:"uniqueIndex:idx_sku_tenant,tenant_id;size:50;not null"` // 商品编码
	Name      string         `json:"name" gorm:"size:200;not null"`
	Category  string         `json:"category" gorm:"size:50"`      // 分类
	Brand     string         `json:"brand" gorm:"size:50"`         // 品牌
	Barcode   string         `json:"barcode" gorm:"size:50;index"` // 条形码
	ImageURL  string         `json:"image_url" gorm:"size:500"`    // 主图
	Unit      string         `json:"unit" gorm:"size:20;default:'件'"`
	CostPrice float64        `json:"cost_price"` // 成本价
	SalePrice float64        `json:"sale_price"` // 销售价
	Specs     JSONB          `json:"specs" gorm:"type:jsonb"`
	Status    int            `json:"status" gorm:"default:1"` // 1-上架 0-下架
	Remark    string         `json:"remark" gorm:"size:500"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (Product) TableName() string {
	return "products"
}

// ProductStatus 商品状态
const (
	ProductStatusOffline = 0 // 下架
	ProductStatusOnline  = 1 // 上架
)

// CreateProductRequest 创建商品请求
type CreateProductRequest struct {
	SkuCode   string `json:"sku_code" binding:"required,min=2,max=50"`
	Name      string `json:"name" binding:"required,min=1,max=200"`
	Category  string `json:"category"`
	Brand     string `json:"brand"`
	Barcode   string `json:"barcode"`
	ImageURL  string `json:"image_url"`
	Unit      string `json:"unit"`
	CostPrice float64 `json:"cost_price"`
	SalePrice float64 `json:"sale_price"`
	Specs     JSONB  `json:"specs"`
	Remark    string `json:"remark"`
}

// UpdateProductRequest 更新商品请求
type UpdateProductRequest struct {
	Name      string  `json:"name"`
	Category  string  `json:"category"`
	Brand     string  `json:"brand"`
	Barcode   string  `json:"barcode"`
	ImageURL  string  `json:"image_url"`
	Unit      string  `json:"unit"`
	CostPrice float64 `json:"cost_price"`
	SalePrice float64 `json:"sale_price"`
	Specs     JSONB   `json:"specs"`
	Status    *int    `json:"status"`
	Remark    string  `json:"remark"`
}

// ProductWithInventory 带库存信息的商品
type ProductWithInventory struct {
	Product
	TotalQuantity int `json:"total_quantity"` // 总库存
	AlertCount    int `json:"alert_count"`    // 预警仓库数
}
