package repository

import (
	"github.com/MorantHP/OURERP/internal/models"
	"gorm.io/gorm"
)

// TenantRepository 租户仓库
type TenantRepository struct {
	db *gorm.DB
}

// NewTenantRepository 创建租户仓库
func NewTenantRepository(db *gorm.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

// FindByID 根据ID查询租户
func (r *TenantRepository) FindByID(id int64) (*models.Tenant, error) {
	var tenant models.Tenant
	err := r.db.First(&tenant, id).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

// FindByCode 根据编码查询租户
func (r *TenantRepository) FindByCode(code string) (*models.Tenant, error) {
	var tenant models.Tenant
	err := r.db.Where("code = ?", code).First(&tenant).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

// List 分页查询租户列表
func (r *TenantRepository) List(page, size int, status *int, platform string) ([]models.Tenant, int64, error) {
	var tenants []models.Tenant
	var total int64

	query := r.db.Model(&models.Tenant{})
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if platform != "" {
		query = query.Where("platform = ?", platform)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * size
	err := query.Order("id DESC").Offset(offset).Limit(size).Find(&tenants).Error
	return tenants, total, err
}

// ListByUserID 查询用户可访问的租户列表
func (r *TenantRepository) ListByUserID(userID int64) ([]models.Tenant, error) {
	var tenants []models.Tenant
	err := r.db.Table("tenants").
		Select("tenants.*, tenant_users.role").
		Joins("INNER JOIN tenant_users ON tenants.id = tenant_users.tenant_id").
		Where("tenant_users.user_id = ? AND tenants.status = ?", userID, models.TenantStatusEnabled).
		Order("tenant_users.created_at ASC").
		Find(&tenants).Error
	return tenants, err
}

// Create 创建租户
func (r *TenantRepository) Create(tenant *models.Tenant) error {
	return r.db.Create(tenant).Error
}

// Update 更新租户
func (r *TenantRepository) Update(tenant *models.Tenant) error {
	return r.db.Save(tenant).Error
}

// Delete 删除租户（软删除）
func (r *TenantRepository) Delete(id int64) error {
	return r.db.Delete(&models.Tenant{}, id).Error
}

// GetUserCount 获取租户用户数量
func (r *TenantRepository) GetUserCount(tenantID int64) (int64, error) {
	var count int64
	err := r.db.Model(&models.TenantUser{}).Where("tenant_id = ?", tenantID).Count(&count).Error
	return count, err
}

// TenantUserRepository 租户用户关联仓库
type TenantUserRepository struct {
	db *gorm.DB
}

// NewTenantUserRepository 创建租户用户关联仓库
func NewTenantUserRepository(db *gorm.DB) *TenantUserRepository {
	return &TenantUserRepository{db: db}
}

// AddUser 添加用户到租户
func (r *TenantUserRepository) AddUser(tenantID, userID int64, role string) error {
	tu := &models.TenantUser{
		TenantID: tenantID,
		UserID:   userID,
		Role:     role,
	}
	return r.db.Create(tu).Error
}

// RemoveUser 从租户移除用户
func (r *TenantUserRepository) RemoveUser(tenantID, userID int64) error {
	return r.db.Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Delete(&models.TenantUser{}).Error
}

// UpdateRole 更新用户角色
func (r *TenantUserRepository) UpdateRole(tenantID, userID int64, role string) error {
	return r.db.Model(&models.TenantUser{}).
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Update("role", role).Error
}

// HasAccess 检查用户是否有权限访问租户
func (r *TenantUserRepository) HasAccess(userID, tenantID int64) bool {
	var count int64
	r.db.Model(&models.TenantUser{}).
		Where("user_id = ? AND tenant_id = ?", userID, tenantID).
		Count(&count)
	return count > 0
}

// GetDefaultTenantID 获取用户的默认租户ID（第一个可访问的租户）
func (r *TenantUserRepository) GetDefaultTenantID(userID int64) int64 {
	var tu models.TenantUser
	err := r.db.Where("user_id = ?", userID).
		Order("created_at ASC").
		First(&tu).Error
	if err != nil {
		return 0
	}
	return tu.TenantID
}

// GetUserRole 获取用户在租户中的角色
func (r *TenantUserRepository) GetUserRole(tenantID, userID int64) string {
	var tu models.TenantUser
	err := r.db.Where("tenant_id = ? AND user_id = ?", tenantID, userID).First(&tu).Error
	if err != nil {
		return ""
	}
	return tu.Role
}

// ListUsers 获取租户的用户列表
func (r *TenantUserRepository) ListUsers(tenantID int64) ([]models.TenantUser, error) {
	var users []models.TenantUser
	err := r.db.Where("tenant_id = ?", tenantID).Find(&users).Error
	return users, err
}
