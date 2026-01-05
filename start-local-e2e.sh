#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== 3kstory 本地端到端验证启动脚本 ===${NC}\n"

# 检查依赖
echo -e "${YELLOW}检查依赖...${NC}"
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go 未安装${NC}"
    exit 1
fi

if ! command -v ffmpeg &> /dev/null; then
    echo -e "${YELLOW}⚠️  ffmpeg 未安装，正在安装...${NC}"
    brew install ffmpeg
fi

if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker 未安装${NC}"
    echo "请访问 https://www.docker.com/products/docker-desktop 安装 Docker Desktop"
    exit 1
fi

# 确保 Docker daemon 已启动（macOS 下通常需要 Docker Desktop 运行）
if ! docker info >/dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  Docker daemon 未运行，尝试启动 Docker Desktop...${NC}"
    if command -v open >/dev/null 2>&1; then
        open -a Docker >/dev/null 2>&1 || true
    fi
    DOCKER_STARTUP_TIMEOUT=${DOCKER_STARTUP_TIMEOUT:-60}
    for i in $(seq 1 "$DOCKER_STARTUP_TIMEOUT"); do
        if docker info >/dev/null 2>&1; then
            echo -e "${GREEN}✅ Docker daemon 已就绪${NC}"
            break
        fi
        if [ "$i" -eq "$DOCKER_STARTUP_TIMEOUT" ]; then
            echo -e "${RED}❌ 无法连接 Docker daemon（docker.sock）${NC}"
            echo "请确认 Docker Desktop 已启动并完成初始化后重试。"
            echo "验证命令：docker info"
            exit 1
        fi
        sleep 1
    done
fi

echo -e "${GREEN}✅ 依赖检查完成${NC}\n"

# 切换到 backend 目录
cd "$(dirname "$0")/backend"

# 避免 docker-compose 对未设置变量发出 WARN（本地 E2E 默认值即可）
export JWT_SECRET="${JWT_SECRET:-dev-local-jwt-secret}"
export QWEN_API_KEY="${QWEN_API_KEY-}"

# 启动数据库和 Redis
echo -e "${YELLOW}启动 PostgreSQL 和 Redis...${NC}"
if command -v docker-compose >/dev/null 2>&1; then
    docker-compose up -d postgres redis
else
    docker compose up -d postgres redis
fi

# 等待数据库就绪
echo -e "${YELLOW}等待数据库就绪...${NC}"
for i in {1..30}; do
    if command -v docker-compose >/dev/null 2>&1; then
        POSTGRES_CID=$(docker-compose ps -q postgres)
    else
        POSTGRES_CID=$(docker compose ps -q postgres)
    fi
    if docker exec "$POSTGRES_CID" pg_isready -U postgres &> /dev/null; then
        echo -e "${GREEN}✅ PostgreSQL 已就绪${NC}"
        break
    fi
    if [ $i -eq 30 ]; then
        echo -e "${RED}❌ PostgreSQL 启动超时${NC}"
        exit 1
    fi
    sleep 1
done

# 启动后端服务
echo -e "${YELLOW}启动后端服务...${NC}"
export AI_VIDEO_SERVICE_URL="http://localhost:8003/v1/generate"
go run ./cmd/server/main.go &
BACKEND_PID=$!

# 启动本地视频生成服务
echo -e "${YELLOW}启动本地视频生成服务...${NC}"
export LOCAL_VIDEO_PORT=8003
export LOCAL_VIDEO_OUTPUT_DIR=".local/videos"
go run ./cmd/local-video-service/entry.go &
VIDEO_SERVICE_PID=$!

# 等待服务启动
echo -e "${YELLOW}等待服务启动...${NC}"
sleep 3

# 检查服务健康状态
echo -e "${YELLOW}检查服务健康状态...${NC}"
for i in {1..10}; do
    if curl -s http://localhost:8080/api/v1/health &> /dev/null; then
        echo -e "${GREEN}✅ 后端服务已就绪${NC}"
        break
    fi
    sleep 1
done

for i in {1..10}; do
    if curl -s http://localhost:8003/health &> /dev/null; then
        echo -e "${GREEN}✅ 视频生成服务已就绪${NC}"
        break
    fi
    sleep 1
done

echo ""
echo -e "${GREEN}=== 所有服务已启动 ===${NC}"
echo ""
echo -e "${YELLOW}服务地址：${NC}"
echo "  后端 API: http://localhost:8080/api/v1"
echo "  视频生成: http://localhost:8003"
echo "  PostgreSQL: localhost:5432"
echo "  Redis: localhost:6379"
echo ""
echo -e "${YELLOW}下一步：${NC}"
echo "  运行 e2e 测试脚本: ../e2e-test.sh"

echo ""
echo -e "${YELLOW}最小 curl 验证（生成并下载 mp4）：${NC}"
cat <<'EOF'
    resp=$(curl -sS -X POST http://localhost:8003/v1/generate \
        -H 'Content-Type: application/json' \
        -d '{"prompt":"Hello from local ffmpeg","duration":3,"aspect_ratio":"16:9"}')

    echo "$resp"
    video_url=$(python3 -c 'import json,sys; print(json.loads(sys.stdin.read())["video_url"])' <<<"$resp")
    curl -fL "$video_url" -o /tmp/3kstory-local.mp4
    ls -lh /tmp/3kstory-local.mp4
EOF
echo ""
echo "按 Ctrl+C 停止所有服务"

# 清理
cleanup() {
    echo -e "\n${YELLOW}正在停止服务...${NC}"
    kill $BACKEND_PID $VIDEO_SERVICE_PID 2>/dev/null || true
    if command -v docker-compose >/dev/null 2>&1; then
        docker-compose down
    else
        docker compose down
    fi
    echo -e "${GREEN}✅ 服务已停止${NC}"
}

trap cleanup EXIT
wait
