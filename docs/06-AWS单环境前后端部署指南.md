# 06 AWS 单环境前后端部署指南

本文档面向单环境、前期自用场景，目标是把 3kstory 前后端稳定部署到一台 AWS EC2，并可对接外星人 AI 服务节点。

## 1. 架构说明

单环境推荐拓扑：

- AWS EC2：Nginx + Frontend + Backend + Local Video Service + PostgreSQL + Redis
- Windows 外星人：AI 推理服务（Ollama/vLLM/OpenCalw 相关服务）
- EC2 通过公网或 VPN 调用外星人 AI 接口（建议 VPN）

仓库中已提供以下部署文件：

- deploy/docker-compose.aws-single.yml
- deploy/nginx/aws-single.conf
- frontend/Dockerfile
- backend/Dockerfile
- backend/Dockerfile.video-service

## 2. EC2 规格建议

前期自用建议：

- 实例：t3.large 或 t3.xlarge
- 系统盘：100GB gp3
- 系统：Ubuntu 22.04 LTS
- 安全组：
  - 22 (仅你的固定 IP)
  - 80 (对外)
  - 443 (建议后续启用)

## 3. 服务器初始化

```bash
sudo apt update && sudo apt upgrade -y
sudo apt install -y ca-certificates curl gnupg git

# Docker
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo $VERSION_CODENAME) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
sudo usermod -aG docker $USER
```

执行后重新登录一次 SSH。

## 4. 拉取代码与配置环境变量

```bash
git clone <your-repo-url> 3kstory
cd 3kstory
cp backend/.env.example backend/.env
```

编辑 backend/.env，至少设置以下字段：

- `ENV=production`
- `PORT=8080`
- `BASE_URL=http://[EC2公网IP或域名]`
- `FRONTEND_URL=http://[EC2公网IP或域名]`
- `DB_HOST=postgres`
- `DB_PORT=5432`
- `DB_USER=postgres`
- `DB_PASSWORD=[强密码]`
- `DB_NAME=3kstory`
- `REDIS_HOST=redis`
- `REDIS_PORT=6379`
- `JWT_SECRET=[强随机字符串]`
- `AI_VIDEO_SERVICE_URL=http://video-service:8003/v1/generate`

如需把文本/图像能力指向外星人 AI 节点，可补充：

- `AI_PROVIDER=local_ollama` 或 `local_vllm` 或 `hybrid`
- `OLLAMA_BASE_URL=http://[alienware-ip]:11434`
- `VLLM_BASE_URL=http://[alienware-ip]:8000`

## 5. 启动服务

在仓库根目录执行：

```bash
cd deploy
docker compose -f docker-compose.aws-single.yml up -d --build
```

查看状态：

```bash
docker compose -f docker-compose.aws-single.yml ps
```

查看日志：

```bash
docker compose -f docker-compose.aws-single.yml logs -f backend
docker compose -f docker-compose.aws-single.yml logs -f frontend
docker compose -f docker-compose.aws-single.yml logs -f video-service
```

## 6. 验证清单

```bash
# 健康检查
curl http://<EC2公网IP>/health

# 首页
curl -I http://<EC2公网IP>/

# API 示例
curl -I http://<EC2公网IP>/api/v1/auth/login
```

浏览器检查：

- http://<EC2公网IP>/
- http://<EC2公网IP>/platforms
- http://<EC2公网IP>/dashboard

## 7. HTTPS（建议）

前期可先 HTTP，自用阶段建议尽快上 HTTPS。

两种方式：

- 方式 A：AWS ALB + ACM（推荐）
- 方式 B：EC2 内 Nginx + Certbot

若使用 ALB：

- ALB 监听 443，证书用 ACM
- 目标组转发到 EC2:80
- `BASE_URL` 与 `FRONTEND_URL` 改为 `https://[你的域名]`
- 同步更新各 OAuth 平台回调地址

## 8. 日常运维

### 8.1 更新版本

```bash
cd ~/3kstory
git pull
cd deploy
docker compose -f docker-compose.aws-single.yml up -d --build
```

### 8.2 备份数据库

```bash
cd ~/3kstory/deploy
docker exec 3kstory-postgres pg_dump -U postgres 3kstory > backup-$(date +%F).sql
```

### 8.3 清理无用镜像

```bash
docker image prune -f
```

## 9. 生产化增强建议（后续）

- PostgreSQL/Redis 改为 RDS + ElastiCache
- 视频文件改存 S3，避免本地盘风险
- 接入 CloudWatch/Prometheus 告警
- 引入 CI/CD（GitHub Actions）自动部署
- 用 VPN（Tailscale/WireGuard）访问外星人 AI 服务，避免公网暴露

## 10. 常见问题

1. 前端能开，接口 502

- 检查 backend 容器是否启动
- 检查 backend/.env 是否正确

1. OAuth 回调失败

- 检查 BASE_URL、FRONTEND_URL 是否线上域名
- 检查开放平台配置的回调地址是否一致

1. 视频生成失败

- 检查 video-service 日志
- 检查 ffmpeg 是否可执行（容器镜像已内置）

1. 外星人 AI 服务连不上

- 优先通过 VPN 打通
- 检查安全组/防火墙
- 检查目标端口是否监听
