package services

import (
	"github.com/MorantHP/OURERP/internal/models"
	"github.com/MorantHP/OURERP/internal/repository"
)

// UserPermissions 用户权限汇总
type UserPermissions struct {
	Role         *models.Role                `json:"role"`
	Permissions  []string                    `json:"permissions"`
	Shops        []int64                     `json:"shop_ids"`
	Warehouses   []int64                     `json:"warehouse_ids"`
	AllShops     bool                        `json:"all_shops"`
	AllWarehouses bool                       `json:"all_warehouses"`
	ResourcePerms []models.UserResourcePermission `json:"resource_permissions"`
}

// PermissionService 权限服务
type PermissionService struct {
	permRepo *repository.PermissionRepository
}

// NewPermissionService 创建权限服务
func NewPermissionService(permRepo *repository.PermissionRepository) *PermissionService {
	return &PermissionService{permRepo: permRepo}
}

// ==================== 角色管理 ====================

// ListRoles 获取角色列表
func (s *PermissionService) ListRoles(tenantID int64) ([]models.Role, error) {
	return s.permRepo.ListRoles(tenantID)
}

// GetRoleByID 获取角色详情
func (s *PermissionService) GetRoleByID(roleID int64) (*models.Role, error) {
	return s.permRepo.GetRoleByID(roleID)
}

// CreateRole 创建自定义角色
func (s *PermissionService) CreateRole(tenantID int64, role *models.Role, permissionCodes []string) error {
	// 设置租户ID
	role.TenantID = tenantID
	role.IsSystem = false

	// 获取权限
	perms, err := s.permRepo.GetPermissionsByCodes(permissionCodes)
	if err != nil {
		return err
	}
	role.Permissions = perms

	return s.permRepo.CreateRole(role)
}

// UpdateRole 更新角色
func (s *PermissionService) UpdateRole(tenantID, roleID int64, role *models.Role, permissionCodes []string) error {
	// 获取现有角色
	existingRole, err := s.permRepo.GetRoleByID(roleID)
	if err != nil {
		return err
	}

	// 检查是否可以修改
	if existingRole.IsSystem {
		return &PermissionError{Msg: "系统预设角色不能修改"}
	}
	if existingRole.TenantID != tenantID {
		return &PermissionError{Msg: "无权修改此角色"}
	}

	// 更新信息
	existingRole.Name = role.Name
	existingRole.Description = role.Description

	// 获取权限
	perms, err := s.permRepo.GetPermissionsByCodes(permissionCodes)
	if err != nil {
		return err
	}
	existingRole.Permissions = perms

	return s.permRepo.UpdateRole(existingRole)
}

// DeleteRole 删除角色
func (s *PermissionService) DeleteRole(tenantID, roleID int64) error {
	return s.permRepo.DeleteRole(roleID, tenantID)
}

// ==================== 权限查询 ====================

// ListPermissions 获取所有权限列表
func (s *PermissionService) ListPermissions() ([]models.Permission, error) {
	return s.permRepo.ListPermissions()
}

// GetUserPermissions 获取用户权限汇总
func (s *PermissionService) GetUserPermissions(tenantID, userID int64) (*UserPermissions, error) {
	result := &UserPermissions{
		AllShops:      true,  // 默认全部访问
		AllWarehouses: true,
	}

	// 获取用户角色
	userRole, err := s.permRepo.GetUserRole(tenantID, userID)
	if err == nil && userRole != nil {
		result.Role = userRole.Role

		// 提取权限代码
		for _, perm := range userRole.Role.Permissions {
			result.Permissions = append(result.Permissions, perm.Code)
		}
	}

	// 获取资源级权限
	resourcePerms, err := s.permRepo.GetUserResourcePermissions(tenantID, userID)
	if err == nil && len(resourcePerms) > 0 {
		result.ResourcePerms = resourcePerms
		result.AllShops = false
		result.AllWarehouses = false

		for _, rp := range resourcePerms {
			if rp.ResourceType == models.ResourceTypeShop && rp.CanRead {
				result.Shops = append(result.Shops, rp.ResourceID)
			}
			if rp.ResourceType == models.ResourceTypeWarehouse && rp.CanRead {
				result.Warehouses = append(result.Warehouses, rp.ResourceID)
			}
		}
	}

	return result, nil
}

// HasPermission 检查用户是否有指定权限
func (s *PermissionService) HasPermission(tenantID, userID int64, permissionCode string) bool {
	has, _ := s.permRepo.HasPermission(tenantID, userID, permissionCode)
	return has
}

// HasAnyPermission 检查用户是否有任一权限
func (s *PermissionService) HasAnyPermission(tenantID, userID int64, permissionCodes []string) bool {
	has, _ := s.permRepo.HasAnyPermission(tenantID, userID, permissionCodes)
	return has
}

