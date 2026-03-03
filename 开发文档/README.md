# OURERP - 多渠道电商ERP系统

## 目录

- [项目简介](#项目简介)
- [系统架构](#系统架构)
- [技术栈](#技术栈)
- [功能模块](#功能模块)
- [快速开始](#快速开始)
- [开发指南](#开发指南)
- [API接口文档](#api接口文档)
- [部署指南](#部署指南)
- [配置说明](#配置说明)

---

## 项目简介

OURERP 是一个面向电商卖家的多渠道 ERP 管理系统，支持淘宝、天猫、抖音、快手、微信视频号等主流电商平台。系统采用前后端分离架构，提供订单管理、库存管理、财务管理、数据分析等核心功能。

### 核心特性

- **多平台支持**: 支持淘宝/天猫、抖音、快手、微信视频号等主流电商平台
- **多租户架构**: 支持多账套管理，数据完全隔离
- **实时数据推送**: WebSocket 实时推送订单、库存等数据变更
- **OAuth 授权**: 一键授权电商平台，自动同步订单数据
- **权限管理**: 细粒度的角色权限控制
- **数据缓存**: Redis 缓存层，提升系统性能

---

## 系统架构

### 整体架构图

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           客户端层 (Frontend)                             │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │                     Vue 3 + TypeScript + Pinia                     │   │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐     │   │
│  │  │  订单   │ │  库存   │ │  财务   │ │ 数据中心 │ │  设置   │     │   │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────┘ └─────────┘     │   │
│  └──────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
                                      │
                                      │ HTTP/WebSocket
                                      ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                           服务端层 (Backend)                              │
│  ┌──────────────────────────────────────────────────────────────────┐   │
│  │                      Go + Gin + GORM                               │   │
│  │  ┌─────────────────────────────────────────────────────────────┐  │   │
│  │  │                      中间件层                                 │  │   │
│  │  │  JWT认证 │ 租户隔离 │ 权限控制 │ 速率限制 │ CORS │ 日志     │  │   │
│  │  └─────────────────────────────────────────────────────────────┘  │   │
│  │  ┌─────────────────────────────────────────────────────────────┐  │   │
│  │  │                      处理器层 (Handlers)                      │  │   │
│  │  │  Auth │ Order │ Shop │ Inventory │ Finance │ Datacenter     │  │   │
│  │  └─────────────────────────────────────────────────────────────┘  │   │
│  │  ┌─────────────────────────────────────────────────────────────┐  │   │
│  │  │                      服务层 (Services)                        │  │   │
│  │  │  Sync │ Statistics │ Permission │ OAuth │ WebSocket │ Cache │  │   │
│  │  └─────────────────────────────────────────────────────────────┘  │   │
│  │  ┌─────────────────────────────────────────────────────────────┐  │   │
│  │  │                     数据访问层 (Repository)                   │  │   │
│  │  └─────────────────────────────────────────────────────────────┘  │   │
│  └──────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
                                      │
                    ┌─────────────────┼─────────────────┐
                    ▼                 ▼                 ▼
            ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
            │  PostgreSQL  │  │    Redis     │  │ 电商平台API  │
            │   (主数据库)  │  │   (缓存)     │  │ (淘宝/抖音等) │
            └──────────────┘  └──────────────┘  └──────────────┘
```

### 目录结构

```
OURERP/
├── backend/                      # 后端代码
│   ├── cmd/
│   │   ├── server/               # 主程序入口
│   │   │   └── main.go
│   │   └── fix/                  # 修复脚本
│   ├── internal/
│   │   ├── cache/                # 缓存服务
│   │   ├── config/               # 配置管理
│   │   ├── handlers/             # HTTP处理器
│   │   ├── middleware/           # 中间件
│   │   ├── models/               # 数据模型
│   │   ├── pkg/                  # 公共包
│   │   │   ├── errors/           # 错误处理
│   │   │   ├── logger/           # 日志系统
│   │   │   ├── security/         # 安全工具
│   │   │   └── validator/        # 输入验证
│   │   ├── platform/             # 平台适配
│   │   │   └── clients/          # 平台客户端
│   │   ├── repository/           # 数据访问层
│   │   ├── seed/                 # 种子数据
│   │   ├── services/             # 业务服务层
│   │   └── tests/                # 测试文件
│   ├── docs/                     # API文档
│   ├── scripts/                  # 脚本文件
│   ├── Dockerfile                # 生产Dockerfile
│   ├── Dockerfile.dev            # 开发Dockerfile
│   ├── go.mod
│   └── go.sum
│
├── frontend/                     # 前端代码
│   ├── src/
│   │   ├── api/                  # API接口
│   │   ├── assets/               # 静态资源
│   │   ├── components/           # 公共组件
│   │   ├── router/               # 路由配置
│   │   ├── stores/               # 状态管理
│   │   ├── utils/                # 工具函数
│   │   └── views/                # 页面视图
│   │       ├── dashboard/        # 仪表盘
│   │       ├── orders/           # 订单管理
│   │       ├── shops/            # 店铺管理
│   │       ├── inventory/        # 库存管理
│   │       ├── tenants/          # 账套管理
│   │       ├── finance/          # 财务管理
│   │       ├── datacenter/       # 数据中心
│   │       └── settings/         # 系统设置
│   ├── package.json
│   └── vite.config.ts
│
├── .github/
│   └── workflows/
│       └── ci.yml                # CI/CD配置
│
├── docker-compose.yml            # Docker编排
└── 开发文档/
    └── readme.md                 # 本文档
```

---

## 技术栈

### 后端

| 技术 | 版本 | 说明 |
|------|------|------|
| Go | 1.22+ | 主开发语言 |
| Gin | 1.9+ | Web框架 |
| GORM | 1.25+ | ORM框架 |
| PostgreSQL | 15+ | 主数据库 |
| Redis | 7+ | 缓存数据库 |
| JWT | - | 身份认证 |
| WebSocket | - | 实时通信 |
| Argon2id | - | 密码哈希 |

### 前端

| 技术 | 版本 | 说明 |
|------|------|------|
| Vue | 3.4+ | 前端框架 |
| TypeScript | 5.0+ | 类型支持 |
| Pinia | 2.1+ | 状态管理 |
| Vue Router | 4.2+ | 路由管理 |
| Element Plus | 2.4+ | UI组件库 |
| Vite | 5.0+ | 构建工具 |
| Axios | 1.6+ | HTTP客户端 |
| ECharts | 5.4+ | 图表库 |

---

## 功能模块

### 1. 订单管理

- **订单列表**: 分页查询、筛选、搜索
- **订单详情**: 查看订单完整信息
- **订单审核**: 审核待处理订单
- **订单发货**: 填写物流信息
- **状态流转**: 待付款 → 待发货 → 已发货 → 已完成

```
订单状态流转:
┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐
│ 待付款    │───▶│ 待发货    │───▶│ 已发货    │───▶│ 已完成    │
│   (0)    │    │   (1)    │    │   (2)    │    │   (3)    │
└──────────┘    └──────────┘    └──────────┘    └──────────┘
      │                              │
      │                              ▼
      │                        ┌──────────┐
      └───────────────────────▶│ 已取消    │
                               │   (4)    │
                               └──────────┘
```

### 2. 库存管理

- **商品管理**: 商品CRUD、SKU管理
- **仓库管理**: 多仓库支持
- **库存查询**: 实时库存、库存预警
- **入库管理**: 入库单、入库审核
- **出库管理**: 出库单、出库审核
- **库存盘点**: 盘点单、差异处理
- **库存调拨**: 跨仓库调拨

### 3. 店铺管理

- **店铺列表**: 多平台店铺管理
- **OAuth授权**: 一键授权电商平台
- **授权状态**: Token有效期管理
- **手动同步**: 触发订单同步
- **同步日志**: 同步记录查询

### 4. 财务管理

- **收支管理**: 收入支出记录
- **供应商管理**: 供应商信息维护
- **商品成本**: 成本价管理
- **订单成本**: 订单成本核算
- **月度结算**: 财务月结

### 5. 数据中心

- **实时监控**: 实时销售数据
- **客户分析**: 客户画像、RFM分析
- **商品分析**: 商品销量、利润分析
- **对比分析**: 同比、环比分析
- **预警管理**: 库存预警、销售预警

### 6. 系统设置

- **用户管理**: 用户CRUD、审批
- **角色管理**: 角色定义、权限分配
- **权限管理**: 细粒度权限控制
- **账套管理**: 多租户管理

---

## 快速开始

### 环境要求

- Go 1.22+
- Node.js 18+
- PostgreSQL 15+
- Redis 7+ (可选)

### 使用 Docker Compose (推荐)

```bash
# 克隆项目
git clone https://github.com/MorantHP/OURERP.git
cd OURERP

# 启动开发环境
docker-compose --profile development up -d

# 访问
# 前端: http://localhost:5173
# 后端: http://localhost:8080
```

### 手动安装

#### 1. 安装后端依赖

```bash
cd backend
go mod download
```

#### 2. 配置数据库

```bash
# 创建数据库
createdb ourerp

# 创建 .env 文件
cat > .env << EOF
DB_HOST=localhost
DB_PORT=5432
DB_USER=ourerp
DB_PASSWORD=your_password
DB_NAME=ourerp
DB_SSLMODE=disable
JWT_SECRET=your-secret-key-at-least-32-characters
JWT_EXPIRE=24
ROOT_PASSWORD=your-root-password
EOF
```

#### 3. 启动后端

```bash
go run cmd/server/main.go
```

#### 4. 安装前端依赖

```bash
cd frontend
npm install
```

#### 5. 启动前端

```bash
npm run dev
```

### 默认账号

- **邮箱**: root@ourerp.com
- **密码**: 由 `ROOT_PASSWORD` 环境变量指定（开发环境默认: root123456）

---

## 开发指南

### 开发流程

```
┌─────────────────────────────────────────────────────────────────┐
│                        开发流程图                                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  1. 需求分析                                                      │
│     └──▶ 明确功能需求和技术方案                                    │
│                                                                  │
│  2. 数据库设计                                                    │
│     └──▶ 设计模型 (internal/models/)                              │
│                                                                  │
│  3. 数据访问层                                                    │
│     └──▶ 实现Repository (internal/repository/)                    │
│                                                                  │
│  4. 业务服务层                                                    │
│     └──▶ 实现Service (internal/services/)                         │
│                                                                  │
│  5. HTTP处理器                                                    │
│     └──▶ 实现Handler (internal/handlers/)                         │
│                                                                  │
│  6. 路由注册                                                      │
│     └──▶ 注册路由 (cmd/server/main.go)                            │
│                                                                  │
│  7. 前端开发                                                      │
│     └──▶ 页面组件 (frontend/src/views/)                           │
│     └──▶ API接口 (frontend/src/api/)                              │
│                                                                  │
│  8. 测试验证                                                      │
│     └──▶ 单元测试 (internal/tests/)                               │
│     └──▶ 集成测试                                                 │
│                                                                  │
│  9. 代码提交                                                      │
│     └──▶ Git commit & push                                        │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 添加新功能示例

#### 1. 定义模型 (models/product.go)

```go
type Product struct {
    ID          int64   `gorm:"primaryKey" json:"id"`
    TenantID    int64   `gorm:"index" json:"tenant_id"`
    Name        string  `gorm:"size:200" json:"name"`
    Price       float64 `json:"price"`
    Status      int     `json:"status"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

#### 2. 创建仓库 (repository/product_repository.go)

```go
type ProductRepository struct {
    db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
    return &ProductRepository{db: db}
}

func (r *ProductRepository) List(tenantID int64, page, size int) ([]models.Product, int64, error) {
    var products []models.Product
    var total int64

    query := r.db.Model(&models.Product{}).Where("tenant_id = ?", tenantID)
    query.Count(&total)
    err := query.Offset((page - 1) * size).Limit(size).Find(&products).Error

    return products, total, err
}
```

#### 3. 创建处理器 (handlers/product_handler.go)

```go
type ProductHandler struct {
    repo *repository.ProductRepository
}

func (h *ProductHandler) List(c *gin.Context) {
    tenantID := middleware.GetTenantIDFromGin(c)
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

    products, total, err := h.repo.List(tenantID, page, size)
    if err != nil {
        c.JSON(500, gin.H{"error": "查询失败"})
        return
    }

    c.JSON(200, gin.H{"list": products, "total": total})
}
```

#### 4. 注册路由 (main.go)

```go
productRepo := repository.NewProductRepository(db)
productHandler := handlers.NewProductHandler(productRepo)

products := tenantRequired.Group("/products")
{
    products.GET("", productHandler.List)
    products.POST("", productHandler.Create)
    products.GET("/:id", productHandler.Get)
    products.PUT("/:id", productHandler.Update)
    products.DELETE("/:id", productHandler.Delete)
}
```

### 代码规范

1. **Go 代码规范**
   - 遵循 [Effective Go](https://golang.org/doc/effective_go)
   - 使用 `gofmt` 格式化代码
   - 错误处理使用 `internal/pkg/errors`
   - 日志使用 `internal/pkg/logger`

2. **TypeScript 代码规范**
   - 使用 ESLint + Prettier
   - 所有 API 响应定义类型
   - 组件使用 `<script setup>` 语法

---

## API接口文档

### 基础信息

- **Base URL**: `http://localhost:8080/api/v1`
- **认证方式**: JWT Bearer Token
- **租户标识**: `X-Tenant-ID` 请求头

### 通用响应格式

```json
// 成功响应
{
  "data": {},
  "message": "success"
}

// 列表响应
{
  "list": [],
  "pagination": {
    "total": 100,
    "page": 1,
    "size": 20,
    "total_pages": 5
  }
}

// 错误响应
{
  "error": "错误信息",
  "code": "ERROR_CODE"
}
```

---

### 认证接口 (Auth)

#### 用户登录

```
POST /api/v1/auth/login
```

**请求体:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**响应:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "用户名",
    "is_root": false,
    "is_approved": true
  }
}
```

#### 用户注册

```
POST /api/v1/auth/register
```

**请求体:**
```json
{
  "email": "newuser@example.com",
  "password": "password123",
  "name": "新用户"
}
```

**响应:**
```json
{
  "message": "注册成功，请等待审核"
}
```

#### 获取当前用户

```
GET /api/v1/auth/me
Authorization: Bearer <token>
```

**响应:**
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "用户名",
    "is_root": false,
    "is_approved": true,
    "status": 1
  }
}
```

#### 用户列表 (仅root)

```
GET /api/v1/users
Authorization: Bearer <token>
```

**查询参数:**
| 参数 | 类型 | 说明 |
|------|------|------|
| page | int | 页码 |
| size | int | 每页数量 |
| status | int | 状态筛选 |

**响应:**
```json
{
  "list": [
    {
      "id": 1,
      "email": "user@example.com",
      "name": "用户名",
      "is_approved": true,
      "status": 1
    }
  ],
  "pagination": {
    "total": 10,
    "page": 1,
    "size": 20
  }
}
```

#### 审核用户 (仅root)

```
PUT /api/v1/users/:id/approve
Authorization: Bearer <token>
```

**请求体:**
```json
{
  "approved": true
}
```

---

### 租户接口 (Tenants)

#### 租户列表

```
GET /api/v1/tenants
Authorization: Bearer <token>
```

**响应:**
```json
{
  "list": [
    {
      "id": 1,
      "name": "默认账套",
      "status": 1,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

#### 创建租户

```
POST /api/v1/tenants
Authorization: Bearer <token>
```

**请求体:**
```json
{
  "name": "新账套",
  "description": "账套描述"
}
```

#### 切换租户

```
POST /api/v1/tenants/switch
Authorization: Bearer <token>
```

**请求体:**
```json
{
  "tenant_id": 1
}
```

---

### 订单接口 (Orders)

> 以下接口需要 `X-Tenant-ID` 请求头

#### 订单列表

```
GET /api/v1/orders
Authorization: Bearer <token>
X-Tenant-ID: 1
```

**查询参数:**
| 参数 | 类型 | 说明 |
|------|------|------|
| page | int | 页码，默认1 |
| size | int | 每页数量，默认20，最大100 |
| status | int | 订单状态 |
| platform | string | 平台筛选 |
| keyword | string | 关键词搜索 |

**响应:**
```json
{
  "list": [
    {
      "id": 1,
      "order_no": "ORD202401010001",
      "platform": "taobao",
      "status": 1,
      "total_amount": 199.00,
      "pay_amount": 189.00,
      "buyer_nick": "买家昵称",
      "receiver_name": "收货人",
      "receiver_phone": "13800138000",
      "created_at": "2024-01-01T10:00:00Z",
      "items": [
        {
          "id": 1,
          "sku_name": "商品名称",
          "quantity": 2,
          "price": 99.50
        }
      ]
    }
  ],
  "pagination": {
    "total": 100,
    "page": 1,
    "size": 20,
    "total_pages": 5
  }
}
```

#### 订单详情

```
GET /api/v1/orders/:id
Authorization: Bearer <token>
X-Tenant-ID: 1
```

**响应:**
```json
{
  "order": {
    "id": 1,
    "order_no": "ORD202401010001",
    "platform": "taobao",
    "status": 1,
    "total_amount": 199.00,
    "pay_amount": 189.00,
    "buyer_nick": "买家昵称",
    "receiver_name": "收货人",
    "receiver_phone": "13800138000",
    "receiver_address": "详细地址",
    "logistics_company": "",
    "logistics_no": "",
    "created_at": "2024-01-01T10:00:00Z",
    "paid_at": "2024-01-01T10:05:00Z",
    "items": []
  }
}
```

#### 创建订单

```
POST /api/v1/orders
Authorization: Bearer <token>
X-Tenant-ID: 1
```

**请求体:**
```json
{
  "platform": "taobao",
  "platform_order_id": "TB123456",
  "shop_id": 1,
  "total_amount": 199.00,
  "pay_amount": 189.00,
  "buyer_nick": "买家昵称",
  "receiver_name": "收货人",
  "receiver_phone": "13800138000",
  "receiver_address": "详细地址",
  "items": [
    {
      "sku_id": 1,
      "sku_name": "商品名称",
      "quantity": 2,
      "price": 99.50
    }
  ]
}
```

#### 订单审核

```
POST /api/v1/orders/:id/audit
Authorization: Bearer <token>
X-Tenant-ID: 1
```

#### 订单发货

```
POST /api/v1/orders/:id/ship
Authorization: Bearer <token>
X-Tenant-ID: 1
```

**请求体:**
```json
{
  "logistics_company": "顺丰速运",
  "logistics_no": "SF1234567890"
}
```

---

### 店铺接口 (Shops)

> 以下接口需要 `X-Tenant-ID` 请求头

#### 店铺列表

```
GET /api/v1/shops
Authorization: Bearer <token>
X-Tenant-ID: 1
```

**响应:**
```json
{
  "list": [
    {
      "id": 1,
      "name": "我的淘宝店",
      "platform": "taobao",
      "platform_shop_id": "shop123",
      "status": 1,
      "last_sync_at": "2024-01-01T10:00:00Z",
      "token_expires_at": "2024-02-01T10:00:00Z"
    }
  ]
}
```

#### 创建店铺

```
POST /api/v1/shops
Authorization: Bearer <token>
X-Tenant-ID: 1
```

**请求体:**
```json
{
  "name": "我的淘宝店",
  "platform": "taobao",
  "platform_shop_id": "shop123",
  "app_key": "your_app_key",
  "app_secret": "your_app_secret"
}
```

#### 手动同步

```
POST /api/v1/shops/:id/sync
Authorization: Bearer <token>
X-Tenant-ID: 1
```

**响应:**
```json
{
  "message": "同步完成",
  "shop_id": 1,
  "total_sync": 50,
  "total_fail": 2,
  "duration_ms": 3500
}
```

---

### 库存接口 (Inventory)

> 以下接口需要 `X-Tenant-ID` 请求头

#### 商品列表

```
GET /api/v1/products
Authorization: Bearer <token>
X-Tenant-ID: 1
```

#### 库存查询

```
GET /api/v1/inventory
Authorization: Bearer <token>
X-Tenant-ID: 1
```

**查询参数:**
| 参数 | 类型 | 说明 |
|------|------|------|
| page | int | 页码 |
| size | int | 每页数量 |
| warehouse_id | int | 仓库ID |
| keyword | string | 关键词 |
| low_stock | bool | 仅显示低库存 |

**响应:**
```json
{
  "list": [
    {
      "id": 1,
      "product_id": 1,
      "product_name": "商品名称",
      "sku": "SKU001",
      "warehouse_id": 1,
      "warehouse_name": "主仓库",
      "quantity": 100,
      "available": 80,
      "locked": 20,
      "low_stock_threshold": 10
    }
  ]
}
```

#### 仓库列表

```
GET /api/v1/warehouses
Authorization: Bearer <token>
X-Tenant-ID: 1
```

---

### 财务接口 (Finance)

> 以下接口需要 `X-Tenant-ID` 请求头

#### 收支记录

```
GET /api/v1/finance/income-expense
Authorization: Bearer <token>
X-Tenant-ID: 1
```

#### 供应商列表

```
GET /api/v1/finance/suppliers
Authorization: Bearer <token>
X-Tenant-ID: 1
```

#### 商品成本

```
GET /api/v1/finance/product-costs
Authorization: Bearer <token>
X-Tenant-ID: 1
```

---

### 数据中心接口 (Datacenter)

> 以下接口需要 `X-Tenant-ID` 请求头

#### 实时数据

```
GET /api/v1/datacenter/realtime
Authorization: Bearer <token>
X-Tenant-ID: 1
```

**响应:**
```json
{
  "overview": {
    "today_orders": 50,
    "today_amount": 5000.00,
    "pending_orders": 10,
    "low_stock_items": 5
  },
  "trend": [
    {"hour": "00:00", "orders": 2, "amount": 200},
    {"hour": "01:00", "orders": 5, "amount": 500}
  ]
}
```

#### 客户分析

```
GET /api/v1/datacenter/customer
Authorization: Bearer <token>
X-Tenant-ID: 1
```

**查询参数:**
| 参数 | 类型 | 说明 |
|------|------|------|
| start_date | string | 开始日期 |
| end_date | string | 结束日期 |

#### 商品分析

```
GET /api/v1/datacenter/product
Authorization: Bearer <token>
X-Tenant-ID: 1
```

#### 预警规则

```
GET /api/v1/datacenter/alerts/rules
Authorization: Bearer <token>
X-Tenant-ID: 1
```

**创建预警规则:**
```
POST /api/v1/datacenter/alerts/rules
Authorization: Bearer <token>
X-Tenant-ID: 1
```

**请求体:**
```json
{
  "name": "库存预警",
  "type": "inventory_low",
  "condition": "quantity < 10",
  "level": "warning",
  "notify_type": "email",
  "enabled": true
}
```

---

### OAuth接口

#### 获取授权URL

```
GET /api/v1/oauth/auth-url?shop_id=1
Authorization: Bearer <token>
```

**响应:**
```json
{
  "auth_url": "https://oauth.taobao.com/authorize?...",
  "state": "abc123"
}
```

#### OAuth回调

```
GET /api/v1/oauth/callback?code=xxx&state=xxx
```

#### 刷新Token

```
POST /api/v1/oauth/refresh?shop_id=1
Authorization: Bearer <token>
```

---

### WebSocket接口

#### 连接

```
ws://localhost:8080/api/v1/ws
Authorization: Bearer <token>
X-Tenant-ID: 1 (通过查询参数或Header)
```

#### 消息格式

```json
{
  "type": "order_new",
  "tenant_id": 1,
  "timestamp": "2024-01-01T10:00:00Z",
  "data": {
    "order_no": "ORD202401010001",
    "amount": 199.00
  }
}
```

#### 消息类型

| 类型 | 说明 |
|------|------|
| connected | 连接成功 |
| order_new | 新订单 |
| order_update | 订单更新 |
| inventory_alert | 库存预警 |
| sync_status | 同步状态 |
| notification | 系统通知 |
| heartbeat | 心跳 |

---

## 部署指南

### Docker 部署

```bash
# 生产环境
docker-compose --profile production up -d

# 查看日志
docker-compose logs -f backend
```

### 手动部署

#### 1. 构建后端

```bash
cd backend
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ourerp-server ./cmd/server
```

#### 2. 构建前端

```bash
cd frontend
npm run build
# 产物在 dist/ 目录
```

#### 3. 配置 Nginx

```nginx
server {
    listen 80;
    server_name your-domain.com;

    # 前端
    location / {
        root /var/www/ourerp/dist;
        try_files $uri $uri/ /index.html;
    }

    # 后端API
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # WebSocket
    location /api/v1/ws {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

---

## 配置说明

### 环境变量

| 变量名 | 必需 | 说明 | 默认值 |
|--------|------|------|--------|
| `DB_HOST` | 是 | 数据库主机 | localhost |
| `DB_PORT` | 是 | 数据库端口 | 5432 |
| `DB_USER` | 是 | 数据库用户 | - |
| `DB_PASSWORD` | 是 | 数据库密码 | - |
| `DB_NAME` | 是 | 数据库名 | ourerp |
| `DB_SSLMODE` | 否 | SSL模式 | disable |
| `REDIS_HOST` | 否 | Redis主机 | localhost |
| `REDIS_PORT` | 否 | Redis端口 | 6379 |
| `JWT_SECRET` | **是** | JWT密钥(≥32字符) | - |
| `JWT_EXPIRE` | 否 | Token有效期(小时) | 24 |
| `ROOT_PASSWORD` | **生产必需** | root密码 | - |
| `CORS_ALLOWED_ORIGINS` | 否 | CORS允许的源 | - |

### 数据库索引

系统自动创建以下索引以优化查询性能：

```sql
-- 用户表
CREATE INDEX idx_users_email ON users(email);

-- 订单表
CREATE INDEX idx_orders_tenant_id ON orders(tenant_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at);

-- 库存表
CREATE INDEX idx_inventory_tenant_product ON inventories(tenant_id, product_id);
CREATE INDEX idx_inventory_warehouse ON inventories(warehouse_id);
```

---

## 许可证

MIT License

---

## 联系方式

- GitHub: https://github.com/MorantHP/OURERP
- Issues: https://github.com/MorantHP/OURERP/issues
