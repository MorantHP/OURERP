package repository

import (
	"context"

	"gorm.io/gorm"
)

// TenantScope 租户查询作用域
// 用于自动添加租户过滤条件

// contextKey 上下文键类型
type contextKey string

// tenantIDKey 租户ID上下文键
const tenantIDKey contextKey = "tenant_id"

// WithTenant 添加租户过滤条件
func WithTenant(tenantID int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if tenantID > 0 {
			return db.Where("tenant_id = ?", tenantID)
		}
		return db
	}
}

// WithTenantFromContext 从上下文获取租户ID并添加过滤
func WithTenantFromContext(ctx context.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		tenantID := GetTenantIDFromContext(ctx)
		if tenantID > 0 {
			return db.Where("tenant_id = ?", tenantID)
		}
		return db
	}
}

// WithTenantIDs 多租户过滤（用于跨租户查询）
func WithTenantIDs(tenantIDs []int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if len(tenantIDs) > 0 {
			return db.Where("tenant_id IN ?", tenantIDs)
		}
		return db
	}
}

// GetTenantIDFromContext 从上下文获取租户ID
func GetTenantIDFromContext(ctx context.Context) int64 {
	if tid, ok := ctx.Value(tenantIDKey).(int64); ok {
		return tid
	}
	return 0
}

// SetTenantIDToContext 设置租户ID到上下文
func SetTenantIDToContext(ctx context.Context, tenantID int64) context.Context {
	return context.WithValue(ctx, tenantIDKey, tenantID)
}

// TenantDB 租户数据库助手
type TenantDB struct {
	db *gorm.DB
}

// NewTenantDB 创建租户数据库助手
func NewTenantDB(db *gorm.DB) *TenantDB {
	return &TenantDB{db: db}
}

// WithTenant 带租户过滤的查询
func (t *TenantDB) WithTenant(tenantID int64) *gorm.DB {
	return t.db.Scopes(WithTenant(tenantID))
}

// WithContext 带上下文租户过滤的查询
func (t *TenantDB) WithContext(ctx context.Context) *gorm.DB {
	return t.db.WithContext(ctx).Scopes(WithTenantFromContext(ctx))
}

// CreateWithTenant 创建记录并自动填充租户ID
func (t *TenantDB) CreateWithTenant(ctx context.Context, model interface{}) error {
	return t.db.WithContext(ctx).Create(model).Error
}

// TenantModel 租户模型接口
type TenantModel interface {
	GetTenantID() int64
	SetTenantID(tenantID int64)
}

// BaseTenantModel 基础租户模型（可嵌入）
type BaseTenantModel struct {
	TenantID int64 `json:"tenant_id" gorm:"index;not null"`
}

// GetTenantID 获取租户ID
func (m *BaseTenantModel) GetTenantID() int64 {
	return m.TenantID
}

// SetTenantID 设置租户ID
func (m *BaseTenantModel) SetTenantID(tenantID int64) {
	m.TenantID = tenantID
}
