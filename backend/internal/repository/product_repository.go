package repository

import (
	"context"

	"github.com/MorantHP/OURERP/internal/models"
	"gorm.io/gorm"
)

// ProductRepository 商品仓库
type ProductRepository struct {
	db *gorm.DB
}

// NewProductRepository 创建商品仓库
func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// Create 创建商品
func (r *ProductRepository) Create(product *models.Product) error {
	return r.db.Create(product).Error
}

// CreateWithContext 创建商品（带租户上下文）
func (r *ProductRepository) CreateWithContext(ctx context.Context, product *models.Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

// FindByID 根据ID查询商品
func (r *ProductRepository) FindByID(id int64) (*models.Product, error) {
	var product models.Product
	err := r.db.First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// FindByIDWithContext 根据ID查询商品（带租户上下文）
func (r *ProductRepository) FindByIDWithContext(ctx context.Context, id int64) (*models.Product, error) {
	var product models.Product
	err := r.db.WithContext(ctx).
		Scopes(WithTenantFromContext(ctx)).
		First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// FindBySkuCode 根据SKU编码查询商品
func (r *ProductRepository) FindBySkuCode(ctx context.Context, skuCode string) (*models.Product, error) {
	var product models.Product
	err := r.db.WithContext(ctx).
		Scopes(WithTenantFromContext(ctx)).
		Where("sku_code = ?", skuCode).
		First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// FindByBarcode 根据条码查询商品
func (r *ProductRepository) FindByBarcode(ctx context.Context, barcode string) (*models.Product, error) {
	var product models.Product
	err := r.db.WithContext(ctx).
		Scopes(WithTenantFromContext(ctx)).
		Where("barcode = ?", barcode).
		First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// List 分页查询商品列表
func (r *ProductRepository) List(page, size int, category, brand, keyword string, status *int) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	query := r.db.Model(&models.Product{})
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if brand != "" {
		query = query.Where("brand = ?", brand)
	}
	if keyword != "" {
		query = query.Where("sku_code LIKE ? OR name LIKE ? OR barcode LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	err := query.Order("id DESC").Offset(offset).Limit(size).Find(&products).Error
	return products, total, err
}

// ListWithContext 分页查询商品列表（带租户上下文）
func (r *ProductRepository) ListWithContext(ctx context.Context, page, size int, category, brand, keyword string, status *int) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	query := r.db.WithContext(ctx).
		Model(&models.Product{}).
		Scopes(WithTenantFromContext(ctx))
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if brand != "" {
		query = query.Where("brand = ?", brand)
	}
	if keyword != "" {
		query = query.Where("sku_code LIKE ? OR name LIKE ? OR barcode LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	err := query.Order("id DESC").Offset(offset).Limit(size).Find(&products).Error
	return products, total, err
}

// Update 更新商品
func (r *ProductRepository) Update(product *models.Product) error {
	return r.db.Save(product).Error
}

// UpdateWithContext 更新商品（带租户上下文）
func (r *ProductRepository) UpdateWithContext(ctx context.Context, product *models.Product) error {
	return r.db.WithContext(ctx).
		Scopes(WithTenantFromContext(ctx)).
		Save(product).Error
}

// Delete 删除商品（软删除）
func (r *ProductRepository) Delete(id int64) error {
	return r.db.Delete(&models.Product{}, id).Error
}

// DeleteWithContext 删除商品（带租户上下文）
func (r *ProductRepository) DeleteWithContext(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Scopes(WithTenantFromContext(ctx)).
		Delete(&models.Product{}, id).Error
}

// GetCategories 获取所有分类
func (r *ProductRepository) GetCategories(ctx context.Context) ([]string, error) {
	var categories []string
	err := r.db.WithContext(ctx).
		Scopes(WithTenantFromContext(ctx)).
		Model(&models.Product{}).
		Distinct("category").
		Where("category != ''").
		Pluck("category", &categories).Error
	return categories, err
}

// GetBrands 获取所有品牌
func (r *ProductRepository) GetBrands(ctx context.Context) ([]string, error) {
	var brands []string
	err := r.db.WithContext(ctx).
		Scopes(WithTenantFromContext(ctx)).
		Model(&models.Product{}).
		Distinct("brand").
		Where("brand != ''").
		Pluck("brand", &brands).Error
	return brands, err
}

// CountByTenantID 统计租户商品数量
func (r *ProductRepository) CountByTenantID(tenantID int64) (int64, error) {
	var count int64
	err := r.db.Model(&models.Product{}).Where("tenant_id = ?", tenantID).Count(&count).Error
	return count, err
}
