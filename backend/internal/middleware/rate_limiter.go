package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// SimpleRateLimiter 简单速率限制器（计数器算法）
type SimpleRateLimiter struct {
	visitors map[string]*simpleVisitor
	mu       sync.RWMutex
	rate     int           // 每分钟最大请求数
	window   time.Duration // 时间窗口
}

type simpleVisitor struct {
	lastSeen time.Time
	count    int
}

// NewSimpleRateLimiter 创建简单速率限制器
func NewSimpleRateLimiter(rate int, window time.Duration) *SimpleRateLimiter {
	limiter := &SimpleRateLimiter{
		visitors: make(map[string]*simpleVisitor),
		rate:     rate,
		window:   window,
	}

	// 启动后台清理goroutine
	go limiter.cleanupVisitors()

	return limiter
}

// cleanupVisitors 定期清理过期的访问记录
func (rl *SimpleRateLimiter) cleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > rl.window {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// getVisitor 获取或创建访问者记录
func (rl *SimpleRateLimiter) getVisitor(ip string) *simpleVisitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		v = &simpleVisitor{
			lastSeen: time.Now(),
			count:    0,
		}
		rl.visitors[ip] = v
	}

	// 如果超过时间窗口，重置计数
	if time.Since(v.lastSeen) > rl.window {
		v.count = 0
		v.lastSeen = time.Now()
	}

	return v
}

// Allow 检查是否允许请求
func (rl *SimpleRateLimiter) Allow(ip string) bool {
	v := rl.getVisitor(ip)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	v.count++
	v.lastSeen = time.Now()

	return v.count <= rl.rate
}

// SimpleRateLimit 简单速率限制中间件
// rate: 每分钟最大请求数
func SimpleRateLimit(rate int) gin.HandlerFunc {
	limiter := NewSimpleRateLimiter(rate, time.Minute)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !limiter.Allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// LoginRateLimit 登录专用速率限制（更严格）
// 每个IP每分钟最多5次登录尝试
func LoginRateLimit() gin.HandlerFunc {
	return SimpleRateLimit(5)
}
