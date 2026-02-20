package middleware

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/MorantHP/OURERP/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// TokenBlacklist Token黑名单接口
type TokenBlacklist interface {
	Add(ctx context.Context, token string, expiration time.Duration) error
	Exists(ctx context.Context, token string) (bool, error)
}

// AuthMiddleware 认证中间件增强版
type AuthMiddleware struct {
	blacklist TokenBlacklist
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(blacklist TokenBlacklist) *AuthMiddleware {
	return &AuthMiddleware{
		blacklist: blacklist,
	}
}

// CORS 跨域中间件 - 支持从环境变量配置允许的域名
func CORS() gin.HandlerFunc {
	// 从环境变量获取允许的域名，多个域名用逗号分隔
	// 生产环境必须设置 CORS_ALLOWED_ORIGINS，如: "https://example.com,https://admin.example.com"
	allowedOriginsEnv := os.Getenv("CORS_ALLOWED_ORIGINS")
	var allowedOrigins []string
	if allowedOriginsEnv != "" {
		allowedOrigins = strings.Split(allowedOriginsEnv, ",")
		for i := range allowedOrigins {
			allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
		}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowOrigin := ""

		// 检查origin是否在允许列表中
		for _, allowed := range allowedOrigins {
			if allowed == origin {
				allowOrigin = origin
				break
			}
		}

		// 开发环境允许localhost（仅当ENV=development时）
		if config.GlobalConfig.Env == "development" && allowOrigin == "" {
			if strings.HasPrefix(origin, "http://localhost") || strings.HasPrefix(origin, "http://127.0.0.1") {
				allowOrigin = origin
			}
		}

		if allowOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowOrigin)
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Tenant-ID")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证信息", "code": "AUTH_MISSING"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证格式错误", "code": "AUTH_FORMAT_ERROR"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(config.GlobalConfig.JWT.Secret), nil
		})

		if err != nil || !token.Valid {
			errMsg := "无效的token"
			code := "TOKEN_INVALID"
			if errors.Is(err, jwt.ErrTokenExpired) {
				errMsg = "Token已过期"
				code = "TOKEN_EXPIRED"
			} else if errors.Is(err, jwt.ErrTokenMalformed) {
				errMsg = "Token格式错误"
				code = "TOKEN_MALFORMED"
			} else if errors.Is(err, jwt.ErrTokenNotValidYet) {
				errMsg = "Token尚未生效"
				code = "TOKEN_NOT_VALID_YET"
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": errMsg, "code": code})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if userID, ok := claims["user_id"]; ok {
				c.Set("user_id", int64(userID.(float64)))
			}
			if email, ok := claims["email"]; ok {
				c.Set("email", email.(string))
			}
		}

		c.Next()
	}
}

// JWTAuthWithBlacklist 带黑名单检查的JWT认证
func (m *AuthMiddleware) JWTAuthWithBlacklist() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证信息", "code": "AUTH_MISSING"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证格式错误", "code": "AUTH_FORMAT_ERROR"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 检查黑名单
		if m.blacklist != nil {
			inBlacklist, err := m.blacklist.Exists(c.Request.Context(), tokenString)
			if err == nil && inBlacklist {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token已失效", "code": "TOKEN_REVOKED"})
				c.Abort()
				return
			}
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(config.GlobalConfig.JWT.Secret), nil
		})

		if err != nil || !token.Valid {
			errMsg := "无效的token"
			code := "TOKEN_INVALID"
			if errors.Is(err, jwt.ErrTokenExpired) {
				errMsg = "Token已过期"
				code = "TOKEN_EXPIRED"
			} else if errors.Is(err, jwt.ErrTokenMalformed) {
				errMsg = "Token格式错误"
				code = "TOKEN_MALFORMED"
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": errMsg, "code": code})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if userID, ok := claims["user_id"]; ok {
				c.Set("user_id", int64(userID.(float64)))
			}
			if email, ok := claims["email"]; ok {
				c.Set("email", email.(string))
			}
			// 存储token以便后续可以将其加入黑名单
			c.Set("token", tokenString)
		}

		c.Next()
	}
}

// OptionalAuth 可选认证（不强制要求登录）
func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(config.GlobalConfig.JWT.Secret), nil
		})

		if err == nil && token.Valid {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				if userID, ok := claims["user_id"]; ok {
					c.Set("user_id", int64(userID.(float64)))
				}
				if email, ok := claims["email"]; ok {
					c.Set("email", email.(string))
				}
			}
		}

		c.Next()
	}
}
