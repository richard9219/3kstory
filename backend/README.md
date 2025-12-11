# Backend API 文档

Go + Gin 后端服务,提供 3kstory 短剧生成平台的核心 API。

---

## 🚀 快速开始

### 前置要求
- Go 1.21+
- PostgreSQL 14+
- Redis 7+
- Docker & Docker Compose（可选）

### 启动方式

**方式 1：Docker Compose（推荐）**
```bash
cd backend

# 1. 配置环境变量
cp .env.example .env
# 编辑 .env，填入 AI API 密钥（可选，本地模型无需）

# 2. 启动完整堆栈（PostgreSQL + Redis + 后端）
docker-compose up -d

# 3. 验证后端
curl http://localhost:8080/api/v1/health

# 4. 查看日志
docker-compose logs -f backend
```

**✅ 服务已就绪**：
- 📍 后端 API：http://localhost:8080/api/v1
- 📊 PostgreSQL：localhost:5432
- 💾 Redis：localhost:6379

**方式 2：本地开发**
```bash
cd backend
cp .env.example .env
make build
make dev
```

---

## 📊 技术栈

- **框架**：Gin 1.10
- **ORM**：GORM + PostgreSQL
- **缓存**：Redis
- **认证**：JWT + bcrypt
- **部署**：Docker + Docker Compose

### 为什么选择 PostgreSQL？

本项目选择 PostgreSQL 而非 MySQL 或 MongoDB 的原因：

**1. JSON/JSONB 支持**
- 项目中的 `Scene` 模型使用了 `jsonb` 类型存储 `Characters` 数组（`CharacterArray`）
- PostgreSQL 的 JSONB 提供原生 JSON 支持，支持索引和查询，性能优于 MySQL 的 JSON 类型
- 适合存储半结构化的场景数据（角色、对话等）

**2. 数据类型丰富**
- 支持数组、JSON、UUID、全文搜索等高级数据类型
- 更适合复杂的数据结构需求

**3. ACID 事务支持**
- 相比 MongoDB，PostgreSQL 提供完整的 ACID 事务支持
- 对于用户数据、项目数据等需要强一致性的场景更可靠

**4. 性能优势**
- 在复杂查询和并发场景下性能优于 MySQL
- 支持并行查询、分区表等高级特性

**5. 开源生态**
- 完全开源，社区活跃
- 与 Go 的 GORM 集成良好
- Docker 部署简单

**为什么不选 MySQL？**
- MySQL 的 JSON 支持不如 PostgreSQL 的 JSONB 强大
- 复杂查询性能相对较弱

**为什么不选 MongoDB？**
- 项目需要强一致性（用户数据、项目状态等）
- 关系型数据（用户-项目-场景）更适合用关系数据库
- PostgreSQL 的 JSONB 已经能满足半结构化数据需求

---

## 🔌 核心 API 端点

### 认证
- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录
- `GET /api/v1/users/me` - 获取用户信息
- `PUT /api/v1/users/me` - 更新用户信息

### 项目管理
- `POST /api/v1/projects` - 创建项目
- `GET /api/v1/projects` - 列表
- `GET /api/v1/projects/:id` - 详情
- `PUT /api/v1/projects/:id` - 更新
- `DELETE /api/v1/projects/:id` - 删除

### 场景生成
- `GET /api/v1/projects/:id/scenes` - 获取场景
- `POST /api/v1/projects/:id/generate-scenes` - 生成场景
- `POST /api/v1/projects/:id/generate-video` - 生成视频（Milestone 1.1）
- `POST /api/v1/projects/generate-drama` - 完整工作流（Milestone 1.3）
- `WebSocket /ws/projects/:id/progress` - 实时进度

---

## 🏗️ 项目结构

