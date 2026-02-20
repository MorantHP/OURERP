// Package docs OURERP API 文档
package docs

// @title OURERP API
// @version 1.0
// @description 电商ERP系统API文档，支持多租户、多平台订单管理、库存管理、财务管理等功能

// @contact.name API Support
// @contact.email support@ourerp.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT认证令牌，格式: Bearer {token}

// @securityDefinitions.apikey TenantID
// @in header
// @name X-Tenant-ID
// @description 租户ID，用于多租户数据隔离

// @tag.name 认证
// @tag.description 用户认证相关接口

// @tag.name 用户管理
// @tag.description 用户增删改查（仅Root用户）

// @tag.name 租户管理
// @tag.description 租户（账套）管理接口

// @tag.name 店铺管理
// @tag.description 店铺管理接口

// @tag.name 订单管理
// @tag.description 订单管理接口

// @tag.name 商品管理
// @tag.description 商品管理接口

// @tag.name 库存管理
// @tag.description 库存管理接口

// @tag.name 财务管理
// @tag.description 财务管理接口

// @tag.name 数据中心
// @tag.description 数据分析和监控接口

// @tag.name 权限管理
// @tag.description 角色和权限管理接口
