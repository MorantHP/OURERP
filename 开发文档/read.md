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