```
backend/
├── cmd/server/main.go              # 应用入口
├── internal/
│   ├── config/config.go            # 配置管理
│   ├── database/
│   │   ├── db.go                   # PostgreSQL 初始化
│   │   └── redis.go                # Redis 初始化
│   ├── models/
│   │   ├── user.go                 # 用户模型
│   │   ├── project.go              # 项目模型
│   │   └── ai_task.go              # AI 任务模型
│   ├── middleware/
│   │   ├── auth.go                 # JWT 认证
│   │   ├── cors.go                 # CORS 配置
│   │   └── logger.go               # 日志
│   ├── services/
│   │   ├── ai_service.go           # AI 集成（Qwen/Runway/Pika）
│   │   ├── video_service.go        # 视频生成（Milestone 1.1）
│   │   └── project_service.go      # 业务逻辑
│   ├── handlers/
│   │   ├── auth_handler.go         # 认证端点
│   │   └── project_handler.go      # 项目端点
│   └── router/
│       └── router.go               # 路由定义
├── Dockerfile                       # 容器镜像
├── docker-compose.yml              # 编排配置
├── Makefile                        # 编译命令
├── go.mod & go.sum                 # 依赖管理
├── .env.example                    # 环境变量模板
└── server                          # 编译后的二进制 (20MB)
```

---

## 🔧 常用命令

```bash
make build              # 编译
make dev                # 开发运行（监听变化）
make docker-up          # 启动 Docker
make docker-down        # 停止 Docker
make logs               # 查看日志
make test               # 运行测试
make migrate            # 数据库迁移
make clean              # 清理产物
```

---

## 📚 深度文档

- [docs/02-技术架构.md](../docs/02-技术架构.md) - 系统架构设计
- [docs/04-本地千问部署指南.md](../docs/04-本地千问模型部署指南.md) - 模型部署
- [docs/05-AI视频生成原理.md](../docs/05-AI视频生成原理.md) - 视频生成技术

---

## 🎯 开发里程碑

### Milestone 1.1：第三方视频生成集成 ⏳

**目标**：集成 Runway、Pika 等第三方视频生成服务

**关键功能**：
- [ ] Runway API 集成（支持文本→视频）
- [ ] Pika API 集成（支持图文→视频）
- [ ] 异步任务管理和回调机制
- [ ] 生成进度查询端点
- [ ] 视频预处理和上传到 OSS

**验收标准**：
- 可通过 API 调用第三方服务生成视频
- 视频质量 ≥ 720p
- 生成时间 ≤ 3 分钟（30 秒视频）
- 支持多个任务并行处理

**实现的 API**：
```
POST /api/v1/projects/{id}/generate-video
{
  "scene_id": "uuid",
  "script": "场景描述文本",
  "image_url": "分镜配图 URL",
  "video_provider": "runway|pika",
  "duration_sec": 30
}
```

---

### Milestone 1.2：本地 Qwen 模型部署 ⏳

**目标**：使用 vLLM/Ollama 部署开源阿里 Qwen 模型，实现本地 AI 能力

**后端需实现的变更**：
- 修改 `AIService`：`GenerateScript()` 从云 API 改为本地 LLM
- 支持动态切换：`cloud_qwen | local_qwen | local_qwen_vl`

**新增 API**：
- `GET /api/v1/ai/models` - 查询可用模型列表
- `GET /api/v1/ai/health` - 检查本地模型健康状态

**Docker Compose 配置**：
```yaml
services:
  qwen-vllm:
    image: vllm/vllm-openai:latest
    model: Qwen/Qwen2.5-7B-Instruct
    ports:
      - "8001:8000"
    gpu: true  # 需要 GPU
    
  qwen-multimodal:
    image: ollama/ollama:latest
    model: qwen2-vl  # 多模态模型用于内容审核
    ports:
      - "11434:11434"
```

**关键功能**：
- [ ] vLLM 部署和配置
- [ ] 本地 Qwen2.5-7B 脚本生成
- [ ] 本地 Qwen2-VL 内容审核
- [ ] 模型热加载和多并发支持
- [ ] 推理性能监控

**验收标准**：
- 脚本生成 ≤ 30 秒（5 场景）
- GPU 显存占用 ≤ 8GB
- 支持 ≥ 10 并发请求
- 审核准确度 ≥ 90%

