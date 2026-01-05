#!/bin/bash
# 快速视频生成测试脚本 - 用于快速验证视频生成功能
set -e

echo "🎬 快速视频生成测试"
echo "==================="

VIDEO_SERVICE="http://localhost:8003"

# 检查服务
if ! curl -s "$VIDEO_SERVICE/health" &> /dev/null; then
    echo "❌ 视频生成服务未运行 ($VIDEO_SERVICE)"
    echo "请先运行: bash start-local-e2e.sh"
    exit 1
fi

echo "✅ 视频生成服务已就绪"

# 直接调用视频生成 API
echo "📝 生成视频..."
RESPONSE=$(curl -s -X POST "$VIDEO_SERVICE/v1/generate" \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "一个快乐的程序员在敲击键盘，办公室的灯光闪闪发光",
    "duration": 10,
    "aspect_ratio": "16:9"
  }')

echo "响应: $RESPONSE"

VIDEO_ID=$(echo "$RESPONSE" | grep -o '"video_id":"[^"]*' | cut -d'"' -f4)
VIDEO_URL=$(echo "$RESPONSE" | grep -o '"video_url":"[^"]*' | cut -d'"' -f4)

if [ -z "$VIDEO_ID" ]; then
    echo "❌ 视频生成失败"
    exit 1
fi

echo ""
echo "✅ 视频生成成功！"
echo "   Video ID: $VIDEO_ID"
echo "   Video URL: $VIDEO_URL"
echo ""
echo "💾 文件位置: .local/videos/$VIDEO_ID.mp4"
echo ""
echo "📺 在浏览器中打开预览: $VIDEO_URL"
echo ""
echo "🔍 查询视频状态:"
echo "   curl $VIDEO_SERVICE/v1/generate/$VIDEO_ID"
