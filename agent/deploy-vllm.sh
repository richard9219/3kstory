#!/bin/bash

# vLLM 部署脚本
# 用途：快速部署 Qwen2.5-7B-Instruct 模型到本地 vLLM 服务

set -e

echo "================================"
echo "3kstory vLLM 部署脚本"
echo "================================"

# 检查 GPU
if ! command -v nvidia-smi &> /dev/null; then
    echo "❌ 错误: 未检测到 NVIDIA GPU 驱动"
    echo "请确保已安装 NVIDIA 驱动和 CUDA"
    exit 1
fi

echo "✅ GPU 检测通过"
nvidia-smi --query-gpu=name,memory.total --format=csv,noheader

# 检查 Docker
if ! command -v docker &> /dev/null; then
    echo "❌ 错误: 未安装 Docker"
    echo "请先安装 Docker: https://docs.docker.com/get-docker/"
    exit 1
fi

echo "✅ Docker 检测通过"

# 检查 Docker Compose
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "❌ 错误: 未安装 Docker Compose"
    exit 1
fi

echo "✅ Docker Compose 检测通过"

# 创建模型缓存目录
echo ""
echo "📁 创建模型缓存目录..."
mkdir -p ./models

# 启动服务
echo ""
echo "🚀 启动 vLLM 服务..."
echo "首次启动会下载模型 (~15GB)，请耐心等待..."
echo ""

docker-compose -f docker-compose-vllm.yml up -d

echo ""
echo "⏳ 等待服务启动..."
sleep 10

# 检查服务状态
echo ""
echo "📊 服务状态："
docker-compose -f docker-compose-vllm.yml ps

# 等待 vLLM 就绪
echo ""
echo "⏳ 等待 vLLM 模型加载完成（可能需要 2-5 分钟）..."
MAX_RETRIES=30
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -s http://localhost:8000/health > /dev/null 2>&1; then
        echo "✅ vLLM 服务已就绪！"
        break
    fi
    echo "等待中... ($((RETRY_COUNT + 1))/$MAX_RETRIES)"
    sleep 10
    RETRY_COUNT=$((RETRY_COUNT + 1))
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo "❌ vLLM 启动超时，请检查日志："
    echo "docker-compose -f docker-compose-vllm.yml logs vllm"
    exit 1
fi

# 测试 API
echo ""
echo "🧪 测试 vLLM API..."
curl -s -X POST http://localhost:8000/v1/completions \
    -H "Content-Type: application/json" \
    -d '{
        "model": "qwen2.5-7b",
        "prompt": "你好，请介绍一下你自己。",
        "max_tokens": 100,
        "temperature": 0.7
    }' | python3 -m json.tool

echo ""
echo "================================"
echo "✅ vLLM 部署完成！"
echo "================================"
echo ""
echo "📝 服务信息："
echo "  - API 端点: http://localhost:8000"
echo "  - 模型名称: qwen2.5-7b"
echo "  - 健康检查: http://localhost:8000/health"
echo "  - API 文档: http://localhost:8000/docs"
echo ""
echo "🔧 常用命令："
echo "  - 查看日志: docker-compose -f docker-compose-vllm.yml logs -f vllm"
echo "  - 停止服务: docker-compose -f docker-compose-vllm.yml down"
echo "  - 重启服务: docker-compose -f docker-compose-vllm.yml restart"
echo ""
echo "💡 更新后端配置："
echo "  在 backend/.env 中设置:"
echo "  QWEN_API_BASE=http://localhost:8000/v1"
echo "  QWEN_API_KEY=token-abc123"
echo ""
