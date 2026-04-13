# 00 前端总体架构

---

## 1. 架构目标

MVP V1 前端目标是快速交付一个中后台式工作流协同界面，支持固定模板、流程实例、节点处理、评论附件、日志和交付验收。

本阶段不追求复杂低代码能力，不做拖拽画布，不做移动端专项适配。

---

## 2. 推荐目录结构

建议在 `frontend/` 下采用如下结构：

```text
frontend/
  src/
    app/
      App.tsx
      router.tsx
      providers.tsx
    api/
      client.ts
      persons.ts
      agents.ts
      templates.ts
      runs.ts
      runNodes.ts
      comments.ts
      attachments.ts
      deliverables.ts
    components/
      layout/
      common/
      forms/
      status/
      upload/
    features/
      base-data/
      templates/
      flow-runs/
      node-processing/
      collaboration/
      deliverables/
      dashboard/
    pages/
      dashboard/
      templates/
      runs/
      resources/
      deliverables/
    types/
      domain.ts
      api.ts
      enums.ts
    utils/
      format.ts
      permissions.ts
      nodeActions.ts
```

---

## 3. 路由设计

MVP 路由：

| 路由 | 页面 | 说明 |
|---|---|---|
| `/` | 工作台 | 默认首页 |
| `/templates` | 模板列表 | 固定模板列表 |
| `/templates/:templateId` | 模板详情 | 只读模板详情 |
| `/templates/:templateId/start` | 发起流程 | 基于模板发起流程 |
| `/runs` | 流程列表 | 全部流程 |
| `/runs/mine` | 我发起的 | 当前用户发起的流程 |
| `/runs/todo` | 待我处理 | 当前用户待办 |
| `/runs/:runId` | 流程详情 | 流程时间线和节点详情 |
| `/resources/persons` | 人员管理 | 人员列表和编辑 |
| `/resources/agents` | 龙虾管理 | 龙虾列表和编辑 |
| `/deliverables` | 交付中心 | 全部交付 |
| `/deliverables/:deliverableId` | 交付详情 | 交付结果和验收 |

---

## 4. 前端类型模型

核心类型：

```ts
type RunStatus = 'draft' | 'running' | 'waiting' | 'blocked' | 'completed' | 'cancelled';

type NodeStatus =
  | 'not_started'
  | 'ready'
  | 'running'
  | 'waiting_confirm'
  | 'waiting_material'
  | 'rejected'
  | 'done'
  | 'failed';

type NodeType = 'manual' | 'review' | 'notify' | 'execute' | 'archive';

type OperatorType = 'person' | 'agent' | 'system';
```

前端必须把状态枚举集中维护，避免页面中散落硬编码。

---

## 5. API 封装原则

统一 API client 负责：

* base URL 配置
* JSON 序列化
* 错误码处理
* 登录态或用户标识透传
* 上传请求封装
* 请求超时处理

每个模块单独维护 API 文件：

* `api/persons.ts`
* `api/agents.ts`
* `api/templates.ts`
* `api/runs.ts`
* `api/runNodes.ts`
* `api/comments.ts`
* `api/attachments.ts`
* `api/deliverables.ts`

---

## 6. 状态管理方案

服务端状态：

* 使用 query cache 管理列表、详情和字典数据
* 节点操作成功后失效 `runDetail`、`runNodeDetail`、`runList` 查询
* 评论发布成功后失效目标对象评论查询
* 附件上传成功后失效目标对象附件查询

本地 UI 状态：

* 当前选中的节点 ID
* 弹窗打开状态
* 表单草稿
* 列表筛选条件
* 侧栏展开状态

不建议 MVP 阶段引入复杂全局状态库。只有当前用户信息、菜单状态和轻量全局配置需要全局保存。

---

## 7. 权限与按钮控制

前端需要基于后端返回的 `available_actions` 控制按钮展示。

如果后端暂时不返回 `available_actions`，前端可先按以下条件计算：

* 责任人：暂存、提交确认、标记完成、标记异常、运行龙虾
* 审核人：审核通过、驳回、要求补材料、标记异常
* 管理员：全部动作
* 协作者：评论、上传附件
* 观察者：只读

注意：前端隐藏按钮只是体验控制，后端仍必须做权限校验。

---

## 8. 状态展示规范

流程状态：

* `running`：蓝色，进行中
* `waiting`：黄色，等待处理
* `blocked`：红色，阻塞中
* `completed`：绿色，已完成
* `cancelled`：灰色，已取消

节点状态：

* `not_started`：灰色
* `ready`：蓝色
* `running`：蓝色
* `waiting_confirm`：黄色
* `waiting_material`：橙色
* `rejected`：红色
* `failed`：红色
* `done`：绿色

---

## 9. 错误处理

前端需要统一处理：

* 表单校验错误：在字段下方展示
* 业务错误：页面消息提示
* 无权限：隐藏操作或展示无权限提示
* 资源不存在：展示空状态
* 上传失败：保留当前表单，不清空用户输入
* 节点操作失败：展示后端错误，不自动刷新为成功状态

---

## 10. 验收标准

* 前端路由覆盖所有 MVP 页面
* API 封装按模块拆分
* 状态枚举集中维护
* 流程详情和节点详情操作后能正确刷新
* 已取消流程和已完成节点展示只读状态
* 权限按钮展示符合 PRD 规则
