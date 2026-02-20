package models

import (
	"time"

	"gorm.io/gorm"
)

// Tenant 租户/账套模型
type Tenant struct {
	ID          int64          `json:"id" gorm:"primaryKey"`
	Code        string         `json:"code" gorm:"uniqueIndex;size:50;not null"` // 租户编码，如: taobao_001
	Name        string         `json:"name" gorm:"size:100;not null"`             // 租户名称，如: 淘宝业务部
	Platform    string         `json:"platform" gorm:"size:20"`                   // 关联平台: taobao, douyin, kuaishou
	Description string         `json:"description" gorm:"size:500"`               // 描述
	Logo        string         `json:"logo" gorm:"size:500"`                      // Logo URL
	Status      int            `json:"status" gorm:"default:1"`                   // 状态: 1-启用, 0-禁用
	Settings    JSONB          `json:"settings" gorm:"type:jsonb"`                // 租户配置
	OwnerID     int64          `json:"owner_id"`                                  // 负责人ID
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (Tenant) TableName() string {
	return "tenants"
}

// TenantStatus 租户状态
const (
	TenantStatusDisabled = 0
	TenantStatusEnabled  = 1
)

// TenantUser 租户用户关联（用户可以访问多个租户）
type TenantUser struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	TenantID  int64     `json:"tenant_id" gorm:"uniqueIndex:idx_tenant_user;not null"`
	UserID    int64     `json:"user_id" gorm:"uniqueIndex:idx_tenant_user;not null"`
	Role      string    `json:"role" gorm:"size:20;default:'member'"` // owner, admin, member
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (TenantUser) TableName() string {
	return "tenant_users"
}

// TenantRole 租户角色
const (
	TenantRoleOwner  = "owner"  // 所有者
	TenantRoleAdmin  = "admin"  // 管理员
	TenantRoleMember = "member" // 成员
)

// CreateTenantRequest 创建租户请求
type CreateTenantRequest struct {
	Code        string `json:"code" binding:"required,min=2,max=50"`
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Platform    string `json:"platform"`
	Description string `json:"description"`
}

// UpdateTenantRequest 更新租户请求
type UpdateTenantRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Logo        string `json:"logo"`
	Status      *int   `json:"status"`
	Settings    JSONB  `json:"settings"`
}

// TenantWithStats 带统计的租户信息
type TenantWithStats struct {
	Tenant
	ShopCount  int64 `json:"shop_count"`
	OrderCount int64 `json:"order_count"`
	UserCount  int64 `json:"user_count"`
}
