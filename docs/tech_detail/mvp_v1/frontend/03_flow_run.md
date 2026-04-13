# 03 流程实例前端实现总结

## 实现范围

已按 `docs/tech_plan/mvp_v1/frontend/03_flow_run.md` 实现流程实例模块首版，覆盖：

- 流程列表页（全部 / 我发起的 / 待我处理）
- 流程详情页（节点时间线 + 节点详情切换）
- 取消流程弹窗与提交流程
- 发起后跳转流程详情链路承接

## 路由实现

新增路由：

- `/runs` -> 全部流程
- `/runs/mine` -> 我发起的
- `/runs/todo` -> 待我处理
- `/runs/:runId` -> 流程详情

文件：

- `frontend/src/App.tsx`
- `frontend/src/pages/RunListPage.tsx`
- `frontend/src/pages/RunDetailPage.tsx`

## API 与 Hooks

已对接接口：

- `GET /api/runs`
- `GET /api/runs/:id`
- `POST /api/runs/:id/cancel`

实现文件：

- `frontend/src/api/runs.ts`
- `frontend/src/hooks/useRuns.ts`
- `frontend/src/types/api.ts`（补齐 Run/RunNode/RunLog 类型）

新增 hooks：

- `useRuns(params)`
- `useRunDetail(runId)`
- `useCancelRun()`

## 列表页实现细节

1. 范围切换

- `/runs` 固定 `scope=all`
- `/runs/mine` 固定 `scope=initiated_by_me`
- `/runs/todo` 固定 `scope=todo`

2. 列表字段

- 实例名称
- 模板名称（通过模板列表映射模板 ID）
- 当前节点
- 当前责任人
- 状态（`RunStatusTag`）
- 发起人
- 更新时间

3. 筛选与分页

- 状态筛选
- 负责人 ID
- 发起人 ID
- 分页与刷新

## 详情页实现细节

1. 顶部摘要

- 流程标题
- 流程状态
- 发起人
- 当前节点
- 当前责任人
- 开始时间
- 模板信息

2. 节点时间线与详情联动

- 时间线按 `sort_order` 排序
- 当前/完成/异常/等待状态使用不同颜色
- 默认选中当前节点（或首节点）
- 点击时间线节点后右侧详情面板同步切换

3. 取消流程

- `CancelRunModal` 输入取消原因
- 原因为空时前端阻止提交
- 提交成功后刷新详情
- 取消按钮展示规则：
  - 当前 actor 为 admin，或
  - 当前 actor 的 `personId` 等于流程发起人
  - 且流程状态不为 `completed/cancelled`

## Actor 头透传（联调用）

为支持后端 `X-Person-ID` / `X-Role-Type`：

- 新增 `ActorBar` 组件，可在页面顶部设置当前身份
- 请求层统一透传到 Header

文件：

- `frontend/src/components/ActorBar.tsx`
- `frontend/src/lib/actor.ts`
- `frontend/src/lib/http.ts`

## 组件新增

- `frontend/src/components/RunStatusTag.tsx`
- `frontend/src/components/RunNodeTimeline.tsx`
- `frontend/src/components/RunNodeDetailPanel.tsx`
- `frontend/src/components/CancelRunModal.tsx`

## 样式补充

在 `frontend/src/styles.css` 新增：

- 流程状态标签样式
- 节点时间线与节点详情面板样式
- 详情页布局样式（桌面双栏 / 移动端单栏）
- ActorBar 样式

## 边界说明

- 节点真实操作区（提交/审核/驳回/补材料/异常/龙虾执行）将在 `04_node_processing` 模块实现
- 交付中心当前为占位页：
  - `frontend/src/pages/DeliverablePlaceholderPage.tsx`

