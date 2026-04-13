# P4：前端集成

## 目标

实现草案创建/确认页面、节点执行态展示增强、发起入口改造，完成龙虾集成的用户侧闭环。

---

## 1. 新增页面

### 1.1 草案创建页 `DraftCreatePage`

**路由：** `/drafts/create`

**功能：**
- 顶部：页面标题"让龙虾生成流程"
- 主体：大文本输入框，placeholder 为"描述你的业务目标和期望的流程步骤..."
- 底部：提交按钮
- 提交后：显示加载态（"龙虾正在生成流程草案..."），调用 `POST /api/flow-drafts`
- 成功后：跳转到草案确认页 `/drafts/:id/confirm`

### 1.2 草案确认页 `DraftConfirmPage`

**路由：** `/drafts/:id/confirm`

**功能：**
- 顶部：草案标题（可编辑）+ 草案描述
- 主体：节点列表，每个节点以卡片形式展示
  - 节点名称（可编辑）
  - 节点类型标签（human_input / human_review / agent_execute / agent_export / human_acceptance）
  - 执行主体标签（人工 / 龙虾名称）
  - 节点执行责任人（可编辑）
  - 结果责任人标签（可编辑）
  - 输入/输出 schema 摘要
  - 执行契约摘要（完成条件、失败条件）
  - 删除按钮
- 节点排序：支持拖拽调整顺序（可选，MVP 可不做）
- 底部操作栏：
  - "确认创建" → 调用 `POST /api/flow-drafts/:id/confirm`，成功后跳转到 `/templates/:templateId/start`
  - "废弃草案" → 调用 `POST /api/flow-drafts/:id/discard`，返回首页

### 1.3 草案列表页 `DraftListPage`（可选）

**路由：** `/drafts`

展示当前用户的草案列表，状态筛选（draft / confirmed / discarded）。MVP 可暂缓，优先级低。

---

## 2. 现有页面改造

### 2.1 DashboardPage

新增"让龙虾生成流程"入口卡片：
- 卡片样式与现有"从模板发起"卡片并列
- 点击跳转到 `/drafts/create`

### 2.2 RunStartPage

顶部新增入口切换：
- Tab 或按钮组："从模板发起" / "让龙虾生成"
- "让龙虾生成"跳转到 `/drafts/create`

### 2.3 RunDetailPage — 节点卡片增强

#### 执行态展示

根据节点类型和状态，节点卡片增加以下展示：

| 节点状态 | 展示内容 |
|----------|----------|
| `running`（agent 节点） | 执行中动画 + 龙虾名称 + 已耗时 |
| `waiting_confirm`（agent 节点） | 结构化结果展示 + "确认通过"/"驳回" 按钮 |
| `done`（agent 自动完成） | 结构化结果 + 执行耗时 |
| `blocked` / `failed` | 异常信息 + "重试"/"人工接管" 按钮 |

#### 双主体标签

所有节点卡片新增两个标签：
- **执行主体**：人名或龙虾名称（蓝色标签=人、橙色标签=龙虾）
- **结果责任人**：人名（灰色标签）

这里不要继续复用“当前负责人”当作结果责任人。前端类型和接口响应要单独返回 `result_owner_person`。

#### 节点详情展开区

展开节点时显示：
- **执行日志区**：龙虾执行步骤 logs（按时间排列）
- **结构化结果区**：
  - `query` 类型 → 表格展示 records
  - `batch_operation` 类型 → 统计数字（成功/失败数）
  - `export` 类型 → 文件下载链接
- **附件/导出物区**：artifacts 列表（文件名 + 下载链接）
- **操作区**：确认/驳回/重试/人工接管按钮（根据状态显示）

---

## 3. 新增 API 调用层

### `frontend/src/api/drafts.ts`

```typescript
export function createDraft(data: CreateDraftRequest): Promise<FlowDraft>
export function getDraft(id: number): Promise<FlowDraft>
export function listDrafts(params?: DraftListParams): Promise<FlowDraft[]>
export function updateDraft(id: number, data: UpdateDraftRequest): Promise<FlowDraft>
export function confirmDraft(id: number): Promise<ConfirmDraftResponse>
export function discardDraft(id: number): Promise<void>
```

### `frontend/src/api/agentTasks.ts`