**性能指标**：
| 指标 | 目标 | 测试方法 |
|------|------|--------|
| 脚本生成延迟 | ≤ 30s | 5 场景 Prompt |
| 图像审核延迟 | ≤ 5s | 1920x1080 图片 |
| GPU 显存 | ≤ 8GB | Qwen2.5-7B + vLLM |
| 吞吐量 | ≥ 10 req/s | 并发脚本请求 |

详细部署步骤：见 [docs/04-本地千问模型部署指南.md](../docs/04-本地千问模型部署指南.md)

---

### Milestone 1.3：完整工作流 ⏳

**目标**：整合所有服务，实现从 Prompt 到完整网剧的端到端流程

**工作流步骤**：
1. **脚本生成** (Qwen LLM) — 30 秒
   - Prompt → JSON 结构化脚本
   - 角色定义、场景列表、对话和动作、配景描述

2. **分镜设计** (AI 服务) — 60 秒
   - 脚本 → 场景描述
   - 角色、背景、灯光等视觉指导

3. **配图生成** (SDXL/Flux) — 120 秒
   - 场景描述 → 分镜配图
   - 自动拆分、并行生成、质量检查

4. **视频生成** (Runway/Pika) — 180 秒
   - 配图 + 脚本 → 带声音的视频
   - 图片转视频、文字转语音 (TTS)、背景音乐合成、字幕合成

5. **内容审核** (Qwen2-VL) — 30 秒
   - 视频 → 质量评分
   - 敏感内容检测、脸部识别、文本 OCR 审核、音频内容审核

6. **组装导出** — 60 秒
   - 各部分 → 完整 MP4
   - 场景拼接、时间轴对齐、导出多种格式、上传 CDN

**总耗时**：≤ 5 分钟

**关键功能**：
- [ ] ProjectGenerationOrchestrator 服务
- [ ] 工作流状态机管理
- [ ] 错误重试和降级处理
- [ ] WebSocket 实时进度推送
- [ ] 各阶段性结果缓存
- [ ] 多媒体资产合并导出

**验收标准**：
- ✅ 端到端工作流完成
- ✅ 生成时间 ≤ 5 分钟（完整网剧）
- ✅ 支持中断和恢复
- ✅ 错误处理 ≥ 99% 可用性
- ✅ WebSocket 实时进度更新

**新 API 端点**：
```
POST /api/v1/projects/generate-drama
{
  "title": "我的网剧",
  "description": "一个关于程序员的爱情故事",
  "genre": "爱情",
  "episodes": 5,
  "scenes_per_episode": 3,
  "duration_sec": 120,
  "prompt": "北京，2024 年，一个 996 的程序员遇见...",
  "video_provider": "runway",
  "use_local_llm": true
}

WebSocket /ws/projects/{project_id}/progress
```

---

## 📖 API 详细示例

### 用户认证

#### 注册用户
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "Test@12345"
  }'
```

**响应示例**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": "uuid-xxx",
    "email": "test@example.com",
    "token": "eyJhbGc..."
  }
}
```

#### 用户登录
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Test@12345"
  }'
```

#### 获取用户信息
```bash
TOKEN="your_jwt_token"
curl -X GET http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer $TOKEN"
```

---

### 项目管理

#### 创建项目
```bash
TOKEN="your_jwt_token"
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "我的第一部网剧",
    "description": "一个程序员的爱情故事",
    "genre": "爱情",
    "target_episodes": 5,
    "target_duration_sec": 120
  }'
```

#### 获取项目列表
```bash
curl -X GET http://localhost:8080/api/v1/projects \
  -H "Authorization: Bearer $TOKEN"
```

#### 获取项目详情
```bash
PROJECT_ID="xxx"
curl -X GET http://localhost:8080/api/v1/projects/$PROJECT_ID \
  -H "Authorization: Bearer $TOKEN"
```

---

### 场景生成

#### 生成场景
```bash
PROJECT_ID="xxx"
curl -X POST http://localhost:8080/api/v1/projects/$PROJECT_ID/generate-scenes \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "北京 2024，一个996程序员遇见了设计部的同事..."
  }'
