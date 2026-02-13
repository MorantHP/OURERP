package main

import (
	"github.com/MorantHP/OURERP/internal/config"
	"github.com/MorantHP/OURERP/internal/handlers"
	"github.com/MorantHP/OURERP/internal/middleware"
	"github.com/MorantHP/OURERP/internal/mock"
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

	// 创建仓库
	userRepo := repository.NewUserRepository(db)
	orderRepo := repository.NewOrderRepository(db)

	// 创建处理器
	authHandler := handlers.NewAuthHandler(userRepo, cfg)
	orderHandler := handlers.NewOrderHandler(orderRepo)
	mockHandler := mock.NewMockTaobaoHandler()

	// 创建路由
	r := gin.Default()

	// 全局中间件
	r.Use(middleware.CORS())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "OURERP Server Running",
			"env":     cfg.Env,
		})
	})

	// API路由组
	api := r.Group("/api/v1")
	{
		// 公开路由
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)

		// 模拟淘宝API（开发测试用）
		mockHandler.RegisterRoutes(api)

		// 需要认证的路由
		authorized := api.Group("/")
		authorized.Use(middleware.JWTAuth())
		{
			authorized.GET("/auth/me", authHandler.GetCurrentUser)
			
			// 订单
			authorized.GET("/orders", orderHandler.ListOrders)
			authorized.POST("/orders", orderHandler.CreateOrder)
			authorized.GET("/orders/:id", orderHandler.GetOrder)
			authorized.POST("/orders/:id/audit", orderHandler.AuditOrder)
			authorized.POST("/orders/:id/ship", orderHandler.ShipOrder)
		}
	}

	// 启动
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	r.Run(addr)
}