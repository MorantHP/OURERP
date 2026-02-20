package repository

import (
	"context"
	"time"

	"github.com/MorantHP/OURERP/internal/models"
	"gorm.io/gorm"
)

type ShopRepository struct {
	db *gorm.DB
}

func NewShopRepository(db *gorm.DB) *ShopRepository {
	return &ShopRepository{db: db}
}

// FindByID 根据ID查询店铺
func (r *ShopRepository) FindByID(id int64) (*models.Shop, error) {
	var shop models.Shop
	err := r.db.First(&shop, id).Error
	if err != nil {
		return nil, err
	}
	return &shop, nil
}

// FindByIDWithContext 根据ID查询店铺（带租户上下文）
func (r *ShopRepository) FindByIDWithContext(ctx context.Context, id int64) (*models.Shop, error) {
	var shop models.Shop
	err := r.db.WithContext(ctx).
		Scopes(WithTenantFromContext(ctx)).
		First(&shop, id).Error
	if err != nil {
		return nil, err
	}
	return &shop, nil
}

// FindByTenantID 根据租户ID和店铺ID查询
func (r *ShopRepository) FindByTenantID(tenantID, shopID int64) (*models.Shop, error) {
	var shop models.Shop
	err := r.db.Where("id = ? AND tenant_id = ?", shopID, tenantID).First(&shop).Error
	if err != nil {
		return nil, err
	}
	return &shop, nil
}

// FindByPlatform 根据平台查询店铺列表
func (r *ShopRepository) FindByPlatform(platform string) ([]models.Shop, error) {
	var shops []models.Shop
	query := r.db.Where("status = ?", models.ShopStatusEnabled)
	if platform != "" {
		query = query.Where("platform = ?", platform)
	}
	err := query.Find(&shops).Error
	return shops, err
}

// FindByPlatformWithContext 根据平台查询店铺列表（带租户上下文）
func (r *ShopRepository) FindByPlatformWithContext(ctx context.Context, platform string) ([]models.Shop, error) {
	var shops []models.Shop
	query := r.db.WithContext(ctx).
		Scopes(WithTenantFromContext(ctx)).
		Where("status = ?", models.ShopStatusEnabled)
	if platform != "" {
		query = query.Where("platform = ?", platform)
	}
	err := query.Find(&shops).Error
	return shops, err
}

// FindByTenantAndPlatform 根据租户ID和平台查询店铺列表
func (r *ShopRepository) FindByTenantAndPlatform(tenantID int64, platform string) ([]models.Shop, error) {
	var shops []models.Shop
	query := r.db.Where("tenant_id = ? AND status = ?", tenantID, models.ShopStatusEnabled)
	if platform != "" {
		query = query.Where("platform = ?", platform)
	}
	err := query.Find(&shops).Error
	return shops, err
}

// FindNeedSync 查询需要同步的店铺
func (r *ShopRepository) FindNeedSync() ([]models.Shop, error) {
	var shops []models.Shop
	err := r.db.Where("status = ?", models.ShopStatusEnabled).
		Where("last_sync_at IS NULL OR last_sync_at < ?", time.Now().Add(-time.Minute)).
		Find(&shops).Error
	return shops, err
}

// FindNeedSyncWithContext 查询需要同步的店铺（带租户上下文）
func (r *ShopRepository) FindNeedSyncWithContext(ctx context.Context) ([]models.Shop, error) {
	var shops []models.Shop
	err := r.db.WithContext(ctx).
		Scopes(WithTenantFromContext(ctx)).
		Where("status = ?", models.ShopStatusEnabled).
		Where("last_sync_at IS NULL OR last_sync_at < ?", time.Now().Add(-time.Minute)).
		Find(&shops).Error
	return shops, err
}

// FindNeedSyncByTenantID 查询指定租户需要同步的店铺
func (r *ShopRepository) FindNeedSyncByTenantID(tenantID int64) ([]models.Shop, error) {
	var shops []models.Shop
	err := r.db.Where("tenant_id = ? AND status = ?", tenantID, models.ShopStatusEnabled).
		Where("last_sync_at IS NULL OR last_sync_at < ?", time.Now().Add(-time.Minute)).
		Find(&shops).Error
	return shops, err
}

// FindExpiringTokens 查询即将过期的Token
func (r *ShopRepository) FindExpiringTokens(within time.Duration) ([]models.Shop, error) {
	var shops []models.Shop
	err := r.db.Where("status = ?", models.ShopStatusEnabled).
		Where("token_expires_at IS NOT NULL").
		Where("token_expires_at < ?", time.Now().Add(within)).
		Find(&shops).Error
	return shops, err
}

// FindExpiringTokensWithContext 查询即将过期的Token（带租户上下文）
func (r *ShopRepository) FindExpiringTokensWithContext(ctx context.Context, within time.Duration) ([]models.Shop, error) {
	var shops []models.Shop
	err := r.db.WithContext(ctx).
		Scopes(WithTenantFromContext(ctx)).
		Where("status = ?", models.ShopStatusEnabled).
		Where("token_expires_at IS NOT NULL").
		Where("token_expires_at < ?", time.Now().Add(within)).
		Find(&shops).Error
	return shops, err
}

