# OURERP - 企业资源规划系统

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/Vue.js-3.x-4FC08D?style=flat&logo=vue.js" alt="Vue Version">
  <img src="https://img.shields.io/badge/TypeScript-5.x-3178C6?style=flat&logo=typescript" alt="TypeScript">
  <img src="https://img.shields.io/badge/License-MIT-green?style=flat" alt="License">
</p>

## 📖 项目简介

OURERP 是一个基于 Go + Vue.js 构建的现代化企业资源规划系统，旨在为企业提供灵活、高效的业务管理解决方案。

### ✨ 核心特性

- 🔐 **用户权限管理** - 基于RBAC的细粒度权限控制
- 🏢 **组织架构** - 多级组织结构支持
- 📋 **审批流程** - 可配置的工作流引擎
- 📊 **数据字典** - 灵活的数据配置管理
- 📝 **日志审计** - 完整的操作追踪记录
- 🐳 **容器化部署** - Docker 一键部署

## 🏗️ 技术架构

```
┌─────────────────────────────────────────┐
│              Frontend (Vue.js)          │
│  TypeScript + Vite + Pinia + Vue Router │
└──────────────────┬──────────────────────┘
                   │ REST API / WebSocket
┌──────────────────▼──────────────────────┐
│              Backend (Go)                │
│   Gin/Fiber + GORM + Redis + JWT        │
└──────────────────┬──────────────────────┘
                   │
┌──────────────────▼──────────────────────┐
│           Database (PostgreSQL)         │
└─────────────────────────────────────────┘
```

### 后端技术栈

- **语言**: Go 1.21+
- **Web框架**: Gin / Fiber
- **ORM**: GORM
- **认证**: JWT
- **缓存**: Redis
- **数据库**: PostgreSQL / MySQL

### 前端技术栈

- **框架**: Vue.js 3.x
- **构建工具**: Vite
- **语言**: TypeScript
- **状态管理**: Pinia
- **路由**: Vue Router
- **UI组件**: Element Plus / Ant Design Vue
- **测试**: Vitest + Playwright

## 🚀 快速开始

### 环境要求

- Go 1.21+
- Node.js 18+
- Docker & Docker Compose
- PostgreSQL 15+ (或使用 Docker)

### 使用 Docker Compose (推荐)

```bash
# 克隆项目
git clone https://github.com/MorantHP/OURERP.git
cd OURERP

# 启动所有服务
docker-compose up -d

# 访问应用
# 前端: http://localhost:3000
# 后端: http://localhost:8080
```

### 本地开发

#### 后端

```bash
cd backend

# 安装依赖
go mod download

# 复制环境变量
cp .env.example .env

# 启动开发服务 (带热重载)
air

# 运行测试
go test ./...
```

#### 前端

```bash
cd frontend

# 安装依赖
npm install

# 复制环境变量
cp .env.example .env

# 启动开发服务
npm run dev

# 运行测试
npm run test

# E2E 测试
npm run test:e2e
```

## 📁 项目结构

```
OURERP/
├── backend/                # Go 后端服务
│   ├── cmd/               # 应用入口
│   ├── internal/          # 内部业务逻辑
│   │   ├── api/          # API 路由
│   │   ├── service/      # 业务逻辑
│   │   ├── repository/   # 数据访问
│   │   ├── model/        # 数据模型
│   │   └── middleware/   # 中间件
│   ├── docs/             # API 文档
│   ├── scripts/          # 脚本工具
│   └── Dockerfile        # 生产镜像
│
├── frontend/              # Vue.js 前端应用
│   ├── src/
│   │   ├── api/          # API 请求
│   │   ├── components/   # 组件
│   │   ├── views/        # 页面
│   │   ├── stores/       # 状态管理
│   │   └── router/       # 路由配置
│   ├── e2e/              # E2E 测试
│   └── public/           # 静态资源
│
├── 开发文档/              # 中文开发文档
├── docker-compose.yml    # Docker 编排
└── README.md
```

## 🔧 配置说明

### 环境变量

#### 后端 (.env)

```env
# 服务配置
SERVER_PORT=8080
SERVER_MODE=debug

# 数据库
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=ourerp

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRE=24h
```

#### 前端 (.env)

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_APP_TITLE=OURERP
```

## 🧪 测试

```bash
# 后端单元测试
cd backend && go test ./... -cover

# 前端单元测试
cd frontend && npm run test

# E2E 测试
cd frontend && npm run test:e2e
```

## 📚 API 文档

启动后端服务后访问:
- Swagger UI: `http://localhost:8080/swagger/index.html`

## 🤝 贡献指南

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 📄 许可证

本项目基于 [MIT](LICENSE) 许可证开源。

## 📮 联系方式

- 项目地址: [https://github.com/MorantHP/OURERP](https://github.com/MorantHP/OURERP)
- 问题反馈: [Issues](https://github.com/MorantHP/OURERP/issues)

---

<p align="center">
  Made with ❤️ by OURERP Team
</p>
