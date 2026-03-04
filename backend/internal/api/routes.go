package api

import (
	"github.com/MorantHP/OURERP/internal/cache"
	"github.com/MorantHP/OURERP/internal/handlers"
	"github.com/MorantHP/OURERP/internal/middleware"
	"github.com/MorantHP/OURERP/internal/repository"
	"github.com/MorantHP/OURERP/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Router 路由器
type Router struct {
	db           *gorm.DB
	cache        cache.CacheService
	authMiddleware *middleware.AuthMiddleware
}

// NewRouter 创建路由器
func NewRouter(db *gorm.DB, cache cache.CacheService) *Router {
	return &Router{
		db:    db,
		cache: cache,
		authMiddleware: middleware.NewAuthMiddleware(cache),
	}
}

// Setup 设置路由
func (r *Router) Setup(engine *gin.Engine) {
	// 全局中间件
	engine.Use(middleware.Recovery())
	engine.Use(middleware.ErrorHandler())
	engine.Use(middleware.CORS())
	engine.Use(middleware.DefaultRateLimit())
	engine.Use(middleware.DefaultAuditMiddleware())

	// 健康检查
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1
	v1 := engine.Group("/api/v1")
	{
		// 认证路由（不需要认证）
		auth := v1.Group("/auth")
		{
			authHandler := handlers.NewAuthHandler(r.db)
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// 需要认证的路由
		protected := v1.Group("")
		protected.Use(middleware.JWTAuth())
		protected.Use(middleware.TenantMiddleware())
		{
			// 用户路由
			userHandler := handlers.NewUserHandler(r.db)
			protected.GET("/users/me", userHandler.GetCurrentUser)
			protected.PUT("/users/me", userHandler.UpdateCurrentUser)

			// 商品路由
			productRepo := repository.NewProductRepository(r.db)
			inventoryRepo := repository.NewInventoryRepository(r.db)
			productService := services.NewProductService(productRepo, inventoryRepo, r.cache)
			productHandler := handlers.NewProductHandlerV2(productService)
			
			products := protected.Group("/products")
			{
				products.GET("", productHandler.List)
				products.POST("", productHandler.Create)
				products.GET("/categories", productHandler.GetCategories)
				products.GET("/brands", productHandler.GetBrands)
				products.GET("/:id", productHandler.Get)
				products.PUT("/:id", productHandler.Update)
				products.DELETE("/:id", productHandler.Delete)
			}

			// 库存路由
			inventoryService := services.NewInventoryService(inventoryRepo, productRepo, nil, r.cache)
			inventoryHandler := handlers.NewInventoryHandlerV2(inventoryService)
			
			inventories := protected.Group("/inventories")
			{
				inventories.GET("", inventoryHandler.List)
				inventories.GET("/alerts", inventoryHandler.GetAlerts)
				inventories.POST("/adjust", inventoryHandler.Adjust)
				inventories.POST("/transfer", inventoryHandler.Transfer)
				inventories.GET("/:product_id/:warehouse_id", inventoryHandler.Get)
			}

			// 订单路由
			orderRepo := repository.NewOrderRepository(r.db)
			orderService := services.NewOrderService(orderRepo, inventoryRepo, r.cache)
			orderHandler := handlers.NewOrderHandlerV2(orderService)
			
			orders := protected.Group("/orders")
			{
				orders.GET("", orderHandler.List)
				orders.GET("/statistics", orderHandler.Statistics)
				orders.POST("", orderHandler.Create)
				orders.GET("/:order_no", orderHandler.Get)
				orders.POST("/:order_no/audit", orderHandler.Audit)
				orders.POST("/:order_no/ship", orderHandler.Ship)
				orders.POST("/:order_no/cancel", orderHandler.Cancel)
			}

			// 仓库路由
			warehouseRepo := repository.NewWarehouseRepository(r.db)
			warehouseHandler := handlers.NewWarehouseHandler(warehouseRepo)
			
			warehouses := protected.Group("/warehouses")
			{
				warehouses.GET("", warehouseHandler.List)
				warehouses.POST("", warehouseHandler.Create)
				warehouses.GET("/:id", warehouseHandler.Get)
				warehouses.PUT("/:id", warehouseHandler.Update)
				warehouses.DELETE("/:id", warehouseHandler.Delete)
			}

			// 租户路由
			tenantRepo := repository.NewTenantRepository(r.db)
			tenantHandler := handlers.NewTenantHandler(tenantRepo)

			tenants := protected.Group("/tenants")
			{
				tenants.GET("", tenantHandler.List)
				tenants.GET("/:id", tenantHandler.Get)
				tenants.POST("", tenantHandler.Create)
				tenants.PUT("/:id", tenantHandler.Update)
				tenants.DELETE("/:id", tenantHandler.Delete)
			}

			// API 同步路由（用于外部系统推送数据）
			shopRepo := repository.NewShopRepository(r.db)
			apiSyncService := services.NewApiSyncService(r.db, orderRepo, shopRepo, productRepo, r.cache)
			apiSyncHandler := handlers.NewApiSyncHandler(apiSyncService)

			sync := protected.Group("/sync")
			{
				sync.POST("/orders", apiSyncHandler.SyncOrders)
				sync.GET("/statistics", apiSyncHandler.GetSyncStatistics)
			}
		}

		// 管理员路由（需要管理员权限）
		admin := v1.Group("/admin")
		admin.Use(middleware.JWTAuth())
		{
			// 系统配置
			// 用户管理
			// 权限管理
		}
	}

	// API v2 (预留)
	v2 := engine.Group("/api/v2")
	{
		v2.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"version": "v2", "status": "ok"})
		})
	}
}

