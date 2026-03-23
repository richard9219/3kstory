# 15 低成本视频生成与 LLM 接入调研报告

面向目标：

- 以低成本验证 [`docs/10-药命效应-第一期历史化实战稿`](/Users/richard/Documents/3kstory/3kstory/docs/10-药命效应-第一期历史化实战稿.md) 的可生产性
- 给出 `Seedance / MiniMax / Runway / ComfyUI` 的使用路线判断
- 给出视频生成、剧本生成、分镜生成、旁白生成的多模型接入最佳实践

---

## 1. 结论先行

当前阶段不建议把 `Runway / Seedance / MiniMax` 直接当作主生产引擎长期硬接。

更适合这个项目的路线是：

1. `剧本 / 分镜 / 镜头提示词 / 旁白` 用便宜文本模型分层完成
2. `关键视觉` 先用图像模型或首帧方案定调
3. `视频` 优先走 `Image-to-Video`，不要直接大规模 `Text-to-Video`
4. `批量验证` 优先走 `ComfyUI/Comfy Cloud` 工作流
5. 只有在关键镜头、招商 Demo、对外展示时，再调用 `Seedance / MiniMax / Runway`

一句话判断：

- `Runway` 适合高质量样片，不适合低成本大批量验证
- `MiniMax` 比 `Runway` 更适合做 API 化验证，但仍然是云成本路线
- `Seedance` 质量和叙事潜力很强，但官方接入门槛、区域性和价格透明度不如 `ComfyUI` 灵活
- `ComfyUI` 不是“单个视频模型”，而是一条更优的生成路线：尤其适合“先定首帧、再控镜头、再批量出片”

---

## 2. 现有项目现状

当前代码里已经有一套不错的雏形：

- 视频提供商：[`backend/internal/services/video_service.go`](/Users/richard/Documents/3kstory/3kstory/backend/internal/services/video_service.go)
  - 已接入 `runway / pika / local`
- 文本生成：[`backend/internal/services/ai_service.go`](/Users/richard/Documents/3kstory/3kstory/backend/internal/services/ai_service.go)
  - 已支持 `cloud_qwen / local_vllm / local_ollama / hybrid`
- 模型健康中心：[`backend/internal/services/model_center_service.go`](/Users/richard/Documents/3kstory/3kstory/backend/internal/services/model_center_service.go)

问题不在“能不能接”，而在“接入颗粒度太粗”：

- 文本模型目前主要按 `一个 provider -> 一个 GenerateScript` 组织
- 视频模型目前主要按 `一个 provider -> 一个 GenerateVideo` 组织
- 还没有把“剧本、分镜、镜头提示词、文案压缩、旁白润色、首帧提示词、视频生成”拆成不同任务层

这会直接导致两件事：

- 好模型被用在不该花钱的地方
- 便宜模型没被放到最合适的环节

---

## 3. 主流视频模型调研

### 3.1 Runway

官方 Runway API 文档显示，视频按秒计费，`gen4.5` 为 `12 credits/s`，`gen4_turbo` 和 `gen3a_turbo` 为 `5 credits/s`；Runway 官方同时说明 `1 credit = $0.01`。这意味着：

- `5 credits/s` 约等于 `$0.05/s`
- `12 credits/s` 约等于 `$0.12/s`
- 一个 `10s` 视频大约在 `$0.5 ~ $1.2`

判断：

- 优点：成片感强、品牌认知高、适合高端样片
- 缺点：按秒成本太高，不适合前期大规模试错
- 适合场景：最终展示镜头、预告片、招商页 hero 片段

### 3.2 MiniMax

MiniMax 官方文档当前更偏向“视频资源包 + 单次扣减单位”模式。官方公开资料显示：

- `MiniMax-Hailuo-2.3-Fast` 生成一个 `768p 6s` 视频扣 `0.7 unit`
- `MiniMax-Hailuo-2.3` 或 `MiniMax-Hailuo-02` 生成一个 `768p 6s` 视频扣 `1 unit`
- 官方资源包示例是 `$1000` 对应 `3760 units`

按这个口径粗算：

- `0.7 unit` 约 `$0.186`
- `1 unit` 约 `$0.266`

判断：

- 优点：比 Runway 更适合批量 API 调用，I2V 路线更实用
- 缺点：采购模型不是简单按次结算，预算控制要围绕资源包设计
- 适合场景：中等规模验证、批量镜头生成、I2V 驱动的短片生产

### 3.3 Seedance

BytePlus 官方产品页已经公开了 `Seedance 1.5 pro` 的产品能力和价格计算方式：

- 支持 `文本或图片转 5-10 秒视频`
- 支持原生音视频联合生成
- 公开价格为：
  - `with audio`: `$2.4 / 1M tokens`
  - `without audio`: `$1.2 / 1M tokens`
- 官方给出的例子：
  - `5s 720p 16:9` 无音频约 `$0.494`
  - `5s 720p 16:9` 含音频约 `$0.988`

判断：

