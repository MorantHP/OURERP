package middleware

import (
	"fmt"

	"github.com/MorantHP/OURERP/internal/pkg/errors"
	"github.com/MorantHP/OURERP/internal/pkg/response"
	"github.com/MorantHP/OURERP/internal/repository"
	"github.com/gin-gonic/gin"
)

// TenantMiddleware 租户中间件 - 强制要求租户上下文
func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := GetTenantIDFromGin(c)

		// 对于需要租户的接口，必须有租户ID
		if tenantID == 0 {
			response.Error(c, errors.ErrTenantNotSelected)
			c.Abort()
			return
		}

		// 将租户ID注入到 Gin 上下文
		c.Set("tenant_id", tenantID)

		// 将租户ID注入到 request context
		ctx := repository.SetTenantIDToContext(c.Request.Context(), tenantID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// OptionalTenantMiddleware 可选租户中间件 - 不强制要求
func OptionalTenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := GetTenantIDFromGin(c)
		
		if tenantID > 0 {
			ctx := repository.SetTenantIDToContext(c.Request.Context(), tenantID)
			c.Request = c.Request.WithContext(ctx)
		}
		
		c.Next()
	}
}

// GetTenantIDFromGin 从Gin上下文获取租户ID
func GetTenantIDFromGin(c *gin.Context) int64 {
	// 优先从请求头获取 X-Tenant-ID
	if tid := c.GetHeader("X-Tenant-ID"); tid != "" {
		var id int64
		if _, err := fmt.Sscanf(tid, "%d", &id); err == nil {
			return id
		}
	}

	// 从 Cookie 获取
	if tid, err := c.Cookie("tenant_id"); err == nil {
		var id int64
		if _, err := fmt.Sscanf(tid, "%d", &id); err == nil {
			return id
		}
	}

	// 从上下文获取
	if tid, exists := c.Get("tenant_id"); exists {
		if intID, ok := tid.(int64); ok {
			return intID
		}
	}

	return 0
}
