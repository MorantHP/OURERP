package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/MorantHP/OURERP/internal/config"
	"github.com/MorantHP/OURERP/internal/handlers"
	"github.com/MorantHP/OURERP/internal/kafka"
	"github.com/MorantHP/OURERP/internal/middleware"
	"github.com/MorantHP/OURERP/internal/mock"
	"github.com/MorantHP/OURERP/internal/models"
	"github.com/MorantHP/OURERP/internal/repository"
	"github.com/MorantHP/OURERP/internal/seed"
	"github.com/MorantHP/OURERP/internal/services"
	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 验证配置
	if err := cfg.Validate(); err != nil {
		fmt.Printf("配置验证失败: %v\n", err)
		os.Exit(1)
	}

	// 设置运行模式
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化数据库
	db, err := repository.InitDB(&cfg.Database)
	if err != nil {
		fmt.Printf("数据库连接失败: %v\n", err)
		os.Exit(1)
	}

	// 自动迁移
	db.AutoMigrate(
		&models.User{},
		&models.Tenant{},
		&models.TenantUser{},
		&models.Order{},
		&models.OrderItem{},
		&models.Shop{},
		// 库存管理模型
		&models.Product{},
		&models.Warehouse{},
		&models.Inventory{},
		&models.InventoryLog{},
		&models.InboundOrder{},
		&models.InboundItem{},
		&models.OutboundOrder{},
		&models.OutboundItem{},
		&models.Stocktake{},
		&models.StocktakeItem{},
		&models.TransferOrder{},
		&models.TransferItem{},
		// 权限管理模型
		&models.Role{},
		&models.Permission{},
		&models.UserRole{},
		&models.UserResourcePermission{},
		// 财务管理模型
		&models.FinanceRecord{},
		&models.PlatformBill{},
		&models.PlatformBillDetail{},
		&models.Supplier{},
		&models.PurchaseSettlement{},
		&models.PurchaseSettlementDetail{},
		&models.PurchasePayment{},
		&models.ProductCost{},
		&models.OrderCost{},
		&models.InventoryCostSnapshot{},
		&models.FinancialSettlement{},
		&models.FinanceBankAccount{},
		// 数据中心模型
		&models.AlertRule{},
		&models.AlertRecord{},
		&models.ReportTemplate{},
		&models.Customer{},
		&models.RealtimeSnapshot{},
		&models.ProductAnalysis{},
		&models.CustomerAnalysis{},
		&models.RegionAnalysis{},
		&models.CompareAnalysis{},
		&models.DashboardWidget{},
	)

	// 创建仓库
	userRepo := repository.NewUserRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	shopRepo := repository.NewShopRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	tenantUserRepo := repository.NewTenantUserRepository(db)
	// 库存管理仓库
	productRepo := repository.NewProductRepository(db)
	warehouseRepo := repository.NewWarehouseRepository(db)
	inventoryRepo := repository.NewInventoryRepository(db)

	// 权限管理仓库
	permissionRepo := repository.NewPermissionRepository(db)

	// 财务管理仓库
	financeRepo := repository.NewFinanceRepository(db)

	// 数据中心仓库
	datacenterRepo := repository.NewDatacenterRepository(db)

	// 创建中间件
	tenantMiddleware := middleware.NewTenantMiddleware(tenantRepo, tenantUserRepo)

	// 创建同步服务
	syncService := services.NewSyncService(orderRepo, shopRepo)

	// 创建处理器
	authHandler := handlers.NewAuthHandler(userRepo, cfg)
	orderHandler := handlers.NewOrderHandler(orderRepo)
	shopHandler := handlers.NewShopHandler(shopRepo, syncService)
	platformHandler := handlers.NewPlatformHandler()
	tenantHandler := handlers.NewTenantHandler(tenantRepo, tenantUserRepo)
	// 库存管理处理器
	productHandler := handlers.NewProductHandler(productRepo, inventoryRepo)
	warehouseHandler := handlers.NewWarehouseHandler(warehouseRepo)
	inventoryHandler := handlers.NewInventoryHandler(inventoryRepo, productRepo, warehouseRepo)

	// 统计服务
	statsService := services.NewStatisticsService(db)
	statisticsHandler := handlers.NewStatisticsHandler(statsService)

	// 权限服务
	permissionService := services.NewPermissionService(permissionRepo)
	permissionHandler := handlers.NewPermissionHandler(permissionService, permissionRepo, shopRepo, warehouseRepo)

	// 财务服务
	financeService := services.NewFinanceService(db, financeRepo)
	financeHandler := handlers.NewFinanceHandler(financeService)

	// 数据中心服务
	realtimeService := services.NewRealtimeService(datacenterRepo)
	customerAnalysisService := services.NewCustomerAnalysisService(datacenterRepo)
	productAnalysisService := services.NewProductAnalysisService(datacenterRepo)
	compareAnalysisService := services.NewCompareAnalysisService(datacenterRepo)
	alertService := services.NewAlertService(datacenterRepo, productRepo, orderRepo)
	datacenterHandler := handlers.NewDatacenterHandler(
		realtimeService,
		customerAnalysisService,
		productAnalysisService,
		compareAnalysisService,
		alertService,
	)

	// 初始化权限数据（权限和角色）
	permissionService.SeedData()

	// 初始化默认管理员用户
	seedDefaultUser(userRepo)

	// 生成演示数据
	seeder := seed.NewSeeder(db)
	if err := seeder.SeedAll(); err != nil {
		fmt.Println("生成演示数据失败:", err)
	}

	// 设置全局权限服务
	permMiddleware := middleware.NewPermissionMiddleware(permissionService)
	middleware.SetGlobalPermissionService(permissionService)

	mockHandler := mock.NewMockTaobaoHandler()

	// OAuth回调URL
	baseURL := fmt.Sprintf("http://%s:%s", cfg.Server.Host, cfg.Server.Port)
	if cfg.Env == "production" {
		baseURL = "https://your-domain.com" // 生产环境需要配置
	}
	oauthHandler := handlers.NewOAuthHandler(shopRepo, baseURL)

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
		// 公开路由 - 带速率限制
		auth := api.Group("/auth")
		auth.Use(middleware.LoginRateLimit()) // 登录速率限制：每IP每分钟最多5次
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// 平台信息（公开）
		api.GET("/platforms", platformHandler.List)
		api.GET("/platforms/:code", platformHandler.Get)

		// OAuth回调（公开）
		oauth := api.Group("/oauth")
		{
			oauth.GET("/auth-url", oauthHandler.GetAuthURL)
			oauth.GET("/callback", oauthHandler.Callback)
			oauth.POST("/refresh", oauthHandler.RefreshToken)
		}

		// 模拟淘宝API（开发测试用）
		mockHandler.RegisterRoutes(api)

		// 需要认证的路由
		authorized := api.Group("/")
		authorized.Use(middleware.JWTAuth())
		{
			authorized.GET("/auth/me", authHandler.GetCurrentUser)

			// 用户管理（仅 root）
			users := authorized.Group("/users")
			{
				users.GET("", authHandler.ListUsers)
				users.PUT("/:id/approve", authHandler.ApproveUser)
				users.PUT("/:id/status", authHandler.SetUserStatus)
				users.DELETE("/:id", authHandler.DeleteUser)
			}

			// 租户管理（账套）
			tenants := authorized.Group("/tenants")
			{
				tenants.GET("", tenantHandler.List)
				tenants.GET("/my", tenantHandler.MyTenants)
				tenants.POST("", tenantHandler.Create)
				tenants.GET("/:id", tenantHandler.Get)
				tenants.PUT("/:id", tenantHandler.Update)
				tenants.DELETE("/:id", tenantHandler.Delete)
				tenants.POST("/:id/users", tenantHandler.AddUser)
				tenants.DELETE("/:id/users/:user_id", tenantHandler.RemoveUser)
				tenants.PUT("/:id/users/:user_id/role", tenantHandler.UpdateUserRole)
				tenants.POST("/switch", tenantHandler.SwitchTenant)
			}

			// 需要租户上下文的路由
			tenantRequired := authorized.Group("/")
			tenantRequired.Use(tenantMiddleware.Handle())
			{
				// 订单
				tenantRequired.GET("/orders", orderHandler.ListOrders)
				tenantRequired.POST("/orders", orderHandler.CreateOrder)
				tenantRequired.GET("/orders/:id", orderHandler.GetOrder)
				tenantRequired.POST("/orders/:id/audit", orderHandler.AuditOrder)
				tenantRequired.POST("/orders/:id/ship", orderHandler.ShipOrder)

				// 店铺管理
				tenantRequired.GET("/shops", shopHandler.List)
				tenantRequired.POST("/shops", shopHandler.Create)
				tenantRequired.GET("/shops/:id", shopHandler.Get)
				tenantRequired.PUT("/shops/:id", shopHandler.Update)
				tenantRequired.DELETE("/shops/:id", shopHandler.Delete)
				tenantRequired.POST("/shops/:id/sync", shopHandler.TriggerSync)
				tenantRequired.GET("/shops/:id/auth-url", shopHandler.GetAuthURL)

				// 商品管理
				products := tenantRequired.Group("/products")
				{
					products.GET("", productHandler.List)
					products.GET("/categories", productHandler.GetCategories)
					products.GET("/brands", productHandler.GetBrands)
					products.GET("/:id", productHandler.Get)
					products.POST("", productHandler.Create)
					products.PUT("/:id", productHandler.Update)
					products.DELETE("/:id", productHandler.Delete)
				}

				// 仓库管理
				warehouses := tenantRequired.Group("/warehouses")
				{
					warehouses.GET("", warehouseHandler.List)
					warehouses.GET("/:id", warehouseHandler.Get)
					warehouses.POST("", warehouseHandler.Create)
					warehouses.PUT("/:id", warehouseHandler.Update)
					warehouses.DELETE("/:id", warehouseHandler.Delete)
					warehouses.POST("/:id/default", warehouseHandler.SetDefault)
				}

				// 库存管理
				inventory := tenantRequired.Group("/inventory")
				{
					inventory.GET("", inventoryHandler.List)
					inventory.GET("/alert", inventoryHandler.Alert)
					inventory.GET("/logs", inventoryHandler.Logs)
					inventory.GET("/:id", inventoryHandler.Get)
					inventory.PUT("/:id", inventoryHandler.Update)
					inventory.POST("/adjust", inventoryHandler.Adjust)
				}

				// 数据统计
				statistics := tenantRequired.Group("/statistics")
				{
					statistics.GET("/overview", statisticsHandler.Overview)
					statistics.GET("/sales-trend", statisticsHandler.SalesTrend)
					statistics.GET("/by-platform", statisticsHandler.ByPlatform)
					statistics.GET("/by-shop", statisticsHandler.ByShop)
					statistics.GET("/by-category", statisticsHandler.ByCategory)
					statistics.GET("/by-brand", statisticsHandler.ByBrand)
					statistics.GET("/order-funnel", statisticsHandler.OrderFunnel)
					statistics.GET("/top-products", statisticsHandler.TopProducts)
				}

				// 权限管理
				permissions := tenantRequired.Group("/permissions")
				{
					// 角色管理
					permissions.GET("/roles", permissionHandler.ListRoles)
					permissions.GET("/roles/:id", permissionHandler.GetRole)
					permissions.POST("/roles", permMiddleware.RequirePermission(models.PermRoleWrite), permissionHandler.CreateRole)
					permissions.PUT("/roles/:id", permMiddleware.RequirePermission(models.PermRoleWrite), permissionHandler.UpdateRole)
					permissions.DELETE("/roles/:id", permMiddleware.RequirePermission(models.PermRoleWrite), permissionHandler.DeleteRole)

					// 权限查询
					permissions.GET("", permissionHandler.ListPermissions)
					permissions.GET("/my", permissionHandler.GetMyPermissions)
					permissions.GET("/users/:id", permMiddleware.RequirePermissionAssign(), permissionHandler.GetUserPermissions)

					// 用户授权
					permissions.PUT("/users/:id/role", permMiddleware.RequirePermissionAssign(), permissionHandler.SetUserRole)
					permissions.PUT("/users/:id/resources", permMiddleware.RequirePermissionAssign(), permissionHandler.SetResourcePermissions)
					permissions.POST("/users/:id/resources", permMiddleware.RequirePermissionAssign(), permissionHandler.AddResourcePermission)
					permissions.DELETE("/users/:id/resources/:rid", permMiddleware.RequirePermissionAssign(), permissionHandler.RemoveResourcePermission)

					// 可授权资源
					permissions.GET("/resources/shops", permMiddleware.RequirePermissionAssign(), permissionHandler.ListShops)
					permissions.GET("/resources/warehouses", permMiddleware.RequirePermissionAssign(), permissionHandler.ListWarehouses)
				}

				// 财务管理
				finance := tenantRequired.Group("/finance")
				{
					// 收支记录
					finance.GET("/records", financeHandler.ListFinanceRecords)
					finance.POST("/records", financeHandler.CreateFinanceRecord)
					finance.GET("/records/:id", financeHandler.GetFinanceRecord)
					finance.PUT("/records/:id", financeHandler.UpdateFinanceRecord)
					finance.DELETE("/records/:id", financeHandler.DeleteFinanceRecord)
					finance.POST("/records/:id/approve", financeHandler.ApproveFinanceRecord)

					// 平台账单
					finance.GET("/bills", financeHandler.ListPlatformBills)
					finance.POST("/bills", financeHandler.CreatePlatformBill)
					finance.GET("/bills/:id", financeHandler.GetPlatformBill)
					finance.GET("/bills/:id/details", financeHandler.GetBillDetails)
					finance.POST("/bills/:id/details/:detail_id/reconcile", financeHandler.ReconcileBillDetail)

					// 供应商
					finance.GET("/suppliers", financeHandler.ListSuppliers)
					finance.POST("/suppliers", financeHandler.CreateSupplier)
					finance.GET("/suppliers/:id", financeHandler.GetSupplier)
					finance.PUT("/suppliers/:id", financeHandler.UpdateSupplier)
					finance.DELETE("/suppliers/:id", financeHandler.DeleteSupplier)

					// 采购结算
					finance.GET("/settlements", financeHandler.ListPurchaseSettlements)
					finance.POST("/settlements", financeHandler.CreatePurchaseSettlement)
					finance.GET("/settlements/:id", financeHandler.GetPurchaseSettlement)
					finance.POST("/settlements/:id/pay", financeHandler.PaySettlement)
					finance.GET("/settlements/:id/payments", financeHandler.GetSettlementPayments)

					// 商品成本
					finance.GET("/product-costs", financeHandler.ListProductCosts)
					finance.PUT("/product-costs/:id", financeHandler.UpdateProductCost)
					finance.POST("/product-costs/batch", financeHandler.BatchUpdateProductCosts)

					// 订单成本
					finance.GET("/order-costs", financeHandler.ListOrderCosts)
					finance.POST("/order-costs/:id/calculate", financeHandler.CalculateOrderCost)
					finance.GET("/order-costs/profit", financeHandler.GetProfitAnalysis)

					// 库存成本快照
					finance.GET("/inventory-snapshots", financeHandler.ListInventorySnapshots)
					finance.POST("/inventory-snapshots/generate", financeHandler.GenerateInventorySnapshot)

					// 月度结算
					finance.GET("/monthly-settlements", financeHandler.ListMonthlySettlements)
					finance.POST("/monthly-settlements/generate", financeHandler.GenerateMonthlySettlement)
					finance.POST("/monthly-settlements/:period/confirm", financeHandler.ConfirmMonthlySettlement)

					// 年度结算
					finance.GET("/yearly-settlements", financeHandler.ListYearlySettlements)
					finance.POST("/yearly-settlements/generate", financeHandler.GenerateYearlySettlement)

					// 结算账户
					finance.GET("/bank-accounts", financeHandler.ListBankAccounts)
					finance.POST("/bank-accounts", financeHandler.CreateBankAccount)
					finance.PUT("/bank-accounts/:id", financeHandler.UpdateBankAccount)
					finance.DELETE("/bank-accounts/:id", financeHandler.DeleteBankAccount)
				}

				// 数据中心
				datacenter := tenantRequired.Group("/datacenter")
				{
					// 实时监控
					datacenter.GET("/realtime/overview", datacenterHandler.GetRealtimeOverview)
					datacenter.GET("/realtime/inventory", datacenterHandler.GetRealtimeInventory)
					datacenter.GET("/realtime/hourly-trend", datacenterHandler.GetHourlyTrend)

					// 客户分析
					datacenter.GET("/customers/analysis", datacenterHandler.GetCustomerAnalysis)
					datacenter.GET("/customers/value-distribution", datacenterHandler.GetCustomerValueDistribution)
					datacenter.GET("/customers/geography", datacenterHandler.GetGeographyDistribution)
					datacenter.GET("/customers/city", datacenterHandler.GetCityDistribution)
					datacenter.GET("/customers/repurchase", datacenterHandler.GetRepurchaseAnalysis)

					// 商品分析
					datacenter.GET("/products/turnover", datacenterHandler.GetProductTurnover)
					datacenter.GET("/products/inventory-level", datacenterHandler.GetInventoryLevel)
					datacenter.GET("/products/purchase-strategy", datacenterHandler.GetPurchaseStrategy)
					datacenter.GET("/products/low-stock", datacenterHandler.GetLowStockProducts)
					datacenter.GET("/products/inventory-summary", datacenterHandler.GetInventorySummary)

					// 对比分析
					datacenter.GET("/compare/period", datacenterHandler.GetPeriodCompare)
					datacenter.GET("/compare/yoy", datacenterHandler.GetYOYCompare)
					datacenter.GET("/compare/mom", datacenterHandler.GetMOMCompare)
					datacenter.GET("/compare/shop", datacenterHandler.GetShopCompare)
					datacenter.GET("/compare/platform", datacenterHandler.GetPlatformCompare)

					// 预警管理
					datacenter.GET("/alerts/rules", datacenterHandler.ListAlertRules)
					datacenter.POST("/alerts/rules", datacenterHandler.CreateAlertRule)
					datacenter.PUT("/alerts/rules/:id", datacenterHandler.UpdateAlertRule)
					datacenter.DELETE("/alerts/rules/:id", datacenterHandler.DeleteAlertRule)
					datacenter.POST("/alerts/rules/:id/toggle", datacenterHandler.ToggleAlertRule)
					datacenter.GET("/alerts/summary", datacenterHandler.GetAlertSummary)
					datacenter.GET("/alerts/records", datacenterHandler.ListAlertRecords)
					datacenter.POST("/alerts/records/:id/handle", datacenterHandler.HandleAlertRecord)
					datacenter.POST("/alerts/records/:id/ignore", datacenterHandler.IgnoreAlertRecord)
					datacenter.POST("/alerts/check", datacenterHandler.CheckAlerts)
					datacenter.GET("/alerts/types", datacenterHandler.GetAlertTypes)
					datacenter.GET("/alerts/levels", datacenterHandler.GetNotifyLevels)
				}
			}
		}
	}

	// 启动Kafka消费者（如果配置了Kafka）
	kafkaConfig := kafka.LoadConfig()
	if len(kafkaConfig.Brokers) > 0 && kafkaConfig.Brokers[0] != "" {
		go startKafkaConsumer(kafkaConfig, orderRepo, productRepo, shopRepo)
	}

	// 启动HTTP服务
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	fmt.Printf("🚀 服务器启动: http://%s\n", addr)
	r.Run(addr)
}

