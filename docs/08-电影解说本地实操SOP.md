# 08 电影解说本地实操 SOP

本文档面向当前阶段的实际目标：在本地 Mac 上完成电影解说视频生产，并在需要时调用 Windows 外星人提供的模型能力。

适用范围：

- 本地开发验证
- 快速试做几期电影解说视频
- 手动发布到 B 站 / 小红书 / 抖音 / 微博前的内容生产流程

---

## 1. 结果目标

执行完本文档后，你应该可以完成这条链路：

1. 在 Mac 上启动前端、后端、数据库、Redis、本地视频服务
2. 在 Windows 外星人上启动 Ollama 或 vLLM
3. 在网页里填写电影名称、简介、素材视频路径或直链
4. 系统生成解说文案、旁白音频、字幕和成片
5. 在本地找到最终 mp4 文件，人工审核并发布

---

## 2. 机器分工

### Mac 主机

- 前端页面
- Go 后端
- PostgreSQL
- Redis
- `local-video-service`
- 本地 TTS
- 视频导出和手动发布

### Windows 外星人

- Ollama 或 vLLM
- 后续可承接更重的模型任务

---

## 3. 前置准备

### 3.1 Mac 依赖

确保已安装：

- Go
- Node.js / npm
- Docker Desktop
- ffmpeg

安装 ffmpeg：

```bash
brew install ffmpeg
```

确认系统 TTS 可用：

```bash
say "3kstory 本地电影解说测试"
```

### 3.2 Windows 外星人依赖

建议已安装：

- Docker Desktop 或直接安装 Ollama
- NVIDIA 驱动
- Ollama 或 vLLM 环境

---

## 4. 启动 Windows 外星人模型服务

二选一即可，先追求稳定可用。

### 方案 A：Ollama

在 Windows 上运行：

```powershell
ollama serve
ollama pull qwen2.5:7b
```

健康检查：

```powershell
curl http://127.0.0.1:11434/api/tags
```

### 方案 B：vLLM

在仓库根目录下的 `agent` 目录执行：

```powershell
cd agent
docker compose -f docker-compose-vllm.yml up -d
```

健康检查：

```powershell
curl http://127.0.0.1:8000/health
```

---

## 5. 配置 Mac 后端环境

在 Mac 上准备 `backend/.env`，至少补齐这组配置：

```bash
ENV=development
PORT=8080
BASE_URL=http://localhost:8080
FRONTEND_URL=http://localhost:3000

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres123
DB_NAME=3kstory

REDIS_HOST=localhost
REDIS_PORT=6379

JWT_SECRET=dev-local-jwt-secret

AI_PROVIDER=local_ollama
OLLAMA_BASE_URL=http://<Windows局域网IP>:11434
VLLM_BASE_URL=http://<Windows局域网IP>:8000

AI_VIDEO_SERVICE_URL=http://localhost:8003/v1/generate
TTS_OUTPUT_DIR=.local/tts
```

如果先不走外星人，也可以临时改成：

```bash
AI_PROVIDER=cloud_qwen
```

---

## 6. 启动 Mac 本地服务

### 6.1 启动后端、数据库、Redis、视频服务

在项目根目录执行：

```bash
bash start-local-e2e.sh
```

看到以下类似输出说明基本正常：

```text
✅ PostgreSQL 已就绪
✅ 后端服务已就绪
✅ 视频生成服务已就绪
```

### 6.2 启动前端

新开一个终端，在项目根目录执行：

```bash
npm --prefix frontend install
npm --prefix frontend run dev
```

默认前端地址：

```text
http://localhost:3000
```

---

## 7. 准备电影素材

当前版本推荐两种素材方式。

### 方式 A：本地绝对路径

例如：

```text
/Users/yourname/Movies/test-movie.mp4
```

这是当前最稳定的方式，推荐优先使用。

### 方式 B：在线直链

例如：

```text
https://example.com/assets/movie-clip.mp4
```

注意：

- 必须是 ffmpeg 能直接读取的媒体文件 URL
- 不能直接填 YouTube、B 站、小红书页面地址
- 不能是需要登录鉴权的私有资源链接

---

## 8. 用网页生成电影解说视频

### 8.1 打开页面

访问：

```text
http://localhost:3000
```

先注册 / 登录，然后进入：

```text
http://localhost:3000/dashboard
```

### 8.2 先准备一个项目

如果项目列表为空，先在首页创建一个项目。

### 8.3 填写解说表单

在“创建解说视频”区域填写：

- 项目：选择已有项目
- 影片/剧名：例如 `让子弹飞`
- 剧情简介：尽量写清楚主线剧情
- 本地素材视频路径：推荐优先填本地 mp4 绝对路径
- 在线素材视频直链：如果不用本地路径可填这里
- 风格：`深度分析` / `搞笑` / `情感向`
- 目标时长：例如 `90`
- 旁白声音：先用默认女声
- Provider：选 `local`
- 画幅：横版可选 `16:9`，短视频可选 `9:16`

点击：

```text
生成解说视频
```

### 8.4 成功后的表现

页面会显示：

- 已提交任务
- `video_id`
- 当前状态

成功完成后，视频会出现在“最近记录”中。

---

## 9. 用 curl 直接生成电影解说

如果你想绕过前端，直接用接口调试，可以按下面流程。

### 9.1 注册账号

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "movie@example.com",
    "username": "movieuser",
    "password": "Test@123"
  }'
```

### 9.2 登录拿 token

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "movie@example.com",
    "password": "Test@123"
  }'
```

