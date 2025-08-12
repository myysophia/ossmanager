#!/bin/bash

# 构建前端脚本

set -e

echo "开始构建前端..."

# 进入前端目录
cd frontend

# 安装依赖（如果需要）
if [ ! -d "node_modules" ]; then
    echo "安装前端依赖..."
    npm install
fi

# 构建前端
echo "构建前端应用..."
npm run build

# 回到根目录
cd ..

# 创建目标目录
echo "创建web/build目录..."
mkdir -p web/build

# 清空并复制构建文件
echo "复制前端构建文件到web/build..."
rm -rf web/build/*
cp -r frontend/out/* web/build/

echo "前端构建完成！"
echo "文件已复制到: web/build/"

# 验证文件
echo "验证构建文件："
ls -la web/build/ | head -10