// FindExpiringTokensByTenantID 查询指定租户即将过期的Token
func (r *ShopRepository) FindExpiringTokensByTenantID(tenantID int64, within time.Duration) ([]models.Shop, error) {
	var shops []models.Shop
	err := r.db.Where("tenant_id = ? AND status = ?", tenantID, models.ShopStatusEnabled).
		Where("token_expires_at IS NOT NULL").
		Where("token_expires_at < ?", time.Now().Add(within)).
		Find(&shops).Error
	return shops, err
}

// List 分页查询店铺列表
func (r *ShopRepository) List(page, size int, platform string, status *int) ([]models.Shop, int64, error) {
	var shops []models.Shop
	var total int64

	query := r.db.Model(&models.Shop{})
	if platform != "" {
		query = query.Where("platform = ?", platform)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	err := query.Order("id DESC").Offset(offset).Limit(size).Find(&shops).Error
	return shops, total, err
}

// ListWithContext 分页查询店铺列表（带租户上下文）
func (r *ShopRepository) ListWithContext(ctx context.Context, page, size int, platform string, status *int) ([]models.Shop, int64, error) {
	var shops []models.Shop
	var total int64

	query := r.db.WithContext(ctx).
		Model(&models.Shop{}).
		Scopes(WithTenantFromContext(ctx))
	if platform != "" {
		query = query.Where("platform = ?", platform)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	err := query.Order("id DESC").Offset(offset).Limit(size).Find(&shops).Error
	return shops, total, err
}

// ListByTenantID 根据租户ID分页查询店铺列表
func (r *ShopRepository) ListByTenantID(tenantID int64, page, size int, platform string, status *int) ([]models.Shop, int64, error) {
	var shops []models.Shop
	var total int64

	query := r.db.Model(&models.Shop{}).Where("tenant_id = ?", tenantID)
	if platform != "" {
		query = query.Where("platform = ?", platform)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	err := query.Order("id DESC").Offset(offset).Limit(size).Find(&shops).Error
	return shops, total, err
}

// Create 创建店铺
func (r *ShopRepository) Create(shop *models.Shop) error {
	return r.db.Create(shop).Error
}

// CreateWithContext 创建店铺（带租户上下文）
func (r *ShopRepository) CreateWithContext(ctx context.Context, shop *models.Shop) error {
	return r.db.WithContext(ctx).Scopes(WithTenantFromContext(ctx)).Create(shop).Error
}

// Update 更新店铺
func (r *ShopRepository) Update(shop *models.Shop) error {
	return r.db.Save(shop).Error
}

// UpdateWithContext 更新店铺（带租户上下文）
func (r *ShopRepository) UpdateWithContext(ctx context.Context, shop *models.Shop) error {
	return r.db.WithContext(ctx).
		Scopes(WithTenantFromContext(ctx)).
		Save(shop).Error
}

// UpdateToken 更新店铺Token
func (r *ShopRepository) UpdateToken(shopID int64, accessToken, refreshToken string, expiresAt time.Time) error {
	return r.db.Model(&models.Shop{}).Where("id = ?", shopID).Updates(map[string]interface{}{
		"access_token":     accessToken,
		"refresh_token":    refreshToken,
		"token_expires_at": expiresAt,
	}).Error
}

// UpdateTokenWithContext 更新店铺Token（带租户上下文）
func (r *ShopRepository) UpdateTokenWithContext(ctx context.Context, shopID int64, accessToken, refreshToken string, expiresAt time.Time) error {
	return r.db.WithContext(ctx).
		Scopes(WithTenantFromContext(ctx)).
		Model(&models.Shop{}).
		Where("id = ?", shopID).
		Updates(map[string]interface{}{
			"access_token":     accessToken,
			"refresh_token":    refreshToken,
			"token_expires_at": expiresAt,
		}).Error
}

// UpdateLastSyncAt 更新最后同步时间
func (r *ShopRepository) UpdateLastSyncAt(shopID int64) error {
	return r.db.Model(&models.Shop{}).Where("id = ?", shopID).Update("last_sync_at", time.Now()).Error
}

// UpdateLastSyncAtWithContext 更新最后同步时间（带租户上下文）
func (r *ShopRepository) UpdateLastSyncAtWithContext(ctx context.Context, shopID int64) error {
	return r.db.WithContext(ctx).
		Scopes(WithTenantFromContext(ctx)).
		Model(&models.Shop{}).
		Where("id = ?", shopID).
		Update("last_sync_at", time.Now()).Error
}

// Delete 删除店铺（软删除）
func (r *ShopRepository) Delete(id int64) error {
	return r.db.Delete(&models.Shop{}, id).Error
}

// DeleteWithContext 删除店铺（带租户上下文）
func (r *ShopRepository) DeleteWithContext(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Scopes(WithTenantFromContext(ctx)).
		Delete(&models.Shop{}, id).Error
}

// CountByTenantID 统计租户店铺数量
func (r *ShopRepository) CountByTenantID(tenantID int64) (int64, error) {
	var count int64
	err := r.db.Model(&models.Shop{}).Where("tenant_id = ?", tenantID).Count(&count).Error
	return count, err
}
