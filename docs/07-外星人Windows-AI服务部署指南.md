# 07 外星人 Windows AI 服务部署指南

本文档用于把 Windows 外星人作为常开 AI 执行节点，承载 OpenCalw/本地模型推理/离线任务执行，并与 AWS 后端协同。

## 1. 目标与职责

外星人节点职责：

- 承载 AI 推理服务（Ollama 或 vLLM，按你的 OpenCalw 方案）
- 执行视频/离线任务（可选）
- 在 AWS 不可达时继续本地执行任务
- 网络恢复后回传结果到 AWS

不建议把外星人直接作为公网 API 主节点。

## 2. 推荐安装清单（Windows）

- Docker Desktop（启用 WSL2）
- NVIDIA 驱动（最新稳定版）
- CUDA Toolkit（如 OpenCalw/vLLM 需要）
- Git
- Python 3.10+
- FFmpeg（如需本机做视频处理）
- Tailscale 或 WireGuard（建议）

## 3. AI 服务部署方式

你可以二选一：

### 3.1 Ollama（简单稳妥）

适合快速落地。

```powershell
# 启动 Ollama（若已安装为 Windows 服务可跳过）
ollama serve

# 拉模型
ollama pull qwen2.5:7b

# 测试
curl http://127.0.0.1:11434/api/generate -Method Post -Body '{"model":"qwen2.5:7b","prompt":"hello","stream":false}' -ContentType 'application/json'
```

### 3.2 vLLM（性能更高）

建议在 Docker + GPU 模式运行。

```powershell
cd <repo>/agent
docker compose -f docker-compose-vllm.yml up -d
```

健康检查：

```powershell
curl http://127.0.0.1:8000/health
```

## 4. 与 AWS 后端联通方式

推荐：VPN 内网（Tailscale/WireGuard）

- AWS EC2 和外星人加入同一私网
- AWS 后端通过私网 IP 调用外星人 AI 端口
- 不需要家庭路由器开放公网端口

如果暂时不用 VPN，最少做：

- 仅白名单 AWS 固定 EIP
- 外星人防火墙只开放 AI 端口给该 EIP
- 启用鉴权 token 与请求签名

## 5. 离线任务设计（重点）

为了应对 AWS 暂时不可用，建议在外星人部署本地 worker，采用双队列。

### 5.1 状态机建议

- queued
- running
- done_local
- synced
- failed_local
- retry_pending
- synced_failed

### 5.2 本地持久化建议

- 使用 SQLite 保存任务元数据
- 产物落盘目录：D:/3kstory/jobs/output
- 原始输入与日志目录：D:/3kstory/jobs/inbox, D:/3kstory/jobs/logs

### 5.3 同步机制

- worker 周期性探测 AWS API
- 在线：拉取云任务并执行，回写状态
- 离线：执行本地队列
- 恢复：批量补传 done_local 任务

### 5.4 幂等回传

回传接口按 task_id 幂等，重复上传不产生重复结果。

## 6. Windows 服务化（开机自启）

建议用 NSSM 或 Windows Task Scheduler。

### 6.1 使用 NSSM（推荐）

为以下进程建立服务：

- ollama serve 或 vLLM 启动脚本
- worker.exe（你的离线执行器）
- sync-agent.exe（你的补传同步器）

关键参数：

- Startup type：Automatic
- Restart on failure：开启
- Working directory：固定到项目目录
- 日志输出到本地文件

## 7. AWS 侧对接参数

后端 .env 中可设置：

- `AI_PROVIDER=local_ollama` 或 `local_vllm` 或 `hybrid`
- `OLLAMA_BASE_URL=http://[alienware-vpn-ip]:11434`
- `VLLM_BASE_URL=http://[alienware-vpn-ip]:8000`

建议额外增加：

- `AI_SERVICE_TOKEN=[强随机]`
- `WORKER_HEARTBEAT_SECONDS=30`
- `WORKER_SYNC_INTERVAL_SECONDS=60`

## 8. 安全建议

- 不暴露外星人数据库与管理端口到公网
- 全链路 HTTPS 或 VPN 内网
- API 调用加 token 与时间戳签名
- 本地磁盘开启 BitLocker（避免设备丢失风险）
- 周期备份模型配置与任务目录

## 9. 监控建议

最少监控：

- GPU 占用/显存
- worker 心跳
- 未同步任务数量
- 同步失败次数
- 磁盘可用空间

建议每 1 分钟采样并推送到本地或云端告警。

## 10. 故障处理手册

1. AWS 不可用

- worker 切换到离线队列
- 继续处理本地任务

1. 网络恢复

- sync-agent 自动补传
- 校验云端状态是否 synced

1. AI 服务崩溃

- Windows 服务自动重启
- 超过阈值后发告警

1. 磁盘满

- 停止新任务接入
- 先上传并清理历史产物

## 11. 你当前最推荐实施顺序

1. 先在外星人部署 Ollama/vLLM 并跑通健康检查
2. 再打通 EC2 到外星人的 VPN 通信
3. 接入后端 AI_PROVIDER 与 BASE_URL
4. 最后上 worker + sync-agent 的离线机制

这样可以先用起来，再逐步增加鲁棒性。
