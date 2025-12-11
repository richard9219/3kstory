#!/bin/bash

# 性能测试脚本
# 用途：测试 vLLM/Ollama 的推理性能

set -e

echo "================================"
echo "3kstory 模型性能测试"
echo "================================"

# 检测服务类型
SERVICE_TYPE=""
if curl -s http://localhost:8000/health > /dev/null 2>&1; then
    SERVICE_TYPE="vllm"
    API_ENDPOINT="http://localhost:8000/v1/completions"
    echo "✅ 检测到 vLLM 服务"
elif curl -s http://localhost:11434/api/tags > /dev/null 2>&1; then
    SERVICE_TYPE="ollama"
    API_ENDPOINT="http://localhost:11434/api/generate"
    echo "✅ 检测到 Ollama 服务"
else
    echo "❌ 错误: 未检测到任何运行中的服务"
    echo "请先启动 vLLM 或 Ollama"
    exit 1
fi

# 测试提示词
PROMPTS=(
    "请生成一个都市爱情短剧的剧本大纲，包含3个场景。"
    "描述一个悬疑推理短剧的开场场景，要有紧张感。"
    "创作一个喜剧短剧的经典桥段，需要有笑点。"
)

echo ""
echo "📊 测试配置："
echo "  - 服务类型: $SERVICE_TYPE"
echo "  - 测试轮数: ${#PROMPTS[@]}"
echo "  - 最大 tokens: 500"
echo ""

# 初始化统计
TOTAL_TIME=0
SUCCESS_COUNT=0
FAIL_COUNT=0

# 测试函数
test_inference() {
    local prompt=$1
    local index=$2
    
    echo "----------------------------------------"
    echo "🧪 测试 $index/${#PROMPTS[@]}: $(echo $prompt | cut -c1-30)..."
    
    START_TIME=$(date +%s.%N)
    
    if [ "$SERVICE_TYPE" = "vllm" ]; then
        # vLLM API 调用
        RESPONSE=$(curl -s -X POST $API_ENDPOINT \
            -H "Content-Type: application/json" \
            -d "{
                \"model\": \"qwen2.5-7b\",
                \"prompt\": \"$prompt\",
                \"max_tokens\": 500,
                \"temperature\": 0.7
            }")
    else
        # Ollama API 调用
        RESPONSE=$(curl -s $API_ENDPOINT -d "{
            \"model\": \"qwen2.5:7b\",
            \"prompt\": \"$prompt\",
            \"stream\": false,
            \"options\": {
                \"num_predict\": 500,
                \"temperature\": 0.7
            }
        }")
    fi
    
    END_TIME=$(date +%s.%N)
    ELAPSED=$(echo "$END_TIME - $START_TIME" | bc)
    
    # 检查响应
    if [ -n "$RESPONSE" ] && [ "$RESPONSE" != "null" ]; then
        SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
        TOTAL_TIME=$(echo "$TOTAL_TIME + $ELAPSED" | bc)
        
        # 提取生成的文本长度
        if [ "$SERVICE_TYPE" = "vllm" ]; then
            TEXT_LENGTH=$(echo $RESPONSE | python3 -c "import sys, json; print(len(json.load(sys.stdin)['choices'][0]['text']))" 2>/dev/null || echo "N/A")
        else
            TEXT_LENGTH=$(echo $RESPONSE | python3 -c "import sys, json; print(len(json.load(sys.stdin)['response']))" 2>/dev/null || echo "N/A")
        fi
        
        printf "✅ 成功 | 耗时: %.2f 秒 | 生成长度: %s 字符\n" $ELAPSED $TEXT_LENGTH
        
        # 显示部分响应（前100个字符）
        if [ "$SERVICE_TYPE" = "vllm" ]; then
            echo "   响应: $(echo $RESPONSE | python3 -c "import sys, json; print(json.load(sys.stdin)['choices'][0]['text'][:100])" 2>/dev/null || echo "解析失败")..."
        else
            echo "   响应: $(echo $RESPONSE | python3 -c "import sys, json; print(json.load(sys.stdin)['response'][:100])" 2>/dev/null || echo "解析失败")..."
        fi
    else
        FAIL_COUNT=$((FAIL_COUNT + 1))
        echo "❌ 失败 | 耗时: $ELAPSED 秒"
        echo "   错误: 空响应或请求失败"
    fi
}

# 执行测试
echo "开始测试..."
echo ""

for i in "${!PROMPTS[@]}"; do
    test_inference "${PROMPTS[$i]}" $((i + 1))
    sleep 2  # 间隔2秒避免过载
done

# 计算平均值
if [ $SUCCESS_COUNT -gt 0 ]; then
    AVG_TIME=$(echo "scale=2; $TOTAL_TIME / $SUCCESS_COUNT" | bc)
else
    AVG_TIME=0
fi

# 输出统计结果
echo ""
echo "================================"
echo "📊 测试结果统计"
echo "================================"
echo "  总测试数: ${#PROMPTS[@]}"
echo "  成功: $SUCCESS_COUNT"
echo "  失败: $FAIL_COUNT"
echo "  总耗时: $(printf "%.2f" $TOTAL_TIME) 秒"
echo "  平均耗时: $(printf "%.2f" $AVG_TIME) 秒/次"
echo ""

# 性能评级
if (( $(echo "$AVG_TIME < 10" | bc -l) )); then
    echo "🎉 性能评级: 优秀"
elif (( $(echo "$AVG_TIME < 30" | bc -l) )); then
    echo "✅ 性能评级: 良好"
elif (( $(echo "$AVG_TIME < 60" | bc -l) )); then
    echo "⚠️  性能评级: 一般"
else
    echo "❌ 性能评级: 需要优化"
fi

echo ""
echo "💡 性能优化建议："
if [ "$SERVICE_TYPE" = "vllm" ]; then
    echo "  - 使用 INT8 量化可减少显存占用"
    echo "  - 调整 --max-model-len 参数优化吞吐量"
    echo "  - 使用 --tensor-parallel-size 启用多卡推理"
else
    echo "  - 使用 ollama run 命令时添加 --num-gpu 参数"
    echo "  - 考虑使用量化版本如 qwen2.5:7b-q4_0"
    echo "  - 增加 num_ctx 参数以支持更长上下文"
fi

echo ""
