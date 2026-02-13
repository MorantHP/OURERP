OURERP/
├── backend/                    # Go后端
│   ├── cmd/server/            # 入口
│   ├── internal/              # 内部代码
│   │   ├── config/           # 配置
│   │   ├── middleware/       # 中间件
│   │   ├── models/           # 模型
│   │   ├── handlers/         # HTTP处理器
│   │   ├── services/         # 业务逻辑
│   │   ├── repository/       # 数据访问
│   │   └── pkg/              # 工具包
│   │       ├── utils/
│   │       └── platform/     # 平台对接
│   ├── scripts/
│   ├── deployments/
│   └── test/
├── frontend/                   # Vue前端
│   ├── src/
│   │   ├── api/              # API接口
│   │   ├── views/            # 页面
│   │   ├── components/       # 组件
│   │   ├── stores/           # 状态管理
│   │   ├── router/           # 路由
│   │   ├── utils/            # 工具
│   │   └── types/            # 类型定义
│   └── public/
├── docker/                     # Docker配置
└── docs/                       # 文档

初始化后端
# 进入后端目录
cd ~/projects/OURERP/backend

# 初始化Go模块
go mod init github.com/MorantHP/OURERP

# 安装核心依赖
go get -u github.com/gin-gonic/gin
go get -u gorm.io/gorm
go get -u gorm.io/driver/postgres
go get -u github.com/redis/go-redis/v9
go get -u github.com/golang-jwt/jwt/v5
go get -u github.com/sirupsen/logrus
go get -u github.com/spf13/viper
go get -u github.com/hibiken/asynq

# 整理依赖
go mod tidy

OURERP/
├── backend/
│   ├── cmd/
│   │   ├── server/           # 主服务
│   │   └── mock/             # 模拟数据生成器
│   │       └── main.go
│   └── internal/
│       └── mock/             # 模拟数据模块
│           ├── generator.go  # 订单生成器
│           ├── data.go       # 模拟数据
│           └── api.go        # 模拟API

# 1. 确保数据库运行
cd ~/下载/projects/OURERP
docker-compose ps

# 2. 启动后端
cd backend
go mod tidy
go run cmd/server/main.go

# 3. 新终端启动前端
cd frontend
npm run dev

# 4. 访问 http://localhost:5173
# 注册账号 -> 登录 -> 查看订单页面

cd ~/下载/projects/OURERP/backend

# 1. 安装依赖
go mod tidy

# 2. 启动后端
go run cmd/server/main.go

# 3. 新终端，生成测试数据
go run cmd/mock/main.go generate

# 4. 查看生成的订单
go run cmd/mock/main.go list

# 5. 启动实时生成（可选）
go run cmd/mock/main.go realtime

# 1. 启动所有服务
cd ~/下载/projects/OURERP
docker-compose up -d postgres redis

cd backend
go run cmd/server/main.go

# 2. 新终端生成数据
cd backend
go run cmd/mock/main.go generate

# 3. 前端查看
cd frontend
npm run dev
# 访问 http://localhost:5173
# 登录后查看订单列表，应该有100个模拟订单