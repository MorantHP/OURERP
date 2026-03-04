package middleware

import (
	"log"
	"runtime/debug"

	"github.com/MorantHP/OURERP/internal/pkg/errors"
	"github.com/MorantHP/OURERP/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

// ErrorHandler 全局错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic recovered: %v\n%s", r, debug.Stack())
				c.JSON(500, gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "服务器内部错误",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// Recovery 错误恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[Recovery] panic recovered:\n%s", debug.Stack())
				
				if appErr, ok := err.(*errors.AppError); ok {
					response.Error(c, appErr)
				} else {
					response.Error(c, errors.ErrInternal)
				}
				c.Abort()
			}
		}()
		c.Next()
	}
}
