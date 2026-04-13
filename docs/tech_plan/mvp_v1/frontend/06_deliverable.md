# 06 交付前端技术方案

---

## 1. 模块目标

实现已完成流程的交付物生成、交付列表、交付详情和验收操作。

---

## 2. 页面与路由

| 路由 | 页面 | 说明 |
|---|---|---|
| `/deliverables` | 交付中心 | 全部交付 |
| `/deliverables?review_status=pending` | 待验收 | 待验收交付 |
| `/deliverables?review_status=approved` | 已归档 | 已验收通过交付 |
| `/deliverables/:deliverableId` | 交付结果页 | 交付详情和验收 |

生成交付物入口位于：

* `/runs/:runId`

---

## 3. 组件拆分

建议组件：

* `DeliverableListPage`
* `DeliverableTable`
* `DeliverableDetailPage`
* `DeliverableSummary`
* `NodeCompletionSummary`
* `DeliverableAttachmentList`
* `CreateDeliverableModal`
* `ReviewDeliverableModal`
* `DeliverableStatusTag`

---

## 4. API 对接

接口：

* `POST /api/deliverables`
* `GET /api/deliverables`
* `GET /api/deliverables/{id}`
* `POST /api/deliverables/{id}/review`

hook：

```ts
export function useDeliverables(params: DeliverableQuery) {}
export function useDeliverableDetail(id: string) {}
export function useCreateDeliverable() {}
export function useReviewDeliverable() {}
```

---

## 5. 生成交付物实现

展示条件：

* 流程状态为 `completed`
* 当前用户是发起人、管理员或最终节点责任人

表单字段：

* 交付标题
* 交付摘要
* 关键结论
* 异常说明
* 验收人
* 关键附件

默认值：

* 交付标题默认使用流程标题加“交付结果”
* 关键附件可从流程和节点附件中选择，提交时使用 `attachment_ids`
* 节点完成情况由后端返回或前端从流程详情中展示

提交规则：

* 必填字段缺失时不能提交
* 提交成功后跳转交付详情页
* 提交失败时保留表单内容

---

## 6. 交付中心实现

列表字段：

* 交付标题
* 关联流程
* 发起人
* 验收人
* 验收状态
* 创建时间
* 操作

筛选项：

* 验收状态
* 验收人

交互：

* 点击交付标题进入交付详情
* 待验收 tab 使用 `review_status=pending`
* 已归档 tab 使用 `review_status=approved`

---

## 7. 交付详情实现

展示内容：

* 交付标题
* 流程摘要
* 流程发起人
* 流程完成时间
* 节点完成情况
* 关键结论
* 异常说明
* 关键附件
* 验收人
* 验收状态
* 验收意见

操作：

* 验收通过
* 验收驳回

交互规则：

* 只有验收人或管理员展示验收按钮
* 已验收通过或已驳回后不重复展示验收按钮
* 验收操作成功后刷新交付详情
* MVP 不展示 PDF / Excel 导出按钮

---

## 8. 验收标准

* 已完成流程展示生成交付物入口
* 未完成流程不展示生成交付物入口
* 交付物生成表单必填校验有效
* 生成成功后跳转交付详情页
* 交付中心能展示交付列表
* 验收人能执行通过和驳回
* 验收后状态刷新正确
* 验收操作不改变流程状态