- 优点：叙事感、镜头感、人物表现力很强，尤其适合“单镜头像成片”的场景
- 缺点：仍然偏贵；如果直接把它作为全流程主力，镜头试错成本不低
- 适合场景：关键镜头、结尾致敬镜头、样片升级版

补充说明：

- 我能核实到的是 BytePlus 官方 `Seedance 1.5 pro` 产品页和价格示例
- 如果你目标是更晚一代 Seedance 型号，当前公开口径和可用接入区域可能还会变化，正式采购前要再核价

### 3.4 ComfyUI / Comfy Cloud

`ComfyUI` 的价值不在“它是不是最强视频模型”，而在于它把生成过程变成了可复用工作流。

Comfy 官方信息能确认几点：

- `ComfyUI` 是开源、节点式、可本地运行的生成系统
- 官方明确强调它适合本地快速迭代、降低成本、完整掌控工作流
- 官方 Cloud 方案支持把 workflow 当成 API 运行
- Cloud 定价页公开给出一个很关键的参考：
  - `Free` 计划 `400 credits`，约可跑 `35` 个 `5s` 的 `Wan 2.2 Image-to-Video` 模板视频
  - `Standard $20/month`，约 `380` 个 `5s` 视频
  - `Creator $35/month`，约 `670` 个 `5s` 视频

判断：

- 优点：最适合低成本试错、首帧可控、风格一致性好、工作流可复用
- 缺点：需要自己搭工作流；最终质量取决于模型、节点和提示词工程
- 适合场景：验证期主路线、批量镜头工厂、分镜驱动短剧、风格统一的解说视频

---

## 4. 参考 comfy.org，是否有更优视频生成路线

有，而且更优的不是“换一个更便宜的单模型”，而是换路线。

### 4.1 不推荐路线

`长文本 -> 直接文生视频 -> 反复抽卡`

问题：

- 贵
- 难控
- 一旦人物、服装、构图漂移，要整段重来

### 4.2 更优路线

`剧本 -> 分镜 -> 首帧 -> Image-to-Video -> 剪辑拼接`

更具体一点：

1. 用文本模型把长稿拆成 `6-10` 个镜头段
2. 每段先产 `首帧提示词 + 参考图`
3. 视觉确认后再做 `I2V`
4. 长片不要一次生成，统一拆成 `4-6s` 或 `6-10s` 镜头
5. 后期用配音、字幕、音效和转场把它们拼成成片

这条路线的优势：

- 每次只花小额成本验证一个镜头
- 角色一致性、服装一致性、场景一致性更容易控
- 失败只重做某个镜头，不重做整片
- 适合 [`docs/10-药命效应-第一期历史化实战稿`](/Users/richard/Documents/3kstory/3kstory/docs/10-药命效应-第一期历史化实战稿.md) 这种“强叙事、强历史氛围、强镜头设计”的项目

### 4.3 对《药命效应》最适合的 Comfy 路线

推荐组合：

- `文本模型`：写解说、拆段、提取镜头目标
- `图像模型`：做“诸葛亮老年独白 / 隆中草庐 / 赤壁火光 / 五丈原病榻”的首帧
- `Comfy I2V`：把首帧扩成短镜头
- `后期`：旁白 + 音乐 + 字幕 + 镜头排序

为什么这条路线优于直接云视频模型：

- 《药命效应》不是纯动作炫技片
- 它更依赖“气氛、叙事、人物、象征物”
- 这类内容首帧定调成功后，短镜头扩展的性价比远高于直接文生整段视频

---

## 5. 面向低成本验证的最佳实践

### 5.1 模型按任务分层，不按厂商分层

推荐把模型接入拆成 6 类任务：

1. `story_ideation`
2. `longform_script`
3. `storyboard_structuring`
4. `shot_prompt_rendering`
5. `narration_polish`
6. `video_generation`

不要让一个大模型同时做全部工作。

### 5.2 好模型只放在“最后一跳”

推荐资源分配：

- `创意探索`：用本地 `Qwen / Ollama`
- `剧本成稿`：用云端中档文本模型
- `分镜结构化`：优先用便宜且稳定的 JSON 输出模型
- `关键镜头生成`：才调用高成本视频模型

原则：

- 便宜模型负责“多轮发散”
- 稳定模型负责“结构化输出”
- 贵模型负责“最后成像”

### 5.3 视频优先 I2V，不优先 T2V

低成本阶段的优先级建议：

1. `参考图/首帧 -> I2V`
2. `首尾帧 -> 视频补间`
3. `T2V`

原因：

- `I2V` 更稳
- 更易统一视觉
- 更适合历史题材人物与服化道控制

### 5.4 长片拆成镜头资产，不要一次生成

对于 10 分钟解说，不要理解成“生成一个 10 分钟视频”。

应该理解成：

- `40-80` 个可复用镜头资产
- 每个镜头资产 `3-8s`
- 后期用剪辑拼成完整节目

### 5.5 给每层任务都定义输出 schema

例如：

- 长稿输出：段落、情绪、节奏
- 分镜输出：镜头编号、场景、构图、运动、时长
- 首帧输出：主体、服装、材质、色调、镜头语言、负向词
- 视频输出：provider、model、seed、duration、resolution、cost_estimate