// NewAuthHandler 创建认证处理器（临时实现）
func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

// AuthHandler 认证处理器
type AuthHandler struct {
	db *gorm.DB
}

func (h *AuthHandler) Login(c *gin.Context) {
	c.JSON(200, gin.H{"message": "login"})
}

func (h *AuthHandler) Register(c *gin.Context) {
	c.JSON(200, gin.H{"message": "register"})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	c.JSON(200, gin.H{"message": "refresh"})
}

// NewUserHandler 创建用户处理器（临时实现）
func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

// UserHandler 用户处理器
type UserHandler struct {
	db *gorm.DB
}

func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	c.JSON(200, gin.H{"user": nil})
}

func (h *UserHandler) UpdateCurrentUser(c *gin.Context) {
	c.JSON(200, gin.H{"message": "updated"})
}

// NewWarehouseHandler 创建仓库处理器（临时实现）
func NewWarehouseHandler(repo *repository.WarehouseRepository) *WarehouseHandler {
	return &WarehouseHandler{repo: repo}
}

// WarehouseHandler 仓库处理器
type WarehouseHandler struct {
	repo *repository.WarehouseRepository
}

func (h *WarehouseHandler) List(c *gin.Context) {
	c.JSON(200, gin.H{"warehouses": []interface{}{}})
}

func (h *WarehouseHandler) Create(c *gin.Context) {
	c.JSON(201, gin.H{"message": "created"})
}

func (h *WarehouseHandler) Get(c *gin.Context) {
	c.JSON(200, gin.H{"warehouse": nil})
}

func (h *WarehouseHandler) Update(c *gin.Context) {
	c.JSON(200, gin.H{"message": "updated"})
}

func (h *WarehouseHandler) Delete(c *gin.Context) {
	c.JSON(200, gin.H{"message": "deleted"})
}

// NewTenantHandler 创建租户处理器（临时实现）
func NewTenantHandler(repo *repository.TenantRepository) *TenantHandler {
	return &TenantHandler{repo: repo}
}

// TenantHandler 租户处理器
type TenantHandler struct {
	repo *repository.TenantRepository
}

func (h *TenantHandler) List(c *gin.Context) {
	c.JSON(200, gin.H{"tenants": []interface{}{}})
}

func (h *TenantHandler) Get(c *gin.Context) {
	c.JSON(200, gin.H{"tenant": nil})
}

func (h *TenantHandler) Create(c *gin.Context) {
	c.JSON(201, gin.H{"message": "created"})
}

func (h *TenantHandler) Update(c *gin.Context) {
	c.JSON(200, gin.H{"message": "updated"})
}

func (h *TenantHandler) Delete(c *gin.Context) {
	c.JSON(200, gin.H{"message": "deleted"})
}
