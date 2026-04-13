# 02 模板前端实现总结

## 实现范围

已按 `docs/tech_plan/mvp_v1/frontend/02_template.md` 完成模板模块首版实现，包含：

- 模板列表页：`/templates`
- 模板详情页：`/templates/:templateId`
- 流程发起页：`/templates/:templateId/start`

并补了最小流程详情占位页承接跳转：

- `/runs/:runId`

## 路由与页面

新增页面文件：

- `frontend/src/pages/TemplateListPage.tsx`
- `frontend/src/pages/TemplateDetailPage.tsx`
- `frontend/src/pages/RunStartPage.tsx`
- `frontend/src/pages/RunDetailPlaceholderPage.tsx`

路由接入：

- `frontend/src/App.tsx`

说明：

- 首页默认跳转模板中心 `/templates`
- 模板详情页只读，不提供编辑/拖拽入口

## API 对接

已接接口：

- `GET /api/templates`（模板列表）
- `GET /api/templates/:id`（模板详情）
- `POST /api/runs`（发起流程）
- `GET /api/runs/:id`（发起后跳转页展示占位数据）

对应实现：

- `frontend/src/api/templates.ts`
- `frontend/src/api/runs.ts`
- `frontend/src/hooks/useTemplates.ts`
- `frontend/src/hooks/useRuns.ts`

## 组件拆分落地

新增组件：

- `frontend/src/components/TemplateNodeTimeline.tsx`
- `frontend/src/components/TemplateNodeCard.tsx`
- `frontend/src/components/RunStartForm.tsx`

复用组件：

- `PersonSelect`（可选发起人选择）
- `Modal`、基础按钮和表单样式

## 关键交互实现

1. 模板列表页

- 展示模板名称、编码、版本、分类、状态、说明、更新时间
- 支持关键字搜索与分页
- 操作支持“查看详情”和“使用模板”
- 仅 `status=published` 展示“使用模板”按钮

2. 模板详情页（只读）

- 展示模板基础信息
- 展示节点时间线（按 `sort_order` 排序）
- 节点卡展示：节点编码、节点类型、默认责任人规则、默认龙虾、输入输出结构摘要
- 无编辑、无拖拽、无连线配置入口

3. 流程发起页

- 表单字段：
  - `title`
  - `reason`
  - `class_info`
  - `current_teacher`
  - `expected_time`
  - `extra_note`
  - `initiator_person_id`（可选，便于无登录态联调）
- 提交时组装：
  - `template_id`
  - `title`
  - `initiator_person_id`
  - `input_payload_json.form_data`
- 发起成功后跳转 `/runs/:runId`

## 异常处理

- 模板 ID 不合法：直接提示错误
- 模板不存在/下线：展示错误 + 重试
- 发起失败：保留表单输入并展示后端错误信息
- 网络失败：页面可重试

## 样式补充

新增模板模块样式：

- `kv-grid`
- `template-node-timeline`
- `template-node-card`
- `pill`

文件：

- `frontend/src/styles.css`

## 当前边界

- 流程详情仍是占位页，完整实现将在 `03_flow_run` 模块继续补齐
- 发起页“附件”目前为占位，待 `05_collaboration_trace` 模块接入
- 未执行 `npm install` / `npm run build` 编译验证（依赖安装后可执行）
