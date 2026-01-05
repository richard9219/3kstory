# 本地端到端验证指南

## 概述

这个指南将指导你如何在 Mac 本地无需 GPU 的情况下，完整地验证 3kstory 从剧本生成到视频输出的整个链路。

## 快速开始（5 分钟）

### 第 1 步：安装依赖

```bash
# 1. 安装 ffmpeg（用于本地视频生成）
brew install ffmpeg

# 2. 验证安装
ffmpeg -version
```

### 第 2 步：启动服务（3 分钟）

```bash
# 进入项目根目录
cd /path/to/3kstory

# 启动所有服务（后端 + 视频服务 + 数据库）
bash start-local-e2e.sh

# 你会看到：
# ✅ PostgreSQL 已就绪
# ✅ Redis 已就绪  
# ✅ 后端服务已就绪 (localhost:8080)
# ✅ 视频生成服务已就绪 (localhost:8003)
```

**不要关闭此终端！** 保持服务运行。

### 第 3 步：运行 E2E 测试（2 分钟）

在另一个终端中：

```bash
cd /path/to/3kstory
bash e2e-test.sh
```

你将看到完整的测试流程：

```
╔════════════════════════════════════════════════════╗
║   3kstory 本地端到端验证测试 (E2E Test)         ║
║   测试流程: 剧本 → 后端 → 视频任务 → mp4 URL    ║
╚════════════════════════════════════════════════════╝

✅ 步骤 1: 检查服务健康状态
✅ 步骤 2: 用户注册
✅ 步骤 3: 用户登录  
✅ 步骤 4: 创建项目
✅ 步骤 5: 创建场景
✅ 步骤 6: 请求视频生成
✅ 步骤 7: 等待视频生成完成
✅ 步骤 8: 验证视频文件

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   ✅ 完整的端到端验证流程已成功完成！
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

流程数据总结：
  用户账户:        test@example.com
  项目 ID:        123
  场景 ID:        456
  视频 ID:        abc12345def
  最终视频 URL:    http://localhost:8003/files/abc12345def.mp4

验证内容：
  ✅ 用户注册和登录
  ✅ 项目创建
  ✅ 场景生成
  ✅ 视频生成请求
  ✅ 视频生成完成
  ✅ 视频文件可访问
```

## 完整流程说明

### 整体架构

```
┌──────────────────────────────────────────────────────────┐
│                     E2E 验证流程                         │
└──────────────────────────────────────────────────────────┘
                            │
          ┌─────────────────┼─────────────────┐
          │                 │                 │
      ┌───▼──┐          ┌───▼──┐         ┌───▼──┐
      │ 用户 │          │ 后端 │         │ 视频 │
      │ 认证 │          │ API  │         │ 生成 │
      └───┬──┘          └───┬──┘         └───┬──┘
          │                 │                 │
          │  1. 注册/登录   │                 │
          ├────────────────►│                 │
          │                 │                 │
          │  2. 创建项目    │                 │
          ├────────────────►│                 │
          │                 │                 │
          │  3. 生成场景    │                 │
          ├────────────────►│                 │
          │                 │                 │
          │  4. 请求视频    │                 │
          ├────────────────►│  5. 调用生成   │
          │                 ├────────────────►│
          │                 │                 │
          │                 │  6. 返回 URL   │
          │                 │◄────────────────┤
          │  7. 查询视频    │                 │
          ├────────────────►│                 │
          │                 │ 8. 返回 URL    │
          │◄────────────────┤                 │
```

### 各步骤详解

#### 步骤 1-3：用户认证
- **注册**：创建测试账户 `test@example.com`
- **登录**：获取 JWT token
- **目的**：验证后端认证系统正常工作

#### 步骤 4：创建项目
- **API**：`POST /api/v1/projects`
- **参数**：项目名称、描述、分类
- **输出**：项目 ID
- **目的**：验证项目管理功能

#### 步骤 5：生成场景
- **API**：`POST /api/v1/projects/:id/generate`
- **逻辑**：AI 生成短剧脚本（本地或云）
- **输出**：场景列表和脚本内容
- **目的**：验证 AI 集成

#### 步骤 6：请求视频生成
- **API**：`POST /api/v1/projects/:id/generate-video`
- **提供者**：使用 `local` 提供者（无需 GPU）
- **参数**：
  - `prompt`：生成的脚本
  - `duration`：10 秒（可配置）
  - `aspect_ratio`：16:9（可选 9:16）
- **输出**：视频 ID
- **目的**：验证视频生成 API

#### 步骤 7：等待完成
- **轮询**：每 2 秒查询一次视频状态
- **超时**：120 秒
- **状态**：`processing` → `completed`
- **输出**：最终视频 URL
- **目的**：验证异步任务管理

#### 步骤 8：验证视频
- **访问**：HTTP GET 视频 URL
- **检查**：文件大小、MIME 类型
- **下载**：可选，保存本地预览
- **目的**：验证视频文件生成和可访问性

## 本地视频生成技术

### ffmpeg 的作用

本服务使用 `ffmpeg` 快速生成视频，无需 GPU：

1. **纯文字模式**（默认）
   ```
   黑色背景 (1280x720 或 720x1280)
   + 白色文字（提示词）
   + 自定义字体（如有可用）
   = 10 秒 MP4 视频
   ```

2. **图片模式**（如提供 image_url）
   ```
   静态图片
   + 循环播放 10 秒
   + 叠加文字
   = 10 秒 MP4 视频
   ```

### 生成过程