// GetAccessibleShops 获取用户可访问的店铺ID列表
func (s *PermissionService) GetAccessibleShops(tenantID, userID int64) ([]int64, bool, error) {
	// 检查是否是所有者或管理员（全部访问）
	isOwner, err := s.permRepo.IsOwner(tenantID, userID)
	if err != nil {
		return nil, false, err
	}
	if isOwner {
		return nil, true, nil // 全部访问
	}

	// 检查是否有店铺管理权限（全部访问）
	hasShopWrite, _ := s.permRepo.HasPermission(tenantID, userID, models.PermShopWrite)
	if hasShopWrite {
		return nil, true, nil
	}

	// 获取资源级权限
	shopIDs, err := s.permRepo.GetAccessibleShops(tenantID, userID)
	if err != nil {
		return nil, false, err
	}

	// 如果没有设置资源级权限，默认全部访问
	if len(shopIDs) == 0 {
		return nil, true, nil
	}

	return shopIDs, false, nil
}

// GetAccessibleWarehouses 获取用户可访问的仓库ID列表
func (s *PermissionService) GetAccessibleWarehouses(tenantID, userID int64) ([]int64, bool, error) {
	// 检查是否是所有者或管理员（全部访问）
	isOwner, err := s.permRepo.IsOwner(tenantID, userID)
	if err != nil {
		return nil, false, err
	}
	if isOwner {
		return nil, true, nil // 全部访问
	}

	// 检查是否有仓库管理权限（全部访问）
	hasWarehouseWrite, _ := s.permRepo.HasPermission(tenantID, userID, models.PermWarehouseWrite)
	if hasWarehouseWrite {
		return nil, true, nil
	}

	// 获取资源级权限
	warehouseIDs, err := s.permRepo.GetAccessibleWarehouses(tenantID, userID)
	if err != nil {
		return nil, false, err
	}

	// 如果没有设置资源级权限，默认全部访问
	if len(warehouseIDs) == 0 {
		return nil, true, nil
	}

	return warehouseIDs, false, nil
}

// ==================== 授权管理 ====================

// SetUserRole 设置用户角色
func (s *PermissionService) SetUserRole(actorID, tenantID, targetUserID, roleID int64) error {
	// 检查是否有授权权限
	canAssign, err := s.CanAssignPermission(actorID, targetUserID, tenantID)
	if err != nil {
		return err
	}
	if !canAssign {
		return &PermissionError{Msg: "无权分配权限"}
	}

	return s.permRepo.SetUserRole(tenantID, targetUserID, roleID)
}

// SetResourcePermissions 设置用户资源权限
func (s *PermissionService) SetResourcePermissions(actorID, tenantID, targetUserID int64, perms []models.UserResourcePermission) error {
	// 检查是否有授权权限
	canAssign, err := s.CanAssignPermission(actorID, targetUserID, tenantID)
	if err != nil {
		return err
	}
	if !canAssign {
		return &PermissionError{Msg: "无权分配权限"}
	}

	return s.permRepo.SetUserResourcePermissions(tenantID, targetUserID, perms)
}

// AddResourcePermission 添加用户资源权限
func (s *PermissionService) AddResourcePermission(actorID, tenantID, targetUserID int64, perm *models.UserResourcePermission) error {
	// 检查是否有授权权限
	canAssign, err := s.CanAssignPermission(actorID, targetUserID, tenantID)
	if err != nil {
		return err
	}
	if !canAssign {
		return &PermissionError{Msg: "无权分配权限"}
	}

	perm.TenantID = tenantID
	perm.UserID = targetUserID

	return s.permRepo.AddUserResourcePermission(perm)
}

// RemoveResourcePermission 移除用户资源权限
func (s *PermissionService) RemoveResourcePermission(actorID, tenantID, targetUserID, permID int64) error {
	// 检查是否有授权权限
	canAssign, err := s.CanAssignPermission(actorID, targetUserID, tenantID)
	if err != nil {
		return err
	}
	if !canAssign {
		return &PermissionError{Msg: "无权分配权限"}
	}

	return s.permRepo.RemoveUserResourcePermission(permID, tenantID, targetUserID)
}

// CanAssignPermission 检查是否可以为目标用户分配权限
func (s *PermissionService) CanAssignPermission(actorID, targetUserID, tenantID int64) (bool, error) {
	// 不能给自己分配权限（通过此接口）
	if actorID == targetUserID {
		return false, nil
	}

	// 检查操作者是否有授权权限
	hasPerm, err := s.permRepo.HasPermission(tenantID, actorID, models.PermPermissionAssign)
	if err != nil {
		return false, err
	}
	if !hasPerm {
		return false, nil
	}

	// 不能修改所有者的权限
	isTargetOwner, err := s.permRepo.IsOwner(tenantID, targetUserID)
	if err != nil {
		return false, err
	}
	if isTargetOwner {
		return false, nil
	}

	return true, nil
}

// IsOwner 检查用户是否是租户所有者
func (s *PermissionService) IsOwner(tenantID, userID int64) bool {
	isOwner, _ := s.permRepo.IsOwner(tenantID, userID)
	return isOwner
}

// ==================== 初始化数据 ====================

// SeedData 初始化权限和角色数据
func (s *PermissionService) SeedData() error {
	// 初始化权限
	if err := s.permRepo.SeedPermissions(); err != nil {
		return err
	}

	// 初始化角色
	if err := s.permRepo.SeedRoles(); err != nil {
		return err
	}

	return nil
}

// PermissionError 权限错误
type PermissionError struct {
	Msg string
}

func (e *PermissionError) Error() string {
	return e.Msg
}