这会直接降低返工。

---

## 6. 《药命效应》最低成本验证方案

### 6.1 验证目标

不是先做完整 10 分钟成片。

第一轮只验证 3 件事：

1. `诸葛亮` 的角色视觉是否能稳定
2. `羽扇机括` 的核心意象是否能成立
3. `历史权谋 + 科幻异物` 的混合气质是否成立

### 6.2 最小可行镜头包

先只做 6 个镜头：

1. 隆中草庐夜读，第一次听见扇中有声
2. 地图沙盘前推演三分天下
3. 赤壁夜色，火光与江风掠过羽扇
4. 蜀锦、盐铁、铸币工坊的制度蒙太奇
5. 祁山军帐，最优方案失效
6. 五丈原风中病榻，羽扇静物收尾

### 6.3 最省钱生产路线

推荐顺序：

1. 文本模型生成 `6 镜头分镜`
2. 图像模型生成每镜头 `2 张首帧候选`
3. 选中后用 `ComfyUI I2V` 生成每镜头 `4-5s`
4. 只对其中 `1-2` 个最关键镜头升级到 `Seedance / MiniMax`
5. 用现成 TTS 加旁白合成 Demo

### 6.4 预算建议

验证期预算建议按三档控制：

- `极省版`
  - 主要走本地文本模型 + ComfyUI
  - 只做 6 个镜头 Demo
- `平衡版`
  - 大多数镜头走 ComfyUI
  - 2 个关键镜头走 MiniMax 或 Seedance
- `展示版`
  - 分镜与批量镜头仍走 ComfyUI
  - Hero 镜头与结尾镜头升级到 Seedance / Runway

推荐当前先走 `平衡版`。

---

## 7. 对当前项目的接入优化建议

### 7.1 新增任务级路由，不再只靠全局 provider

建议新增一层 `TaskRouter`：

- `script`
- `storyboard`
- `shot_prompt`
- `narration`
- `image`
- `video`

每个任务有自己的：

- 首选模型
- 备选模型
- 成本上限
- 超时上限
- 输出 schema

### 7.2 给视频层增加“策略”而不是只给 provider

建议把当前 `VideoGenerationRequest` 扩成：

- `strategy: storyboard_i2v | premium_t2v | template_montage | local_previz`
- `provider`
- `model`
- `reference_images`
- `seed`
- `cost_cap`

因为真正影响成本的往往不是 provider，而是策略。

### 7.3 文本层拆分为三类 prompt

建议不要一个 `GenerateScript()` 包办所有文本任务。

建议拆成：

1. `GenerateNarrativeOutline`
2. `GenerateStoryboardJSON`
3. `GenerateShotPromptPack`

### 7.4 记录每个镜头的成本

每次生成都落库：

- provider
- model
- duration
- resolution
- prompt hash
- retry count
- estimated cost
- actual cost

没有这层账本，就做不到真正的低成本优化。

---

## 8. 推荐落地路线

### 阶段 A：一周内可完成

- 保留现有 `runway / pika / local`
- 不急着把所有云视频模型都硬接进后端
- 先把任务层拆出来
- 先把《药命效应》的 6 镜头验证包跑通

### 阶段 B：验证通过后再扩

- 增加 `MiniMax` provider
- 增加 `Seedance` provider
- 增加镜头级成本统计
- 增加“Comfy workflow API”适配层

### 阶段 C：再考虑自动化生产

- 固化角色 LoRA/参考图
- 把爆款题材变成可复用模板
- 把 `剧本 -> 分镜 -> 首帧 -> 视频 -> 旁白` 做成流水线

---

## 9. 最终建议

如果目标是“低成本验证 `药命效应` 是否值得继续做”，最佳路线不是：

- 直接全面上 `Runway`
- 或者盲目补齐所有视频大模型 API

最佳路线是：

- `文本层便宜化`
- `分镜层结构化`
- `首帧层定调`
- `视频层 I2V 化`
- `高价模型只打关键镜头`

从项目角度，我的建议排序是：

1. 先做 `ComfyUI/Comfy Cloud + I2V` 主路线
2. 再补 `MiniMax` 作为中档云视频能力
3. `Seedance` 用于关键镜头升级
4. `Runway` 只保留给高端展示镜头

---

## 10. 参考资料

- Runway API Pricing: https://docs.dev.runwayml.com/guides/pricing/
- MiniMax Video Packages Pricing: https://platform.minimax.io/docs/guides/pricing-video
- MiniMax Video API Overview: https://platform.minimax.io/docs/api-reference/api-overview
- BytePlus Seedance Product Page: https://www.byteplus.com/en/product/Seedance
- ComfyUI Official Site: https://www.comfy.org/
- ComfyUI Docs: https://docs.comfy.org/
- Comfy Cloud Pricing: https://www.comfy.org/cloud/pricing

说明：

- 本报告优先使用官方公开页面
- `Seedance` 的更晚型号、开放区域和实时价格存在变化可能，正式采购前需要再次核价
