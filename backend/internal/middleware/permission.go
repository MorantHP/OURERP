package middleware

import (
	"net/http"

	"github.com/MorantHP/OURERP/internal/models"
	"github.com/MorantHP/OURERP/internal/repository"
	"github.com/MorantHP/OURERP/internal/services"
	"github.com/gin-gonic/gin"
)

// PermissionMiddleware 权限中间件
type PermissionMiddleware struct {
	permService *services.PermissionService
}

// NewPermissionMiddleware 创建权限中间件
func NewPermissionMiddleware(permService *services.PermissionService) *PermissionMiddleware {
	return &PermissionMiddleware{permService: permService}
}

// RequirePermission 检查用户是否有指定权限
func (m *PermissionMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := GetTenantIDFromGin(c)
		userID := GetUserIDFromGin(c)

		if tenantID == 0 || userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录或未选择账套"})
			c.Abort()
			return
		}

		if !m.permService.HasPermission(tenantID, userID, permission) {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权访问"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermission 检查用户是否有任一权限
func (m *PermissionMiddleware) RequireAnyPermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := GetTenantIDFromGin(c)
		userID := GetUserIDFromGin(c)

		if tenantID == 0 || userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录或未选择账套"})
			c.Abort()
			return
		}

		if !m.permService.HasAnyPermission(tenantID, userID, permissions) {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权访问"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireOwner 要求用户是租户所有者
func (m *PermissionMiddleware) RequireOwner() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := GetTenantIDFromGin(c)
		userID := GetUserIDFromGin(c)

		if tenantID == 0 || userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录或未选择账套"})
			c.Abort()
			return
		}

		if !m.permService.IsOwner(tenantID, userID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "需要主账号权限"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermissionAssign 要求用户有授权权限
func (m *PermissionMiddleware) RequirePermissionAssign() gin.HandlerFunc {
	return m.RequirePermission(models.PermPermissionAssign)
}

// FilterByAccessibleShops 过滤店铺访问权限
// 在 gin context 中设置 accessible_shop_ids 和 all_shops_accessible
func (m *PermissionMiddleware) FilterByAccessibleShops() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := GetTenantIDFromGin(c)
		userID := GetUserIDFromGin(c)

		if tenantID == 0 || userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录或未选择账套"})
			c.Abort()
			return
		}

		shopIDs, allAccessible, err := m.permService.GetAccessibleShops(tenantID, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取店铺权限失败"})
			c.Abort()
			return
		}

		c.Set("accessible_shop_ids", shopIDs)
		c.Set("all_shops_accessible", allAccessible)
		c.Next()
	}
}

// FilterByAccessibleWarehouses 过滤仓库访问权限
// 在 gin context 中设置 accessible_warehouse_ids 和 all_warehouses_accessible
func (m *PermissionMiddleware) FilterByAccessibleWarehouses() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := GetTenantIDFromGin(c)
		userID := GetUserIDFromGin(c)

		if tenantID == 0 || userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录或未选择账套"})
			c.Abort()
			return
		}

		warehouseIDs, allAccessible, err := m.permService.GetAccessibleWarehouses(tenantID, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取仓库权限失败"})
			c.Abort()
			return
		}

		c.Set("accessible_warehouse_ids", warehouseIDs)
		c.Set("all_warehouses_accessible", allAccessible)
		c.Next()
	}
}

// GetUserIDFromGin 从 gin context 获取用户ID
func GetUserIDFromGin(c *gin.Context) int64 {
	userID, exists := c.Get("userID")
	if !exists {
		return 0
	}
	return userID.(int64)
}

// GetAccessibleShopIDs 从 gin context 获取可访问的店铺ID列表
func GetAccessibleShopIDs(c *gin.Context) ([]int64, bool) {
	shopIDs, exists := c.Get("accessible_shop_ids")
	if !exists {
		return nil, false
	}
	allAccessible, _ := c.Get("all_shops_accessible")
	return shopIDs.([]int64), allAccessible.(bool)
}

// GetAccessibleWarehouseIDs 从 gin context 获取可访问的仓库ID列表
func GetAccessibleWarehouseIDs(c *gin.Context) ([]int64, bool) {
	warehouseIDs, exists := c.Get("accessible_warehouse_ids")
	if !exists {
		return nil, false
	}
	allAccessible, _ := c.Get("all_warehouses_accessible")
	return warehouseIDs.([]int64), allAccessible.(bool)
}

// ==================== 独立函数版本（兼容旧代码） ====================

// 全局权限服务实例
var globalPermService *services.PermissionService

// SetGlobalPermissionService 设置全局权限服务
func SetGlobalPermissionService(permService *services.PermissionService) {
	globalPermService = permService
}

// RequirePermissionFunc 检查权限的独立函数
func RequirePermissionFunc(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if globalPermService == nil {
			c.Next()
			return
		}

		tenantID := GetTenantIDFromGin(c)
		userID := GetUserIDFromGin(c)

		if tenantID == 0 || userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录或未选择账套"})
			c.Abort()
			return
		}

		if !globalPermService.HasPermission(tenantID, userID, permission) {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权访问"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CheckResourceAccess 检查资源访问权限
func CheckResourceAccess(c *gin.Context, resourceType string, resourceID int64, action string) bool {
	if globalPermService == nil {
		return true // 如果没有初始化权限服务，默认允许
	}

	tenantID := GetTenantIDFromGin(c)
	userID := GetUserIDFromGin(c)

	// 根据资源类型检查
	switch resourceType {
	case "shop":
		shopIDs, allAccessible := GetAccessibleShopIDs(c)
		if allAccessible {
			return true
		}
		for _, id := range shopIDs {
			if id == resourceID {
				return true
			}
		}
		return false

	case "warehouse":
		warehouseIDs, allAccessible := GetAccessibleWarehouseIDs(c)
		if allAccessible {
			return true
		}
		for _, id := range warehouseIDs {
			if id == resourceID {
				return true
			}
		}
		return false

	default:
		// 其他资源类型使用权限检查
		permCode := resourceType + ":" + action
		return globalPermService.HasPermission(tenantID, userID, permCode)
	}
}

// WithTenantScope 带租户作用域的数据库查询
// 使用 repository 包的 WithTenantFromContext
func WithTenantScope(ctx interface{}) func(db interface{}) interface{} {
	return func(db interface{}) interface{} {
		// 这里需要根据实际情况实现
		// 使用 repository.WithTenantFromContext
		return db
	}
}

// GetUserPermissionsFromContext 从context获取用户权限信息
func GetUserPermissionsFromContext(c *gin.Context) *services.UserPermissions {
	tenantID := GetTenantIDFromGin(c)
	userID := GetUserIDFromGin(c)

	if globalPermService == nil || tenantID == 0 || userID == 0 {
		return nil
	}

	perms, err := globalPermService.GetUserPermissions(tenantID, userID)
	if err != nil {
		return nil
	}
	return perms
}

// 确保接口兼容
var _ = repository.GetTenantIDFromContext
