# 06 本地 Mac + Windows 协同部署指南

本文档替代短期内的 AWS 主线部署方案。当前阶段以本地交付和内容验证为目标，推荐使用「Mac 主开发机 + Windows 外星人辅助推理」的单机协同路线。

## 1. 当前目标

- Mac 负责前端、后端、数据库、Redis、本地视频合成与发布操作
- Windows 外星人负责本地大模型推理与重计算任务
- 先打通「输入电影素材 -> 生成解说视频 -> 发布到社交媒体」闭环
- 暂不追求云上高可用、弹性扩缩容和公网部署

## 2. 推荐拓扑

```text
┌───────────────────────────────────────────────┐
│ Mac（主机）                                   │
│ - Frontend: Next.js                           │
│ - Backend: Go + Gin                           │
│ - PostgreSQL / Redis                          │
│ - local-video-service（ffmpeg 合成）          │
│ - 浏览器 / 运营发布                           │
└───────────────────────────────────────────────┘
                     │ 局域网 / Tailscale
                     ▼
┌───────────────────────────────────────────────┐
│ Windows 外星人（辅助节点）                    │
│ - Ollama 或 vLLM                              │
│ - 可选素材预处理 / 批量任务                   │
│ - 后续可承担图像、视频、ASR 等重计算          │
└───────────────────────────────────────────────┘
```

## 3. 服务职责划分

### 3.1 Mac 主机

- 运行 `start-local-e2e.sh` 或分别启动 backend / frontend / local-video-service
- 持有本地开发数据库与 Redis
- 接收用户输入的电影标题、剧情简介、素材视频路径或直链
- 生成解说稿
- 在 macOS 上用系统 `say` 生成旁白音频
- 用 ffmpeg 从源视频截取片段、叠字幕、拼接成解说视频
- 最终由你手动上传到 B 站 / 小红书 / 抖音 / 微博

### 3.2 Windows 外星人

- 启动 `agent/docker-compose-ollama.yml` 或 `agent/docker-compose-vllm.yml`
- 对外提供 `OLLAMA_BASE_URL` 或 `VLLM_BASE_URL`
- 负责解说稿生成、后续剧情分析、标签生成、多模态能力扩展

## 4. 推荐实施顺序

1. 先在 Mac 上跑通本地后端、PostgreSQL、Redis、video-service
2. 在 Windows 上跑通 Ollama 或 vLLM 健康检查
3. 在 Mac `.env` 中指向外星人模型服务
4. 从 `/dashboard` 提交电影解说任务，优先使用本地素材视频绝对路径
5. 生成成片后先手动发布几期视频，验证选题和效果

## 5. 关键环境变量

Mac 后端建议：

```bash
AI_PROVIDER=local_ollama
OLLAMA_BASE_URL=http://<windows-ip>:11434
VLLM_BASE_URL=http://<windows-ip>:8000
AI_VIDEO_SERVICE_URL=http://localhost:8003/v1/generate
TTS_OUTPUT_DIR=.local/tts
```

如果外星人不稳定，可以先切回：

```bash
AI_PROVIDER=cloud_qwen
```

## 6. 当前版本能力边界

已经支持：

- 电影标题 + 剧情简介生成解说稿
- 填写本地素材视频路径，按片段截取并叠字幕
- macOS 本地 `say` 生成真实旁白音频
- 也可传入 ffmpeg 可访问的直链视频 URL

暂未支持：

- 直接上传大型视频文件到后端
- 从 YouTube / 抖音页面 URL 自动下载素材
- 自动理解整部电影内容并生成高精度分段解说
- 自动发布到社交媒体

## 7. 为什么这条路线更适合现在

- 交付速度快，调试半径最小
- 电影解说最关键的是素材、文案、节奏，本地迭代效率最高
- 外星人只承担重计算，系统更稳，排障更简单
- 在发出几期视频前，不必承担云部署和运维成本
