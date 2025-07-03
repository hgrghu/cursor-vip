#!/bin/bash

# Cursor VIP Windows EXE 构建脚本
# 生成包含图标和清单的Windows可执行文件

set -e

echo "🚀 开始构建 Cursor VIP Windows EXE..."

# 版本信息
VERSION="2.5.8"
BUILD_TIME=$(date -u +"%Y-%m-%d %H:%M:%S UTC")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 创建构建目录
mkdir -p build/windows

# 清理旧的资源文件
rm -f rsrc.syso

echo "📦 生成Windows资源文件..."
# 检查rsrc工具是否安装
if ! command -v ~/go/bin/rsrc &> /dev/null; then
    echo "⚙️  安装rsrc工具..."
    go install github.com/akavel/rsrc@latest
fi

# 生成包含图标和清单的资源文件
~/go/bin/rsrc -manifest rsrc.manifest -ico rsrc.ico -o rsrc.syso

# 构建参数
LDFLAGS="-s -w -X 'main.version=${VERSION}' -X 'main.buildTime=${BUILD_TIME}' -X 'main.gitCommit=${GIT_COMMIT}'"

echo "🔨 构建 Windows 64位版本..."
GOOS=windows GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o build/windows/cursor-vip-v${VERSION}-windows-amd64.exe .

echo "🔨 构建 Windows 32位版本..."
GOOS=windows GOARCH=386 go build -ldflags="${LDFLAGS}" -o build/windows/cursor-vip-v${VERSION}-windows-386.exe .

# 生成便于使用的简化文件名
cp build/windows/cursor-vip-v${VERSION}-windows-amd64.exe build/windows/cursor-vip.exe
cp build/windows/cursor-vip-v${VERSION}-windows-386.exe build/windows/cursor-vip-x86.exe

# 清理临时文件
rm -f rsrc.syso

echo "✅ 构建完成！生成的文件："
ls -lh build/windows/

echo "📁 文件说明："
echo "  cursor-vip.exe                    - 64位版本（推荐）"
echo "  cursor-vip-x86.exe               - 32位版本"
echo "  cursor-vip-v${VERSION}-windows-amd64.exe - 64位完整版本号"
echo "  cursor-vip-v${VERSION}-windows-386.exe   - 32位完整版本号"

echo ""
echo "🎉 Windows EXE 构建成功！"
echo "📋 使用说明："
echo "1. 下载适合您系统的版本（大多数用户选择64位版本）"
echo "2. 双击运行 cursor-vip.exe"
echo "3. 按照提示操作即可享受免费的Cursor VIP功能"