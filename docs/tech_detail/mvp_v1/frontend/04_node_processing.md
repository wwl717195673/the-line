# 04 节点处理前端实现总结

## 实现范围

已按 `docs/tech_plan/mvp_v1/frontend/04_node_processing.md` 在流程详情页内落地节点工作台，覆盖：

- 节点详情查询
- 节点输入暂存
- 节点动作执行（审核、驳回、补材料、完成、异常、运行龙虾）
- 评论发布
- 附件录入（MVP 使用 URL 绑定）
- 节点日志展示

承载页面：

- `/runs/:runId`（嵌入右侧节点工作台）

## API 对接

新增/接入接口：

- `GET /api/run-nodes/:id`
- `PUT /api/run-nodes/:id/input`
- `POST /api/run-nodes/:id/submit`
- `POST /api/run-nodes/:id/approve`
- `POST /api/run-nodes/:id/reject`
- `POST /api/run-nodes/:id/request-material`
- `POST /api/run-nodes/:id/complete`
- `POST /api/run-nodes/:id/fail`
- `POST /api/run-nodes/:id/run-agent`
- `POST /api/comments`
- `POST /api/attachments`

实现文件：

- `frontend/src/api/runNodes.ts`
- `frontend/src/api/collaboration.ts`
- `frontend/src/hooks/useRunNodes.ts`

## 组件与交互

新增组件：

- `frontend/src/components/RunNodeWorkbench.tsx`

接入页面：

- `frontend/src/pages/RunDetailPage.tsx`

工作台分区：

1. 节点基础信息

- 节点名称、编码、状态
- `available_actions` 列表

2. 输入/输出编辑区

- `input_json` 文本区
- `output_json` 文本区
- 暂存调用 `save_input`

3. 动作按钮区

- `submit`
- `approve`
- `reject`
- `request_material`
- `complete`
- `fail`
- `run_agent`

动作展示/可用规则：

- 以前端读取后端 `available_actions` 为准
- 流程 `cancelled` 时统一只读禁用

4. 附件、评论、日志

- 附件：输入 `file_name + file_url` 后创建绑定到当前节点
- 评论：输入文本后发布到当前节点
- 日志：按时间顺序展示 `log_type + content`

## 状态刷新规则

任一节点动作成功后，执行双刷新：

- 刷新当前节点详情（`run-nodes/:id`）
- 刷新流程详情（`runs/:id`）

目的：

- 保证流程当前节点与状态能在串行推进后即时更新
- 保证右侧动作按钮和日志与后端一致

## 代码调整

1. 类型扩展

- 在 `frontend/src/types/api.ts` 增加：
  - `RunNodeDetail`
  - `Attachment`
  - `Comment`

2. 流程详情页替换

- 由只读节点详情面板切换为可操作工作台
- 保留左侧时间线点击切换节点逻辑

3. 样式补充

- 在 `frontend/src/styles.css` 增加：
  - `action-grid`
  - `inline-form`
  - `plain-list`

## 当前边界

- 驳回/补材料/异常/审核备注目前使用 `window.prompt` 收集参数，后续可替换为独立 modal
- 附件目前只实现 JSON 方式 URL 绑定，未接本地文件上传表单
- 节点详情仍以 JSON 文本编辑为主，固定节点专用表单将在后续模块进一步收敛
