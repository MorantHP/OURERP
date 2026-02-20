package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// SecurityConfig 安全配置
type SecurityConfig struct {
	// 速率限制
	RateLimitRequests int           // 请求数限制
	RateLimitWindow   time.Duration // 时间窗口

	// 安全头
	EnableHSTS          bool   // HTTP Strict Transport Security
	HSTSMaxAge          int    // HSTS 最大年龄（秒）
	EnableCSP           bool   // Content Security Policy
	CSPPolicy           string // CSP 策略
	EnableXSSProtection bool   // XSS 保护
	EnableFrameGuard    bool   // 点击劫持保护
	FrameGuardPolicy    string // frame guard 策略

	// CORS
	EnableCORS      bool
	AllowedOrigins  []string
	AllowedMethods  []string
	AllowedHeaders  []string
	ExposedHeaders  []string
	AllowCredentials bool
	MaxAge          int

	// 请求限制
	MaxRequestBodySize int64 // 最大请求体大小
}

// DefaultSecurityConfig 默认安全配置
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		RateLimitRequests:  100,
		RateLimitWindow:    time.Minute,
		EnableHSTS:         true,
		HSTSMaxAge:         31536000, // 1年
		EnableCSP:          false,    // 默认关闭，可能影响前端
		EnableXSSProtection: true,
		EnableFrameGuard:   true,
		FrameGuardPolicy:   "DENY",
		EnableCORS:         true,
		AllowedOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowedMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Tenant-ID", "X-Request-ID"},
		ExposedHeaders:     []string{"Content-Length", "X-Request-ID"},
		AllowCredentials:   true,
		MaxAge:             86400,
		MaxRequestBodySize: 10 * 1024 * 1024, // 10MB
	}
}

// SecurityMiddleware 安全中间件
type SecurityMiddleware struct {
	config       *SecurityConfig
	rateLimiters sync.Map // IP -> *rateLimiter
}

type rateLimiter struct {
	requests int
	window   time.Time
	mu       sync.Mutex
}

// NewSecurityMiddleware 创建安全中间件
func NewSecurityMiddleware(config *SecurityConfig) *SecurityMiddleware {
	if config == nil {
		config = DefaultSecurityConfig()
	}
	return &SecurityMiddleware{
		config: config,
	}
}

// SecurityHeaders 添加安全响应头
func (m *SecurityMiddleware) SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// X-Content-Type-Options
		c.Header("X-Content-Type-Options", "nosniff")

		// X-XSS-Protection
		if m.config.EnableXSSProtection {
			c.Header("X-XSS-Protection", "1; mode=block")
		}

		// X-Frame-Options
		if m.config.EnableFrameGuard {
			c.Header("X-Frame-Options", m.config.FrameGuardPolicy)
		}

		// Strict-Transport-Security
		if m.config.EnableHSTS {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// Content-Security-Policy
		if m.config.EnableCSP && m.config.CSPPolicy != "" {
			c.Header("Content-Security-Policy", m.config.CSPPolicy)
		}

		// Referrer-Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions-Policy
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}

// CORS 跨域中间件
func (m *SecurityMiddleware) CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.config.EnableCORS {
			c.Next()
			return
		}

		origin := c.GetHeader("Origin")
		allowedOrigin := ""

		// 检查是否允许的源
		for _, o := range m.config.AllowedOrigins {
			if o == "*" || o == origin {
				allowedOrigin = o
				break
			}
			// 支持通配符匹配
			if strings.HasPrefix(o, "*.") && strings.HasSuffix(origin, o[1:]) {
				allowedOrigin = origin
				break
			}
		}

		if allowedOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowedOrigin)
			c.Header("Access-Control-Allow-Methods", strings.Join(m.config.AllowedMethods, ", "))
			c.Header("Access-Control-Allow-Headers", strings.Join(m.config.AllowedHeaders, ", "))
			c.Header("Access-Control-Expose-Headers", strings.Join(m.config.ExposedHeaders, ", "))

			if m.config.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}

			if m.config.MaxAge > 0 {
				c.Header("Access-Control-Max-Age", string(rune(m.config.MaxAge)))
			}
		}

		// 预检请求直接返回
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RateLimit 速率限制中间件
func (m *SecurityMiddleware) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := getClientIP(c)

		limiter := m.getLimiter(ip)
		limiter.mu.Lock()
		defer limiter.mu.Unlock()

		now := time.Now()
		if now.Sub(limiter.window) > m.config.RateLimitWindow {
			limiter.requests = 0
			limiter.window = now
		}

		limiter.requests++
		if limiter.requests > m.config.RateLimitRequests {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "请求过于频繁，请稍后再试",
				"code":  "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// StrictRateLimit 严格速率限制（用于登录等敏感接口）
func (m *SecurityMiddleware) StrictRateLimit(requests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := getClientIP(c)
		key := "strict:" + ip

		limiter := m.getLimiter(key)
		limiter.mu.Lock()
		defer limiter.mu.Unlock()

		now := time.Now()
		if now.Sub(limiter.window) > window {
			limiter.requests = 0
			limiter.window = now
		}

		limiter.requests++
		if limiter.requests > requests {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "登录尝试过于频繁，请稍后再试",
				"code":  "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (m *SecurityMiddleware) getLimiter(key string) *rateLimiter {
	if limiter, ok := m.rateLimiters.Load(key); ok {
		return limiter.(*rateLimiter)
	}

	limiter := &rateLimiter{
		requests: 0,
		window:   time.Now(),
	}
	actual, _ := m.rateLimiters.LoadOrStore(key, limiter)
	return actual.(*rateLimiter)
}

// RequestSizeLimit 请求大小限制
func (m *SecurityMiddleware) RequestSizeLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if m.config.MaxRequestBodySize > 0 {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, m.config.MaxRequestBodySize)
		}
		c.Next()
	}
}

// RequestID 添加请求ID
func (m *SecurityMiddleware) RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// getClientIP 获取客户端IP
func getClientIP(c *gin.Context) string {
	// 检查代理头
	if forwarded := c.GetHeader("X-Forwarded-For"); forwarded != "" {
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	if realIP := c.GetHeader("X-Real-IP"); realIP != "" {
		return realIP
	}

	return c.ClientIP()
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 16)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}

// CleanRateLimiters 定期清理过期的限速器
func (m *SecurityMiddleware) CleanRateLimiters() {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			m.rateLimiters.Range(func(key, value interface{}) bool {
				limiter := value.(*rateLimiter)
				limiter.mu.Lock()
				if time.Since(limiter.window) > m.config.RateLimitWindow*2 {
					m.rateLimiters.Delete(key)
				}
				limiter.mu.Unlock()
				return true
			})
		}
	}()
}
