#!/bin/bash

# Ollama 部署脚本
# 用途：快速部署 Qwen2.5-7B-Instruct 模型到本地 Ollama 服务

set -e

echo "================================"
echo "3kstory Ollama 部署脚本"
echo "================================"

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

# 启动 Ollama 服务
echo ""
echo "🚀 启动 Ollama 服务..."
docker-compose -f docker-compose-ollama.yml up -d ollama

echo ""
echo "⏳ 等待 Ollama 服务启动..."
sleep 5

# 检查服务状态
MAX_RETRIES=30
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -s http://localhost:11434/api/tags > /dev/null 2>&1; then
        echo "✅ Ollama 服务已启动！"
        break
    fi
    echo "等待中... ($((RETRY_COUNT + 1))/$MAX_RETRIES)"
    sleep 2
    RETRY_COUNT=$((RETRY_COUNT + 1))
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo "❌ Ollama 启动失败，请检查日志："
    echo "docker-compose -f docker-compose-ollama.yml logs ollama"
    exit 1
fi

# 拉取 Qwen 模型
echo ""
echo "📥 下载 Qwen2.5:7b 模型（~4.7GB，首次需要几分钟）..."
docker exec 3kstory-ollama ollama pull qwen2.5:7b

echo ""
echo "✅ 模型下载完成！"

# 测试模型
echo ""
echo "🧪 测试模型推理..."
curl -s http://localhost:11434/api/generate -d '{
  "model": "qwen2.5:7b",
  "prompt": "你好，请用一句话介绍你自己。",
  "stream": false
}' | python3 -c "import sys, json; print(json.load(sys.stdin)['response'])"

# 启动其他服务
echo ""
echo "🚀 启动 PostgreSQL 和 Redis..."
docker-compose -f docker-compose-ollama.yml up -d postgres redis

echo ""
echo "================================"
echo "✅ Ollama 部署完成！"
echo "================================"
echo ""
echo "📝 服务信息："
echo "  - Ollama API: http://localhost:11434"
echo "  - 已安装模型: qwen2.5:7b"
echo "  - PostgreSQL: localhost:5432"
echo "  - Redis: localhost:6379"
echo ""
echo "🔧 常用命令："
echo "  - 查看日志: docker-compose -f docker-compose-ollama.yml logs -f ollama"
echo "  - 停止服务: docker-compose -f docker-compose-ollama.yml down"
echo "  - 列出模型: docker exec 3kstory-ollama ollama list"
echo "  - 交互测试: docker exec -it 3kstory-ollama ollama run qwen2.5:7b"
echo ""
echo "💡 更新后端配置："
echo "  在 backend/.env 中设置:"
echo "  QWEN_API_BASE=http://localhost:11434"
echo "  使用 Ollama API 格式"
echo ""
echo "📚 API 使用示例："
echo "  curl http://localhost:11434/api/generate -d '{"
echo "    \"model\": \"qwen2.5:7b\","
echo "    \"prompt\": \"你好\","
echo "    \"stream\": false"
echo "  }'"
echo ""
