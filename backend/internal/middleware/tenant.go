package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/MorantHP/OURERP/internal/models"
	"github.com/MorantHP/OURERP/internal/repository"
	"github.com/gin-gonic/gin"
)

// TenantContextKey 租户上下文键
type TenantContextKey string

const (
	// TenantKey 租户信息上下文键
	TenantKey TenantContextKey = "tenant"
)

// TenantMiddleware 租户中间件
type TenantMiddleware struct {
	tenantRepo *repository.TenantRepository
	userRepo   *repository.TenantUserRepository
}

// NewTenantMiddleware 创建租户中间件
func NewTenantMiddleware(tenantRepo *repository.TenantRepository, userRepo *repository.TenantUserRepository) *TenantMiddleware {
	return &TenantMiddleware{
		tenantRepo: tenantRepo,
		userRepo:   userRepo,
	}
}

// Handle 租户识别中间件
// 从以下来源获取租户ID（优先级从高到低）:
// 1. Query参数: ?tenant_id=xxx
// 2. Header: X-Tenant-ID
// 3. Cookie: tenant_id
// 4. 用户默认租户
func (m *TenantMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tenantID int64
		var found bool

		// 1. 从Query参数获取
		if tid := c.Query("tenant_id"); tid != "" {
			if id, err := strconv.ParseInt(tid, 10, 64); err == nil {
				tenantID = id
				found = true
			}
		}

		// 2. 从Header获取
		if !found {
			if tid := c.GetHeader("X-Tenant-ID"); tid != "" {
				if id, err := strconv.ParseInt(tid, 10, 64); err == nil {
					tenantID = id
					found = true
				}
			}
		}

		// 3. 从Cookie获取
		if !found {
			if tid, err := c.Cookie("tenant_id"); err == nil {
				if id, err := strconv.ParseInt(tid, 10, 64); err == nil {
					tenantID = id
					found = true
				}
			}
		}

		// 4. 从用户信息获取默认租户
		if !found {
			if userID, exists := c.Get("user_id"); exists {
				if uid, ok := userID.(int64); ok {
					// 获取用户的第一个可用租户作为默认
					if defaultTenantID := m.userRepo.GetDefaultTenantID(uid); defaultTenantID > 0 {
						tenantID = defaultTenantID
						found = true
					}
				}
			}
		}

		// 验证租户是否存在且用户有权限
		if found && tenantID > 0 {
			tenant, err := m.tenantRepo.FindByID(tenantID)
			if err != nil || tenant.Status != models.TenantStatusEnabled {
				c.JSON(http.StatusBadRequest, gin.H{"error": "租户不存在或已禁用"})
				c.Abort()
				return
			}

			// 检查用户是否有权限访问该租户
			if userID, exists := c.Get("user_id"); exists {
				if uid, ok := userID.(int64); ok {
					if !m.userRepo.HasAccess(uid, tenantID) {
						c.JSON(http.StatusForbidden, gin.H{"error": "无权访问该租户"})
						c.Abort()
						return
					}
				}
			}

			// 设置租户上下文 - 使用repository的函数设置tenant_id
			ctx := repository.SetTenantIDToContext(c.Request.Context(), tenantID)
			ctx = context.WithValue(ctx, TenantKey, tenant)
			c.Request = c.Request.WithContext(ctx)
			c.Set("tenant_id", tenantID)
			c.Set("tenant", tenant)
		}

		c.Next()
	}
}

// Optional 可选租户中间件（不强制要求租户）
func (m *TenantMiddleware) Optional() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tenantID int64

		// 从各种来源尝试获取租户ID
		if tid := c.Query("tenant_id"); tid != "" {
			if id, err := strconv.ParseInt(tid, 10, 64); err == nil {
				tenantID = id
			}
		}

		if tenantID == 0 {
			if tid := c.GetHeader("X-Tenant-ID"); tid != "" {
				if id, err := strconv.ParseInt(tid, 10, 64); err == nil {
					tenantID = id
				}
			}
		}

		if tenantID > 0 {
			tenant, err := m.tenantRepo.FindByID(tenantID)
			if err == nil && tenant.Status == models.TenantStatusEnabled {
				ctx := repository.SetTenantIDToContext(c.Request.Context(), tenantID)
				ctx = context.WithValue(ctx, TenantKey, tenant)
				c.Request = c.Request.WithContext(ctx)
				c.Set("tenant_id", tenantID)
				c.Set("tenant", tenant)
			}
		}

		c.Next()
	}
}

// GetTenantID 从上下文获取租户ID
func GetTenantID(ctx context.Context) int64 {
	return repository.GetTenantIDFromContext(ctx)
}

// GetTenant 从上下文获取租户信息
func GetTenant(ctx context.Context) *models.Tenant {
	if tenant, ok := ctx.Value(TenantKey).(*models.Tenant); ok {
		return tenant
	}
	return nil
}

// GetTenantIDFromGin 从Gin上下文获取租户ID
func GetTenantIDFromGin(c *gin.Context) int64 {
	if tid, exists := c.Get("tenant_id"); exists {
		if id, ok := tid.(int64); ok {
			return id
		}
	}
	return 0
}

// RequireTenant 要求租户的中间件（必须指定租户）
func (m *TenantMiddleware) RequireTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		tid := GetTenantIDFromGin(c)
		if tid == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请选择账套"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// ParseTenantIDs 解析租户ID列表（用于跨租户查询）
func ParseTenantIDs(s string) []int64 {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	ids := make([]int64, 0, len(parts))
	for _, p := range parts {
		if id, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64); err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}
