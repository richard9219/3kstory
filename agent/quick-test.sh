#!/bin/bash

# 快速测试脚本
# 用途：测试本地模型服务是否正常工作

echo "================================"
echo "3kstory 模型快速测试"
echo "================================"
echo ""

# 测试 vLLM
if curl -s http://localhost:8000/health > /dev/null 2>&1; then
    echo "✅ vLLM 服务运行中"
    echo ""
    echo "🧪 测试 vLLM API..."
    curl -X POST http://localhost:8000/v1/completions \
        -H "Content-Type: application/json" \
        -d '{
            "model": "qwen2.5-7b",
            "prompt": "你好，请用一句话介绍你自己。",
            "max_tokens": 50,
            "temperature": 0.7
        }' | python3 -m json.tool
    echo ""
    echo "✅ vLLM 测试完成"
    exit 0
fi

# 测试 Ollama
if curl -s http://localhost:11434/api/tags > /dev/null 2>&1; then
    echo "✅ Ollama 服务运行中"
    echo ""
    echo "🧪 测试 Ollama API..."
    curl http://localhost:11434/api/generate -d '{
        "model": "qwen2.5:7b",
        "prompt": "你好，请用一句话介绍你自己。",
        "stream": false
    }' | python3 -m json.tool
    echo ""
    echo "✅ Ollama 测试完成"
    exit 0
fi

echo "❌ 未检测到运行中的服务"
echo ""
echo "请先启动服务："
echo "  vLLM:   ./deploy-vllm.sh"
echo "  Ollama: ./deploy-ollama.sh"
exit 1
