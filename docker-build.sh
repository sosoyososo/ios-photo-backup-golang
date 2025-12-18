#!/bin/bash
# Docker 构建脚本：使用 Go 1.24.0 编译项目并输出二进制文件到当前目录

set -e

echo "=== Photo Backup Server Docker Build ==="
echo "使用 Go 1.24.0 编译 photo-backup-server 和 photo-backup-cli"
echo ""

# 清理旧的二进制文件
echo "清理旧的二进制文件..."
rm -f photo-backup-server photo-backup-cli

# 构建 Docker 镜像
echo "构建 Docker 镜像..."
docker build -t photo-backup-builder .

# 创建临时容器并复制二进制文件
echo "提取编译后的二进制文件..."

# 提取 server
docker run --rm -v "$PWD":/output photo-backup-builder sh -c 'cp /photo-backup-server /output/photo-backup-server'

# 提取 cli
docker run --rm -v "$PWD":/output photo-backup-builder sh -c 'cp /photo-backup-cli /output/photo-backup-cli'

echo ""
echo "=== 构建完成 ==="
echo "输出文件："
ls -lh photo-backup-server photo-backup-cli 2>/dev/null || echo "未找到编译后的文件"

echo ""
echo "使用说明："
echo "1. photo-backup-server - 服务器二进制文件（Linux x86_64，带 CGO 支持）"
echo "2. photo-backup-cli - CLI 工具二进制文件（Linux x86_64，带 CGO 支持）"
echo ""
echo "可以将这些文件传输到 Linux 服务器上运行"
