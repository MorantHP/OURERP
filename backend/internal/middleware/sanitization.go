package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// SanitizationConfig 清理配置
type SanitizationConfig struct {
	// 是否移除HTML标签
	StripHTML bool
	// 是否转义HTML
	EscapeHTML bool
	// 是否移除危险字符
	RemoveDangerousChars bool
	// 最大请求体大小
	MaxBodySize int64
}

// DefaultSanitizationConfig 默认清理配置
func DefaultSanitizationConfig() *SanitizationConfig {
	return &SanitizationConfig{
		StripHTML:            true,
		EscapeHTML:           false,
		RemoveDangerousChars: true,
		MaxBodySize:          10 * 1024 * 1024, // 10MB
	}
}

// SanitizationMiddleware 输入清理中间件
type SanitizationMiddleware struct {
	config *SanitizationConfig
}

// NewSanitizationMiddleware 创建清理中间件
func NewSanitizationMiddleware(config *SanitizationConfig) *SanitizationMiddleware {
	if config == nil {
		config = DefaultSanitizationConfig()
	}
	return &SanitizationMiddleware{config: config}
}

// Handle 处理请求
func (m *SanitizationMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只处理JSON请求
		if !strings.Contains(c.GetHeader("Content-Type"), "application/json") {
			c.Next()
			return
		}

		// 限制请求体大小
		if m.config.MaxBodySize > 0 {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, m.config.MaxBodySize)
		}

		// 读取请求体
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无法读取请求体"})
			c.Abort()
			return
		}

		// 清理JSON数据
		var data interface{}
		if len(bodyBytes) > 0 {
			if err := json.Unmarshal(bodyBytes, &data); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "无效的JSON格式"})
				c.Abort()
				return
			}

			// 递归清理数据
			data = m.sanitizeValue(data)

			// 重新编码
			cleanBytes, err := json.Marshal(data)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "数据处理失败"})
				c.Abort()
				return
			}

			// 替换请求体
			c.Request.Body = io.NopCloser(bytes.NewBuffer(cleanBytes))
		}

		c.Next()
	}
}

// sanitizeValue 递归清理值
func (m *SanitizationMiddleware) sanitizeValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		return m.sanitizeString(v)
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, val := range v {
			// 清理键名
			cleanKey := m.sanitizeString(key)
			result[cleanKey] = m.sanitizeValue(val)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, val := range v {
			result[i] = m.sanitizeValue(val)
		}
		return result
	default:
		return value
	}
}

// sanitizeString 清理字符串
func (m *SanitizationMiddleware) sanitizeString(s string) string {
	if s == "" {
		return s
	}

	// 移除危险字符
	if m.config.RemoveDangerousChars {
		s = removeDangerousChars(s)
	}

	// 移除HTML标签
	if m.config.StripHTML {
		s = stripHTML(s)
	}

	// 转义HTML
	if m.config.EscapeHTML {
		s = escapeHTML(s)
	}

	return s
}

// stripHTML 移除HTML标签
func stripHTML(s string) string {
	// 移除所有HTML标签
	re := regexp.MustCompile(`<[^>]*>`)
	s = re.ReplaceAllString(s, "")

	// 移除HTML实体
	re = regexp.MustCompile(`&[a-zA-Z]+;`)
	s = re.ReplaceAllString(s, "")

	return s
}

// escapeHTML 转义HTML特殊字符
func escapeHTML(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&#39;",
		"/", "&#47;",
	)
	return replacer.Replace(s)
}

// removeDangerousChars 移除危险字符
func removeDangerousChars(s string) string {
	// 移除null字节
	s = strings.ReplaceAll(s, "\x00", "")

	// 移除控制字符
	var result strings.Builder
	for _, r := range s {
		if r >= 32 || r == '\n' || r == '\r' || r == '\t' {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// SQLInjectionCheck 简单的SQL注入检测（GORM已经使用参数化查询，这只是额外保护）
func SQLInjectionCheck() gin.HandlerFunc {
	// 常见SQL注入模式
	sqlPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(\b(SELECT|INSERT|UPDATE|DELETE|DROP|UNION|ALTER|CREATE|TRUNCATE)\b)`),
		regexp.MustCompile(`(?i)(--|\/\*|\*\/|;|@@|@)`),
		regexp.MustCompile(`(?i)(\b(OR|AND)\b\s+\d+\s*=\s*\d+)`),
		regexp.MustCompile(`(?i)(\b(OR|AND)\b\s+['"]\w+['"]\s*=\s*['"]\w+['"])`),
	}

	return func(c *gin.Context) {
		// 检查查询参数
		for _, values := range c.Request.URL.Query() {
			for _, value := range values {
				for _, pattern := range sqlPatterns {
					if pattern.MatchString(value) && !isSafeValue(value) {
						c.JSON(http.StatusBadRequest, gin.H{
							"error": "检测到潜在的SQL注入",
							"code":  "SQL_INJECTION_DETECTED",
						})
						c.Abort()
						return
					}
				}
			}
		}

		c.Next()
	}
}

// isSafeValue 检查是否是安全值（例如正常的搜索词）
func isSafeValue(value string) bool {
	// 如果值很短且不包含特殊字符，可能是正常的搜索词
	if len(value) < 20 && !strings.ContainsAny(value, `'"\\;`) {
		return true
	}
	return false
}

// XSSProtection XSS保护中间件
func XSSProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置XSS保护头
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("X-Content-Type-Options", "nosniff")

		c.Next()
	}
}
