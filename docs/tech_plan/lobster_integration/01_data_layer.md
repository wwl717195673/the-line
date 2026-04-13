# P1：数据层扩展

## 目标

新增 3 个核心数据模型，扩展模板/实例节点责任字段、节点类型和状态常量，为 P2（真实执行）和 P3（辅助编排）提供数据基础。

---

## 1. 新增数据模型

### 1.1 FlowDraft（流程草案）

用于承接自然语言生成的流程草案。

```go
type FlowDraft struct {
    ID                  uint64         `gorm:"primaryKey" json:"id"`
    Title               string         `gorm:"size:256;not null" json:"title"`
    Description         string         `gorm:"type:text" json:"description"`
    SourcePrompt        string         `gorm:"type:text;not null" json:"source_prompt"`
    CreatorPersonID     uint64         `gorm:"not null;index" json:"creator_person_id"`
    PlannerAgentID      *uint64        `gorm:"index" json:"planner_agent_id"`
    Status              string         `gorm:"size:32;not null;index" json:"status"`
    StructuredPlanJSON  datatypes.JSON `gorm:"type:json" json:"structured_plan_json"`
    ConfirmedTemplateID *uint64        `gorm:"index" json:"confirmed_template_id"`
    CreatedAt           time.Time      `json:"created_at"`
    UpdatedAt           time.Time      `json:"updated_at"`
    ConfirmedAt         *time.Time     `json:"confirmed_at"`
}
```

**字段说明：**

| 字段 | 说明 |
|------|------|
| `SourcePrompt` | 用户原始自然语言输入 |
| `PlannerAgentID` | 生成草案的编排龙虾 ID |
| `Status` | `draft` / `confirmed` / `discarded` |
| `StructuredPlanJSON` | 结构化流程草案（节点列表、类型、责任人等） |
| `ConfirmedTemplateID` | 确认后生成的模板 ID |

### 1.2 AgentTask（节点执行任务）

记录虾线向龙虾发起的一次真实执行请求。

```go
type AgentTask struct {
    ID            uint64         `gorm:"primaryKey" json:"id"`
    RunID         uint64         `gorm:"not null;index" json:"run_id"`
    RunNodeID     uint64         `gorm:"not null;index" json:"run_node_id"`
    AgentID       uint64         `gorm:"not null;index" json:"agent_id"`
    TaskType      string         `gorm:"size:64;not null" json:"task_type"`
    InputJSON     datatypes.JSON `gorm:"type:json" json:"input_json"`
    Status        string         `gorm:"size:32;not null;index" json:"status"`
    StartedAt     *time.Time     `json:"started_at"`
    FinishedAt    *time.Time     `json:"finished_at"`
    ErrorMessage  string         `gorm:"type:text" json:"error_message"`
    ResultJSON    datatypes.JSON `gorm:"type:json" json:"result_json"`
    ArtifactsJSON datatypes.JSON `gorm:"type:json" json:"artifacts_json"`
    CreatedAt     time.Time      `json:"created_at"`
    UpdatedAt     time.Time      `json:"updated_at"`
}
```

**字段说明：**

| 字段 | 说明 |
|------|------|
| `TaskType` | `query` / `batch_operation` / `export` |
| `Status` | `queued` / `running` / `completed` / `needs_review` / `failed` / `blocked` / `cancelled` |
| `ResultJSON` | 结构化执行结果 |
| `ArtifactsJSON` | 执行产出物（文件、报表等） |

### 1.3 AgentTaskReceipt（执行回执）

保存龙虾回传的原始回执，便于审计和回放。

```go
type AgentTaskReceipt struct {
    ID            uint64         `gorm:"primaryKey" json:"id"`
    AgentTaskID   uint64         `gorm:"not null;index" json:"agent_task_id"`
    RunID         uint64         `gorm:"not null;index" json:"run_id"`
    RunNodeID     uint64         `gorm:"not null;index" json:"run_node_id"`
    AgentID       uint64         `gorm:"not null;index" json:"agent_id"`
    ReceiptStatus string         `gorm:"size:32;not null" json:"receipt_status"`
    PayloadJSON   datatypes.JSON `gorm:"type:json" json:"payload_json"`
    ReceivedAt    time.Time      `json:"received_at"`
}
```

**字段说明：**

| 字段 | 说明 |
|------|------|
| `ReceiptStatus` | `completed` / `needs_review` / `failed` / `blocked` |
| `PayloadJSON` | 龙虾原始回执全文 |

---

## 2. 模板与实例节点字段扩展

当前系统里的 `DefaultOwnerRule` / `OwnerPersonID` 更偏向“谁来处理节点”，还不能稳定表达“谁对结果负责”。

为了落实“执行主体与结果责任人分离”，建议在现有模型上新增字段，而不是复用旧字段：

### 2.1 FlowTemplateNode 扩展

```go
type FlowTemplateNode struct {
    // ... existing ...
    DefaultOwnerRule    string  `json:"default_owner_rule"`      // 谁执行
    DefaultAgentID      *uint64 `json:"default_agent_id"`        // 哪个龙虾执行
    ResultOwnerRule     string  `json:"result_owner_rule"`       // 谁对结果负责
    ResultOwnerPersonID *uint64 `json:"result_owner_person_id"`  // specified_person 时使用
}
```

### 2.2 FlowRunNode 扩展

```go
type FlowRunNode struct {
    // ... existing ...
    OwnerPersonID       *uint64 `json:"owner_person_id"`         // 当前执行责任人
    ReviewerPersonID    *uint64 `json:"reviewer_person_id"`      // 当前审核人
    ResultOwnerPersonID *uint64 `json:"result_owner_person_id"`  // 结果责任人
}
```

