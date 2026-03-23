# Comfy Workflow Templates

这些模板给 `ProviderComfy` 使用。

使用方式：

1. 设置环境变量：

```env
COMFY_BASE_URL=http://localhost:8188
COMFY_WORKFLOW_DIR=workflows/comfy
```

2. 调用 `/api/v1/projects/:id/generate-video`
3. 在请求里传：

```json
{
  "provider": "comfy",
  "workflow_path": "wan_i2v_shot_template.json",
  "extra_data": {
    "prompt_node_id": "12",
    "prompt_input_name": "text",
    "image_node_id": "13",
    "image_input_name": "image"
  }
}
```

## 模板说明

- `wan_i2v_shot_template.json`
  - 首帧到视频
  - 用于《药命效应》第一批 6 个验证镜头
- `wan_t2v_shot_template.json`
  - 纯文生视频
  - 用于没有首帧时的概念预演

## 依赖

这些模板按常见 Comfy 视频工作流命名，通常需要：

- Wan 或同级别视频模型节点
- VideoHelperSuite 的视频输出节点
- 基础文本编码/采样节点

如果你的 Comfy 节点名称不同，直接修改模板里的 `class_type` 或用 `extra_data.workflow_inputs` 覆盖节点输入即可。
