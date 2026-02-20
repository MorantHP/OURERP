package repository

import (
	"github.com/MorantHP/OURERP/internal/models"
	"gorm.io/gorm"
)

type PermissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

// ==================== 角色管理 ====================

// ListRoles 获取角色列表（系统预设 + 租户自定义）
func (r *PermissionRepository) ListRoles(tenantID int64) ([]models.Role, error) {
	var roles []models.Role
	err := r.db.Where("tenant_id = 0 OR tenant_id = ?", tenantID).
		Preload("Permissions").
		Order("tenant_id ASC, id ASC").
		Find(&roles).Error
	return roles, err
}

// GetRoleByID 根据ID获取角色
func (r *PermissionRepository) GetRoleByID(roleID int64) (*models.Role, error) {
	var role models.Role
	err := r.db.Preload("Permissions").First(&role, roleID).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// GetRoleByCode 根据Code获取角色
func (r *PermissionRepository) GetRoleByCode(code string, tenantID int64) (*models.Role, error) {
	var role models.Role
	err := r.db.Where("code = ? AND (tenant_id = 0 OR tenant_id = ?)", code, tenantID).
		Preload("Permissions").
		First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// CreateRole 创建自定义角色
func (r *PermissionRepository) CreateRole(role *models.Role) error {
	return r.db.Create(role).Error
}

// UpdateRole 更新角色
func (r *PermissionRepository) UpdateRole(role *models.Role) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 更新角色基本信息
		if err := tx.Save(role).Error; err != nil {
			return err
		}
		// 更新权限关联
		if err := tx.Model(role).Association("Permissions").Replace(role.Permissions); err != nil {
			return err
		}
		return nil
	})
}