### 2.3 设计约定

* `OwnerPersonID` 表示节点执行责任人
* `ResultOwnerPersonID` 表示节点结果责任人
* 两者允许相同，但不能假定永远相同
* `human_acceptance` 节点通常要求显式指定 `ResultOwnerPersonID`

---

## 3. 节点类型扩展

在 `internal/domain/template.go` 中新增节点类型常量：

```go
// 新增节点类型（龙虾集成）
NodeTypeHumanInput      = "human_input"       // 人工输入
NodeTypeHumanReview     = "human_review"       // 人工审核
NodeTypeAgentExecute    = "agent_execute"      // 自动执行
NodeTypeAgentExport     = "agent_export"       // 自动导出
NodeTypeHumanAcceptance = "human_acceptance"   // 最终签收
```

**兼容策略：** 保留现有 5 个常量（`manual`, `review`, `notify`, `execute`, `archive`）不删除。现有固定模板继续使用旧类型，新建流程使用新类型。在 service 层做映射兼容，`execute` 和 `agent_execute` 共享同一处理逻辑。

---

## 4. 节点状态扩展

在 `internal/domain/run.go` 中新增：

```go
NodeStatusBlocked   = "blocked"    // 龙虾无法完成，需人工接管
NodeStatusCancelled = "cancelled"  // 节点被取消
```

---

## 5. Domain 常量新增

在 `internal/domain/` 下新增文件 `draft.go` 和 `agent_task.go`：

### draft.go

```go
const (
    DraftStatusDraft     = "draft"
    DraftStatusConfirmed = "confirmed"
    DraftStatusDiscarded = "discarded"
)
```

### agent_task.go

```go
const (
    // AgentTask 状态
    AgentTaskStatusQueued      = "queued"
    AgentTaskStatusRunning     = "running"
    AgentTaskStatusCompleted   = "completed"
    AgentTaskStatusNeedsReview = "needs_review"
    AgentTaskStatusFailed      = "failed"
    AgentTaskStatusBlocked     = "blocked"
    AgentTaskStatusCancelled   = "cancelled"

    // AgentTask 任务类型
    AgentTaskTypeQuery          = "query"
    AgentTaskTypeBatchOperation = "batch_operation"
    AgentTaskTypeExport         = "export"

    // 回执状态
    ReceiptStatusCompleted   = "completed"
    ReceiptStatusNeedsReview = "needs_review"
    ReceiptStatusFailed      = "failed"
    ReceiptStatusBlocked     = "blocked"
)
```

---

## 6. 数据库迁移

GORM AutoMigrate 会自动创建新表。需要在 `internal/db/` 的 migration 注册中加入：

```go
db.AutoMigrate(
    // 现有...
    &model.FlowDraft{},
    &model.AgentTask{},
    &model.AgentTaskReceipt{},
)
```

并同步扩展现有模型字段：

* `model.FlowTemplateNode` — 新增 `result_owner_rule` / `result_owner_person_id`
* `model.FlowRunNode` — 新增 `result_owner_person_id`

---

## 7. Repository 层

新增 3 个 repository：

| Repository | 方法 |
|------------|------|
| `FlowDraftRepository` | Create, GetByID, List(creatorID, status), Update, UpdateStatus |
| `AgentTaskRepository` | Create, GetByID, GetByRunNodeID, GetActiveByRunNodeID, List(runID), UpdateStatus, UpdateResult |
| `AgentTaskReceiptRepository` | Create, GetByTaskID, List(runID) |

`GetActiveByRunNodeID` 用于防止自动节点重复调度，活跃状态定义为：

* `queued`
* `running`

---

## 8. Service 层

新增 3 个 service 的基础 CRUD 部分：

| Service | 基础方法 |
|---------|----------|
| `FlowDraftService` | Create, Get, List, Update |
| `AgentTaskService` | Create, Get, ListByRun, ListByNode |
| `AgentTaskReceiptService` | Create, GetByTask |

> 业务逻辑（触发执行、处理回执、确认草案等）在 P2/P3 中实现。

---

## 9. Handler + 路由

新增 handler 和基础路由注册，具体 API 路径见 P2/P3 文档。

新增接口路径统一沿用当前项目风格，全部挂在 `/api/...` 下，不额外引入 `/api/v1/...`。

---

## 10. 交付清单

- [ ] `internal/model/flow_draft.go`
- [ ] `internal/model/agent_task.go`
- [ ] `internal/model/agent_task_receipt.go`
- [ ] `internal/model/flow_template_node.go` — 新增结果责任人字段
- [ ] `internal/model/flow_run_node.go` — 新增结果责任人字段
- [ ] `internal/domain/draft.go`
- [ ] `internal/domain/agent_task.go`
- [ ] `internal/domain/template.go` — 新增节点类型常量
- [ ] `internal/domain/run.go` — 新增 `blocked` / `cancelled` 状态
- [ ] `internal/repository/flow_draft_repository.go`
- [ ] `internal/repository/agent_task_repository.go`
- [ ] `internal/repository/agent_task_receipt_repository.go`
- [ ] `internal/service/flow_draft_service.go`（基础 CRUD）
- [ ] `internal/service/agent_task_service.go`（基础 CRUD）
- [ ] `internal/service/agent_task_receipt_service.go`（基础 CRUD）
- [ ] `internal/db/` migration 注册
- [ ] `internal/dto/` 新增请求/响应 DTO
