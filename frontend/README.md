# Frontend - 3kstory 官网 & 应用

Next.js 14 + React 18 前端，提供酷炫官网展示和网剧创作应用。

---

## 🚀 快速开始

### 前置要求
- Node.js 18+
- npm / yarn / pnpm

### 安装和运行

```bash
cd frontend

# 1. 安装依赖
npm install

# 2. 配置环境变量（可选）
# 创建 .env.local 文件
# NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
# NEXT_PUBLIC_WS_URL=ws://localhost:8080
# NEXT_PUBLIC_DEMO_MODE=auto

# 3. 启动开发服务器
npm run dev
# 访问 http://localhost:3000

# 4. 生产构建
npm run build
npm start

# 5. 代码检查
npm run lint
npm run type-check
```

**✅ 前端访问**：http://localhost:3000

---

## 📊 技术栈

- **框架**：Next.js 14 (App Router)
- **UI**：React 18 + TypeScript
- **样式**：Tailwind CSS
- **组件库**：Ant Design
- **动画**：Framer Motion
- **HTTP**：Axios
- **状态管理**：Zustand

---

## 📁 项目结构

```
frontend/
├── app/                            # Next.js 14 App Router
│   ├── layout.tsx                  # 全局布局
│   ├── page.tsx                    # 首页
│   ├── (auth)/                     # 认证页面
│   │   ├── login/page.tsx
│   │   └── signup/page.tsx
│   ├── (app)/                      # 应用页面
│   │   ├── dashboard/page.tsx      # 项目仪表板
│   │   ├── projects/[id]/edit      # 编辑器
│   │   └── projects/[id]/preview   # 预览
│   └── api/                        # 后端路由（可选）
├── components/                     # React 组件
│   ├── Header.tsx                  # 顶部导航
│   ├── Footer.tsx
│   ├── ProjectForm.tsx             # 项目创建表单
│   ├── ScenePreview.tsx            # 场景预览
│   ├── Editor/                     # 编辑器组件
│   └── ...
├── lib/                            # 工具函数
│   ├── api.ts                      # API 客户端
│   ├── store.ts                    # Zustand 状态管理
│   └── utils.ts                    # 工具函数
├── types/                          # TypeScript 类型
│   └── index.ts
├── styles/                         # 全局样式
│   └── globals.css
├── public/                         # 静态资产
├── package.json
├── tsconfig.json
├── next.config.js
├── tailwind.config.js
└── .gitignore
```

---

## 🎨 页面规划

### 官网
- **首页** (`/`) - Hero 区域 + 功能展示 + 用户案例
- **定价** (`/pricing`) - 不同套餐选择
- **博客** (`/blog`) - 使用案例和教程
- **文档** (`/docs`) - 用户指南

### 应用
- **登录/注册** (`/login`, `/signup`) - 用户认证
- **仪表板** (`/dashboard`) - 项目列表和统计
- **编辑器** (`/projects/:id/edit`) - 三列布局编辑器
  - 左列：脚本编辑
  - 中列：实时预览
  - 右列：参数调整
- **预览** (`/projects/:id/preview`) - 完整预览

---

## 🔧 常用命令

```bash
npm run dev             # 开发服务器
npm run build           # 生产构建
npm start               # 启动生产版
npm run lint            # ESLint 检查
npm run type-check      # TypeScript 类型检查
npm run test            # 运行测试
```

---

## 📚 深度文档

- [docs/03-短剧创作流程.md](../docs/03-短剧创作流程.md) - 用户流程设计

---

## �� 开发里程碑

### 官网开发（Phase 1）
- Hero 演示视频和动画
- 功能展示区（Framer Motion）
- 用户案例和评价
- 定价表
- 登录/注册入口

### 应用后台（Phase 2）
- 项目列表和管理
- 统计仪表板
- 用户信息管理

### 网剧编辑器（Phase 3）
- 三列布局（脚本编辑 + 实时预览 + 参数调整）
- 脚本编辑器和格式化
- 分镜预览
- 进度监听（WebSocket）
- 视频下载和分享

详细开发计划请查看：[backend/README.md](../backend/README.md#-开发里程碑)

---

## 🌐 环境变量

创建 `.env.local`：

```env
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
NEXT_PUBLIC_WS_URL=ws://localhost:8080
NEXT_PUBLIC_DEMO_MODE=auto
```

### 演示模式开关

- `NEXT_PUBLIC_DEMO_MODE=on`：强制前端演示模式（不请求后端，直接返回前端静态兜底数据）
- `NEXT_PUBLIC_DEMO_MODE=auto`：默认模式（先请求后端，后端不可达时自动兜底）
- `NEXT_PUBLIC_DEMO_MODE=off`：关闭兜底（必须有后端）

---

## 🎬 官网特色

- **酷炫动画**：使用 Framer Motion 实现页面动效
- **演示视频**：展示完整短剧生成流程
- **性能优化**：Next.js 14 原生优化
- **响应式设计**：完美适配全平台

---

## 🎯 开发里程碑

### 官网开发（Phase 1）🚀

**特点**：
- 酷炫的 Hero 区域展示演示视频
- 功能展示区（自动化流程、质量、速度）
- 用户案例和评价
- 定价表
- 登录/注册入口

**技术栈**：
- Next.js 14 (App Router)
- Framer Motion（动画）
- Tailwind CSS
- Shadcn/ui 组件库

**核心页面**：
1. `/` - 首页（Hero + 功能展示）
2. `/features` - 功能详情
3. `/pricing` - 定价
4. `/blog` - 案例分享
5. `/login` - 登录页面
6. `/signup` - 注册页面

---

### 应用后台（Phase 2）

**用户仪表板**：
- 项目列表（卡片展示，支持搜索和筛选）
- 快速统计（已生成网剧数、总时长、积分等）
- 创建新项目按钮

**网剧编辑器**：
- 左侧：脚本编辑器（可视化场景设计）
- 中间：实时预览（分镜配图）
- 右侧：参数调整（模型选择、效果参数）

**导出和分享**：
- 视频下载和预览
- 社交分享（微博、微信、小红书）
- 作品库管理和展示

---

## 🛠️ 故障排除

### 前端 npm 安装失败
```bash
# 清理 node_modules 和锁文件
rm -rf node_modules package-lock.json

# 重新安装
npm install

# 如果仍失败，尝试
npm install --legacy-peer-deps
```

### 开发服务器启动失败
```bash
# 检查端口是否被占用
lsof -i :3000

# 清理 Next.js 缓存
rm -rf .next

# 重新启动
npm run dev
```

---

**更新日期**：2024 年 12 月
