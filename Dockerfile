# 多阶段构建 Dockerfile
# 使用 Go 1.22 编译项目并输出二进制文件到当前目录

FROM golang:1.22 AS builder

# 设置 Go 模块代理
ENV GOPROXY='https://goproxy.cn'

# 设置工作目录
WORKDIR /app

# 复制 go mod 和 sum 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源码
COPY . .

# 编译 server 二进制文件（Linux x86_64，带 CGO 支持）
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o photo-backup-server cmd/server/main.go

# 编译 CLI 二进制文件（Linux x86_64，带 CGO 支持）
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o photo-backup-cli cmd/cli/main.go

# 创建输出阶段
FROM alpine:latest AS output

# 从 builder 阶段复制编译好的二进制文件
COPY --from=builder /app/photo-backup-server /photo-backup-server
COPY --from=builder /app/photo-backup-cli /photo-backup-cli

# 设置入口点（可选）
CMD ["sh"]
