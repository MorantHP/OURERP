// cmd/server/main.go（更新版）
package main

import (
	"net/http"

	"github.com/MorantHP/OURERP/internal/config"
	"github.com/MorantHP/OURERP/internal/models"
	"github.com/MorantHP/OURERP/internal/repository"
	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 设置运行模式
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化数据库
	db := repository.InitDB(&cfg.Database)

	// 自动迁移
	db.AutoMigrate(&models.User{}, &models.Order{}, &models.OrderItem{})

	// 创建路由
	r := gin.Default()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "OURERP Server Running",
			"env":     cfg.Env,
		})
	})

	// API路由组
	api := r.Group("/api/v1")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "pong"})
		})
	}

	// 启动
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	r.Run(addr)
}
