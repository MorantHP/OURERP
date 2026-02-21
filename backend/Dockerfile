# OURERP Backend Dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app

# 安装依赖
RUN apk add --no-cache git ca-certificates tzdata

# 复制go mod文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o ourerp-server ./cmd/server

# 最终镜像
FROM alpine:3.19

WORKDIR /app

# 安装必要工具
RUN apk add --no-cache ca-certificates tzdata

# 从builder复制二进制文件
COPY --from=builder /app/ourerp-server .
COPY --from=builder /app/docs ./docs

# 设置时区
ENV TZ=Asia/Shanghai

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 运行
ENTRYPOINT ["./ourerp-server"]
