#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   3kstory 本地端到端验证测试 (E2E Test)         ║${NC}"
echo -e "${BLUE}║   测试流程: 剧本 → 后端 → 视频任务 → mp4 URL    ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════╝${NC}\n"

# API 基础 URL
API_BASE="http://localhost:8080/api/v1"
VIDEO_SERVICE="http://localhost:8003"

# 测试用户凭证
TEST_EMAIL="test@example.com"
TEST_PASSWORD="Test@123"

# 颜色输出函数
info() { echo -e "${BLUE}ℹ️  $1${NC}"; }
success() { echo -e "${GREEN}✅ $1${NC}"; }
error() { echo -e "${RED}❌ $1${NC}"; }
warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
step() { echo -e "\n${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"; echo -e "${YELLOW}$1${NC}"; echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"; }

# 检查服务是否运行
check_service() {
    local url=$1
    local name=$2
    if curl -s "$url" &> /dev/null; then
        success "$name 已就绪"
        return 0
    else
        error "$name 未响应: $url"
        return 1
    fi
}

# 1. 检查服务健康状态
step "步骤 1: 检查服务健康状态"
check_service "$API_BASE/health" "后端服务" || exit 1
check_service "$VIDEO_SERVICE/health" "视频生成服务" || exit 1

# 2. 用户注册
step "步骤 2: 用户注册"
info "注册用户: $TEST_EMAIL"
REGISTER_RESPONSE=$(curl -s -X POST "$API_BASE/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$TEST_EMAIL\",
    \"username\": \"testuser\",
    \"password\": \"$TEST_PASSWORD\"
  }")

# 检查注册响应
if echo "$REGISTER_RESPONSE" | grep -q "error\|already exists"; then
    warning "用户已存在，跳过注册"
else
    success "用户注册成功"
fi

# 3. 用户登录
step "步骤 3: 用户登录"
info "登录用户: $TEST_EMAIL"
LOGIN_RESPONSE=$(curl -s -X POST "$API_BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$TEST_EMAIL\",
    \"password\": \"$TEST_PASSWORD\"
  }")

TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
if [ -z "$TOKEN" ]; then
    error "登录失败，无法获取 token"
    error "响应: $LOGIN_RESPONSE"
    exit 1
fi
success "登录成功，Token: ${TOKEN:0:20}..."

# 4. 创建项目
step "步骤 4: 创建项目"
info "项目类型: 短剧"
PROJECT_RESPONSE=$(curl -s -X POST "$API_BASE/projects" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "测试短剧项目",
    "description": "本地端到端验证测试项目",
    "category": "comedy",
    "target_platform": "short_video"
  }')

PROJECT_ID=$(echo "$PROJECT_RESPONSE" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
if [ -z "$PROJECT_ID" ]; then
    error "项目创建失败"
    error "响应: $PROJECT_RESPONSE"
    exit 1
fi
success "项目创建成功，ID: $PROJECT_ID"

# 5. 创建场景 (Scene)
step "步骤 5: 创建场景"
info "场景内容: AI 生成的短剧剧本"
SCENE_RESPONSE=$(curl -s -X POST "$API_BASE/projects/$PROJECT_ID/generate" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "scene_count": 1,
    "style": "comedy"
  }')

SCENE_ID=$(echo "$SCENE_RESPONSE" | grep -o '"scene_id":[0-9]*' | head -1 | cut -d':' -f2)
if [ -z "$SCENE_ID" ]; then
    # 尝试从生成的场景列表中获取
    SCENES=$(curl -s -X GET "$API_BASE/projects/$PROJECT_ID/scenes" \
      -H "Authorization: Bearer $TOKEN")
    SCENE_ID=$(echo "$SCENES" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
fi

if [ -z "$SCENE_ID" ]; then
    error "场景创建失败"
    error "响应: $SCENE_RESPONSE"
    exit 1
fi
success "场景创建成功，ID: $SCENE_ID"

# 获取场景内容用于视频生成
info "获取场景详情..."
SCENE_DETAIL=$(curl -s -X GET "$API_BASE/projects/$PROJECT_ID/scenes?id=$SCENE_ID" \
  -H "Authorization: Bearer $TOKEN")
PROMPT=$(echo "$SCENE_DETAIL" | grep -o '"script":"[^"]*' | head -1 | cut -d'"' -f4)

if [ -z "$PROMPT" ]; then
    # 使用默认提示词
    PROMPT="一个搞笑的故事：小王在办公室里做了一个有趣的事情，逗得同事们哈哈大笑。"
    warning "使用默认提示词"
fi
info "场景提示词: ${PROMPT:0:50}..."

# 6. 调用视频生成服务
step "步骤 6: 请求视频生成"
info "提示词: $PROMPT"
info "时长: 10 秒"
info "分辨率: 16:9"

VIDEO_RESPONSE=$(curl -s -X POST "$API_BASE/projects/$PROJECT_ID/generate-video" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"scene_id\": $SCENE_ID,
    \"prompt\": \"$PROMPT\",
    \"provider\": \"local\",
    \"duration\": 10,
    \"aspect_ratio\": \"16:9\"
  }")

VIDEO_ID=$(echo "$VIDEO_RESPONSE" | grep -o '"video_id":"[^"]*' | head -1 | cut -d'"' -f4)
if [ -z "$VIDEO_ID" ]; then
    error "视频生成请求失败"
    error "响应: $VIDEO_RESPONSE"
    exit 1
fi
success "视频生成请求已提交，Video ID: $VIDEO_ID"

# 7. 轮询检查视频生成状态
step "步骤 7: 等待视频生成完成"
info "轮询间隔: 2 秒，最多等待 2 分钟..."

VIDEO_URL=""
for i in {1..60}; do
    STATUS_RESPONSE=$(curl -s -X GET "$API_BASE/projects/$PROJECT_ID/video-status" \
      -H "Authorization: Bearer $TOKEN" \
      -d "{\"video_id\": \"$VIDEO_ID\"}")
    
    STATUS=$(echo "$STATUS_RESPONSE" | grep -o '"status":"[^"]*' | head -1 | cut -d'"' -f4)
    VIDEO_URL=$(echo "$STATUS_RESPONSE" | grep -o '"video_url":"[^"]*' | head -1 | cut -d'"' -f4)
    
    if [ "$STATUS" = "completed" ]; then
        success "视频生成完成！"
        success "视频 URL: $VIDEO_URL"
        break
    elif [ "$STATUS" = "failed" ]; then
        error "视频生成失败"
        ERROR_MSG=$(echo "$STATUS_RESPONSE" | grep -o '"message":"[^"]*' | head -1 | cut -d'"' -f4)
        error "错误: $ERROR_MSG"
        exit 1
    else
        printf "  进度: [$i/60] 状态: $STATUS\r"
    fi
    
    sleep 2
done

if [ -z "$VIDEO_URL" ]; then
    error "视频生成超时"
    exit 1
fi

# 8. 验证视频文件
step "步骤 8: 验证视频文件"
info "检查视频 URL: $VIDEO_URL"

# 尝试下载视频文件头部确认文件有效
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$VIDEO_URL")
if [ "$HTTP_CODE" = "200" ]; then
    success "视频文件可访问 (HTTP $HTTP_CODE)"
    
    # 获取文件大小
    FILE_SIZE=$(curl -s -I "$VIDEO_URL" | grep -i "content-length" | awk '{print $2}' | tr -d '\r')
    if [ -n "$FILE_SIZE" ]; then
        FILE_SIZE_MB=$(echo "scale=2; $FILE_SIZE / 1048576" | bc)
        success "视频文件大小: ${FILE_SIZE_MB} MB"
    fi
else
    error "视频文件访问失败 (HTTP $HTTP_CODE)"
    exit 1
fi

# 9. 完整流程总结
step "步骤 9: 测试完成总结"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}   ✅ 完整的端到端验证流程已成功完成！${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${BLUE}流程数据总结：${NC}"
echo "  用户账户:        $TEST_EMAIL"
echo "  项目 ID:        $PROJECT_ID"
echo "  场景 ID:        $SCENE_ID"
echo "  视频 ID:        $VIDEO_ID"
echo "  最终视频 URL:    $VIDEO_URL"
echo ""
echo -e "${BLUE}验证内容：${NC}"
echo "  ✅ 用户注册和登录"
echo "  ✅ 项目创建"
echo "  ✅ 场景生成"
echo "  ✅ 视频生成请求"
echo "  ✅ 视频生成完成"
echo "  ✅ 视频文件可访问"
echo ""
echo -e "${YELLOW}测试地址：${NC}"
echo "  后端 API:       $API_BASE"
echo "  视频生成服务:   $VIDEO_SERVICE"
echo ""
echo -e "${YELLOW}下一步操作：${NC}"
echo "  1. 在浏览器中打开视频 URL 进行预览"
echo "  2. 检查 .local/videos 目录查看生成的视频文件"
echo "  3. 在前端应用中测试完整的用户界面"
echo ""
success "E2E 测试验证完成！🎉"
