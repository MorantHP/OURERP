package repository

import (
	"context"

	"github.com/MorantHP/OURERP/internal/models"
	"gorm.io/gorm"
)

// WarehouseRepository 仓库
type WarehouseRepository struct {
	db *gorm.DB
}

// NewWarehouseRepository 创建仓库
func NewWarehouseRepository(db *gorm.DB) *WarehouseRepository {
	return &WarehouseRepository{db: db}
}

// Create 创建仓库
func (r *WarehouseRepository) Create(warehouse *models.Warehouse) error {
	return r.db.Create(warehouse).Error
}

// FindByID 根据ID查询仓库
func (r *WarehouseRepository) FindByID(id int64) (*models.Warehouse, error) {
	var warehouse models.Warehouse
	err := r.db.First(&warehouse, id).Error
	if err != nil {
		return nil, err
	}
	return &warehouse, nil
}

// FindByIDWithContext 根据ID查询仓库（带租户上下文）
func (r *WarehouseRepository) FindByIDWithContext(ctx context.Context, id int64) (*models.Warehouse, error) {
	var warehouse models.Warehouse
	err := r.db.WithContext(ctx).
		Scopes(WithTenantFromContext(ctx)).
		First(&warehouse, id).Error
	if err != nil {
		return nil, err
	}
	return &warehouse, nil
}

// FindByCode 根据编码查询仓库
func (r *WarehouseRepository) FindByCode(ctx context.Context, code string) (*models.Warehouse, error) {
	var warehouse models.Warehouse
	err := r.db.WithContext(ctx).
		Scopes(WithTenantFromContext(ctx)).
		Where("code = ?", code).
		First(&warehouse).Error
	if err != nil {
		return nil, err
	}
	return &warehouse, nil
}

// FindDefault 查询默认仓库
func (r *WarehouseRepository) FindDefault(ctx context.Context) (*models.Warehouse, error) {
	var warehouse models.Warehouse
	err := r.db.WithContext(ctx).
		Scopes(WithTenantFromContext(ctx)).
		Where("is_default = ? AND status = ?", true, models.WarehouseStatusEnabled).
		First(&warehouse).Error
	if err != nil {
		return nil, err
	}
	return &warehouse, nil
}

// List 查询仓库列表
func (r *WarehouseRepository) List(status *int) ([]models.Warehouse, error) {
	var warehouses []models.Warehouse
	query := r.db.Model(&models.Warehouse{})
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	err := query.Order("is_default DESC, id ASC").Find(&warehouses).Error
	return warehouses, err
}

// ListWithContext 查询仓库列表（带租户上下文）
func (r *WarehouseRepository) ListWithContext(ctx context.Context, status *int) ([]models.Warehouse, error) {
	var warehouses []models.Warehouse
	query := r.db.WithContext(ctx).
		Model(&models.Warehouse{}).
		Scopes(WithTenantFromContext(ctx))
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	err := query.Order("is_default DESC, id ASC").Find(&warehouses).Error
	return warehouses, err
}

// Update 更新仓库
func (r *WarehouseRepository) Update(warehouse *models.Warehouse) error {
	return r.db.Save(warehouse).Error
}

// UpdateWithContext 更新仓库（带租户上下文）
func (r *WarehouseRepository) UpdateWithContext(ctx context.Context, warehouse *models.Warehouse) error {
	return r.db.WithContext(ctx).
		Scopes(WithTenantFromContext(ctx)).
		Save(warehouse).Error
}

// Delete 删除仓库（软删除）
func (r *WarehouseRepository) Delete(id int64) error {
	return r.db.Delete(&models.Warehouse{}, id).Error
}

// DeleteWithContext 删除仓库（带租户上下文）
func (r *WarehouseRepository) DeleteWithContext(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Scopes(WithTenantFromContext(ctx)).
		Delete(&models.Warehouse{}, id).Error
}

// SetDefault 设置默认仓库
func (r *WarehouseRepository) SetDefault(ctx context.Context, tenantID, warehouseID int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 先清除其他默认仓库
		if err := tx.Model(&models.Warehouse{}).
			Where("tenant_id = ?", tenantID).
			Update("is_default", false).Error; err != nil {
			return err
		}
		// 设置新的默认仓库
		return tx.Model(&models.Warehouse{}).
			Where("id = ? AND tenant_id = ?", warehouseID, tenantID).
			Update("is_default", true).Error
	})
}

// CountByTenantID 统计租户仓库数量
func (r *WarehouseRepository) CountByTenantID(tenantID int64) (int64, error) {
	var count int64
	err := r.db.Model(&models.Warehouse{}).Where("tenant_id = ?", tenantID).Count(&count).Error
	return count, err
}
