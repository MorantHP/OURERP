package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter 限流器接口
type RateLimiter interface {
	Allow(key string) bool
}

// IPRateLimiter IP限流器
type IPRateLimiter struct {
	mu       sync.RWMutex
	visitors map[string]*visitor
	rate     int           // 每秒请求数
	burst    int           // 突发请求数
	cleanup  time.Duration // 清理周期
}

type visitor struct {
	lastSeen time.Time
	tokens   int
}

// NewIPRateLimiter 创建IP限流器
func NewIPRateLimiter(rate, burst int) *IPRateLimiter {
	limiter := &IPRateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		burst:    burst,
		cleanup:  time.Minute,
	}

	// 启动清理goroutine
	go limiter.cleanupVisitors()

	return limiter
}

func (l *IPRateLimiter) cleanupVisitors() {
	for {
		time.Sleep(l.cleanup)
		l.mu.Lock()
		for ip, v := range l.visitors {
			if time.Since(v.lastSeen) > l.cleanup {
				delete(l.visitors, ip)
			}
		}
		l.mu.Unlock()
	}
}

func (l *IPRateLimiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	v, exists := l.visitors[ip]
	if !exists {
		v = &visitor{
			lastSeen: time.Now(),
			tokens:   l.burst - 1,
		}
		l.visitors[ip] = v
		return true
	}

	// 补充令牌
	elapsed := time.Since(v.lastSeen)
	tokensToAdd := int(elapsed.Seconds() * float64(l.rate))
	if tokensToAdd > 0 {
		v.tokens += tokensToAdd
		if v.tokens > l.burst {
			v.tokens = l.burst
		}
		v.lastSeen = time.Now()
	}

	if v.tokens > 0 {
		v.tokens--
		return true
	}

	return false
}

// RateLimit 限流中间件
func RateLimit(limiter RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.Allow(ip) {
			c.JSON(429, gin.H{
				"code":    "RATE_LIMIT_EXCEEDED",
				"message": "请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// DefaultRateLimit 默认限流中间件 (100 req/s, burst 200)
func DefaultRateLimit() gin.HandlerFunc {
	limiter := NewIPRateLimiter(100, 200)
	return RateLimit(limiter)
}

// StrictRateLimit 严格限流中间件 (20 req/s, burst 50)
func StrictRateLimit() gin.HandlerFunc {
	limiter := NewIPRateLimiter(20, 50)
	return RateLimit(limiter)
}
