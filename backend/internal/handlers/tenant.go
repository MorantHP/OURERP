package handlers

import (
	"net/http"
	"strconv"

	"github.com/MorantHP/OURERP/internal/middleware"
	"github.com/MorantHP/OURERP/internal/models"
	"github.com/MorantHP/OURERP/internal/repository"
	"github.com/gin-gonic/gin"
)

// TenantHandler 租户处理器
type TenantHandler struct {
	tenantRepo     *repository.TenantRepository
	tenantUserRepo *repository.TenantUserRepository
}

// NewTenantHandler 创建租户处理器
func NewTenantHandler(tenantRepo *repository.TenantRepository, tenantUserRepo *repository.TenantUserRepository) *TenantHandler {
	return &TenantHandler{
		tenantRepo:     tenantRepo,
		tenantUserRepo: tenantUserRepo,
	}
}

// List 获取租户列表
// GET /api/v1/tenants
func (h *TenantHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	platform := c.Query("platform")

	var status *int
	if s := c.Query("status"); s != "" {
		v, err := strconv.Atoi(s)
		if err == nil {
			status = &v
		}
	}

	tenants, total, err := h.tenantRepo.List(page, size, status, platform)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list": tenants,
		"pagination": gin.H{
			"page":  page,
			"size": size,
			"total": total,
		},
	})
}

// MyTenants 获取当前用户可访问的租户列表
// GET /api/v1/tenants/mine
func (h *TenantHandler) MyTenants(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	uid := userID.(int64)
	tenants, err := h.tenantRepo.ListByUserID(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}

	// 获取当前选中的租户
	currentTenantID := middleware.GetTenantIDFromGin(c)

	c.JSON(http.StatusOK, gin.H{
		"list":             tenants,
		"current_tenant_id": currentTenantID,
	})
}

// Get 获取租户详情
// GET /api/v1/tenants/:id
func (h *TenantHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	tenant, err := h.tenantRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "租户不存在"})
		return
	}

	// 获取统计信息
	userCount, _ := h.tenantRepo.GetUserCount(id)

	c.JSON(http.StatusOK, gin.H{
		"tenant":     tenant,
		"user_count": userCount,
	})
}

// Create 创建租户
// POST /api/v1/tenants
func (h *TenantHandler) Create(c *gin.Context) {
	var req models.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查编码是否已存在
	if _, err := h.tenantRepo.FindByCode(req.Code); err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "租户编码已存在"})
		return
	}

	// 获取当前用户ID作为所有者
	userID, _ := c.Get("user_id")
	uid := userID.(int64)

	tenant := &models.Tenant{
		Code:        req.Code,
		Name:        req.Name,
		Platform:    req.Platform,
		Description: req.Description,
		Status:      models.TenantStatusEnabled,
		OwnerID:     uid,
	}

	if err := h.tenantRepo.Create(tenant); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		return
	}

	// 将创建者添加为租户所有者
	if err := h.tenantUserRepo.AddUser(tenant.ID, uid, models.TenantRoleOwner); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "添加用户关联失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "创建成功",
		"tenant":  tenant,
	})
}

// Update 更新租户
// PUT /api/v1/tenants/:id
func (h *TenantHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var req models.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenant, err := h.tenantRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "租户不存在"})
		return
	}

	// 更新字段
	if req.Name != "" {
		tenant.Name = req.Name
	}
	if req.Description != "" {
		tenant.Description = req.Description
	}
	if req.Logo != "" {
		tenant.Logo = req.Logo
	}
	if req.Status != nil {
		tenant.Status = *req.Status
	}

	if err := h.tenantRepo.Update(tenant); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
		"tenant":  tenant,
	})
}

// Delete 删除租户
// DELETE /api/v1/tenants/:id
func (h *TenantHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if err := h.tenantRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// AddUser 添加用户到租户
// POST /api/v1/tenants/:id/users
func (h *TenantHandler) AddUser(c *gin.Context) {
	tenantID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的租户ID"})
		return
	}

	var req struct {
		UserID int64  `json:"user_id" binding:"required"`
		Role   string `json:"role" binding:"required,oneof=owner admin member"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.tenantUserRepo.AddUser(tenantID, req.UserID, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "添加失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "添加成功"})
}

// RemoveUser 从租户移除用户
// DELETE /api/v1/tenants/:id/users/:user_id
func (h *TenantHandler) RemoveUser(c *gin.Context) {
	tenantID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的租户ID"})
		return
	}

	userID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	if err := h.tenantUserRepo.RemoveUser(tenantID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "移除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "移除成功"})
}

// UpdateUserRole 更新用户在租户中的角色
// PUT /api/v1/tenants/:id/users/:user_id/role
func (h *TenantHandler) UpdateUserRole(c *gin.Context) {
	tenantID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的租户ID"})
		return
	}

	userID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var req struct {
		Role string `json:"role" binding:"required,oneof=owner admin member"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.tenantUserRepo.UpdateRole(tenantID, userID, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

// SwitchTenant 切换当前租户
// POST /api/v1/tenants/switch
func (h *TenantHandler) SwitchTenant(c *gin.Context) {
	var req struct {
		TenantID int64 `json:"tenant_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	uid := userID.(int64)

	// 检查用户是否有权限访问该租户
	if !h.tenantUserRepo.HasAccess(uid, req.TenantID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权访问该租户"})
		return
	}

	// 验证租户是否存在
	tenant, err := h.tenantRepo.FindByID(req.TenantID)
	if err != nil || tenant.Status != models.TenantStatusEnabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "租户不存在或已禁用"})
		return
	}

	// 获取用户在该租户中的角色
	role := h.tenantUserRepo.GetUserRole(req.TenantID, uid)

	// 设置Cookie
	c.SetCookie("tenant_id", strconv.FormatInt(req.TenantID, 10), 86400*30, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"message":   "切换成功",
		"tenant_id": req.TenantID,
		"tenant": gin.H{
			"id":          tenant.ID,
			"code":        tenant.Code,
			"name":        tenant.Name,
			"platform":    tenant.Platform,
			"description": tenant.Description,
			"logo":        tenant.Logo,
			"status":      tenant.Status,
			"settings":    tenant.Settings,
			"owner_id":    tenant.OwnerID,
			"created_at":  tenant.CreatedAt,
			"updated_at":  tenant.UpdatedAt,
			"role":        role,
		},
	})
}
