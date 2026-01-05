# 快速参考：本地 E2E 验证

## 📋 30 秒快速开始

```bash
# 1. 安装依赖（第一次运行）
brew install ffmpeg

# 2. 启动服务（终端 1）
cd /path/to/3kstory
bash start-local-e2e.sh

# 3. 运行测试（终端 2）
bash e2e-test.sh
```

## 🎯 完整链路验证

```
用户认证 → 项目创建 → 场景生成 → 视频生成 → 验证输出
   ✅        ✅        ✅        ✅        ✅
```

## 📍 关键 URL

| 服务 | URL | 检查 |
|------|-----|------|
| 后端 API | http://localhost:8080 | `curl http://localhost:8080/api/v1/health` |
| 视频生成 | http://localhost:8003 | `curl http://localhost:8003/health` |
| 数据库 | localhost:5432 | `docker-compose ps` |
| 生成的视频 | `.local/videos/` | `ls -lh .local/videos/` |

## 🔧 故障排除

| 问题 | 解决方案 |
|------|--------|
| Port 占用 | `lsof -i :8080` 然后 `kill -9 <PID>` |
| ffmpeg 缺失 | `brew install ffmpeg` |
| DB 连接失败 | `docker-compose restart postgres` |
| API 错误 | 查看 `start-local-e2e.sh` 中的服务日志 |

## 📊 时间预期

| 步骤 | 耗时 |
|------|------|
| ffmpeg 安装 | 3 分钟（第一次） |
| 启动服务 | 30 秒 |
| 完整 E2E 测试 | 3-5 分钟 |
| **总计** | **5-10 分钟** |

## 🎬 E2E 测试流程

```
✅ 检查服务健康状态
✅ 用户注册 (test@example.com)
✅ 用户登录 (获取 token)
✅ 创建项目
✅ 生成场景
✅ 请求视频生成
✅ 轮询等待完成 (最多 2 分钟)
✅ 验证视频文件可访问
```

## 📁 文件结构

```
3kstory/
├── start-local-e2e.sh           ← 启动脚本
├── e2e-test.sh                   ← 测试脚本
├── docs/local-e2e/LOCAL_E2E_GUIDE.md  ← 完整文档
├── backend/
│   └── ...
└── README.md                     ← 项目说明
```

## 🚀 生成的视频信息

```json
{
  "video_id": "abc12345def67890",
  "status": "completed",
  "video_url": "http://localhost:8003/files/abc12345def67890.mp4",
  "duration": 10,
  "resolution": "1280x720",
  "file_size": "25.5 MB"
}
```

## 🔄 修改测试参数

编辑 `e2e-test.sh` 中的这些变量：

```bash
# 用户邮箱
TEST_EMAIL="test@example.com"

# 视频时长（秒）- 在 e2e-test.sh 的 generateRequest 中
"duration": 10,

# 视频比例
"aspect_ratio": "16:9"  # 或 "9:16"
```

## 💡 典型用途

1. **首次本地验证** ← 你在这里
2. **快速功能测试**
3. **CI/CD 集成测试**
4. **演示不依赖 GPU**
5. **多人协作验证**

## 🎓 下一步

- ✅ 完成本地 E2E 验证
- 📚 阅读 [LOCAL_E2E_GUIDE.md](./docs/local-e2e/LOCAL_E2E_GUIDE.md) 了解详细信息
- 🚀 部署到生产环境（见 backend/README.md）
- 🎬 集成真实的视频生成 API（Runway/Pika）
- 🖥️ 启动前端应用进行完整 UI 测试