// startKafkaConsumer 启动Kafka订单消费者
func startKafkaConsumer(kafkaConfig *kafka.Config, orderRepo *repository.OrderRepository, productRepo *repository.ProductRepository, shopRepo *repository.ShopRepository) {
	// 创建生产者（用于死信队列）
	producer, err := kafka.NewProducer(kafkaConfig)
	if err != nil {
		fmt.Printf("❌ Kafka生产者创建失败: %v\n", err)
		return
	}

	// 创建ERP订单处理器
	handler := kafka.NewERPOrderHandler(orderRepo, productRepo, shopRepo)

	// 创建消费者
	consumer, err := kafka.NewConsumer(kafkaConfig, handler, producer)
	if err != nil {
		fmt.Printf("❌ Kafka消费者创建失败: %v\n", err)
		producer.Close()
		return
	}

	// 创建上下文和信号处理
	ctx, cancel := context.WithCancel(context.Background())

	// 处理信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		fmt.Println("\n🛑 收到停止信号，正在关闭Kafka消费者...")
		cancel()
		consumer.Stop()
		producer.Close()
	}()

	// 启动消费者
	fmt.Printf("📥 Kafka消费者已启动，订阅Topics: %v\n", kafkaConfig.GetAllTopics())
	if err := consumer.Start(ctx, kafkaConfig.GetAllTopics()); err != nil {
		fmt.Printf("Kafka消费者错误: %v\n", err)
	}
}

