# 06 交付前端实现总结

## 实现范围

已按 `docs/tech_plan/mvp_v1/frontend/06_deliverable.md` 实现交付模块首版，覆盖：

- 已完成流程内的“生成交付物”入口
- 交付中心列表页
- 交付详情页
- 验收通过 / 驳回操作

## 路由与页面

新增页面：

- `frontend/src/pages/DeliverableListPage.tsx` -> `/deliverables`
- `frontend/src/pages/DeliverableDetailPage.tsx` -> `/deliverables/:deliverableId`

路由接入：

- `frontend/src/App.tsx`

## API 与 Hooks

新增 API：

- `POST /api/deliverables`
- `GET /api/deliverables`
- `GET /api/deliverables/:id`
- `POST /api/deliverables/:id/review`

实现文件：

- `frontend/src/api/deliverables.ts`
- `frontend/src/hooks/useDeliverables.ts`
- `frontend/src/types/api.ts`（新增 Deliverable 类型）

## 生成交付物

入口位置：

- 流程详情页 `RunDetailPage`（流程状态为 `completed`）

展示规则：

- 已有交付物：展示“查看交付物”
- 无交付物：
  - 发起人 / 管理员 / 最终节点责任人显示“生成交付物”
  - 其他用户显示“进入交付中心”

交付生成弹窗：

- 组件：`CreateDeliverableModal`
- 字段：
  - 标题
  - 摘要
  - 关键结论
  - 异常说明
  - 验收人
  - 关键附件
- 附件来源：
  - `flow_run` 附件
  - 所有 `flow_run_node` 附件
  - 通过 `attachment_ids` 提交

## 交付中心列表

列表字段：

- 交付标题
- 关联流程
- 发起人
- 验收人
- 验收状态
- 创建时间

筛选能力：

- `review_status`（全部 / 待验收 / 已归档）
- `reviewer_person_id`

状态展示：

- `DeliverableStatusTag`

## 交付详情与验收

交付详情展示：

- 交付标题、摘要
- 关联流程、发起人
- 验收人、验收状态、验收时间
- 验收意见（从 `result_json.review_comment` 展示）
- 节点完成情况（优先 `result_json.node_summary`）
- 关键附件列表

验收操作：

- 组件：`ReviewDeliverableModal`
- 仅当 `review_status=pending` 且用户为验收人或管理员显示按钮
- 支持：
  - 验收通过
  - 验收驳回
- 操作后刷新详情

## 关联改动

流程详情页：

- `frontend/src/pages/RunDetailPage.tsx`
- 接入 `CreateDeliverableModal`
- 生成成功后跳转到 `/deliverables/:id`

新增组件：

- `frontend/src/components/DeliverableStatusTag.tsx`
- `frontend/src/components/CreateDeliverableModal.tsx`
- `frontend/src/components/ReviewDeliverableModal.tsx`

样式扩展：

- `frontend/src/styles.css`
- 增加交付状态标签和附件选择区样式

## 当前边界

- 交付关键结论、异常说明保存在 `result_json`，未拆成独立结构化字段
- 验收操作仅支持单次更新状态（符合后端当前实现）
- 未实现 PDF / Excel 导出按钮
