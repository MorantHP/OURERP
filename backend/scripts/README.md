# 数据中心模块测试脚本

本目录包含数据中心模块的API测试脚本和工具。

## 文件说明

| 文件 | 说明 |
|------|------|
| `test_datacenter.sh` | Linux/Mac Bash测试脚本 |
| `test_datacenter.bat` | Windows批处理测试脚本 |
| `postman_datacenter.json` | Postman集合文件，可导入Postman使用 |
| `datacenter_service_test.go` | Go单元测试文件 |

## 使用方法

### 1. 获取JWT Token

首先需要登录获取JWT token：

```bash
# 登录获取token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password"}'
```

### 2. Linux/Mac 测试

```bash
# 添加执行权限
chmod +x test_datacenter.sh

# 运行测试
./test_datacenter.sh http://localhost:8080 "your-jwt-token"
```

### 3. Windows 测试

```cmd
# 运行测试
test_datacenter.bat http://localhost:8080 "your-jwt-token"
```

### 4. Postman 测试

1. 打开Postman
2. 点击 Import 按钮
3. 选择 `postman_datacenter.json` 文件导入
4. 在集合变量中设置：
   - `base_url`: 服务器地址 (如 `http://localhost:8080`)
   - `token`: JWT token
5. 运行集合

### 5. Go 单元测试

```bash
# 运行单元测试
cd backend
go test ./internal/services/... -v

# 运行基准测试
go test ./internal/services/... -bench=.
```

## API 端点列表

### 实时监控

| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/api/v1/datacenter/realtime/overview` | 获取实时概览 |
| GET | `/api/v1/datacenter/realtime/inventory` | 获取实时库存状态 |
| GET | `/api/v1/datacenter/realtime/hourly-trend` | 获取小时趋势 |

### 客户分析

| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/api/v1/datacenter/customers/analysis` | 获取客户分析 |
| GET | `/api/v1/datacenter/customers/value-distribution` | 获取客户价值分布 |
| GET | `/api/v1/datacenter/customers/geography` | 获取地域分布 |
| GET | `/api/v1/datacenter/customers/city` | 获取城市分布 |
| GET | `/api/v1/datacenter/customers/repurchase` | 获取复购分析 |

### 商品分析

| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/api/v1/datacenter/products/turnover` | 获取商品动销率 |
| GET | `/api/v1/datacenter/products/inventory-level` | 获取库存水位 |
| GET | `/api/v1/datacenter/products/purchase-strategy` | 获取进货策略 |
| GET | `/api/v1/datacenter/products/low-stock` | 获取低库存商品 |
| GET | `/api/v1/datacenter/products/inventory-summary` | 获取库存汇总 |

### 对比分析

| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/api/v1/datacenter/compare/period` | 期间对比 |
| GET | `/api/v1/datacenter/compare/yoy` | 同比分析 |
| GET | `/api/v1/datacenter/compare/mom` | 环比分析 |
| GET | `/api/v1/datacenter/compare/shop` | 店铺对比 |
| GET | `/api/v1/datacenter/compare/platform` | 平台对比 |

### 预警管理

| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/api/v1/datacenter/alerts/rules` | 预警规则列表 |
| POST | `/api/v1/datacenter/alerts/rules` | 创建预警规则 |
| PUT | `/api/v1/datacenter/alerts/rules/:id` | 更新预警规则 |
| DELETE | `/api/v1/datacenter/alerts/rules/:id` | 删除预警规则 |
| POST | `/api/v1/datacenter/alerts/rules/:id/toggle` | 启用/停用规则 |
| GET | `/api/v1/datacenter/alerts/summary` | 预警汇总 |
| GET | `/api/v1/datacenter/alerts/records` | 预警记录列表 |
| POST | `/api/v1/datacenter/alerts/records/:id/handle` | 处理预警 |
| POST | `/api/v1/datacenter/alerts/records/:id/ignore` | 忽略预警 |
| POST | `/api/v1/datacenter/alerts/check` | 检查预警 |
| GET | `/api/v1/datacenter/alerts/types` | 预警类型 |
| GET | `/api/v1/datacenter/alerts/levels` | 预警级别 |

## 测试结果示例

```
========================================
    数据中心模块 API 测试
========================================
Base URL: http://localhost:8080
========================================

>>> 实时监控模块测试 <<<

测试: 获取实时概览
请求: GET /api/v1/datacenter/realtime/overview
✓ 通过 (HTTP 200)
----------------------------------------

...

========================================
    测试结果汇总
========================================
通过: 25
失败: 0
总计: 25
========================================
所有测试通过!
```
