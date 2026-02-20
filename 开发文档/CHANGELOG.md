# 安全修复日志 (2026-02-15)

## 修复的安全问题

### 1. 硬编码敏感信息 (严重)
**文件**: `backend/internal/config/config.go`

**问题**:
- 数据库密码 `ourerp123` 硬编码
- JWT密钥 `your-secret-key` 硬编码

**修复**:
- 移除所有敏感信息的默认值
- 必须通过环境变量设置 `DB_PASSWORD`, `REDIS_PASSWORD`, `JWT_SECRET`

---

### 2. CORS配置过于宽松 (严重)
**文件**: `backend/internal/middleware/auth.go`

**问题**:
- `Access-Control-Allow-Origin: *` 允许任意来源访问

**修复**:
- 新增环境变量 `CORS_ALLOWED_ORIGINS` 配置允许的域名
- 开发环境自动允许 localhost
- 生产环境必须显式配置允许的域名

---

### 3. 数据库文件泄露 (严重)
**文件**: `backend/ourerp.db`

**问题**:
- SQLite数据库文件被提交到git仓库

**修复**:
- 删除数据库文件
- 更新 `.gitignore` 忽略所有 `*.db`, `*.sqlite` 文件

---

### 4. Vue语法错误 (严重)
**文件**: `frontend/src/views/orders/OrderListView.vue`

**问题**:
- `</style>` 标签后有多余代码导致解析错误

**修复**:
- 删除文件末尾的错误代码（第297-312行）

---

### 5. 登录速率限制缺失 (严重)
**新文件**: `backend/internal/middleware/rate_limiter.go`

**问题**:
- 登录接口无速率限制，存在暴力破解风险

**修复**:
- 新增 `RateLimiter` 中间件
- 登录/注册接口限制每IP每分钟最多5次请求
- 返回 HTTP 429 状态码提示用户

---

### 6. 重复代码定义 (中等)
**文件**: `backend/internal/models/order.go`

**问题**:
- `LoginRequest` 在 `order.go` 和 `user.go` 中重复定义

**修复**:
- 删除 `order.go` 中的重复定义，保留 `user.go` 中的版本

---

### 7. 重复Store文件 (中等)
**文件**: `frontend/src/stores/stores/user.ts`

**问题**:
- 存在重复的mock版本Store文件

**修复**:
- 删除整个 `frontend/src/stores/stores/` 目录

---

### 8. 前端硬编码API地址 (中等)
**文件**: `frontend/src/utils/request.ts`

**问题**:
- API地址硬编码为 `http://localhost:8080/api/v1`

**修复**:
- 改为从环境变量 `VITE_API_BASE_URL` 读取
- 保留默认值作为开发环境使用

---

## 新增文件

| 文件 | 说明 |
|------|------|
| `backend/.env.example` | 后端环境变量配置示例 |
| `frontend/.env.example` | 前端环境变量配置示例 |
| `backend/internal/middleware/rate_limiter.go` | 速率限制中间件 |

---

## 更新文件

| 文件 | 修改内容 |
|------|----------|
| `backend/internal/config/config.go` | 移除硬编码敏感信息 |
| `backend/internal/middleware/auth.go` | 改进CORS配置 |
| `backend/cmd/server/main.go` | 添加登录速率限制 |
| `backend/internal/models/order.go` | 删除重复定义 |
| `frontend/src/utils/request.ts` | API地址改用环境变量 |
| `frontend/src/views/orders/OrderListView.vue` | 修复语法错误 |
| `.gitignore` | 添加数据库、日志等忽略规则 |

---

## 部署说明

### 后端配置
```bash
cd backend
cp .env.example .env
# 编辑 .env 文件，设置安全的密码和密钥
```

### 前端配置
```bash
cd frontend
cp .env.example .env.local
# 编辑 .env.local 文件，设置API地址
```

### 生产环境注意事项
1. `JWT_SECRET` 必须使用至少32位随机字符串
2. `DB_PASSWORD` 必须使用强密码
3. `CORS_ALLOWED_ORIGINS` 必须配置为实际域名，不要使用通配符
4. 确保 `.env` 文件不会被提交到版本控制

---

# 多电商平台对接功能 (2026-02-15)

## 新增功能

### 支持的电商平台

| 平台 | 认证方式 | 状态 |
|------|---------|------|
| 淘宝/天猫 | OAuth2.0 | ✅ 已实现 |
| 抖音电商 | OAuth2.0 | ✅ 已实现 |
| 快手电商 | OAuth2.0 | ✅ 已实现 |
| 微信视频号 | OAuth2.0 | ✅ 已实现 |
| 自定义平台 | API Key/Webhook | ✅ 已实现 |
| TikTok小店 | OAuth2.0 | 🔜 待实现 |
| 京企直卖 | API Key | 🔜 待实现 |
| 京东 | OAuth2.0 | 🔜 待实现 |
| 小红书 | OAuth2.0 | 🔜 待实现 |
| 唯品会 | API Key | 🔜 待实现 |
| 1688 | OAuth2.0 | 🔜 待实现 |