```typescript
export function listAgentTasks(params: { run_id?: number }): Promise<AgentTask[]>
export function getAgentTask(id: number): Promise<AgentTask>
export function getAgentTaskReceipt(taskId: number): Promise<AgentTaskReceipt>
export function confirmAgentResult(nodeId: number, data: ConfirmResultRequest): Promise<void>
export function takeoverNode(nodeId: number, data: TakeoverRequest): Promise<void>
```

---

## 4. 新增 Hooks

### `frontend/src/hooks/useDrafts.ts`

```typescript
export function useDrafts(params?: DraftListParams): { drafts, loading, error, refresh }
export function useDraftDetail(id: number): { draft, loading, error, refresh }
```

### `frontend/src/hooks/useAgentTasks.ts`

```typescript
export function useAgentTasks(runId: number): { tasks, loading, error, refresh }
export function useAgentTaskDetail(taskId: number): { task, receipt, loading }
```

---

## 5. TypeScript 类型定义扩展

在 `frontend/src/types/api.ts` 中新增：

```typescript
// 流程草案
interface FlowDraft {
  id: number
  title: string
  description: string
  source_prompt: string
  creator_person_id: number
  planner_agent_id?: number
  status: 'draft' | 'confirmed' | 'discarded'
  structured_plan_json: DraftPlan
  confirmed_template_id?: number
  created_at: string
  updated_at: string
  confirmed_at?: string
}

interface DraftPlan {
  title: string
  description: string
  nodes: DraftNode[]
  final_deliverable: string
}

interface DraftNode {
  node_code: string
  node_name: string
  node_type: string
  sort_order: number
  executor_type: 'agent' | 'human'
  owner_rule: string
  owner_person_id?: number
  executor_agent_code?: string
  result_owner_rule: string
  result_owner_person_id?: number
  task_type?: string
  input_schema: Record<string, unknown>
  output_schema: Record<string, unknown>
  completion_condition?: string
  failure_condition?: string
  escalation_rule?: string
}

// AgentTask
interface AgentTask {
  id: number
  run_id: number
  run_node_id: number
  agent_id: number
  task_type: string
  input_json: Record<string, unknown>
  status: 'queued' | 'running' | 'completed' | 'needs_review' | 'failed' | 'cancelled'
  started_at?: string
  finished_at?: string
  error_message?: string
  result_json: Record<string, unknown>
  artifacts_json: Artifact[]
  created_at: string
  updated_at: string
}

interface Artifact {
  name: string
  url: string
  type: string
}

interface AgentTaskReceipt {
  id: number
  agent_task_id: number
  run_id: number
  run_node_id: number
  agent_id: number
  receipt_status: 'completed' | 'needs_review' | 'failed' | 'blocked'
  payload_json: Record<string, unknown>
  received_at: string
}
```

在运行态节点类型里也要补充：

```typescript
interface RunNode {
  // ... existing ...
  result_owner_person_id?: number
  result_owner_person?: Person
}
```

---

## 6. 路由注册

在 `frontend/src/App.tsx` 中新增：

```typescript
<Route path="/drafts/create" element={<DraftCreatePage />} />
<Route path="/drafts/:id/confirm" element={<DraftConfirmPage />} />
// 可选
<Route path="/drafts" element={<DraftListPage />} />
```

---

## 7. 交付清单

- [ ] `frontend/src/pages/DraftCreatePage.tsx` — 草案创建页
- [ ] `frontend/src/pages/DraftConfirmPage.tsx` — 草案确认页
- [ ] `frontend/src/api/drafts.ts` — 草案 API 调用
- [ ] `frontend/src/api/agentTasks.ts` — AgentTask API 调用
- [ ] `frontend/src/hooks/useDrafts.ts` — 草案数据 hooks
- [ ] `frontend/src/hooks/useAgentTasks.ts` — AgentTask 数据 hooks
- [ ] `frontend/src/types/api.ts` — 类型扩展
- [ ] `frontend/src/pages/RunDetailPage.tsx` — 节点执行态增强
- [ ] `frontend/src/pages/RunStartPage.tsx` — 发起入口改造
- [ ] `frontend/src/pages/DashboardPage.tsx` — 入口新增
- [ ] `frontend/src/App.tsx` — 路由注册
