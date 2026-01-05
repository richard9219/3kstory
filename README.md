# 3kstory: AI 影视作品生成平台 🎬

一个专注于**AI生成影视作品**（视频、电影、短剧等）的创作平台，提供强大的AI工具集，同时也会生成优质作品。

> 从 Prompt → 完整影视作品，只需 5 分钟

---

## 📋 快速导航

| 用途 | 文档 |
|------|------|
| 🚀 **快速启动** | 见下方"5 分钟启动"章节 |
| 🔧 **后端开发** | [backend/README.md](./backend/README.md) - API 文档和开发指南 |
| 🎨 **前端开发** | [frontend/README.md](./frontend/README.md) - 前端开发指南 |
| 📚 **深度文档** | [docs/](./docs/) 文件夹 |

---

## 🎯 项目目标

**Phase 1（🔥 进行中）**：AI 影视生成 MVP + 本地 AI 模型
- Milestone 1.1：集成第三方视频生成（Runway/Pika）
- Milestone 1.2：本地部署 Qwen2.5-7B 模型
- Milestone 1.3：端到端完整工作流

**Phase 2-3**：自托管优化、性能提升、作品库扩展

详细开发里程碑请查看：
- 后端里程碑：[backend/README.md](./backend/README.md#-开发里程碑)
- 前端里程碑：[frontend/README.md](./frontend/README.md#-开发里程碑)

---

## 🏗️ 项目结构

```
3kstory/
├── README.md                          # 本文件(项目总览)
├── .gitignore                         # Git 配置
│
├── backend/                           # Go 后端
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── database/
│   │   ├── handlers/
│   │   ├── middleware/
│   │   ├── models/
│   │   ├── router/
│   │   └── services/
│   ├── Dockerfile
│   ├── docker-compose.yml
│   ├── Makefile
│   ├── go.mod & go.sum
│   ├── .env.example
│   └── server                         # ✅ 编译后的二进制 (20MB)
│
├── frontend/                          # Next.js 前端
│   ├── package.json
│   ├── tsconfig.json
│   ├── next.config.js
│   ├── tailwind.config.js
│   ├── app/
│   ├── components/
│   ├── lib/
│   └── public/
│
├── docs/                              # 技术文档
    ├── 01-竞品分析.md
    ├── 02-技术架构.md
    ├── 03-短剧创作流程.md
    ├── 04-本地千问模型部署指南.md     # NEW
    └── 05-AI视频生成原理.md           # NEW
│
└── works/                             # 作品库
    └── 01-重生马斯克/                 # 第一个作品
        └── script.md                  # 剧本
```

---

## 📊 技术栈

| 层级 | 技术 | 版本 |
|------|------|------|
| **后端** | Go + Gin | 1.21 + 1.10 |
| **前端** | Next.js + React | 14 + 18 |
| **数据库** | PostgreSQL + Redis | 14 + 7 |
| **AI 模型** | Qwen2.5-7B | 开源 |
| **视频生成** | Runway / Pika | API |
| **推理框架** | vLLM / Ollama | 最新 |
| **部署** | Docker Compose | - |

---

## 🚀 5 分钟快速启动

### 前置要求

```bash
# 检查依赖
docker --version          # Docker 20.10+
docker-compose --version  # Docker Compose 2.0+
go version                # Go 1.21+
node --version            # Node.js 18+
```

### 方式 1：完整堆栈启动（推荐）

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

### 方式 2：仅启动后端（本地开发）

```bash
cd backend

# 编译
make build

# 运行
make dev
```

### 方式 3：本地端到端验证（⭐ 推荐先用这个）

这是一个**无需 GPU**、**最小依赖**的完整验证方案，使用 ffmpeg 本地生成视频。适合快速验证整个系统链路。

```bash
# 1. 安装 ffmpeg（第一次）
brew install ffmpeg

# 2. 启动所有服务（包括后端、视频生成、数据库）
bash start-local-e2e.sh

# 3. 在另一个终端运行完整 E2E 测试
bash e2e-test.sh

# 🎉 完成！查看输出的 mp4 URL，在浏览器中预览视频
```

**完整链路验证**：
```
用户注册 → 登录 → 创建项目 → 生成场景 → 视频生成 → 获取 mp4 URL
   ✅       ✅       ✅        ✅        ✅        ✅
```

详细文档：
- [QUICKSTART.md](./QUICKSTART.md) - 30 秒快速参考
- [docs/local-e2e/START_HERE.md](./docs/local-e2e/START_HERE.md) - 本地 E2E 文档入口
- [docs/local-e2e/LOCAL_E2E_GUIDE.md](./docs/local-e2e/LOCAL_E2E_GUIDE.md) - 完整详细指南

**快速测试视频生成服务**（跳过认证流程）：
```bash
bash quick-video-test.sh
```

---

## 🔌 API 文档

详细的 API 接口文档请查看：[backend/README.md](./backend/README.md)

主要功能模块：
- 🔐 用户认证（注册、登录、用户信息）
- 📁 项目管理（创建、列表、详情、更新、删除）
- 🎬 场景生成（AI 生成场景、视频生成）
- 🚀 完整工作流（端到端影视作品生成）

---

## 📖 详细文档

### 核心功能文档

| 文档 | 内容 |
|------|------|
| [docs/01-竞品分析.md](./docs/01-竞品分析.md) | 国内外竞品分析，差异化定位 |
| [docs/02-技术架构.md](./docs/02-技术架构.md) | 系统架构、数据库设计 |
| [docs/03-短剧创作流程.md](./docs/03-短剧创作流程.md) | 用户流程、算法设计 |

### 部署和集成文档

| 文档 | 内容 |
|------|------|
| [docs/04-本地千问模型部署指南.md](./docs/04-本地千问模型部署指南.md) | **NEW** - vLLM/Ollama 部署、量化优化 |
| [docs/05-AI视频生成原理.md](./docs/05-AI视频生成原理.md) | **NEW** - 扩散模型、视频生成技术原理 |

---

## 🔧 开发指南

### 后端开发

详细的后端开发文档和命令请查看：[backend/README.md](./backend/README.md)

主要命令：
- `make build` - 编译
- `make dev` - 开发运行
- `make docker-up` - 启动容器
- `make test` - 运行测试

### 前端开发

详细的前端开发文档请查看：[frontend/README.md](./frontend/README.md)

主要命令：
- `npm run dev` - 开发服务器
- `npm run build` - 生产构建
- `npm run lint` - 代码检查

---

## 🌟 核心功能

### Milestone 1.1：第三方视频生成集成

- ✅ 集成 Runway（文本→视频，30-120 秒）
- ✅ 集成 Pika（图文→视频，高保真）
- 🔜 本地模型支持（Phase 2）

**验收标准**：720p 以上质量，≤ 3 分钟生成时间，并行处理 ≥ 10 个任务

### Milestone 1.2：本地 Qwen 模型部署

- 支持 vLLM（推荐，高性能）和 Ollama（轻量级）
- Qwen2.5-7B-Instruct（脚本生成）
- Qwen2-VL（内容审核）

**性能指标**：脚本生成 ≤ 30 秒，GPU 显存 ≤ 8GB，吞吐量 ≥ 10 req/s

详细部署步骤：见 [docs/04-本地千问模型部署指南.md](./docs/04-本地千问模型部署指南.md)

### Milestone 1.3：端到端工作流

完整工作流包括：脚本生成 → 分镜设计 → 配图生成 → 视频生成 → 内容审核 → 组装导出

**总耗时**：≤ 5 分钟

详细 API 文档：见 [backend/README.md](./backend/README.md)

---

## 📱 前端功能

### 官网

- 🎬 Hero 区域展示演示视频
- ✨ 功能展示（动画效果）
- 💬 用户案例和评价
- 💰 定价表
- 🔐 登录/注册

### 应用后台

- 📊 项目列表和仪表板
- ✏️ 影视作品编辑器（三列布局）
- 🎥 实时预览和进度监听
- 📥 视频下载和分享
- 🎬 作品库管理和展示

---

## 🛠️ 故障排除

### PostgreSQL 连接失败

```bash
cd backend
docker-compose ps          # 检查容器状态
docker-compose logs postgres  # 查看日志
docker-compose restart postgres  # 重启
```

### 后端编译错误

```bash
cd backend
make clean
go mod download
go mod tidy
make build
```

### 前端依赖问题

```bash
cd frontend
rm -rf node_modules package-lock.json
npm install --legacy-peer-deps
```

### 模型部署失败

见 [docs/04-本地千问模型部署指南.md](./docs/04-本地千问模型部署指南.md) 的故障排除章节

---

## 📊 性能指标（Phase 1）

| 指标 | 目标 | 优先级 |
|------|------|--------|
| 脚本生成 | ≤ 30s | P0 |
| 视频生成 | ≤ 180s | P0 |
| **总耗时** | **≤ 5 min** | **P0** |
| 系统可用性 | ≥ 99.5% | P0 |
| 生成质量 | ≥ 720p | P0 |
| 并发用户 | ≥ 100 | P1 |

---

## 🚀 下一步

### 立即可做

1. **启动本地开发环境**
   ```bash
   cd backend && docker-compose up -d
   cd frontend && npm install && npm run dev
   ```

2. **读相关文档**
   - [backend/README.md](./backend/README.md) - 后端 API 和开发指南
   - [frontend/README.md](./frontend/README.md) - 前端开发指南
   - [docs/04-本地千问模型部署指南.md](./docs/04-本地千问模型部署指南.md) - 部署本地模型

3. **测试 API**
   查看 [backend/README.md](./backend/README.md) 中的 API 示例

---

## 📞 联系和支持

- 📧 技术支持:[创建 GitHub Issue](https://github.com/richard9219/3kstory/issues)
- 💬 讨论和建议:[Discussions](https://github.com/richard9219/3kstory/discussions)
- 📖 完整文档：见 [docs/](./docs/) 文件夹

---

## 📄 许可证

MIT License

---

**最后更新**：2025 年 12 月
