package handlers

import (
	"net/http"
	"strconv"

	"github.com/MorantHP/OURERP/internal/middleware"
	"github.com/MorantHP/OURERP/internal/models"
	"github.com/MorantHP/OURERP/internal/repository"
	"github.com/MorantHP/OURERP/internal/services"
	"github.com/gin-gonic/gin"
)

type PermissionHandler struct {
	permService   *services.PermissionService
	permRepo      *repository.PermissionRepository
	shopRepo      *repository.ShopRepository
	warehouseRepo *repository.WarehouseRepository
}

func NewPermissionHandler(
	permService *services.PermissionService,
	permRepo *repository.PermissionRepository,
	shopRepo *repository.ShopRepository,
	warehouseRepo *repository.WarehouseRepository,
) *PermissionHandler {
	return &PermissionHandler{
		permService:   permService,
		permRepo:      permRepo,
		shopRepo:      shopRepo,
		warehouseRepo: warehouseRepo,
	}
}

// ==================== 角色管理 ====================

// ListRoles 获取角色列表
func (h *PermissionHandler) ListRoles(c *gin.Context) {
	tenantID := middleware.GetTenantIDFromGin(c)

	roles, err := h.permService.ListRoles(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取角色列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"roles": roles})
}

// GetRole 获取角色详情
func (h *PermissionHandler) GetRole(c *gin.Context) {
	roleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的角色ID"})
		return
	}

	role, err := h.permService.GetRoleByID(roleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "角色不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"role": role})
}

// CreateRole 创建自定义角色
func (h *PermissionHandler) CreateRole(c *gin.Context) {
	tenantID := middleware.GetTenantIDFromGin(c)
	userID := middleware.GetUserIDFromGin(c)

	// 检查是否有角色管理权限
	if !h.permService.HasPermission(tenantID, userID, models.PermRoleWrite) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权创建角色"})
		return
	}

	var req struct {
		Code            string   `json:"code" binding:"required"`
		Name            string   `json:"name" binding:"required"`
		Description     string   `json:"description"`
		PermissionCodes []string `json:"permission_codes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role := &models.Role{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.permService.CreateRole(tenantID, role, req.PermissionCodes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建角色失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"role": role})
}

// UpdateRole 更新角色
func (h *PermissionHandler) UpdateRole(c *gin.Context) {
	tenantID := middleware.GetTenantIDFromGin(c)
	userID := middleware.GetUserIDFromGin(c)
	roleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的角色ID"})
		return
	}

	// 检查是否有角色管理权限
	if !h.permService.HasPermission(tenantID, userID, models.PermRoleWrite) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权修改角色"})
		return
	}

	var req struct {
		Name            string   `json:"name" binding:"required"`
		Description     string   `json:"description"`
		PermissionCodes []string `json:"permission_codes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role := &models.Role{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.permService.UpdateRole(tenantID, roleID, role, req.PermissionCodes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新角色失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// DeleteRole 删除角色
func (h *PermissionHandler) DeleteRole(c *gin.Context) {
	tenantID := middleware.GetTenantIDFromGin(c)
	userID := middleware.GetUserIDFromGin(c)
	roleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的角色ID"})
		return
	}

	// 检查是否有角色管理权限
	if !h.permService.HasPermission(tenantID, userID, models.PermRoleWrite) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权删除角色"})
		return
	}

	if err := h.permService.DeleteRole(tenantID, roleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除角色失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// ==================== 权限查询 ====================

// ListPermissions 获取所有权限列表
func (h *PermissionHandler) ListPermissions(c *gin.Context) {
	permissions, err := h.permService.ListPermissions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取权限列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"permissions": permissions})
}

// GetMyPermissions 获取当前用户权限
func (h *PermissionHandler) GetMyPermissions(c *gin.Context) {
	tenantID := middleware.GetTenantIDFromGin(c)
	userID := middleware.GetUserIDFromGin(c)

	perms, err := h.permService.GetUserPermissions(tenantID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取权限失败"})
		return
	}

	c.JSON(http.StatusOK, perms)
}

// GetUserPermissions 获取指定用户权限
func (h *PermissionHandler) GetUserPermissions(c *gin.Context) {
	tenantID := middleware.GetTenantIDFromGin(c)
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	perms, err := h.permService.GetUserPermissions(tenantID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取权限失败"})
		return
	}

	c.JSON(http.StatusOK, perms)
}

// ==================== 用户授权 ====================

// SetUserRole 设置用户角色
func (h *PermissionHandler) SetUserRole(c *gin.Context) {
	actorID := middleware.GetUserIDFromGin(c)
	tenantID := middleware.GetTenantIDFromGin(c)
	targetUserID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var req struct {
		RoleID int64 `json:"role_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.permService.SetUserRole(actorID, tenantID, targetUserID, req.RoleID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "角色设置成功"})
}

// SetResourcePermissions 设置用户资源权限
func (h *PermissionHandler) SetResourcePermissions(c *gin.Context) {
	actorID := middleware.GetUserIDFromGin(c)
	tenantID := middleware.GetTenantIDFromGin(c)
	targetUserID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var req struct {
		Permissions []models.UserResourcePermission `json:"permissions"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.permService.SetResourcePermissions(actorID, tenantID, targetUserID, req.Permissions); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "资源权限设置成功"})
}

// AddResourcePermission 添加用户资源权限
func (h *PermissionHandler) AddResourcePermission(c *gin.Context) {
	actorID := middleware.GetUserIDFromGin(c)
	tenantID := middleware.GetTenantIDFromGin(c)
	targetUserID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var perm models.UserResourcePermission
	if err := c.ShouldBindJSON(&perm); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.permService.AddResourcePermission(actorID, tenantID, targetUserID, &perm); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"permission": perm})
}

// RemoveResourcePermission 移除用户资源权限
func (h *PermissionHandler) RemoveResourcePermission(c *gin.Context) {
	actorID := middleware.GetUserIDFromGin(c)
	tenantID := middleware.GetTenantIDFromGin(c)
	targetUserID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	permID, err := strconv.ParseInt(c.Param("rid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的权限ID"})
		return
	}

	if err := h.permService.RemoveResourcePermission(actorID, tenantID, targetUserID, permID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// ==================== 可授权资源 ====================

// ListShops 获取可授权的店铺列表
func (h *PermissionHandler) ListShops(c *gin.Context) {
	tenantID := middleware.GetTenantIDFromGin(c)

	shops, _, err := h.shopRepo.ListByTenantID(tenantID, 1, 1000, "", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取店铺列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"shops": shops})
}

// ListWarehouses 获取可授权的仓库列表
func (h *PermissionHandler) ListWarehouses(c *gin.Context) {
	ctx := c.Request.Context()
	ctx = repository.SetTenantIDToContext(ctx, middleware.GetTenantIDFromGin(c))

	warehouses, err := h.warehouseRepo.ListWithContext(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取仓库列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"warehouses": warehouses})
}
