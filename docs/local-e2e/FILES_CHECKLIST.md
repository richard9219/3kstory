# 本地 E2E 验证 - 文件清单和使用指南

## 📦 创建的文件清单

### 核心脚本

| 文件 | 说明 | 用途 |
|------|------|------|
| `start-local-e2e.sh` | 启动脚本 | 一键启动所有服务（数据库、后端、视频服务） |
| `e2e-test.sh` | E2E 测试脚本 | 执行完整的 8 步端到端验证流程 |
| `quick-video-test.sh` | 快速视频测试 | 快速验证视频生成功能（跳过认证） |

### 文档

| 文件 | 说明 | 适用人群 |
|------|------|---------|
| `QUICKSTART.md` | 快速参考 | ⭐ 新手优先看这个（30 秒快速开始） |
| `LOCAL_E2E_GUIDE.md` | 完整指南 | 需要深入了解的开发者 |
| `LOCAL_E2E_IMPLEMENTATION.md` | 实现总结 | 想了解整个实现方案的人 |
| `README.md` | 项目总览 | 整体项目说明 |
| `backend/README.md` | 后端文档 | 后端开发和 API 文档 |

### 关键实现

| 文件 | 说明 |
|------|------|
| `backend/cmd/local-video-service/entry.go` | 本地视频生成服务入口（ffmpeg 生成 mp4） |

---

## 🚀 快速使用流程

### 第一次使用（5 分钟）

```bash
# 1. 切换到项目目录
cd /path/to/3kstory

# 2. 安装 ffmpeg
brew install ffmpeg

# 3. 启动服务（保持此终端运行）
bash start-local-e2e.sh

# 4. 打开新终端，运行测试
bash e2e-test.sh

# 5. 等待测试完成，查看结果
# 输出：最终视频 URL: http://localhost:8003/files/{video_id}.mp4
```

### 后续使用

```bash
# 启动服务
bash start-local-e2e.sh

# 测试
bash e2e-test.sh

# 快速视频测试（可选）
bash quick-video-test.sh
```

---

## 📚 文档阅读指南

### 如果你是...

#### 👤 第一次使用的开发者
1. **必读**：[QUICKSTART.md](../../QUICKSTART.md) - 5 分钟理解全貌
2. **参考**：本清单下方的"常见问题"

#### 👨‍💻 想深入了解的开发者
1. **阅读**：[LOCAL_E2E_GUIDE.md](./LOCAL_E2E_GUIDE.md) - 完整详细指南
2. **参考**：[backend/README.md](../../backend/README.md) - API 和架构文档

#### 🏗️ 想了解实现方案的架构师
1. **阅读**：[LOCAL_E2E_IMPLEMENTATION.md](./LOCAL_E2E_IMPLEMENTATION.md) - 实现总结
2. **查看**：[docs/02-技术架构.md](../02-技术架构.md) - 系统架构

#### 🚀 想快速集成的工程师
1. **参考**：[QUICKSTART.md](../../QUICKSTART.md) - 快速命令
2. **查看**：下方的"API 快速参考"

---

## 🔗 关键 URL 汇总

### 开发服务 URL

| 服务 | URL | 检查命令 |
|------|-----|---------|
| 后端 API | `http://localhost:8080/api/v1` | `curl http://localhost:8080/api/v1/health` |
| 视频服务 | `http://localhost:8003` | `curl http://localhost:8003/health` |
| PostgreSQL | `localhost:5432` | `docker-compose ps` |
| Redis | `localhost:6379` | `docker-compose logs redis` |

### 数据存储

| 类型 | 位置 | 说明 |
|------|------|------|
| 生成的视频 | `.local/videos/` | MP4 视频文件 |
| 数据库数据 | Docker volume | PostgreSQL 数据 |
| 缓存数据 | Docker container | Redis 内存 |

---

## 🔍 常见问题快速解答

### Q1: 如何启动服务？
```bash
bash start-local-e2e.sh
```

### Q2: 如何运行 E2E 测试？
```bash
bash e2e-test.sh
```

### Q3: 生成的视频在哪里？
```bash
ls -lh .local/videos/
```

### Q4: 如何在浏览器预览视频？
- 从 `e2e-test.sh` 的输出找到 `video_url`
- 或访问：`http://localhost:8003/files/{video_id}.mp4`

### Q5: 报错"ffmpeg not found"怎么办？
```bash
brew install ffmpeg
```

### Q6: 报错"Port 8080 already in use"怎么办？
```bash
lsof -i :8080  # 找到占用进程
kill -9 {PID}  # 关闭进程
```

### Q7: 如何修改视频参数（时长、分辨率）？
编辑 `e2e-test.sh`，找到这些行：
```bash
"duration": 10,
"aspect_ratio": "16:9"
```

### Q8: 测试中途中断如何恢复？
```bash
# 1. 停止所有服务（Ctrl+C 在启动脚本终端）
# 2. 重新启动
bash start-local-e2e.sh
bash e2e-test.sh
```

### Q9: 我想只测试视频生成，不需要测试认证流程？
```bash
bash quick-video-test.sh
```

### Q10: 如何看服务的日志？
```bash
# 在启动脚本的终端中可以看到日志
# 或查看 Docker 日志
docker-compose logs -f backend
docker-compose logs -f postgres
```

---

## 📊 测试结果示例

### 成功的 E2E 测试输出

```
╔════════════════════════════════════════════════════╗
║   3kstory 本地端到端验证测试 (E2E Test)         ║
║   测试流程: 剧本 → 后端 → 视频任务 → mp4 URL    ║
╚════════════════════════════════════════════════════╝

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
步骤 1: 检查服务健康状态
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ 后端服务已就绪
✅ 视频生成服务已就绪

... [步骤 2-7 输出] ...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
步骤 8: 验证视频文件
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

ℹ️  检查视频 URL: http://localhost:8003/files/abc12345def.mp4
✅ 视频文件可访问 (HTTP 200)
✅ 视频文件大小: 42.50 MB

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

---

返回入口：见 [START_HERE.md](./START_HERE.md)