// seedDefaultUser 创建 root 超级管理员用户
func seedDefaultUser(userRepo *repository.UserRepository) {
	// 检查是否已存在 root 用户
	existingUser, _ := userRepo.FindByEmail("root@ourerp.com")
	if existingUser != nil {
		return // 已存在，跳过
	}

	// 从环境变量获取 root 密码，如果未设置则生成随机密码
	rootPassword := os.Getenv("ROOT_PASSWORD")
	if rootPassword == "" {
		// 生产环境必须设置 ROOT_PASSWORD
		if config.GlobalConfig.Env == "production" {
			fmt.Println("错误: 生产环境必须设置 ROOT_PASSWORD 环境变量")
			return
		}
		// 开发环境使用默认密码
		rootPassword = "root123456"
		fmt.Println("警告: 使用默认 root 密码，请在生产环境设置 ROOT_PASSWORD 环境变量")
	}

	// 创建 root 超级管理员
	root := &models.User{
		Email:      "root@ourerp.com",
		Name:       "Root",
		IsRoot:     true,
		IsApproved: true, // root 默认已审核
		Status:     1,
	}
	if err := root.SetPassword(rootPassword); err != nil {
		fmt.Println("Failed to set root password:", err)
		return
	}

	if err := userRepo.Create(root); err != nil {
		fmt.Println("Failed to create root user:", err)
		return
	}

	fmt.Println("Root user created: root@ourerp.com")
	fmt.Println("请登录后立即修改默认密码!")
}