### 同步策略
- **定时轮询**: 每5-60分钟自动拉取订单
- **Webhook**: 支持平台推送订单变更通知

---

## 新增文件

### 后端

| 文件 | 说明 |
|------|------|
| `backend/internal/models/shop.go` | 店铺模型 |
| `backend/internal/platform/clients/base.go` | 基础HTTP客户端（含重试） |
| `backend/internal/platform/clients/taobao.go` | 淘宝/天猫API客户端 |
| `backend/internal/platform/clients/douyin.go` | 抖音电商API客户端 |
| `backend/internal/platform/clients/kuaishou.go` | 快手电商API客户端 |
| `backend/internal/platform/clients/wechat_video.go` | 微信视频号API客户端 |
| `backend/internal/platform/clients/custom.go` | 自定义平台客户端 |
| `backend/internal/handlers/shop.go` | 店铺管理Handler |
| `backend/internal/handlers/platform.go` | 平台信息Handler |
| `backend/internal/handlers/oauth.go` | OAuth授权Handler |

### 前端

| 文件 | 说明 |
|------|------|
| `frontend/src/api/shop.ts` | 店铺和平台API |
| `frontend/src/views/shops/ShopListView.vue` | 店铺管理页面 |

---

## 更新文件

| 文件 | 修改内容 |
|------|----------|
| `backend/internal/platform/types.go` | 新增平台类型：TikTok、京企直卖、1688等 |
| `backend/internal/platform/registry.go` | 所有平台配置 |
| `backend/internal/repository/shop_repository.go` | 扩展店铺仓库方法 |
| `backend/cmd/server/main.go` | 添加店铺、平台、OAuth路由 |
| `frontend/src/router/index.ts` | 添加店铺管理路由 |
| `frontend/src/views/LayoutView.vue` | 添加店铺管理菜单 |

---

## API接口

### 平台信息
```
GET  /api/v1/platforms           - 获取支持的平台列表
GET  /api/v1/platforms/:code     - 获取单个平台配置
```

### 店铺管理
```
GET    /api/v1/shops             - 店铺列表
POST   /api/v1/shops             - 创建店铺
GET    /api/v1/shops/:id         - 店铺详情
PUT    /api/v1/shops/:id         - 更新店铺
DELETE /api/v1/shops/:id         - 删除店铺
POST   /api/v1/shops/:id/sync    - 手动同步
GET    /api/v1/shops/:id/auth-url - 获取授权URL
```

### OAuth授权
```
GET  /api/v1/oauth/auth-url?shop_id=xxx - 获取授权URL
GET  /api/v1/oauth/callback              - OAuth回调
POST /api/v1/oauth/refresh?shop_id=xxx   - 刷新Token
```

---

## 使用说明

### 1. 添加店铺
1. 进入「店铺管理」页面
2. 点击「添加店铺」
3. 选择平台、填写店铺名称和App Key/Secret
4. 保存后点击「授权」完成OAuth授权

### 2. 同步订单
- **自动同步**: 店铺授权成功后，系统按设定间隔自动同步
- **手动同步**: 点击店铺列表中的「同步」按钮

### 3. 自定义平台
- 支持API Key认证方式
- 支持Webhook接收订单推送
- 需要配置API地址和密钥

---

# 同步服务重构 (2026-02-16)

## 新增功能

### 订单同步服务
- 重构 `SyncService` 支持多平台订单同步
- 支持淘宝/天猫、抖音、快手、微信视频号、自定义平台
- 自动检测Token过期并提示重新授权
- 订单状态自动映射到内部统一状态

### 定时调度服务
- 新增 `SchedulerService` 定时同步调度器
- 按店铺配置的同步间隔自动拉取订单
- 支持立即手动同步
- 支持批量同步所有店铺

---

## 新增文件

| 文件 | 说明 |
|------|------|
| `backend/internal/services/scheduler_service.go` | 定时同步调度服务 |

---

## 更新文件

| 文件 | 修改内容 |
|------|----------|
| `backend/internal/services/sync_service.go` | 重构支持多平台订单同步 |

---

## 同步服务接口

```go
// 同步店铺订单
syncService.SyncShopOrders(ctx, shopID, startTime, endTime) (*SyncResult, error)

// 调度服务
schedulerService.Start()                              // 启动调度
schedulerService.Stop()                               // 停止调度
schedulerService.SyncShopNow(shopID)                  // 立即同步
schedulerService.UpdateShopInterval(shopID, interval) // 更新间隔
schedulerService.SyncAllShops()                       // 同步所有店铺
```

---

## 同步结果结构

```go
type SyncResult struct {
    ShopID       int64
    ShopName     string
    Platform     string
    TotalSynced  int
    TotalFailed  int
    ErrorMessage string
    SyncedAt     time.Time
    Duration     time.Duration
}
```