把返回里的 `token` 保存到环境变量：

```bash
export TOKEN='<你的token>'
```

### 9.3 创建项目

```bash
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "电影解说测试项目",
    "prompt": "本地电影解说视频测试"
  }'
```

记下返回的项目 `id`：

```bash
export PROJECT_ID=<项目ID>
```

### 9.4 提交电影解说任务

使用本地素材视频路径：

```bash
curl -X POST http://localhost:8080/api/v1/projects/$PROJECT_ID/generate-narration \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "movie_title": "让子弹飞",
    "synopsis": "张麻子假扮县长进鹅城，与黄四郎展开斗智斗勇，故事兼具黑色幽默与权力隐喻。",
    "style": "深度分析",
    "target_duration": 90,
    "voice": "female_cn",
    "provider": "local",
    "aspect_ratio": "16:9",
    "source_video_path": "/Users/yourname/Movies/rangzidanfei.mp4"
  }'
```

或使用直链：

```bash
curl -X POST http://localhost:8080/api/v1/projects/$PROJECT_ID/generate-narration \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "movie_title": "让子弹飞",
    "synopsis": "张麻子假扮县长进鹅城，与黄四郎展开斗智斗勇，故事兼具黑色幽默与权力隐喻。",
    "style": "深度分析",
    "target_duration": 90,
    "voice": "female_cn",
    "provider": "local",
    "aspect_ratio": "16:9",
    "source_video_url": "https://example.com/rangzidanfei.mp4"
  }'
```

返回里会有：

- `task_id`
- `video_id`
- `status`
- `video_url`

---

## 10. 导出和定位成片

### 10.1 浏览器直接打开

如果接口返回了：

```text
http://localhost:8003/files/<video_id>.mp4
```

可以直接在浏览器打开预览。

### 10.2 本地文件目录

最终视频默认输出在：

```text
backend/.local/videos/
```

旁白音频默认输出在：

```text
backend/.local/tts/
```

如果你是通过 `start-local-e2e.sh` 启动，视频和音频都在 `backend` 目录下面的 `.local/` 子目录里。

---

## 11. 发布前检查清单

建议每次导出后检查：

- 解说文案是否通顺，是否明显事实错误
- 素材片段是否和旁白内容基本匹配
- 字幕是否遮挡关键画面
- 旁白音量是否过小或机械感过强
- 总时长是否适合目标平台
- 是否存在版权、违规画面、敏感台词

建议先人工发 3 到 5 期，观察：

- 完播率
- 点赞率
- 评论关键词
- 哪种风格转化更好

---

## 12. 当前版本的真实能力边界

当前已经实现：

- 基于电影名和简介生成解说稿
- 本地生成旁白音频
- 从源视频切片
- 自动叠字幕
- 输出成片 mp4

当前还没有实现：

- 自动解析整部电影并理解剧情细节
- 自动做 ASR 转录和镜头切分
- 自动识别高光片段
- 自动发布到各平台
- 自动规避版权风险

所以当前最适合的工作流是：

- 你先自己准备一份素材视频
- 写好较准确的剧情简介
- 先用系统自动生成初稿
- 人工审核后再发布

---

## 13. 常见故障排查

### 问题 1：Windows 模型服务调用失败

检查：

- Windows 上 Ollama / vLLM 是否真的在运行
- Mac 能否访问 Windows IP 和端口
- 防火墙是否拦截 11434 或 8000

Mac 上可测试：

```bash
curl http://<Windows局域网IP>:11434/api/tags
curl http://<Windows局域网IP>:8000/health
```

### 问题 2：生成失败，提示找不到素材视频

检查：

- `source_video_path` 是否是绝对路径
- 文件是否真实存在
- 启动后端的当前机器是否能访问这个路径

验证：

```bash
ls -lh /Users/yourname/Movies/test-movie.mp4
```

### 问题 3：在线 URL 无法读取

检查：

- URL 是否直达 mp4 / mov / m3u8 等媒体资源
- 链接是否需要 cookie 或鉴权
- 是否被源站限流

可先本地验证：

```bash
ffprobe "https://example.com/test.mp4"
```

### 问题 4：旁白音频没有生成

Mac 上验证：

```bash
say "测试语音"
```

然后检查目录：

```bash
ls -lh backend/.local/tts
```

### 问题 5：视频生成成功但内容不准

这通常不是部署问题，而是输入不够准。

优化方式：

- 把剧情简介写得更具体
- 先缩短目标时长，例如 45 秒
- 一次只做一个明确主题
- 手动准备更贴近旁白内容的素材视频

---

## 14. 推荐首批试做策略

为了尽快发出几期视频，建议不要一开始就追求“整部电影自动理解”。

更务实的做法是：

1. 先挑 3 部你熟悉、素材容易准备的电影
2. 每部先做 30 到 90 秒版本
3. 每部测试两种风格：
   - 深度分析
   - 情绪/悬念型
4. 先发到一个平台验证
5. 观察反馈后再决定要不要加 ASR、镜头切分、自动高光抽取

---

## 15. 相关文档

- [技术架构](/Users/richard/Documents/3kstory/3kstory/docs/02-技术架构.md)
- [本地 Mac + Windows 协同部署指南](/Users/richard/Documents/3kstory/3kstory/docs/06-本地Mac-Windows协同部署指南.md)
- [本地 E2E 启动说明](/Users/richard/Documents/3kstory/3kstory/docs/local-e2e/START_HERE.md)