```

#### 生成视频（Milestone 1.1）
```bash
PROJECT_ID="xxx"
curl -X POST http://localhost:8080/api/v1/projects/$PROJECT_ID/generate-video \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "scene_id": "scene-1",
    "script": "北京办公室，程序员正在工作",
    "image_url": "https://example.com/storyboard.jpg",
    "video_provider": "runway",
    "duration_sec": 30
  }'
```

**响应示例**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "task_id": "uuid-xxx",
    "status": "pending",
    "video_url": null,
    "created_at": "2024-12-11T..."
  }
}
```

#### 完整工作流生成（Milestone 1.3）
```bash
curl -X POST http://localhost:8080/api/v1/projects/generate-drama \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "我的网剧",
    "description": "一个关于程序员的爱情故事",
    "genre": "爱情",
    "episodes": 5,
    "scenes_per_episode": 3,
    "duration_sec": 120,
    "prompt": "北京，2024 年，一个 996 的程序员遇见...",
    "video_provider": "runway",
    "use_local_llm": true
  }'
```

#### WebSocket 实时进度（Milestone 1.3）
```javascript
const ws = new WebSocket('ws://localhost:8080/ws/projects/{project_id}/progress');
ws.onmessage = (event) => {
  const progress = JSON.parse(event.data);
  console.log('进度更新:', progress);
};
```

---

### 错误处理

所有 API 遵循统一的错误响应格式：

```json
{
  "code": 400,
  "message": "错误描述",
  "data": null
}
```

常见错误码：
- `400` - 请求参数错误
- `401` - 未授权（需要登录）
- `403` - 无权限
- `404` - 资源不存在
- `500` - 服务器内部错误

---

## 🌍 环境变量配置

创建 `.env` 文件（参考 `.env.example`）：

```env
# 数据库
DATABASE_URL=postgres://postgres:password@postgres:5432/3k_vedio
REDIS_URL=redis://redis:6379

# AI 服务
QWEN_API_KEY=sk-xxx
QWEN_MODEL=qwen-max-latest

# 应用
JWT_SECRET=your-super-secret-jwt-key-change-in-production
SERVER_PORT=8080
LOG_LEVEL=debug

# OSS 文件存储
OSS_ENDPOINT=https://oss-cn-hangzhou.aliyuncs.com
OSS_ACCESS_KEY=xxx
OSS_SECRET_KEY=xxx
OSS_BUCKET=3kstory
```

---

## 🛠️ 故障排除

### PostgreSQL 连接失败
```bash
cd backend
docker-compose ps          # 检查容器状态
docker-compose logs postgres  # 查看日志
docker-compose restart postgres  # 重启
```

### Redis 连接失败
```bash
# 测试 Redis 连接
docker-compose exec redis redis-cli ping
# 应返回 PONG
```

### 后端编译错误
```bash
cd backend
make clean
go mod download
go mod tidy
make build
```

### 模型部署失败
见 [docs/04-本地千问模型部署指南.md](../docs/04-本地千问模型部署指南.md) 的故障排除章节

---

## 📊 性能指标（Phase 1）

| 指标 | 目标 | 优先级 |
|------|------|--------|
| 脚本生成 | ≤ 30s | P0 |
| 分镜设计 | ≤ 60s | P1 |
| 配图生成 | ≤ 120s | P1 |
| 视频生成 | ≤ 180s | P0 |
| 内容审核 | ≤ 30s | P2 |
| **总耗时** | **≤ 5 min** | **P0** |
| 系统可用性 | ≥ 99.5% | P0 |
| 生成质量 | ≥ 720p | P0 |
| 并发用户 | ≥ 100 | P1 |

---

## 🚀 未来规划

### Phase 2: 自托管 LLM 优化（Q2 2024）
- 部署专用 GPU 集群（A100）
- 模型量化和加速（INT4 / INT8）
- vLLM 分布式推理
- 实时性能监控和自动扩缩容

### Phase 3: 多模态审核 + 加速（Q3 2024）
- Qwen2-VL 多模态审核优化
- CDN 全球加速
- OSS 智能转码
- 用户审核反馈系统

---

**更新日期**：2024 年 12 月
