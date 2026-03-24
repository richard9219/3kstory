# 21 导演 Agent 使用手册与 API 示例

本文档说明导演 Agent 的操作入口、典型工作流，以及后端接口的示例请求体。

## 1. 页面入口

- 工厂导航进入：`导演 Agent`
- 页面路径：`/factory/director-agents`

页面主要模块：

- 项目与模板池：查看内置模板与自定义模板
- 新增模板：配置导演风格参数与五维权重
- 编辑模板：回填已有模板并保存更新
- 自动导演策略器：按题材自动选模板并可批量应用
- A/B 双导演并行对比：同剧本双模板评分选优

## 2. 典型使用流程

1. 进入目标项目，确认已有分镜镜头。
2. 在模板池中创建或编辑导演模板。
3. 执行自动导演策略：
   - 输入题材（genre）
   - 可选择手动模板或让系统自动挑选
   - 设置 tune_percent 微调强度
   - 选择是否直接应用到镜头
4. 执行 A/B 对比：
   - 选择模板 A、模板 B
   - 查看 score_a / score_b 与 winner_template
   - 可选：应用最佳模板
   - 可选：应用后自动渲染最佳导演版

## 3. 权重字段说明

模板五维权重（建议总和约等于 1，后端会归一化处理）：

- `weight_narrative`: 叙事
- `weight_visual`: 视觉
- `weight_emotion`: 情绪
- `weight_rhythm`: 节奏
- `weight_continuity`: 连贯

取值建议范围：`0 ~ 1`。

## 4. API 清单

- `GET /api/v1/projects/:id/director-templates`
- `POST /api/v1/projects/:id/director-templates`
- `PUT /api/v1/projects/:id/director-templates/:templateID`
- `DELETE /api/v1/projects/:id/director-templates/:templateID`
- `POST /api/v1/projects/:id/director-agent/auto-strategy`
- `POST /api/v1/projects/:id/director-agent/ab-compare`

## 5. API 示例请求体

以下示例均为 JSON 请求体。

### 5.1 创建导演模板

接口：`POST /api/v1/projects/101/director-templates`

```json
{
  "name": "赛博黑色电影",
  "slug": "cyber-noir",
  "prompt_prefix": "neon rain, high contrast, reflective surfaces, moody atmosphere",
  "camera_language": "低机位跟拍 + 慢推 + 反射构图",
  "emotion_tone": "冷峻、压抑、悬疑",
  "transition_type": "fade",
  "transition_duration_ms": 260,
  "genre_keywords": "悬疑,犯罪,都市",
  "weight_narrative": 0.24,
  "weight_visual": 0.30,
  "weight_emotion": 0.20,
  "weight_rhythm": 0.16,
  "weight_continuity": 0.10
}
```

### 5.2 编辑导演模板

接口：`PUT /api/v1/projects/101/director-templates/333`

```json
{
  "name": "赛博黑色电影 v2",
  "prompt_prefix": "neon rain, hard rim light, cinematic smoke, high contrast",
  "camera_language": "低机位跟拍 + 斜向推轨",
  "emotion_tone": "冷峻、危机、悬疑",
  "transition_type": "match",
  "transition_duration_ms": 300,
  "genre_keywords": "悬疑,犯罪,都市,心理",
  "weight_narrative": 0.22,
  "weight_visual": 0.33,
  "weight_emotion": 0.20,
  "weight_rhythm": 0.15,
  "weight_continuity": 0.10
}
```

### 5.3 自动导演策略

接口：`POST /api/v1/projects/101/director-agent/auto-strategy`

```json
{
  "genre": "历史",
  "template_id": 0,
  "apply": true,
  "tune_percent": 72
}
```

说明：

- `template_id` 为空或 0 时，系统按题材自动选择。
- `apply=true` 时会批量更新可编辑镜头参数。

### 5.4 A/B 双导演对比

接口：`POST /api/v1/projects/101/director-agent/ab-compare`

```json
{
  "template_a_id": 101,
  "template_b_id": 102,
  "genre": "历史",
  "apply_best": true,
  "tune_percent": 70,
  "render_best_cut": true
}
```

说明：

- `apply_best=true`：把胜出模板应用到镜头。
- `render_best_cut=true`：应用后尝试触发导演版渲染。

## 6. 常见问题

### 6.1 删除模板失败

- 内置模板（`is_builtin=true`）不允许删除。

### 6.2 自动策略没有更新镜头

可能原因：

- 项目下暂无镜头
- 镜头全部被锁定
- `apply=false`

### 6.3 A/B 对比无法执行

请确认：

- `template_a_id` 与 `template_b_id` 都有效
- 两者不是同一个模板

## 7. 建议实践

- 先用自动策略做首轮风格收敛，再用 A/B 对比做精修。
- 对成熟题材沉淀为自定义模板，提升复用效率。
- 对关键项目记录每次模板权重变化，便于复盘与团队协作。
