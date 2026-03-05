# OURERP 后端架构详解

> 作者：系统自动生成
> 版本：v1.0
> 更新时间：2026-03-05
> 代码总量：约 22,759 行 Go 代码

---

## 📚 目录

1. [项目概述](#项目概述)
2. [技术栈](#技术栈)
3. [目录结构](#目录结构)
4. [核心概念](#核心概念)
5. [请求处理流程](#请求处理流程)
6. [分层架构详解](#分层架构详解)
7. [每个文件功能说明](#每个文件功能说明)
8. [多租户架构](#多租户架构)
9. [认证授权](#认证授权)
10. [外部系统对接](#外部系统对接)
11. [学习路径](#学习路径)

---

## 项目概述

OURERP 是一个**多租户电商 ERP 系统**，支持：
- 多平台订单管理（淘宝、京东、抖音等）
- 库存管理
- 财务管理
- 数据分析
- 权限管理

### 核心特性

✅ **多租户架构**：一套系统服务多个企业/账套
✅ **平台对接**：支持淘宝、京东、抖音、快手等主流电商平台
✅ **API 对接**：提供 RESTful API 供外部系统推送数据
✅ **Kafka 集成**：支持消息队列（可选）
✅ **实时分析**：数据中心提供实时监控和分析

---

## 技术栈

| 组件 | 技术 | 说明 |
|-----|------|------|
| **语言** | Go 1.21+ | 高性能、类型安全 |
| **Web 框架** | Gin | 轻量级 HTTP 框架 |
| **ORM** | GORM | Go 最流行的 ORM |
| **数据库** | PostgreSQL 14+ | 关系型数据库 |
| **缓存** | Redis 7+ | 内存数据库 |
| **消息队列** | Kafka | 可选的消息队列 |
| **认证** | JWT | JSON Web Token |
| **API 文档** | Swagger | 自动生成 API 文档 |

---

## 目录结构

```
backend/
├── cmd/                    # 应用程序入口
│   ├── server/            # 主服务器
│   ├── mock/              # 模拟数据生成器
│   ├── kafka-test/        # Kafka 测试工具
│   └── fix/               # 数据修复工具
│
├── internal/              # 内部应用代码
│   ├── config/           # 配置管理
│   ├── models/           # 数据模型 (14个文件)
│   ├── repository/       # 数据访问层 (13个文件)
│   ├── services/         # 业务逻辑层 (20个文件)
│   ├── handlers/         # HTTP处理层 (20个文件)
│   ├── middleware/       # 中间件 (9个文件)
│   ├── kafka/            # 消息队列 (6个文件)
│   ├── platform/         # 平台适配器 (8个文件)
│   ├── cache/            # 缓存服务
│   ├── mock/             # 模拟数据
│   ├── seed/             # 数据种子
│   ├── pkg/              # 公共工具 (6个文件)
│   ├── api/              # 路由注册
│   └── tests/            # 测试
│
├── docs/                  # 文档
│   ├── swagger.go        # API 文档定义
│   └── swagger_models.go # API 模型
│
├── scripts/              # 脚本工具
│   └── migrate.go       # 数据库迁移
│
├── .env                  # 环境变量配置
├── .env.example          # 环境变量示例
├── Dockerfile           # 生产环境镜像
├── Dockerfile.dev       # 开发环境镜像
├── go.mod               # 依赖定义
└── go.sum               # 依赖锁定
```

---

## 核心概念

### 1. 分层架构

OURERP 采用经典的分层架构：

```
┌─────────────────────────────────────┐
│         HTTP 请求                    │
└──────────────┬──────────────────────┘
               ▼
┌─────────────────────────────────────┐
│         Middleware (中间件)          │
│  - CORS (跨域)                       │
│  - JWTAuth (认证)                    │
│  - TenantMiddleware (租户)           │
│  - PermissionMiddleware (权限)       │
└──────────────┬──────────────────────┘
               ▼
┌─────────────────────────────────────┐
│         Handler (处理层)              │
│  - 处理 HTTP 请求/响应                │
│  - 参数验证                          │
│  - 调用 Service                      │
└──────────────┬──────────────────────┘
               ▼
┌─────────────────────────────────────┐
│         Service (业务逻辑层)          │
│  - 实现业务逻辑                       │
│  - 协调多个 Repository                │
│  - 事务管理                          │
│  - 缓存管理                          │
└──────────────┬──────────────────────┘
               ▼
┌─────────────────────────────────────┐
│       Repository (数据访问层)         │
│  - 封装数据库操作                     │
│  - 提供 CRUD 接口                    │
│  - 自动租户过滤                       │
└──────────────┬──────────────────────┘
               ▼
┌─────────────────────────────────────┐
│         PostgreSQL 数据库             │
└─────────────────────────────────────┘
```

### 2. 依赖注入

所有组件通过构造函数注入依赖：

```go
// 示例：创建 OrderHandler
orderHandler := handlers.NewOrderHandler(orderRepo)

// 示例：创建 OrderService
orderService := services.NewOrderService(
    orderRepo,      // 注入订单仓库
    inventoryRepo,  // 注入库存仓库
    cacheService,   // 注入缓存服务
)
```

**优点**：
- 松耦合
- 易于测试
- 依赖关系清晰

### 3. Context 传递

使用 `context.Context` 传递租户 ID：

```go
// 1. 在中间件中设置
ctx = repository.SetTenantIDToContext(c.Request.Context(), tenantID)
c.Request = c.Request.WithContext(ctx)

// 2. 在 Service 中获取
tenantID := repository.GetTenantIDFromContext(ctx)

// 3. 在 Repository 中使用
db.WithContext(ctx).Scopes(WithTenantFromContext(ctx)).Find(&orders)
```

---

## 请求处理流程

以"查询订单列表"为例，完整流程如下：

### 第 1 步：HTTP 请求

```http
GET /api/v1/orders?page=1&size=20
Headers:
  Authorization: Bearer eyJhbGc...
  X-Tenant-ID: 1
```

### 第 2 步：路由匹配

**文件位置**：[cmd/server/main.go:261](cmd/server/main.go)

```go
tenantRequired.GET("/orders", orderHandler.ListOrders)
```

### 第 3 步：中间件链处理

#### 3.1 CORS 中间件
**文件位置**：[internal/middleware/auth.go:35](internal/middleware/auth.go)

```go
func CORS() gin.HandlerFunc {
    // 处理跨域请求
    // 验证 Origin 是否在允许列表中
    c.Header("Access-Control-Allow-Origin", allowOrigin)
    c.Next()
}
```

#### 3.2 JWT 认证中间件
**文件位置**：[internal/middleware/auth.go:83](internal/middleware/auth.go)

```go
func JWTAuth() gin.HandlerFunc {
    // 1. 提取 Token
    tokenString := strings.Split(authHeader, " ")[1]

    // 2. 解析和验证
    token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return []byte(config.GlobalConfig.JWT.Secret), nil
    })

    // 3. 提取 Claims 并存储到 Gin 上下文
    c.Set("user_id", int64(claims["user_id"].(float64)))
    c.Set("email", claims["email"].(string))

    c.Next()
}
```

#### 3.3 租户中间件
**文件位置**：[internal/middleware/tenant.go:11](internal/middleware/tenant.go)

```go
func TenantMiddleware() gin.HandlerFunc {
    // 1. 从 Header 获取租户 ID
    tenantID := GetTenantIDFromGin(c)  // 从 X-Tenant-ID 读取

    // 2. 设置到 Gin 上下文
    c.Set("tenant_id", tenantID)

    // 3. 设置到 Request 上下文
    ctx := repository.SetTenantIDToContext(c.Request.Context(), tenantID)
    c.Request = c.Request.WithContext(ctx)

    c.Next()
}
```

### 第 4 步：Handler 处理
**文件位置**：[internal/handlers/order.go:21](internal/handlers/order.go)

```go
func (h *OrderHandler) ListOrders(c *gin.Context) {
    // 1. 解析查询参数
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

    // 2. 调用 Repository 查询数据
    orders, total, err := h.orderRepo.ListWithContext(
        c.Request.Context(),  // 传递包含租户ID的上下文
        page, size,
        c.Query("status"),
        c.Query("platform"),
        c.Query("keyword"),
    )

    // 3. 返回 JSON 响应
    c.JSON(200, gin.H{
        "list": orders,
        "pagination": gin.H{
            "total":       total,
            "page":        page,
            "size":        size,
            "total_pages": (total + int64(size) - 1) / int64(size),
        },
    })
}
```

### 第 5 步：Repository 数据访问
**文件位置**：[internal/repository/order_repository.go:91](internal/repository/order_repository.go)

```go
func (r *OrderRepository) ListWithContext(ctx context.Context, page, size int, ...) {
    query := r.db.WithContext(ctx).
        Scopes(WithTenantFromContext(ctx)).  // ← 自动添加租户过滤
        Preload("Items")

    // 生成的 SQL 会自动包含:
    // SELECT * FROM orders WHERE tenant_id = 1 LIMIT 20
    query.Find(&orders)
}
```

### 第 6 步：租户隔离
**文件位置**：[internal/repository/tenant_scope.go:29](internal/repository/tenant_scope.go)

```go
func WithTenantFromContext(ctx context.Context) func(db *gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        tenantID := GetTenantIDFromContext(ctx)  // 从上下文提取租户ID
        if tenantID > 0 {
            return db.Where("tenant_id = ?", tenantID)  // 注入过滤条件
        }
        return db
    }
}
```

---

## 分层架构详解

### Models 层 (数据模型层)

**位置**：[internal/models/](internal/models/)

**职责**：
- 定义数据库表结构
- 定义业务实体
- 数据验证规则
- 常量定义

**示例**：订单模型
**文件**：[internal/models/order.go:18](internal/models/order.go)

```go
type Order struct {
    ID               int64          `json:"id" gorm:"primaryKey"`
    TenantID         int64          `json:"tenant_id" gorm:"index;not null"`
    OrderNo          string         `json:"order_no" gorm:"uniqueIndex:idx_order_tenant,tenant_id"`
    Platform         string         `json:"platform"`           // 平台: taobao/jd/douyin
    PlatformOrderID  string         `json:"platform_order_id"`  // 平台订单号
    Status           int            `json:"status"`             // 订单状态
    Items            []OrderItem    `json:"items" gorm:"foreignKey:OrderID"`
    // ... 更多字段
}

// 订单状态常量
const (
    OrderStatusPendingPayment = iota + 1  // 待付款
    OrderStatusPendingShip                // 待发货
    OrderStatusShipped                    // 已发货
    OrderStatusCompleted                  // 已完成
    OrderStatusCancelled                  // 已取消
)
```

**关键点**：
- 所有业务表都有 `TenantID` 字段
- GORM 标签定义数据库约束
- JSON 标签控制 API 响应字段

---

### Repository 层 (数据访问层)

**位置**：[internal/repository/](internal/repository/)

**职责**：
- 封装数据库操作
- 提供 CRUD 接口
- 自动租户过滤
- 事务管理

**示例**：订单仓库
**文件**：[internal/repository/order_repository.go](internal/repository/order_repository.go)

```go
type OrderRepository struct {
    db *gorm.DB
}

// ✅ 推荐: 自动租户隔离
func (r *OrderRepository) CreateWithContext(ctx context.Context, order *models.Order) error {
    return r.db.WithContext(ctx).
        Scopes(WithTenantFromContext(ctx)).  // 自动添加 WHERE tenant_id = ?
        Create(order).Error
}

// ✅ 推荐: 带租户的查询
func (r *OrderRepository) ListWithContext(ctx context.Context, page, size int, ...) {
    query := r.db.WithContext(ctx).
        Scopes(WithTenantFromContext(ctx)).  // 租户过滤
        Preload("Items")

    query.Offset((page - 1) * size).Limit(size).Find(&orders)
}

// ❌ 已弃用: 不安全,可能跨租户操作
func (r *OrderRepository) Create(order *models.Order) error {
    return r.db.Create(order).Error
}
```

**关键设计**：
- 所有方法都应该使用 `*WithContext` 版本
- 通过 `WithTenantFromContext` Scope 自动注入租户条件
- 支持事务操作

---

### Service 层 (业务逻辑层)

**位置**：[internal/services/](internal/services/)

**职责**：
- 实现复杂业务逻辑
- 协调多个 Repository
- 事务管理
- 缓存管理

**示例**：订单服务
**文件**：[internal/services/order_service.go:33](internal/services/order_service.go)

```go
type OrderService struct {
    orderRepo      *repository.OrderRepository
    inventoryRepo  *repository.InventoryRepository
    cacheDecorator *CacheDecorator
}

func (s *OrderService) CreateOrder(ctx context.Context, req *models.CreateOrderRequest) (*models.Order, error) {
    // 1. 从上下文获取租户ID
    tenantID := repository.GetTenantIDFromContext(ctx)

    // 2. 构建订单实体
    order := &models.Order{
        TenantID:    tenantID,
        OrderNo:     models.GenerateOrderNo(),
        Status:      models.OrderStatusPendingPayment,
        // ... 填充其他字段
    }

    // 3. 调用 Repository 保存
    if err := s.orderRepo.CreateWithContext(ctx, order); err != nil {
        return nil, errors.WrapInternal(err, "创建订单失败")
    }

    // 4. 使缓存失效
    _ = s.cacheDecorator.InvalidateOrderCache(ctx, tenantID)

    return order, nil
}
```

**关键功能**：
- 业务逻辑封装
- 跨 Repository 协调
- 错误处理和转换
- 缓存管理

---

### Handler 层 (HTTP 处理层)

**位置**：[internal/handlers/](internal/handlers/)

**职责**：
- 处理 HTTP 请求/响应
- 参数验证
- 调用 Service
- 返回 JSON 响应

**示例**：订单处理器
**文件**：[internal/handlers/order.go:21](internal/handlers/order.go)

```go
type OrderHandler struct {
    orderRepo *repository.OrderRepository
}

func (h *OrderHandler) ListOrders(c *gin.Context) {
    // 1. 解析查询参数
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

    // 限制最大分页大小
    if size > 100 {
        size = 100
    }

    // 2. 调用 Repository
    orders, total, err := h.orderRepo.ListWithContext(
        c.Request.Context(),
        page, size,
        c.Query("status"),
        c.Query("platform"),
        c.Query("keyword"),
    )

    // 3. 返回响应
    c.JSON(http.StatusOK, gin.H{
        "list": orders,
        "pagination": gin.H{
            "total":       total,
            "page":        page,
            "size":        size,
            "total_pages": (total + int64(size) - 1) / int64(size),
        },
    })
}
```

---

## 每个文件功能说明

### 📁 cmd/ 目录

#### [cmd/server/main.go](cmd/server/main.go) (561 行)
**主服务器入口**

**功能**：
1. 加载和验证配置
2. 初始化数据库连接
3. 自动迁移数据库表 (94个模型)
4. 创建所有 Repository (15个)
5. 创建所有 Service (15个)
6. 创建所有 Handler (15个)
7. 配置 Gin 路由和中间件
8. 启动 HTTP 服务器
9. 启动 Kafka 消费者 (可选)

**关键代码**：
```go
func main() {
    // 1. 加载配置
    cfg := config.Load()

    // 2. 初始化数据库
    db, _ := repository.InitDB(&cfg.Database)

    // 3. 创建仓库
    orderRepo := repository.NewOrderRepository(db)
    shopRepo := repository.NewShopRepository(db)
    // ... 其他仓库

    // 4. 创建服务
    orderService := services.NewOrderService(orderRepo, inventoryRepo, cacheService)

    // 5. 创建处理器
    orderHandler := handlers.NewOrderHandler(orderRepo)

    // 6. 配置路由
    r := gin.Default()
    tenantRequired.GET("/orders", orderHandler.ListOrders)

    // 7. 启动服务
    r.Run(":8080")
}
```

---

### 📁 internal/models/ 目录

#### [user.go](internal/models/user.go)
**用户模型**

**功能**：
- `User` 用户实体 (Email, Password, Name, Status)
- `UserStatus` 状态常量
- `SetPassword()` 密码加密 (bcrypt)
- `CheckPassword()` 密码验证

#### [tenant.go](internal/models/tenant.go)
**租户模型**

**功能**：
- `Tenant` 租户/账套实体
- `TenantUser` 租户用户关联
- 租户类型枚举

#### [order.go](internal/models/order.go)
**订单模型**

**功能**：
- `Order` 订单实体
- `OrderItem` 订单明细
- 订单状态常量 (5种状态)
- `CreateOrderRequest` DTO
- `GenerateOrderNo()` 订单号生成器

#### [shop.go](internal/models/shop.go)
**店铺模型**

#### [product.go](internal/models/product.go)
**商品模型**

#### [warehouse.go](internal/models/warehouse.go)
**仓库模型**

#### [inventory.go](internal/models/inventory.go)
**库存模型**

**功能**：
- `Inventory` 库存实体
- `InventoryLog` 库存变动日志
- `InboundOrder` 入库单
- `OutboundOrder` 出库单
- `Stocktake` 盘点单
- `TransferOrder` 调拨单

#### [rbac.go](internal/models/rbac.go)
**权限模型**

**功能**：
- `Role` 角色
- `Permission` 权限
- `UserRole` 用户角色关联
- `UserResourcePermission` 资源权限
- 权限常量定义

#### [finance.go](internal/models/finance.go)
**财务模型**

**功能**：
- 11个财务相关模型
- 收支记录、平台账单
- 供应商、采购结算
- 成本计算、财务结算

#### [datacenter.go](internal/models/datacenter.go)
**数据中心模型**

**功能**：
- 10个数据中心模型
- 预警规则、报表模板
- 客户分析、商品分析
- 对比分析、仪表盘

---

### 📁 internal/repository/ 目录

#### [database.go](internal/repository/database.go)
**数据库初始化**

**功能**：
- `InitDB()` 初始化数据库连接
- 配置 GORM (日志、连接池)
- 健康检查

#### [tenant_scope.go](internal/repository/tenant_scope.go)
**租户隔离核心**

**功能**：
- `WithTenant(tenantID)` 添加租户过滤
- `WithTenantFromContext(ctx)` 从上下文获取租户ID
- `GetTenantIDFromContext(ctx)` 提取租户ID
- `SetTenantIDToContext(ctx, tenantID)` 设置租户ID

**示例**：
```go
// 使用方式
db.Scopes(WithTenantFromContext(ctx)).Find(&orders)
// 生成 SQL: SELECT * FROM orders WHERE tenant_id = 1
```

#### [user_repository.go](internal/repository/user_repository.go)
**用户数据访问**

**功能**：
- `Create()` 创建用户
- `FindByEmail()` 根据邮箱查找
- `List()` 用户列表
- `Update()` 更新用户

#### [order_repository.go](internal/repository/order_repository.go)
**订单数据访问**

**功能**：
- `CreateWithContext()` 创建订单
- `ListWithContext()` 分页查询
- `UpdateStatusWithContext()` 更新状态
- `UpsertWithContext()` 插入或更新
- `FindByPlatformOrderIDWithContext()` 平台订单号查询

#### [shop_repository.go](internal/repository/shop_repository.go)
**店铺数据访问**

#### [product_repository.go](internal/repository/product_repository.go)
**商品数据访问**

#### [warehouse_repository.go](internal/repository/warehouse_repository.go)
**仓库数据访问**

#### [inventory_repository.go](internal/repository/inventory_repository.go)
**库存数据访问**

**功能**：
- `GetStock()` 查询库存
- `Adjust()` 调整库存
- `GetLogs()` 查询日志
- 入库/出库/盘点/调拨

#### [permission_repository.go](internal/repository/permission_repository.go)
**权限数据访问**

#### [finance_repository.go](internal/repository/finance_repository.go)
**财务数据访问**

#### [datacenter_repository.go](internal/repository/datacenter_repository.go)
**数据中心数据访问**

---

### 📁 internal/services/ 目录

#### [order_service.go](internal/services/order_service.go)
**订单服务**

**功能**：
- `CreateOrder()` 创建订单
- `ListOrders()` 查询订单列表
- `AuditOrder()` 审核订单
- `ShipOrder()` 发货
- `CancelOrder()` 取消订单
- `GetOrderStatistics()` 订单统计

#### [inventory_service.go](internal/services/inventory_service.go)
**库存服务**

**功能**：
- `GetStock()` 查询库存
- `AdjustStock()` 调整库存
- `CreateInboundOrder()` 创建入库单
- `CreateOutboundOrder()` 创建出库单
- `ProcessInboundOrder()` 处理入库
- `GetInventoryAlert()` 库存预警

#### [product_service.go](internal/services/product_service.go)
**商品服务**

#### [api_sync_service.go](internal/services/api_sync_service.go)
**API 同步服务** ⭐

**功能**：
- `SyncOrders()` 批量同步订单
- `syncSingleOrder()` 同步单个订单
- `getOrCreateShop()` 自动创建店铺
- `convertToOrder()` 转换订单格式
- `convertOrderStatus()` 映射订单状态
- `GetSyncStatistics()` 获取同步统计

**核心接口**：外部系统通过此服务推送订单数据

#### [sync_service.go](internal/services/sync_service.go)
**平台同步服务**

**功能**：
- 同步店铺订单
- 同步商品信息
- 同步物流信息

#### [oauth_service.go](internal/services/oauth_service.go)
**OAuth 服务**

**功能**：
- `GetAuthURL()` 生成授权 URL
- `HandleCallback()` 处理授权回调
- `RefreshToken()` 刷新令牌

#### [permission_service.go](internal/services/permission_service.go)
**权限服务**

**功能**：
- `CreateRole()` 创建角色
- `AssignUserRole()` 分配角色
- `SetResourcePermissions()` 设置资源权限
- `CheckPermission()` 检查权限
- `SeedData()` 初始化权限数据

#### [finance_service.go](internal/services/finance_service.go)
**财务服务**

#### [statistics_service.go](internal/services/statistics_service.go)
**统计服务**

#### [realtime_service.go](internal/services/realtime_service.go)
**实时数据服务**

#### [customer_analysis_service.go](internal/services/customer_analysis_service.go)
**客户分析服务**

#### [product_analysis_service.go](internal/services/product_analysis_service.go)
**商品分析服务**

#### [compare_analysis_service.go](internal/services/compare_analysis_service.go)
**对比分析服务**

#### [alert_service.go](internal/services/alert_service.go)
**预警服务**

#### [cache_decorator.go](internal/services/cache_decorator.go)
**缓存装饰器**

**功能**：
- `GetOrSet()` 缓存查询模式
- `InvalidateProductCache()` 使商品缓存失效
- `InvalidateInventoryCache()` 使库存缓存失效
- `InvalidateOrderCache()` 使订单缓存失效

#### [cache_service.go](internal/services/cache_service.go)
**缓存服务**

#### [websocket_service.go](internal/services/websocket_service.go)
**WebSocket 服务**

#### [scheduler_service.go](internal/services/scheduler_service.go)
**定时任务服务**

---

### 📁 internal/handlers/ 目录

#### [auth.go](internal/handlers/auth.go)
**认证处理器**

**功能**：
- `Register()` 用户注册
- `Login()` 用户登录
- `GetCurrentUser()` 获取当前用户
- `ListUsers()` 用户列表
- `ApproveUser()` 审核用户

#### [order.go](internal/handlers/order.go)
**订单处理器**

**功能**：
- `ListOrders()` 订单列表
- `GetOrder()` 订单详情
- `CreateOrder()` 创建订单
- `AuditOrder()` 审核订单
- `ShipOrder()` 发货

#### [shop.go](internal/handlers/shop.go)
**店铺处理器**

**功能**：
- `List()` 店铺列表
- `Create()` 创建店铺
- `TriggerSync()` 触发同步
- `GetAuthURL()` 获取授权 URL

#### [product.go](internal/handlers/product.go)
**商品处理器**

#### [warehouse.go](internal/handlers/warehouse.go)
**仓库处理器**

#### [inventory.go](internal/handlers/inventory.go)
**库存处理器**

#### [tenant.go](internal/handlers/tenant.go)
**租户处理器**

**功能**：
- `List()` 租户列表
- `MyTenants()` 我的租户
- `Create()` 创建租户
- `SwitchTenant()` 切换租户

#### [permission.go](internal/handlers/permission.go)
**权限处理器**

#### [statistics.go](internal/handlers/statistics.go)
**统计处理器**

#### [finance_handler.go](internal/handlers/finance_handler.go)
**财务处理器**

#### [datacenter_handler.go](internal/handlers/datacenter_handler.go)
**数据中心处理器**

#### [api_sync_handler.go](internal/handlers/api_sync_handler.go)
**API 同步处理器** ⭐

**功能**：
- `SyncOrders()` 批量同步订单
- `GetSyncStatistics()` 获取同步统计

**核心接口**：`POST /api/v1/sync/orders`

#### [oauth.go](internal/handlers/oauth.go)
**OAuth 处理器**

#### [websocket_handler.go](internal/handlers/websocket_handler.go)
**WebSocket 处理器**

---

### 📁 internal/middleware/ 目录

#### [auth.go](internal/middleware/auth.go)
**认证中间件**

**功能**：
- `CORS()` 跨域处理
- `JWTAuth()` JWT 认证
- `JWTAuthWithBlacklist()` 带黑名单的认证
- `OptionalAuth()` 可选认证

#### [tenant.go](internal/middleware/tenant.go)
**租户中间件**

**功能**：
- `TenantMiddleware()` 租户检查
- `GetTenantIDFromGin()` 从 Gin 上下文获取租户ID

**优先级**：
1. HTTP Header `X-Tenant-ID`
2. Cookie `tenant_id`
3. Gin Context

#### [permission.go](internal/middleware/permission.go)
**权限中间件**

**功能**：
- `RequirePermission()` 需要特定权限
- `RequirePermissionAssign()` 需要授权权限

#### [rate_limiter.go](internal/middleware/rate_limiter.go)
**限流中间件**

**功能**：
- `SimpleRateLimiter()` 简单限流器
- `LoginRateLimit()` 登录限流（每IP每分钟5次）

#### [audit.go](internal/middleware/audit.go)
**审计中间件**

#### [error_handler.go](internal/middleware/error_handler.go)
**错误处理中间件**

#### [sanitization.go](internal/middleware/sanitization.go)
**输入清理中间件**

#### [security.go](internal/middleware/security.go)
**安全中间件**

---

### 📁 internal/kafka/ 目录

#### [config.go](internal/kafka/config.go)
**Kafka 配置**

#### [producer.go](internal/kafka/producer.go)
**Kafka 生产者**

**功能**：
- `NewProducer()` 创建生产者
- `SendMessage()` 发送消息
- `SendOrderMessage()` 发送订单消息

#### [consumer.go](internal/kafka/consumer.go)
**Kafka 消费者**

**功能**：
- `NewConsumer()` 创建消费者
- `Start()` 启动消费者
- `Stop()` 停止消费者

#### [message.go](internal/kafka/message.go)
**消息定义**

#### [erp_handler.go](internal/kafka/erp_handler.go)
**ERP 订单处理器**

**功能**：
- `HandleOrderCreate()` 处理订单创建
- `HandleOrderUpdate()` 处理订单更新
- `HandleOrderCancel()` 处理订单取消

#### [simulator.go](internal/kafka/simulator.go)
**消息模拟器**

---

### 📁 internal/platform/ 目录

#### [registry.go](internal/platform/registry.go)
**平台注册表**

**支持的平台**：
- 淘宝
- 天猫
- 抖音
- 快手
- 微信视频号
- TikTok
- 京东
- 拼多多
- 自定义平台

#### [types.go](internal/platform/types.go)
**平台类型定义**

#### [adapters.go](internal/platform/adapters.go)
**平台适配器**

#### [clients/base.go](internal/platform/clients/base.go)
**基础客户端**

#### [clients/taobao.go](internal/platform/clients/taobao.go)
**淘宝客户端**

#### [clients/douyin.go](internal/platform/clients/douyin.go)
**抖音客户端**

#### [clients/kuaishou.go](internal/platform/clients/kuaishou.go)
**快手客户端**

---

### 📁 internal/pkg/ 目录

#### [errors/errors.go](internal/pkg/errors/errors.go)
**错误处理**

**功能**：
- 统一错误定义
- 错误码常量
- 错误包装函数

#### [logger/logger.go](internal/pkg/logger/logger.go)
**日志服务**

#### [response/response.go](internal/pkg/response/response.go)
**响应封装**

**功能**：
- `Success()` 成功响应
- `Error()` 错误响应
- `BadRequest()` 错误请求响应
- `Unauthorized()` 未授权响应

#### [security/password.go](internal/pkg/security/password.go)
**密码工具**

#### [validator/validator.go](internal/pkg/validator/validator.go)
**参数验证**

---

### 📁 internal/seed/ 目录

#### [seed.go](internal/seed/seed.go)
**数据种子**

**功能**：
- `SeedAll()` 生成所有演示数据
- `SeedUsers()` 生成用户
- `SeedTenants()` 生成租户
- `SeedOrders()` 生成订单
- `clearData()` 清空现有数据

---

### 📁 docs/ 目录

#### [swagger.go](docs/swagger.go)
**API 文档定义**

#### [swagger_models.go](docs/swagger_models.go)
**API 模型定义**

---

## 多租户架构

### 核心概念

OURERP 采用**共享数据库、共享 Schema**的多租户架构：

```
┌─────────────────────────────────────┐
│         PostgreSQL 数据库             │
│                                      │
│  ┌────────────────────────────────┐ │
│  │      orders 表                  │ │
│  ├────────────────────────────────┤ │
│  │ id | tenant_id | order_no     │ │
│  │ 1  |    1      | ORD001       │ │  ← 租户1的订单
│  │ 2  |    1      | ORD002       │ │  ← 租户1的订单
│  │ 3  |    2      | ORD003       │ │  ← 租户2的订单
│  │ 4  |    2      | ORD004       │ │  ← 租户2的订单
│  └────────────────────────────────┘ │
└─────────────────────────────────────┘
```

### 租户隔离机制

#### 1. 数据库层面
- 所有业务表都有 `tenant_id` 字段
- 通过数据库索引隔离不同租户的数据

#### 2. 应用层面
```go
// Context 传递租户ID
type contextKey string
const tenantIDKey contextKey = "tenant_id"

func SetTenantIDToContext(ctx context.Context, tenantID int64) context.Context {
    return context.WithValue(ctx, tenantIDKey, tenantID)
}

func GetTenantIDFromContext(ctx context.Context) int64 {
    if tid, ok := ctx.Value(tenantIDKey).(int64); ok {
        return tid
    }
    return 0
}
```

#### 3. 自动过滤
```go
// 所有数据库查询自动添加租户条件
db.Scopes(WithTenantFromContext(ctx)).Find(&orders)
// 生成 SQL: SELECT * FROM orders WHERE tenant_id = 1
```

### 租户ID来源

**优先级**：
1. HTTP Header `X-Tenant-ID: 1`
2. Cookie `tenant_id=1`
3. Gin Context

**切换租户流程**：
```http
POST /api/v1/tenants/switch
Headers:
  Authorization: Bearer eyJhbGc...

Body:
{
  "tenant_id": 2
}

# 响应
{
  "code": "SUCCESS",
  "message": "Tenant switched"
}
```

---

## 认证授权

### JWT Token 生成

**登录流程**：

```go
// 1. 用户登录
POST /api/v1/auth/login
{
  "email": "root@ourerp.com",
  "password": "root123456"
}

// 2. 验证成功后生成 Token
token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
    "user_id": 3,
    "email":   "root@ourerp.com",
    "exp":     time.Now().Add(24 * time.Hour).Unix(),
})

tokenString, _ := token.SignedString([]byte(JWT_SECRET))

// 3. 返回 Token
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 3,
    "email": "root@ourerp.com",
    "name": "Root"
  }
}
```

### Token 使用

```http
GET /api/v1/orders
Headers:
  Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
  X-Tenant-ID: 1
```

### 权限验证

**基于角色的访问控制 (RBAC)**：

```
用户 → 角色 → 权限
     ↓
  资源级权限
  (某个店铺、某个仓库)
```

**权限示例**：
```go
// 权限常量
const (
    PermOrderWrite    = "order:write"     // 订单写入权限
    PermInventoryView = "inventory:view"  // 库存查看权限
    PermFinanceAudit  = "finance:audit"   // 财务审核权限
)

// 使用中间件保护路由
tenantRequired.POST("/orders",
    permMiddleware.RequirePermission(models.PermOrderWrite),
    orderHandler.CreateOrder)
```

---

## 外部系统对接

### API 对接方式 ⭐

OURERP 提供两种对接方式：

#### 1. REST API 对接 (推荐)

**特点**：
- ✅ 简单直接
- ✅ 实时同步
- ✅ 易于调试

**核心接口**：`POST /api/v1/sync/orders`

**详细文档**：参考 [API对接指南.md](API对接指南.md)

#### 2. Kafka 消息队列 (可选)

**特点**：
- ✅ 高吞吐量
- ✅ 解耦系统
- ✅ 削峰填谷

**适用场景**：
- 大批量订单同步
- 异步处理
- 系统解耦

### API 对接流程

**步骤 1：登录获取 Token**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "root@ourerp.com",
    "password": "root123456"
  }'
```

**步骤 2：切换账套**
```bash
curl -X POST http://localhost:8080/api/v1/tenants/switch \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"tenant_id": 1}'
```

**步骤 3：推送订单**
```bash
curl -X POST http://localhost:8080/api/v1/sync/orders \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "X-Tenant-ID: 1" \
  -H "Content-Type: application/json" \
  -d '{
    "source": "taobao",
    "orders": [{
      "platform": "taobao",
      "platform_order_id": "TB123456789",
      "status": "paid",
      "total_amount": 299.00,
      "pay_amount": 299.00,
      "buyer_nick": "买家昵称",
      "receiver_name": "收货人",
      "receiver_phone": "13800138000",
      "receiver_address": "浙江省杭州市...",
      "items": [{
        "sku_id": "12345",
        "sku_name": "商品名称",
        "quantity": 2,
        "price": 299.00
      }]
    }]
  }'
```

### 数据格式映射

**平台状态 → 系统状态**：

| 平台状态 | 系统状态 | 说明 |
|---------|---------|------|
| pending_payment | 1 | 待付款 |
| paid / pending_ship | 2 | 待发货 |
| shipped | 3 | 已发货 |
| completed | 4 | 已完成 |
| cancelled | 5 | 已取消 |

**支持的平台**：

| 平台代码 | 平台名称 |
|---------|---------|
| taobao | 淘宝 |
| tmall | 天猫 |
| jd | 京东 |
| douyin | 抖音 |
| kuaishou | 快手 |
| pdd | 拼多多 |
| custom | 自定义 |

---

## 学习路径

### 初级阶段 (1-2周)

**目标**：理解项目结构和基础概念

1. **环境搭建**
   ```bash
   # 启动 PostgreSQL
   docker run -d --name postgres \
     -e POSTGRES_PASSWORD=postgres \
     -p 5432:5432 postgres:14

   # 启动 Redis
   docker run -d --name redis \
     -p 6379:6379 redis:7

   # 启动后端
   cd backend
   go run cmd/server/main.go
   ```

2. **阅读顺序**
   - [cmd/server/main.go](cmd/server/main.go) - 了解启动流程
   - [internal/config/config.go](internal/config/config.go) - 了解配置
   - [internal/models/](internal/models/) - 了解数据模型
   - [internal/middleware/auth.go](internal/middleware/auth.go) - 了解认证

3. **实践任务**
   - 创建一个新的 API 端点
   - 添加一个新的数据模型
   - 实现简单的 CRUD 功能

### 中级阶段 (2-4周)

**目标**：掌握核心业务逻辑

1. **深入阅读**
   - [internal/repository/](internal/repository/) - 数据访问层
   - [internal/services/](internal/services/) - 业务逻辑层
   - [internal/handlers/](internal/handlers/) - HTTP 处理层

2. **重点关注**
   - 租户隔离机制
   - 订单处理流程
   - 库存管理逻辑

3. **实践任务**
   - 实现一个完整的业务功能
   - 添加单元测试
   - 优化查询性能

### 高级阶段 (4-8周)

**目标**：掌握高级特性和架构设计

1. **深入学习**
   - Kafka 消息队列
   - 平台适配器
   - 数据中心和分析
   - 权限系统

2. **架构设计**
   - 缓存策略
   - 性能优化
   - 安全加固

3. **实践任务**
   - 接入一个新的电商平台
   - 实现复杂的业务功能
   - 性能调优

### 推荐资源

**Go 语言**：
- [Go 官方教程](https://go.dev/tour/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go by Example](https://gobyexample.com/)

**Gin 框架**：
- [Gin 官方文档](https://gin-gonic.com/docs/)
- [Gin 源码分析](https://github.com/gin-gonic/gin)

**GORM**：
- [GORM 官方文档](https://gorm.io/docs/)
- [GORM 最佳实践](https://gorm.io/docs/docs.html)

**系统设计**：
- [Designing Data-Intensive Applications](https://dataintensive.net/)
- [系统设计面试](https://github.com/checkcheckzz/system-design-interview)

---

## 常见问题

### Q1: 如何调试？

**A**: 使用日志和断点

```go
// 添加日志
logger.Infof("Processing order: %s", orderNo)

// 使用 Delve 调试器
dlv debug cmd/server/main.go
```

### Q2: 如何测试？

**A**: 编写单元测试

```go
func TestOrderService_CreateOrder(t *testing.T) {
    // 1. 准备测试数据
    // 2. 调用被测试函数
    // 3. 验证结果
}
```

### Q3: 如何部署？

**A**: 使用 Docker

```bash
# 构建镜像
docker build -t ourerp-backend .

# 运行容器
docker run -d -p 8080:8080 \
  -e DB_HOST=postgres \
  -e REDIS_HOST=redis \
  ourerp-backend
```

### Q4: 如何扩展？

**A**: 添加新功能

1. 创建 Model
2. 创建 Repository
3. 创建 Service
4. 创建 Handler
5. 注册路由

### Q5: 如何贡献？

**A**:
1. Fork 项目
2. 创建分支
3. 提交代码
4. 发起 Pull Request

---

## 附录

### 文件索引

**按功能分类**：

| 功能模块 | 文件数 | 关键文件 |
|---------|-------|---------|
| 订单管理 | 5 | order.go, order_service.go, order_repository.go |
| 库存管理 | 6 | inventory.go, inventory_service.go |
| 商品管理 | 3 | product.go, product_service.go |
| 店铺管理 | 3 | shop.go, sync_service.go |
| 用户管理 | 3 | user.go, auth.go |
| 租户管理 | 3 | tenant.go, tenant_repository.go |
| 权限管理 | 3 | rbac.go, permission_service.go |
| 财务管理 | 1 | finance.go (含11个子模型) |
| 数据中心 | 1 | datacenter.go (含10个子模型) |
| 平台对接 | 8 | platform/*.go |
| Kafka | 6 | kafka/*.go |
| 中间件 | 9 | middleware/*.go |

### 代码统计

| 目录 | 文件数 | 代码行数 (估算) |
|-----|-------|---------------|
| models | 14 | ~2,500 |
| repository | 13 | ~3,000 |
| services | 20 | ~5,000 |
| handlers | 20 | ~4,000 |
| middleware | 9 | ~1,500 |
| kafka | 6 | ~1,200 |
| platform | 8 | ~2,000 |
| pkg | 6 | ~800 |
| 其他 | 24 | ~2,759 |
| **总计** | **120** | **~22,759** |

### 联系方式

- GitHub: https://github.com/MorantHP/OURERP
- Issues: https://github.com/MorantHP/OURERP/issues
- 文档: [API对接指南.md](API对接指南.md)

---

**文档版本**: v1.0
**最后更新**: 2026-03-05
**维护者**: OURERP Team