```bash
ffmpeg \
  -f lavfi \
  -i "color=c=black:s=1280x720:d=10" \  # 黑色背景 10 秒
  -vf "drawtext=textfile=prompt.txt:fontcolor=white:fontsize=36" \  # 绘制文字
  -r 30 \  # 帧率 30fps
  -y output.mp4  # 输出文件
```

**性能**：
- 生成时间：< 5 秒
- 文件大小：10-50 MB（取决于分辨率）
- CPU 占用：低（无需 GPU）

## 文件位置

```
3kstory/
├── start-local-e2e.sh          # 启动脚本（启动所有服务）
├── e2e-test.sh                 # 测试脚本（执行完整 E2E 验证）
├── backend/
│   ├── cmd/local-video-service/
│   │   └── main.go             # 视频生成服务实现
│   ├── internal/services/
│   │   └── video_service.go    # 视频服务 API
│   ├── internal/handlers/
│   │   └── video_handler.go    # 视频处理端点
│   └── .local/videos/          # ⬅️ 生成的 mp4 文件存储在这里
└── README.md                   # 这个文档
```

## 故障排除

### 问题：ffmpeg not found

**解决方案**：
```bash
brew install ffmpeg
# 验证安装
ffmpeg -version
```

### 问题：Port 8080 or 8003 already in use

**解决方案**：
```bash
# 查找占用进程
lsof -i :8080
lsof -i :8003

# 关闭进程
kill -9 <PID>

# 或重新启动脚本
bash start-local-e2e.sh
```

### 问题：PostgreSQL connection failed

**解决方案**：
```bash
# 检查 Docker
docker-compose ps

# 重启数据库
docker-compose restart postgres

# 查看日志
docker-compose logs postgres
```

### 问题：测试脚本中的 HTTP 错误

**解决方案**：
```bash
# 检查服务是否运行
curl http://localhost:8080/api/v1/health
curl http://localhost:8003/health

# 查看服务日志
# 在启动脚本的终端中查看错误输出
```

## 环境变量配置

如需自定义，编辑 `start-local-e2e.sh`：

```bash
# 后端服务
export AI_VIDEO_SERVICE_URL="http://localhost:8003/v1/generate"

# 视频生成服务
export LOCAL_VIDEO_PORT=8003
export LOCAL_VIDEO_OUTPUT_DIR=".local/videos"
export LOCAL_VIDEO_PUBLIC_BASE="http://localhost:8003"
```

## 下一步

### 1. 预览生成的视频
```bash
# 查看生成的视频文件
ls -lh .local/videos/

# 在浏览器中打开 URL（从 e2e-test.sh 输出获得）
open "http://localhost:8003/files/{video_id}.mp4"
```

### 2. 自定义测试
编辑 `e2e-test.sh`，修改：
- `PROMPT`：改变生成的视频内容
- `DURATION`：调整视频时长（1-60 秒）
- `ASPECT_RATIO`：改为 "9:16" 进行竖屏测试

### 3. 集成 GPU 视频生成
当需要更高质量的视频时：
1. 配置 Runway 或 Pika API 密钥
2. 在 `e2e-test.sh` 中改为 `"provider": "runway"`
3. 重新运行测试

### 4. 部署到生产环境
- 使用 Docker 部署后端
- 部署到云服务（AWS、Azure、Aliyun）
- 配置 CDN 加速

## 性能指标

| 指标 | 实际值 | 目标值 |
|------|--------|--------|
| 视频生成时间 | < 5 秒 | < 10 秒 |
| 文件大小 | 20-50 MB | 无限制 |
| CPU 占用 | 低 | < 50% |
| 生成并发数 | 5-10 | ≥ 10 |
| 可用性 | 100% | ≥ 99.5% |

## 技术细节

### 后端组件
- **语言**：Go 1.21+
- **框架**：Gin
- **数据库**：PostgreSQL
- **缓存**：Redis
- **认证**：JWT + bcrypt

### 视频生成组件
- **工具**：ffmpeg
- **输出格式**：MP4 (H.264 video, AAC audio)
- **支持格式**：16:9, 9:16
- **帧率**：30 fps
- **比特率**：自适应（2-5 Mbps）

### API 端点
```
POST /api/v1/auth/register         # 用户注册
POST /api/v1/auth/login            # 用户登录
POST /api/v1/projects              # 创建项目
GET  /api/v1/projects/:id          # 获取项目
POST /api/v1/projects/:id/generate # 生成场景
POST /api/v1/projects/:id/generate-video  # 生成视频
GET  /api/v1/projects/:id/video-status    # 查询视频状态

POST http://localhost:8003/v1/generate    # 视频生成服务 API
GET  http://localhost:8003/v1/generate/:id  # 查询生成状态
```

## 常见问题 (FAQ)

**Q：为什么使用本地 ffmpeg 而不是云 API？**
A：这个本地方案可以快速验证整个系统流程，无需:
- GPU 硬件
- 云 API 密钥
- 互联网连接（完全离线）
- 生成费用

**Q：本地生成的视频质量如何？**
A：本地方案适合测试流程，质量是基础级别。生产环境应使用 Runway/Pika 获得高质量视频。

**Q：可以并行生成多个视频吗？**
A：可以，ffmpeg 支持多进程，后端也支持异步任务队列。

**Q：如何修改视频时长和分辨率？**
A：编辑 `e2e-test.sh`：
```bash
# 改为 30 秒、9:16 竖屏
"duration": 30,
"aspect_ratio": "9:16"
```

**Q：生成的视频存储在哪里？**
A：`.local/videos/` 目录（会自动创建）

## 许可证

MIT License - 见项目根目录 LICENSE 文件

## 联系方式

- GitHub Issues：[3kstory/issues](https://github.com/your-repo/issues)
- 技术文档：[docs/](../docs/)
