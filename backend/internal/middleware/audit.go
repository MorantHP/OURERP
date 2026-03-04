package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditLog 审计日志
type AuditLog struct {
	UserID      int64       `json:"user_id"`
	TenantID    int64       `json:"tenant_id"`
	Method      string      `json:"method"`
	Path        string      `json:"path"`
	Query       string      `json:"query"`
	Body        interface{} `json:"body,omitempty"`
	IP          string      `json:"ip"`
	UserAgent   string      `json:"user_agent"`
	StatusCode  int         `json:"status_code"`
	Latency     int64       `json:"latency_ms"`
	Errors      string      `json:"errors,omitempty"`
	Timestamp   time.Time   `json:"timestamp"`
}

// AuditLogger 审计日志记录器
type AuditLogger interface {
	Log(log *AuditLog) error
}

// ConsoleAuditLogger 控制台审计日志
type ConsoleAuditLogger struct{}

func (l *ConsoleAuditLogger) Log(auditLog *AuditLog) error {
	data, _ := json.Marshal(auditLog)
	log.Printf("[AUDIT] %s", string(data))
	return nil
}

// AuditMiddleware 审计中间件
func AuditMiddleware(logger AuditLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 读取请求体
		var body interface{}
		if c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			
			// 尝试解析JSON
			var jsonBody interface{}
			if err := json.Unmarshal(bodyBytes, &jsonBody); err == nil {
				body = jsonBody
			}
		}

		// 处理请求
		c.Next()

		// 记录日志
		auditLog := &AuditLog{
			UserID:     getUserIDFromGin(c),
			TenantID:   GetTenantIDFromGin(c),
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			Query:      c.Request.URL.RawQuery,
			Body:       body,
			IP:         c.ClientIP(),
			UserAgent:  c.GetHeader("User-Agent"),
			StatusCode: c.Writer.Status(),
			Latency:    time.Since(start).Milliseconds(),
			Errors:     c.Errors.String(),
			Timestamp:  start,
		}

		// 只记录重要操作
		if shouldAudit(c.Request.Method, c.Request.URL.Path) {
			if err := logger.Log(auditLog); err != nil {
				log.Printf("Failed to log audit: %v", err)
			}
		}
	}
}

func getUserIDFromGin(c *gin.Context) int64 {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(int64); ok {
			return id
		}
	}
	return 0
}

func shouldAudit(method, path string) bool {
	// 记录所有写操作
	switch method {
	case "POST", "PUT", "DELETE", "PATCH":
		return true
	}
	
	// 记录敏感查询
	sensitivePaths := []string{"/login", "/logout", "/password", "/permission"}
	for _, p := range sensitivePaths {
		if len(path) >= len(p) && path[:len(p)] == p {
			return true
		}
	}
	
	return false
}

// DefaultAuditMiddleware 默认审计中间件
func DefaultAuditMiddleware() gin.HandlerFunc {
	return AuditMiddleware(&ConsoleAuditLogger{})
}