// DeleteRole 删除角色（只能删除自定义角色）
func (r *PermissionRepository) DeleteRole(roleID, tenantID int64) error {
	result := r.db.Where("id = ? AND tenant_id = ? AND is_system = false", roleID, tenantID).
		Delete(&models.Role{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ==================== 权限管理 ====================

// ListPermissions 获取所有权限列表
func (r *PermissionRepository) ListPermissions() ([]models.Permission, error) {
	var permissions []models.Permission
	err := r.db.Order("resource, action").Find(&permissions).Error
	return permissions, err
}

// GetPermissionsByCodes 根据权限代码获取权限
func (r *PermissionRepository) GetPermissionsByCodes(codes []string) ([]models.Permission, error) {
	var permissions []models.Permission
	err := r.db.Where("code IN ?", codes).Find(&permissions).Error
	return permissions, err
}

// SeedPermissions 初始化权限数据
func (r *PermissionRepository) SeedPermissions() error {
	// 检查是否已初始化
	var count int64
	r.db.Model(&models.Permission{}).Count(&count)
	if count > 0 {
		return nil
	}

	permissions := []models.Permission{
		// 订单权限
		{Code: models.PermOrderRead, Name: "查看订单", Resource: "order", Action: "read", Description: "查看订单列表和详情"},
		{Code: models.PermOrderWrite, Name: "编辑订单", Resource: "order", Action: "write", Description: "创建和编辑订单"},
		{Code: models.PermOrderDelete, Name: "删除订单", Resource: "order", Action: "delete", Description: "删除订单"},
		{Code: models.PermOrderAudit, Name: "审核订单", Resource: "order", Action: "audit", Description: "审核订单"},
		{Code: models.PermOrderShip, Name: "订单发货", Resource: "order", Action: "ship", Description: "订单发货操作"},

		// 商品权限
		{Code: models.PermProductRead, Name: "查看商品", Resource: "product", Action: "read", Description: "查看商品列表和详情"},
		{Code: models.PermProductWrite, Name: "编辑商品", Resource: "product", Action: "write", Description: "创建和编辑商品"},
		{Code: models.PermProductDelete, Name: "删除商品", Resource: "product", Action: "delete", Description: "删除商品"},

		// 库存权限
		{Code: models.PermInventoryRead, Name: "查看库存", Resource: "inventory", Action: "read", Description: "查看库存列表和详情"},
		{Code: models.PermInventoryWrite, Name: "编辑库存", Resource: "inventory", Action: "write", Description: "调整库存"},

		// 仓库权限
		{Code: models.PermWarehouseRead, Name: "查看仓库", Resource: "warehouse", Action: "read", Description: "查看仓库列表和详情"},
		{Code: models.PermWarehouseWrite, Name: "编辑仓库", Resource: "warehouse", Action: "write", Description: "创建和编辑仓库"},

		// 用户权限
		{Code: models.PermUserRead, Name: "查看用户", Resource: "user", Action: "read", Description: "查看用户列表和详情"},
		{Code: models.PermUserWrite, Name: "编辑用户", Resource: "user", Action: "write", Description: "创建和编辑用户"},
		{Code: models.PermUserDelete, Name: "删除用户", Resource: "user", Action: "delete", Description: "删除用户"},

		// 店铺权限
		{Code: models.PermShopRead, Name: "查看店铺", Resource: "shop", Action: "read", Description: "查看店铺列表和详情"},
		{Code: models.PermShopWrite, Name: "编辑店铺", Resource: "shop", Action: "write", Description: "创建和编辑店铺"},
		{Code: models.PermShopDelete, Name: "删除店铺", Resource: "shop", Action: "delete", Description: "删除店铺"},

		// 财务权限
		{Code: models.PermFinanceRead, Name: "查看财务", Resource: "finance", Action: "read", Description: "查看财务数据"},
		{Code: models.PermFinanceWrite, Name: "编辑财务", Resource: "finance", Action: "write", Description: "编辑财务数据"},
		{Code: models.PermFinanceExport, Name: "导出财务", Resource: "finance", Action: "export", Description: "导出财务报表"},

		// 报表权限
		{Code: models.PermReportRead, Name: "查看报表", Resource: "report", Action: "read", Description: "查看统计报表"},
		{Code: models.PermReportExport, Name: "导出报表", Resource: "report", Action: "export", Description: "导出统计报表"},

		// 角色权限
		{Code: models.PermRoleRead, Name: "查看角色", Resource: "role", Action: "read", Description: "查看角色列表"},
		{Code: models.PermRoleWrite, Name: "编辑角色", Resource: "role", Action: "write", Description: "创建和编辑角色"},

		// 授权权限
		{Code: models.PermPermissionAssign, Name: "分配权限", Resource: "permission", Action: "assign", Description: "为用户分配权限"},

		// 系统权限
		{Code: models.PermSystemConfig, Name: "系统配置", Resource: "system", Action: "config", Description: "系统配置管理"},
		{Code: models.PermSystemLog, Name: "系统日志", Resource: "system", Action: "log", Description: "查看系统日志"},
	}

	return r.db.Create(&permissions).Error
}

// SeedRoles 初始化角色数据
func (r *PermissionRepository) SeedRoles() error {
	// 检查是否已初始化
	var count int64
	r.db.Model(&models.Role{}).Count(&count)
	if count > 0 {
		return nil
	}

	// 获取所有权限
	allPerms, err := r.ListPermissions()
	if err != nil {
		return err
	}
	permMap := make(map[string]models.Permission)
	for _, p := range allPerms {
		permMap[p.Code] = p
	}

	// 创建预设角色
	for roleType, permCodes := range models.RolePermissions {
		role := models.Role{
			Code:        string(roleType),
			Name:        getRoleName(roleType),
			Description: getRoleDescription(roleType),
			IsSystem:    true,
			TenantID:    0,
		}

		// 获取角色权限
		var perms []models.Permission
		for _, code := range permCodes {
			if p, ok := permMap[code]; ok {
				perms = append(perms, p)
			}
		}
		role.Permissions = perms

		if err := r.db.Create(&role).Error; err != nil {
			return err
		}
	}

	return nil
}

func getRoleName(roleType models.RoleType) string {
	names := map[models.RoleType]string{
		models.RoleOwner:     "主账号",
		models.RoleAdmin:     "管理员",
		models.RoleManager:   "经理",
		models.RoleOperator:  "运营",
		models.RoleFinance:   "财务",
		models.RoleWarehouse: "仓库",
		models.RoleCustomer:  "客服",
		models.RoleViewer:    "只读",
	}
	return names[roleType]
}

func getRoleDescription(roleType models.RoleType) string {
	descs := map[models.RoleType]string{
		models.RoleOwner:     "账套所有者，拥有全部权限和授权管理能力",
		models.RoleAdmin:     "管理员，拥有全部业务权限",
		models.RoleManager:   "经理，管理订单、商品、库存等业务",
		models.RoleOperator:  "运营人员，负责订单处理",
		models.RoleFinance:   "财务人员，查看财务和报表",
		models.RoleWarehouse: "仓库人员，管理库存和出入库",
		models.RoleCustomer:  "客服人员，处理订单问题",
		models.RoleViewer:    "只读用户，只能查看数据",
	}
	return descs[roleType]
}

// ==================== 用户角色关联 ====================

// GetUserRole 获取用户在租户中的角色
func (r *PermissionRepository) GetUserRole(tenantID, userID int64) (*models.UserRole, error) {
	var userRole models.UserRole
	err := r.db.Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Preload("Role.Permissions").
		First(&userRole).Error
	if err != nil {
		return nil, err
	}
	return &userRole, nil
}

// SetUserRole 设置用户角色
func (r *PermissionRepository) SetUserRole(tenantID, userID, roleID int64) error {
	var userRole models.UserRole
	err := r.db.Where("tenant_id = ? AND user_id = ?", tenantID, userID).First(&userRole).Error
	if err == gorm.ErrRecordNotFound {
		// 创建新的角色关联
		userRole = models.UserRole{
			TenantID: tenantID,
			UserID:   userID,
			RoleID:   roleID,
		}
		return r.db.Create(&userRole).Error
	}
	if err != nil {
		return err
	}
	// 更新角色
	return r.db.Model(&userRole).Update("role_id", roleID).Error
}

// RemoveUserRole 移除用户角色
func (r *PermissionRepository) RemoveUserRole(tenantID, userID int64) error {
	return r.db.Where("tenant_id = ? AND user_id = ?", tenantID, userID).Delete(&models.UserRole{}).Error
}

// ==================== 用户资源权限 ====================

// GetUserResourcePermissions 获取用户资源权限列表
func (r *PermissionRepository) GetUserResourcePermissions(tenantID, userID int64) ([]models.UserResourcePermission, error) {
	var perms []models.UserResourcePermission
	err := r.db.Where("tenant_id = ? AND user_id = ?", tenantID, userID).Find(&perms).Error
	return perms, err
}

// SetUserResourcePermissions 设置用户资源权限
func (r *PermissionRepository) SetUserResourcePermissions(tenantID, userID int64, perms []models.UserResourcePermission) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 删除旧的权限
		if err := tx.Where("tenant_id = ? AND user_id = ?", tenantID, userID).
			Delete(&models.UserResourcePermission{}).Error; err != nil {
			return err
		}
		// 创建新的权限
		if len(perms) > 0 {
			for i := range perms {
				perms[i].TenantID = tenantID
				perms[i].UserID = userID
			}
			if err := tx.Create(&perms).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// AddUserResourcePermission 添加用户资源权限
func (r *PermissionRepository) AddUserResourcePermission(perm *models.UserResourcePermission) error {
	return r.db.Create(perm).Error
}

// RemoveUserResourcePermission 移除用户资源权限
func (r *PermissionRepository) RemoveUserResourcePermission(id, tenantID, userID int64) error {
	return r.db.Where("id = ? AND tenant_id = ? AND user_id = ?", id, tenantID, userID).
		Delete(&models.UserResourcePermission{}).Error
}

// GetAccessibleShops 获取用户可访问的店铺ID列表
func (r *PermissionRepository) GetAccessibleShops(tenantID, userID int64) ([]int64, error) {
	var perms []models.UserResourcePermission
	err := r.db.Where("tenant_id = ? AND user_id = ? AND resource_type = ? AND can_read = ?",
		tenantID, userID, models.ResourceTypeShop, true).
		Find(&perms).Error
	if err != nil {
		return nil, err
	}

	ids := make([]int64, len(perms))
	for i, p := range perms {
		ids[i] = p.ResourceID
	}
	return ids, nil
}

// GetAccessibleWarehouses 获取用户可访问的仓库ID列表
func (r *PermissionRepository) GetAccessibleWarehouses(tenantID, userID int64) ([]int64, error) {
	var perms []models.UserResourcePermission
	err := r.db.Where("tenant_id = ? AND user_id = ? AND resource_type = ? AND can_read = ?",
		tenantID, userID, models.ResourceTypeWarehouse, true).
		Find(&perms).Error
	if err != nil {
		return nil, err
	}

	ids := make([]int64, len(perms))
	for i, p := range perms {
		ids[i] = p.ResourceID
	}
	return ids, nil
}

// ==================== 权限检查 ====================

// HasPermission 检查用户是否有指定权限
func (r *PermissionRepository) HasPermission(tenantID, userID int64, permissionCode string) (bool, error) {
	// 获取用户角色
	userRole, err := r.GetUserRole(tenantID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}

	// 检查角色权限
	for _, perm := range userRole.Role.Permissions {
		if perm.Code == permissionCode {
			return true, nil
		}
	}

	return false, nil
}

// HasAnyPermission 检查用户是否有任一权限
func (r *PermissionRepository) HasAnyPermission(tenantID, userID int64, permissionCodes []string) (bool, error) {
	for _, code := range permissionCodes {
		has, err := r.HasPermission(tenantID, userID, code)
		if err != nil {
			return false, err
		}
		if has {
			return true, nil
		}
	}
	return false, nil
}

// IsOwner 检查用户是否是租户所有者
func (r *PermissionRepository) IsOwner(tenantID, userID int64) (bool, error) {
	userRole, err := r.GetUserRole(tenantID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return userRole.Role.Code == string(models.RoleOwner), nil
}
