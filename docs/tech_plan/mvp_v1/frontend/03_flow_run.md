# 03 流程实例前端技术方案

---

## 1. 模块目标

实现流程实例的发起、列表、详情、取消和串行推进后的页面刷新。

流程推进本身由后端执行，前端负责展示当前状态、触发节点操作、刷新数据并呈现结果。

---

## 2. 页面与路由

| 路由 | 页面 | 说明 |
|---|---|---|
| `/runs` | 全部流程 | 流程实例列表 |
| `/runs/mine` | 我发起的 | 发起人为当前用户 |
| `/runs/todo` | 待我处理 | 当前节点责任人为当前用户 |
| `/runs/:runId` | 流程详情 | 流程全局事实面 |

---

## 3. 组件拆分

建议组件：

* `RunListPage`
* `RunFilterBar`
* `RunTable`
* `RunStatusTag`
* `RunDetailPage`
* `RunSummaryPanel`
* `RunNodeTimeline`
* `RunNodeTimelineItem`
* `RunNodeDetailDrawer`
* `CancelRunModal`

---

## 4. API 对接

接口：

* `POST /api/runs`
* `GET /api/runs`
* `GET /api/runs/{id}`
* `POST /api/runs/{id}/cancel`

hook：

```ts
export function useRuns(params: RunQuery) {}
export function useRunDetail(runId: string) {}
export function useCancelRun() {}
```

节点操作成功后，需要失效：

* `runs.list`
* `runs.detail`
* `runNodes.detail`

---

## 5. 流程列表实现

列表字段：

* 实例名称
* 模板名称
* 当前节点
* 当前责任人
* 状态
* 发起人
* 最近更新时间
* 操作

筛选：

* 状态
* 负责人
* 发起人

实现规则：

* `/runs` 使用 `scope=all`
* `/runs/mine` 使用 `scope=initiated_by_me`
* `/runs/todo` 使用 `scope=todo`
* 默认按更新时间倒序
* 已完成和已取消流程不展示取消按钮

---

## 6. 流程详情实现

布局建议：

* 顶部：流程标题、状态、发起人、当前节点、当前责任人
* 左侧：流程基础信息
* 中间：节点时间线
* 右侧：节点详情抽屉或侧栏
* 底部或侧边：流程评论区

节点时间线规则：

* 当前节点高亮
* `done` 节点显示绿色
* `failed` 和 `rejected` 显示红色
* `waiting_confirm` 和 `waiting_material` 显示黄色或橙色
* `not_started` 显示灰色

交互规则：

* 进入详情页默认选中当前节点
* 点击时间线节点切换右侧详情
* 已取消流程下所有节点只读
* 已完成流程展示生成交付物入口

---

## 7. 取消流程实现

弹窗字段：

* 取消原因

提交规则：

* 原因不能为空
* 调用 `POST /api/runs/{id}/cancel`
* 成功后刷新列表和详情
* 失败后展示错误

展示规则：

* 只有发起人或管理员展示取消按钮
* `completed` 和 `cancelled` 状态不展示取消按钮

---

## 8. 前端状态刷新规则

以下动作成功后必须刷新流程详情：

* 节点提交确认
* 审核通过
* 驳回
* 要求补材料
* 标记完成
* 标记异常
* 运行龙虾
* 上传附件
* 发布评论
* 取消流程

---

## 9. 验收标准

* 流程列表展示字段完整
* `全部流程`、`我发起的`、`待我处理` 范围正确
* 流程详情展示 9 个节点
* 当前节点高亮正确
* 点击节点能切换节点详情
* 取消流程弹窗校验有效
* 取消后流程详情变为只读
* 节点推进后流程详情自动展示新的当前节点
