# 龙虾集成 Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在虾线 V1 基础上实现龙虾辅助编排（自然语言生成流程草案）和龙虾真实执行（自动节点触发 + 标准回执协议）两大能力。

**Architecture:** 采用 4 阶段分层推进：P1 扩展数据层（新 model/repo/service），P2 和 P3 基于同一数据层分别实现真实执行链路与辅助编排链路，P4 完成前端页面改造。龙虾侧通过统一接口 + Mock 实现对接。

**Tech Stack:** Go + Gin + GORM / React 18 + TypeScript + Vite

---

## 全局约定

- 所有新增后端文件均遵循现有模式：repository 用构造函数注入 `*gorm.DB`、service 用构造函数注入依赖、handler 用构造函数注入 service、路由在 `internal/app/router.go` 注册。
- 前端遵循 `src/api/` → `src/hooks/` → `src/types/api.ts` → `src/pages/` 的增量扩展模式。
- DTO 和响应统一使用现有的 `internal/response/` 和 `internal/dto/` 模式。
- 节点类型兼容：新类型 `agent_execute` / `agent_export` / `human_input` / `human_review` / `human_acceptance` 新增，旧类型保留。
- MVP 自动节点 `task_type` 只支持 `query` / `batch_operation` / `export`。
- 新接口路径统一使用当前项目风格 `/api/...`，不额外引入 `/api/v1/...`。
- 执行主体与结果责任人必须单独建模，不能复用 `DefaultOwnerRule` / `OwnerPersonID` 代替。
- 自动节点必须保证单飞执行：同一 `run_node` 同时最多只有一个活跃 `AgentTask`。

---

## 阶段依赖

```
P1 (数据层扩展)
  ├── P2 (龙虾真实执行) ──┐
  └── P3 (龙虾辅助编排) ──┼── P4 (前端集成)
```

P2 和 P3 可并行。P4 依赖 P2+P3 接口。

---

## Chunk 1: P1 数据层扩展

### 文件结构

| 新建文件 | 职责 |
|----------|------|
| `backend/internal/model/flow_draft.go` | FlowDraft 模型定义 |
| `backend/internal/model/agent_task.go` | AgentTask 模型定义 |
| `backend/internal/model/agent_task_receipt.go` | AgentTaskReceipt 模型定义 |
| `backend/internal/domain/draft.go` | 草案状态常量 |
| `backend/internal/domain/agent_task.go` | AgentTask 状态/类型/回执常量 |
| `backend/internal/repository/flow_draft_repository.go` | FlowDraft CRUD |
| `backend/internal/repository/agent_task_repository.go` | AgentTask CRUD |
| `backend/internal/repository/agent_task_receipt_repository.go` | AgentTaskReceipt CRUD |
| `backend/internal/service/flow_draft_service.go` | FlowDraft 基础 CRUD service |
| `backend/internal/service/agent_task_service.go` | AgentTask 基础 CRUD service |
| `backend/internal/service/agent_task_receipt_service.go` | AgentTaskReceipt 基础 CRUD service |
| `backend/internal/handler/flow_draft_handler.go` | FlowDraft REST handler |
| `backend/internal/handler/agent_task_handler.go` | AgentTask 查询/回执 handler（P1 先做查询部分） |
| `backend/internal/dto/draft.go` | FlowDraft 请求/响应 DTO |
| `backend/internal/dto/agent_task.go` | AgentTask 请求/响应 DTO |

| 修改文件 | 修改内容 |
|----------|----------|
| `backend/internal/domain/template.go` | 新增 5 个节点类型常量 |
| `backend/internal/domain/run.go` | 新增 `blocked` / `cancelled` 节点状态 |
| `backend/internal/model/flow_template_node.go` | 新增结果责任人字段 |
| `backend/internal/model/flow_run_node.go` | 新增结果责任人字段 |
| `backend/internal/db/migrate.go` | AutoMigrate 注册 3 个新 model |
| `backend/internal/app/router.go` | 注册新 handler 路由 |

### Task 1: FlowDraft 模型 + 常量

**Files:**
- Create: `backend/internal/model/flow_draft.go`
- Create: `backend/internal/domain/draft.go`
- Modify: `backend/internal/domain/template.go`

- [ ] **Step 1: 写 FlowDraft model**

```go
package model

import (
	"time"

	"gorm.io/datatypes"
)

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

func (FlowDraft) TableName() string {
	return "flow_drafts"
}
```

- [ ] **Step 2: 写 draft domain 常量**

```go
package domain

const (
	DraftStatusDraft     = "draft"
	DraftStatusConfirmed = "confirmed"
	DraftStatusDiscarded = "discarded"
)
```

- [ ] **Step 3: 扩展 template.go（不删旧常量，只做新增）**

在 `backend/internal/domain/template.go` 末尾追加：

```go
const (
	NodeTypeHumanInput      = "human_input"
	NodeTypeHumanReview     = "human_review"
	NodeTypeAgentExecute    = "agent_execute"
	NodeTypeAgentExport     = "agent_export"
	NodeTypeHumanAcceptance = "human_acceptance"
)
```

- [ ] **Step 4: Commit**

```bash
git add backend/internal/model/flow_draft.go backend/internal/domain/draft.go backend/internal/domain/template.go
git commit -m "feat(p1): add FlowDraft model and domain constants"
```

### Task 2: AgentTask / AgentTaskReceipt 模型 + 常量

**Files:**
- Create: `backend/internal/model/agent_task.go`
- Create: `backend/internal/model/agent_task_receipt.go`
- Create: `backend/internal/domain/agent_task.go`
- Modify: `backend/internal/domain/run.go`

- [ ] **Step 1: 写 AgentTask model**

```go
package model

import (
	"time"

	"gorm.io/datatypes"
)

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

func (AgentTask) TableName() string {
	return "agent_tasks"
}
```

- [ ] **Step 2: 写 AgentTaskReceipt model**

```go
package model

import (
	"time"

	"gorm.io/datatypes"
)

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

func (AgentTaskReceipt) TableName() string {
	return "agent_task_receipts"
}
```

- [ ] **Step 3: 写 agent_task domain 常量**

```go
package domain

const (
	AgentTaskStatusQueued      = "queued"
	AgentTaskStatusRunning     = "running"
	AgentTaskStatusCompleted   = "completed"
	AgentTaskStatusNeedsReview = "needs_review"
	AgentTaskStatusFailed      = "failed"
	AgentTaskStatusBlocked     = "blocked"
	AgentTaskStatusCancelled   = "cancelled"

	AgentTaskTypeQuery          = "query"
	AgentTaskTypeBatchOperation = "batch_operation"
	AgentTaskTypeExport         = "export"

	ReceiptStatusCompleted   = "completed"
	ReceiptStatusNeedsReview = "needs_review"
	ReceiptStatusFailed      = "failed"
	ReceiptStatusBlocked     = "blocked"
)
```

- [ ] **Step 4: 扩展 run.go 节点状态**

在 `backend/internal/domain/run.go` 中，找到 `NodeStatusFailed` 并新增：

```go
const (
	// ... existing ...
	NodeStatusFailed    = "failed"
	NodeStatusBlocked   = "blocked"
	NodeStatusCancelled = "cancelled"
)
```

- [ ] **Step 5: Commit**

```bash
git add backend/internal/model/agent_task.go backend/internal/model/agent_task_receipt.go backend/internal/domain/agent_task.go backend/internal/domain/run.go
git commit -m "feat(p1): add AgentTask, AgentTaskReceipt models and domain constants"
```

### Task 3: Repository 层

**Files:**
- Create: `backend/internal/repository/flow_draft_repository.go`
- Create: `backend/internal/repository/agent_task_repository.go`
- Create: `backend/internal/repository/agent_task_receipt_repository.go`

- [ ] **Step 1: 写 FlowDraftRepository**

```go
package repository

import (
	"context"

	"gorm.io/gorm"
	"the-line/backend/internal/model"
)

type FlowDraftRepository struct {
	db *gorm.DB
}

func NewFlowDraftRepository(database *gorm.DB) *FlowDraftRepository {
	return &FlowDraftRepository{db: database}
}

func (r *FlowDraftRepository) Create(ctx context.Context, draft *model.FlowDraft) error {
	return r.db.WithContext(ctx).Create(draft).Error
}

func (r *FlowDraftRepository) GetByID(ctx context.Context, id uint64) (*model.FlowDraft, error) {
	var draft model.FlowDraft
	if err := r.db.WithContext(ctx).First(&draft, id).Error; err != nil {
		return nil, err
	}
	return &draft, nil
}

func (r *FlowDraftRepository) List(ctx context.Context, creatorID uint64, status string, page, pageSize int) ([]model.FlowDraft, int64, error) {
	var items []model.FlowDraft
	var total int64
	query := r.db.WithContext(ctx).Model(&model.FlowDraft{})
	if creatorID > 0 {
		query = query.Where("creator_person_id = ?", creatorID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *FlowDraftRepository) Update(ctx context.Context, draft *model.FlowDraft) error {
	return r.db.WithContext(ctx).Save(draft).Error
}

func (r *FlowDraftRepository) UpdateStatus(ctx context.Context, id uint64, status string) error {
	return r.db.WithContext(ctx).Model(&model.FlowDraft{}).Where("id = ?", id).Update("status", status).Error
}
```

- [ ] **Step 2: 写 AgentTaskRepository**

```go
package repository

import (
	"context"

	"gorm.io/gorm"
	"the-line/backend/internal/model"
)

type AgentTaskRepository struct {
	db *gorm.DB
}

func NewAgentTaskRepository(database *gorm.DB) *AgentTaskRepository {
	return &AgentTaskRepository{db: database}
}

func (r *AgentTaskRepository) Create(ctx context.Context, task *model.AgentTask) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *AgentTaskRepository) GetByID(ctx context.Context, id uint64) (*model.AgentTask, error) {
	var task model.AgentTask
	if err := r.db.WithContext(ctx).First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *AgentTaskRepository) GetByRunNodeID(ctx context.Context, runNodeID uint64) (*model.AgentTask, error) {
	var task model.AgentTask
	if err := r.db.WithContext(ctx).Where("run_node_id = ?", runNodeID).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *AgentTaskRepository) ListByRunID(ctx context.Context, runID uint64) ([]model.AgentTask, error) {
	var tasks []model.AgentTask
	err := r.db.WithContext(ctx).Where("run_id = ?", runID).Order("created_at DESC").Find(&tasks).Error
	return tasks, err
}

func (r *AgentTaskRepository) Update(ctx context.Context, task *model.AgentTask) error {
	return r.db.WithContext(ctx).Save(task).Error
}

func (r *AgentTaskRepository) UpdateStatus(ctx context.Context, id uint64, status string) error {
	return r.db.WithContext(ctx).Model(&model.AgentTask{}).Where("id = ?", id).Update("status", status).Error
}
```

- [ ] **Step 3: 写 AgentTaskReceiptRepository**

```go
package repository

import (
	"context"

	"gorm.io/gorm"
	"the-line/backend/internal/model"
)

type AgentTaskReceiptRepository struct {
	db *gorm.DB
}

func NewAgentTaskReceiptRepository(database *gorm.DB) *AgentTaskReceiptRepository {
	return &AgentTaskReceiptRepository{db: database}
}

func (r *AgentTaskReceiptRepository) Create(ctx context.Context, receipt *model.AgentTaskReceipt) error {
	return r.db.WithContext(ctx).Create(receipt).Error
}

func (r *AgentTaskReceiptRepository) GetByTaskID(ctx context.Context, taskID uint64) (*model.AgentTaskReceipt, error) {
	var receipt model.AgentTaskReceipt
	if err := r.db.WithContext(ctx).Where("agent_task_id = ?", taskID).First(&receipt).Error; err != nil {
		return nil, err
	}
	return &receipt, nil
}

func (r *AgentTaskReceiptRepository) ListByRunID(ctx context.Context, runID uint64) ([]model.AgentTaskReceipt, error) {
	var receipts []model.AgentTaskReceipt
	err := r.db.WithContext(ctx).Where("run_id = ?", runID).Order("received_at DESC").Find(&receipts).Error
	return receipts, err
}
```

- [ ] **Step 4: Commit**

```bash
git add backend/internal/repository/flow_draft_repository.go backend/internal/repository/agent_task_repository.go backend/internal/repository/agent_task_receipt_repository.go
git commit -m "feat(p1): add repositories for FlowDraft, AgentTask, AgentTaskReceipt"
```

### Task 4: Service 基础层

**Files:**
- Create: `backend/internal/service/flow_draft_service.go`
- Create: `backend/internal/service/agent_task_service.go`
- Create: `backend/internal/service/agent_task_receipt_service.go`

- [ ] **Step 1: 写 FlowDraftService（基础 CRUD）**

```go
package service

import (
	"context"

	"the-line/backend/internal/domain"
	"the-line/backend/internal/dto"
	"the-line/backend/internal/model"
	"the-line/backend/internal/repository"
)

type FlowDraftService struct {
	repo *repository.FlowDraftRepository
}

func NewFlowDraftService(repo *repository.FlowDraftRepository) *FlowDraftService {
	return &FlowDraftService{repo: repo}
}

func (s *FlowDraftService) Create(ctx context.Context, req dto.CreateFlowDraftRequest, creatorID uint64) (*model.FlowDraft, error) {
	draft := &model.FlowDraft{
		Title:           req.Title,
		Description:     req.Description,
		SourcePrompt:    req.SourcePrompt,
		CreatorPersonID: creatorID,
		PlannerAgentID:  req.PlannerAgentID,
		Status:          domain.DraftStatusDraft,
		StructuredPlanJSON: req.StructuredPlanJSON,
	}
	if err := s.repo.Create(ctx, draft); err != nil {
		return nil, err
	}
	return draft, nil
}

func (s *FlowDraftService) GetByID(ctx context.Context, id uint64) (*model.FlowDraft, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *FlowDraftService) List(ctx context.Context, query dto.FlowDraftListQuery) ([]model.FlowDraft, int64, error) {
	page := query.Page
	if page <= 0 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	return s.repo.List(ctx, query.CreatorPersonID, query.Status, page, pageSize)
}

func (s *FlowDraftService) Update(ctx context.Context, id uint64, req dto.UpdateFlowDraftRequest) (*model.FlowDraft, error) {
	draft, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if draft.Status != domain.DraftStatusDraft {
		return nil, domain.ErrInvalidState
	}
	if req.Title != "" {
		draft.Title = req.Title
	}
	if req.Description != "" {
		draft.Description = req.Description
	}
	if len(req.StructuredPlanJSON) > 0 {
		draft.StructuredPlanJSON = req.StructuredPlanJSON
	}
	if err := s.repo.Update(ctx, draft); err != nil {
		return nil, err
	}
	return draft, nil
}
```

- [ ] **Step 2: 写 AgentTaskService（基础 CRUD）**

```go
package service

import (
	"context"

	"the-line/backend/internal/model"
	"the-line/backend/internal/repository"
)

type AgentTaskService struct {
	repo *repository.AgentTaskRepository
}

func NewAgentTaskService(repo *repository.AgentTaskRepository) *AgentTaskService {
	return &AgentTaskService{repo: repo}
}

func (s *AgentTaskService) Create(ctx context.Context, task *model.AgentTask) error {
	return s.repo.Create(ctx, task)
}

func (s *AgentTaskService) GetByID(ctx context.Context, id uint64) (*model.AgentTask, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *AgentTaskService) GetByRunNodeID(ctx context.Context, runNodeID uint64) (*model.AgentTask, error) {
	return s.repo.GetByRunNodeID(ctx, runNodeID)
}

func (s *AgentTaskService) ListByRunID(ctx context.Context, runID uint64) ([]model.AgentTask, error) {
	return s.repo.ListByRunID(ctx, runID)
}

func (s *AgentTaskService) Update(ctx context.Context, task *model.AgentTask) error {
	return s.repo.Update(ctx, task)
}

func (s *AgentTaskService) UpdateStatus(ctx context.Context, id uint64, status string) error {
	return s.repo.UpdateStatus(ctx, id, status)
}
```

- [ ] **Step 3: 写 AgentTaskReceiptService（基础 CRUD）**

```go
package service

import (
	"context"

	"the-line/backend/internal/model"
	"the-line/backend/internal/repository"
)

type AgentTaskReceiptService struct {
	repo *repository.AgentTaskReceiptRepository
}

func NewAgentTaskReceiptService(repo *repository.AgentTaskReceiptRepository) *AgentTaskReceiptService {
	return &AgentTaskReceiptService{repo: repo}
}

func (s *AgentTaskReceiptService) Create(ctx context.Context, receipt *model.AgentTaskReceipt) error {
	return s.repo.Create(ctx, receipt)
}

func (s *AgentTaskReceiptService) GetByTaskID(ctx context.Context, taskID uint64) (*model.AgentTaskReceipt, error) {
	return s.repo.GetByTaskID(ctx, taskID)
}

func (s *AgentTaskReceiptService) ListByRunID(ctx context.Context, runID uint64) ([]model.AgentTaskReceipt, error) {
	return s.repo.ListByRunID(ctx, runID)
}
```

- [ ] **Step 4: Commit**

```bash
git add backend/internal/service/flow_draft_service.go backend/internal/service/agent_task_service.go backend/internal/service/agent_task_receipt_service.go
git commit -m "feat(p1): add base services for FlowDraft, AgentTask, AgentTaskReceipt"
```

### Task 5: DTO + Response + Handler 注册

**Files:**
- Create: `backend/internal/dto/draft.go`
- Create: `backend/internal/dto/agent_task.go`
- Create: `backend/internal/handler/flow_draft_handler.go`
- Create: `backend/internal/handler/agent_task_handler.go`
- Modify: `backend/internal/db/migrate.go`
- Modify: `backend/internal/app/router.go`

- [ ] **Step 1: 写 draft DTO**

```go
package dto

import (
	"encoding/json"
)

type CreateFlowDraftRequest struct {
	Title              string          `json:"title"`
	Description        string          `json:"description"`
	SourcePrompt       string          `json:"source_prompt" binding:"required"`
	PlannerAgentID     *uint64         `json:"planner_agent_id"`
	StructuredPlanJSON json.RawMessage `json:"structured_plan_json"`
}

type UpdateFlowDraftRequest struct {
	Title              string          `json:"title"`
	Description        string          `json:"description"`
	StructuredPlanJSON json.RawMessage `json:"structured_plan_json"`
}

type FlowDraftListQuery struct {
	PageQuery
	CreatorPersonID uint64 `form:"creator_person_id"`
	Status          string `form:"status"`
}
```

- [ ] **Step 2: 写 agent_task DTO**

```go
package dto

import (
	"encoding/json"
	"time"
)

type AgentReceiptRequest struct {
	AgentID      uint64          `json:"agent_id" binding:"required"`
	Status       string          `json:"status" binding:"required,oneof=completed needs_review failed blocked"`
	StartedAt    *time.Time      `json:"started_at"`
	FinishedAt   *time.Time      `json:"finished_at"`
	Summary      string          `json:"summary"`
	Result       json.RawMessage `json:"result"`
	Artifacts    json.RawMessage `json:"artifacts"`
	Logs         []string        `json:"logs"`
	ErrorMessage string          `json:"error_message"`
}

type ConfirmAgentResultRequest struct {
	Action  string `json:"action" binding:"required,oneof=approve reject"`
	Comment string `json:"comment"`
}

type TakeoverNodeRequest struct {
	Action       string          `json:"action" binding:"required,oneof=retry manual_complete"`
	ManualResult json.RawMessage `json:"manual_result"`
}
```

- [ ] **Step 3: 写 FlowDraftHandler**

```go
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"the-line/backend/internal/domain"
	"the-line/backend/internal/dto"
	"the-line/backend/internal/response"
	"the-line/backend/internal/service"
)

type FlowDraftHandler struct {
	service *service.FlowDraftService
}

func NewFlowDraftHandler(svc *service.FlowDraftService) *FlowDraftHandler {
	return &FlowDraftHandler{service: svc}
}

func (h *FlowDraftHandler) Create(c *gin.Context) {
	var req dto.CreateFlowDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.Validation(err.Error()))
		return
	}
	actor := actorFromContext(c)
	draft, err := h.service.Create(c.Request.Context(), req, actor.PersonID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, draft)
}

func (h *FlowDraftHandler) GetByID(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	draft, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, response.NotFound("草案不存在"))
		return
	}
	response.OK(c, draft)
}

func (h *FlowDraftHandler) List(c *gin.Context) {
	var query dto.FlowDraftListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, response.Validation(err.Error()))
		return
	}
	actor := actorFromContext(c)
	if query.CreatorPersonID == 0 {
		query.CreatorPersonID = actor.PersonID
	}
	items, total, err := h.service.List(c.Request.Context(), query)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, items, total, query.Page, query.PageSize)
}

func (h *FlowDraftHandler) Update(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateFlowDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.Validation(err.Error()))
		return
	}
	draft, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, draft)
}
```

- [ ] **Step 4: 写 AgentTaskHandler（P1 只做查询）**

```go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"the-line/backend/internal/response"
	"the-line/backend/internal/service"
)

type AgentTaskHandler struct {
	agentTaskService        *service.AgentTaskService
	agentTaskReceiptService *service.AgentTaskReceiptService
}

func NewAgentTaskHandler(
	agentTaskService *service.AgentTaskService,
	agentTaskReceiptService *service.AgentTaskReceiptService,
) *AgentTaskHandler {
	return &AgentTaskHandler{
		agentTaskService:        agentTaskService,
		agentTaskReceiptService: agentTaskReceiptService,
	}
}

func (h *AgentTaskHandler) GetByID(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	task, err := h.agentTaskService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, response.NotFound("任务不存在"))
		return
	}
	response.OK(c, task)
}

func (h *AgentTaskHandler) ListByRunID(c *gin.Context) {
	runID, ok := parseIDParam(c, "runId")
	if !ok {
		return
	}
	tasks, err := h.agentTaskService.ListByRunID(c.Request.Context(), runID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, tasks)
}

func (h *AgentTaskHandler) GetReceipt(c *gin.Context) {
	taskID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	receipt, err := h.agentTaskReceiptService.GetByTaskID(c.Request.Context(), taskID)
	if err != nil {
		response.Error(c, response.NotFound("回执不存在"))
		return
	}
	response.OK(c, receipt)
}
```

- [ ] **Step 5: 修改 migrate.go 加入 AutoMigrate**

在 `backend/internal/db/migrate.go` 的 `AutoMigrate` 参数列表中追加：

```go
&model.FlowDraft{},
&model.AgentTask{},
&model.AgentTaskReceipt{},
```

- [ ] **Step 6: 修改 router.go 注册路由**

在 `backend/internal/app/router.go` 中：
1. 实例化 repository、service、handler
2. 注册路由（暂无需注册执行/回执接口，那些属于 P2）

```go
flowDraftRepo := repository.NewFlowDraftRepository(database)
agentTaskRepo := repository.NewAgentTaskRepository(database)
agentTaskReceiptRepo := repository.NewAgentTaskReceiptRepository(database)

flowDraftService := service.NewFlowDraftService(flowDraftRepo)
agentTaskService := service.NewAgentTaskService(agentTaskRepo)
agentTaskReceiptService := service.NewAgentTaskReceiptService(agentTaskReceiptRepo)

flowDraftHandler := handler.NewFlowDraftHandler(flowDraftService)
agentTaskHandler := handler.NewAgentTaskHandler(agentTaskService, agentTaskReceiptService)
```

在 `api` 路由组内新增路由组（仿照现有模式，注意路由注册在 `/api` 下而不是 `/api/v1`）：

```go
flowDrafts := api.Group("/flow-drafts")
{
    flowDrafts.POST("", flowDraftHandler.Create)
    flowDrafts.GET("", flowDraftHandler.List)
    flowDrafts.GET("/:id", flowDraftHandler.GetByID)
    flowDrafts.PUT("/:id", flowDraftHandler.Update)
}

agentTasks := api.Group("/agent-tasks")
{
    agentTasks.GET("/:id", agentTaskHandler.GetByID)
    agentTasks.GET("/run/:runId", agentTaskHandler.ListByRunID)
    agentTasks.GET("/:id/receipt", agentTaskHandler.GetReceipt)
}
```

- [ ] **Step 7: 编译测试**

Run: `cd backend && go build ./cmd/api`
Expected: 编译通过，无新增 error

- [ ] **Step 8: Commit**

```bash
git add backend/
git commit -m "feat(p1): add FlowDraft and AgentTask handlers, DTOs, routes and migrations"
```

---

## Chunk 2: P2 龙虾真实执行

> **依赖注意：** `agent_executor.go` 中的 `AgentPlannerExecutor` 接口依赖 `dto.DraftPlan`（在 Chunk 3 Task 1 中创建）。因此 **Chunk 3 Task 1（DraftPlan DTO）必须在 Chunk 2 Task 1 之前或同时完成**，否则 P2 无法编译。如果 P2 和 P3 并行实现，需要先提取 DraftPlan DTO 作为公共步骤。

### 文件结构

| 新建文件 | 职责 |
|----------|------|
| `backend/internal/executor/agent_executor.go` | AgentExecutor / AgentPlannerExecutor 接口 |
| `backend/internal/executor/mock_agent_executor.go` | MockAgentExecutor 实现（异步回调回执） |
| `backend/internal/executor/mock_agent_planner_executor.go` | MockAgentPlannerExecutor 实现 |

| 修改文件 | 修改内容 |
|----------|----------|
| `backend/internal/service/agent_task_service.go` | 新增 `CreateAndDispatch`、`ProcessReceipt`、`handleNodeTransition` |
| `backend/internal/service/run_node_service.go` | 自动节点触发 + 人工确认/接管 |
| `backend/internal/handler/agent_task_handler.go` | 新增 `Receipt` 回执接口 |
| `backend/internal/handler/run_node_handler.go` | 新增 `ConfirmAgentResult`、`TakeoverNode` |
| `backend/internal/app/router.go` | 注册回执、确认、接管路由 |

### Task 1: Executor 接口 + Mock 实现

**Files:**
- Create: `backend/internal/executor/agent_executor.go`
- Create: `backend/internal/executor/mock_agent_executor.go`

- [ ] **Step 1: 写 executor 接口**

```go
package executor

import (
	"context"

	"the-line/backend/internal/dto"
	"the-line/backend/internal/model"
)

type AgentExecutor interface {
	Execute(ctx context.Context, task *model.AgentTask, agent *model.Agent) error
}

type AgentPlannerExecutor interface {
	GenerateDraft(ctx context.Context, prompt string, agent *model.Agent) (*dto.DraftPlan, error)
}
```

- [ ] **Step 2: 写 MockAgentExecutor**

```go
package executor

import (
	"context"
	"encoding/json"
	"math/rand"
	"time"

	"the-line/backend/internal/domain"
	"the-line/backend/internal/dto"
	"the-line/backend/internal/model"
)

type MockAgentExecutor struct {
	receiptCallback func(taskID uint64, receipt *dto.AgentReceiptRequest) error
}

func NewMockAgentExecutor(callback func(taskID uint64, receipt *dto.AgentReceiptRequest) error) *MockAgentExecutor {
	return &MockAgentExecutor{receiptCallback: callback}
}

func (m *MockAgentExecutor) Execute(ctx context.Context, task *model.AgentTask, agent *model.Agent) error {
	go func() {
		delay := 2*time.Second + time.Duration(rand.Intn(1000))*time.Millisecond
		time.Sleep(delay)
		receipt := buildMockReceipt(task)
		_ = m.receiptCallback(task.ID, receipt)
	}()
	return nil
}

func buildMockReceipt(task *model.AgentTask) *dto.AgentReceiptRequest {
	now := time.Now()
	started := now.Add(-2 * time.Second)
	var result json.RawMessage
	switch task.TaskType {
	case domain.AgentTaskTypeQuery:
		result, _ = json.Marshal(map[string]any{
			"records_count": 12,
			"records": []map[string]any{
				{"id": 101, "name": "场次A", "video_bound": false},
			},
		})
	case domain.AgentTaskTypeBatchOperation:
		result, _ = json.Marshal(map[string]any{
			"success_count": 10,
			"failed_count":  2,
			"failed_ids":    []int{102, 103},
		})
	case domain.AgentTaskTypeExport:
		result, _ = json.Marshal(map[string]any{
			"file_url":  "/uploads/mock_export.xlsx",
			"file_name": "mock_export.xlsx",
		})
	}
	return &dto.AgentReceiptRequest{
		AgentID:    task.AgentID,
		Status:     domain.ReceiptStatusCompleted,
		StartedAt:  &started,
		FinishedAt: &now,
		Summary:    "Mock 执行完成",
		Result:     result,
		Artifacts:  []byte("[]"),
		Logs:       []string{"step1: mock", "step2: done"},
	}
}
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/executor/
git commit -m "feat(p2): add AgentExecutor interface and mock implementation"
```

### Task 2: AgentTaskService 增强 — 调度与回执处理

**Files:**
- Modify: `backend/internal/service/agent_task_service.go`

- [ ] **Step 1: 扩展 AgentTaskService 构造函数和字段**

将 `AgentTaskService` 改为：

```go
type AgentTaskService struct {
	db                *gorm.DB
	repo              *repository.AgentTaskRepository
	receiptRepo       *repository.AgentTaskReceiptRepository
	runNodeRepo       *repository.RunNodeRepository
	runService        *RunService
	nodeLogRepo       *repository.NodeLogRepository
	agentRepo         *repository.AgentRepository
	executor          executor.AgentExecutor
}

func NewAgentTaskService(
	db *gorm.DB,
	repo *repository.AgentTaskRepository,
	receiptRepo *repository.AgentTaskReceiptRepository,
	runNodeRepo *repository.RunNodeRepository,
	runService *RunService,
	nodeLogRepo *repository.NodeLogRepository,
	agentRepo *repository.AgentRepository,
	executor executor.AgentExecutor,
) *AgentTaskService {
	return &AgentTaskService{
		db:          db,
		repo:        repo,
		receiptRepo: receiptRepo,
		runNodeRepo: runNodeRepo,
		runService:  runService,
		nodeLogRepo: nodeLogRepo,
		agentRepo:   agentRepo,
		executor:    executor,
	}
}
```

- [ ] **Step 2: 添加 CreateAndDispatch 方法**

```go
func (s *AgentTaskService) CreateAndDispatch(ctx context.Context, node *model.FlowRunNode) error {
	task := &model.AgentTask{
		RunID:     node.RunID,
		RunNodeID: node.ID,
		AgentID:   *node.BoundAgentID,
		TaskType:  inferTaskType(node.NodeType, node.ConfigJSON),
		InputJSON: node.InputJSON,
		Status:    domain.AgentTaskStatusQueued,
	}
	if err := s.repo.Create(ctx, task); err != nil {
		return err
	}

	now := time.Now()
	updates := map[string]any{
		"status":     domain.NodeStatusRunning,
		"started_at": &now,
	}
	if err := s.runNodeRepo.UpdateWithDB(s.db, ctx, node.ID, updates); err != nil {
		return err
	}

	_ = s.appendLog(ctx, node.ID, domain.LogTypeAgentRun, "龙虾任务已创建，开始执行")

	agent, err := s.agentRepo.GetByID(ctx, *node.BoundAgentID)
	if err != nil {
		return err
	}

	task.Status = domain.AgentTaskStatusRunning
	task.StartedAt = &now
	if err := s.repo.Update(ctx, task); err != nil {
		return err
	}

	return s.executor.Execute(ctx, task, agent)
}

func inferTaskType(nodeType string, config datatypes.JSON) string {
	if len(config) > 0 {
		var cfg map[string]any
		_ = json.Unmarshal(config, &cfg)
		if v, ok := cfg["task_type"].(string); ok && v != "" {
			return v
		}
	}
	switch nodeType {
	case domain.NodeTypeAgentExport:
		return domain.AgentTaskTypeExport
	case domain.NodeTypeAgentExecute:
		return domain.AgentTaskTypeQuery
	}
	return domain.AgentTaskTypeQuery
}

func (s *AgentTaskService) appendLog(ctx context.Context, nodeID uint64, logType, content string) error {
	logEntry := &model.FlowRunNodeLog{
		RunNodeID:    nodeID,
		LogType:      logType,
		Content:      content,
		OperatorType: domain.OperatorTypeAgent,
	}
	return s.nodeLogRepo.Create(ctx, logEntry)
}
```

> 注意：`runNodeRepo.UpdateWithDB` 的签名需对齐现有 `run_node_repository.go` 的 `UpdateWithDB(ctx context.Context, database *gorm.DB, id uint64, updates map[string]any) error`。如果不存在，使用 `Update(ctx context.Context, id uint64, updates map[string]any) error` 或直接修改 repository 增加该方法。

- [ ] **Step 3: 添加 ProcessReceipt 方法**

```go
func (s *AgentTaskService) ProcessReceipt(ctx context.Context, taskID uint64, req *dto.AgentReceiptRequest) error {
	task, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}
	if task.AgentID != req.AgentID {
		return response.Validation("agent_id 不匹配")
	}
	if task.Status != domain.AgentTaskStatusRunning {
		return response.InvalidState("任务不在执行中状态")
	}

	receipt := &model.AgentTaskReceipt{
		AgentTaskID:   taskID,
		RunID:         task.RunID,
		RunNodeID:     task.RunNodeID,
		AgentID:       req.AgentID,
		ReceiptStatus: req.Status,
		PayloadJSON:   mustJSON(req),
		ReceivedAt:    time.Now(),
	}
	if err := s.receiptRepo.Create(ctx, receipt); err != nil {
		return err
	}

	task.Status = mapReceiptStatusToTaskStatus(req.Status)
	task.FinishedAt = req.FinishedAt
	task.ResultJSON = req.Result
	task.ArtifactsJSON = req.Artifacts
	task.ErrorMessage = req.ErrorMessage
	if err := s.repo.Update(ctx, task); err != nil {
		return err
	}

	_ = s.appendLog(ctx, task.RunNodeID, domain.LogTypeAgentRun, req.Summary)

	return s.handleNodeTransition(ctx, task, req)
}

func mapReceiptStatusToTaskStatus(receiptStatus string) string {
	switch receiptStatus {
	case domain.ReceiptStatusCompleted:
		return domain.AgentTaskStatusCompleted
	case domain.ReceiptStatusNeedsReview:
		return domain.AgentTaskStatusNeedsReview
	case domain.ReceiptStatusFailed:
		return domain.AgentTaskStatusFailed
	case domain.ReceiptStatusBlocked:
		return domain.AgentTaskStatusFailed
	}
	return domain.AgentTaskStatusFailed
}
```

- [ ] **Step 4: 添加 handleNodeTransition 方法**

```go
func (s *AgentTaskService) handleNodeTransition(ctx context.Context, task *model.AgentTask, req *dto.AgentReceiptRequest) error {
	nodeID := task.RunNodeID
	runID := task.RunID

	switch req.Status {
	case domain.ReceiptStatusCompleted:
		err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&model.FlowRunNode{}).Where("id = ?", nodeID).Updates(map[string]any{
				"status":       domain.NodeStatusDone,
				"output_json":  task.ResultJSON,
				"completed_at": time.Now(),
			}).Error; err != nil {
				return err
			}
			return s.runService.AdvanceAfterNodeDone(tx, runID, nodeID, domain.Actor{})
		})
		return err

	case domain.ReceiptStatusNeedsReview:
		return s.runNodeRepo.UpdateWithDB(s.db, ctx, nodeID, map[string]any{
			"status":      domain.NodeStatusWaitConfirm,
			"output_json": task.ResultJSON,
		})

	case domain.ReceiptStatusFailed, domain.ReceiptStatusBlocked:
		nodeStatus := domain.NodeStatusFailed
		if req.Status == domain.ReceiptStatusBlocked {
			nodeStatus = domain.NodeStatusBlocked
		}
		return s.runNodeRepo.UpdateWithDB(s.db, ctx, nodeID, map[string]any{
			"status": nodeStatus,
		})
	}
	return nil
}
```

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/agent_task_service.go
git commit -m "feat(p2): add AgentTask create/dispatch and receipt processing logic"
```

### Task 3: RunNodeService — 自动触发 + 确认 + 接管

**Files:**
- Modify: `backend/internal/service/run_node_service.go`

- [ ] **Step 1: 在 RunNodeService 构造函数中注入 AgentTaskService**

在 `RunNodeService` struct 中新增字段 `agentTaskService *AgentTaskService`，并在 `NewRunNodeService` 中接收该参数。

> 这需要在 `router.go` 中调整初始化顺序：先创建 `AgentTaskService`，再创建 `RunNodeService`。

- [ ] **Step 2: 在 AdvanceNode 或流程推进处增加自动触发**

找到 `RunService.AdvanceAfterNodeDone` 执行后的逻辑。在 `nextNode` 被设置为 `ready` 后，调用自动触发。最简单的方式是在 `RunService` 中直接调用 `agentTaskService.CreateAndDispatch`（需要通过注入字段），或者在 `RunNodeService` 的方法链中触发。

推荐做法：在 `RunNodeService` 的 `Approve` / `Complete` 成功后，且 `AdvanceAfterNodeDone` 执行完毕后，检查新节点的状态是否为 `ready` 且是 agent 节点，若是则触发。

但为了保持 DRY，更好的做法是在 `RunService` 中新增一个 `AdvanceAfterNodeDone` 的封装方法：

```go
func (s *RunService) advanceAfterNodeDoneWithAgentTrigger(tx *gorm.DB, runID uint64, nodeID uint64, actor domain.Actor) error {
	if err := s.AdvanceAfterNodeDone(tx, runID, nodeID, actor); err != nil {
		return err
	}
	// 触发检查由外部 service 层在 tx 提交后异步进行
	return nil
}
```

更实际的做法：修改 `RunService.AdvanceAfterNodeDone` 的末尾，在更新 `nextNode` 为 `ready` 之后，增加一个 post-commit hook 机制，或者在事务提交后由上层判断。

对于 MVP，采用最简单的方式：保持 `RunService.AdvanceAfterNodeDone` 不变，在 `RunNodeService` 的 `Approve`、`Complete` 方法中，事务提交后检查当前 Run 的最新节点并触发。

一个小技巧：GORM 的 `Transaction` 回调结束后，我们再手动检查：

```go
func (s *RunNodeService) Approve(...) (dto.RunNodeDetailResponse, error) {
    // ... 现有事务逻辑 ...
    
    // 事务提交后触发自动节点
    if err := s.maybeTriggerAgent(ctx, node.RunID); err != nil {
        // 触发失败不打断审批成功，仅记录
        _ = s.appendLog(ctx, node.ID, domain.LogTypeSystem, "自动触发失败: "+err.Error())
    }
    return detail, nil
}

func (s *RunNodeService) maybeTriggerAgent(ctx context.Context, runID uint64) error {
    run, err := s.runRepo.GetByID(ctx, runID)
    if err != nil {
        return err
    }
    if run.CurrentStatus != domain.RunStatusRunning || run.CurrentNodeCode == "" {
        return nil
    }
    nodes, err := s.runNodeRepo.ListByRunID(ctx, runID)
    if err != nil {
        return err
    }
    for _, n := range nodes {
        if n.NodeCode == run.CurrentNodeCode && n.Status == domain.NodeStatusReady {
            if isAgentNode(n.NodeType) && n.BoundAgentID != nil {
                return s.agentTaskService.CreateAndDispatch(ctx, &n)
            }
        }
    }
    return nil
}
```

- [ ] **Step 3: 添加 ConfirmAgentResult 方法**

```go
func (s *RunNodeService) ConfirmAgentResult(ctx context.Context, nodeID uint64, req dto.ConfirmAgentResultRequest, actor domain.Actor) (dto.RunNodeDetailResponse, error) {
	node, err := s.runNodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return dto.RunNodeDetailResponse{}, response.NotFound("节点不存在")
	}
	if node.Status != domain.NodeStatusWaitConfirm {
		return dto.RunNodeDetailResponse{}, response.InvalidState("节点不在待确认状态")
	}
	if !canOperateNode(*node, actor, true) {
		return dto.RunNodeDetailResponse{}, response.Forbidden("无权操作")
	}

	if req.Action == "approve" {
		err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			now := time.Now()
			if err := tx.Model(&model.FlowRunNode{}).Where("id = ?", nodeID).Updates(map[string]any{
				"status":       domain.NodeStatusDone,
				"completed_at": &now,
			}).Error; err != nil {
				return err
			}
			return s.runService.AdvanceAfterNodeDone(tx, node.RunID, nodeID, actor)
		})
		if err != nil {
			return dto.RunNodeDetailResponse{}, err
		}
		_ = s.appendLog(ctx, nodeID, domain.LogTypeApprove, actor.PersonID, req.Comment)
		if err := s.maybeTriggerAgent(ctx, node.RunID); err != nil {
			_ = s.appendLog(ctx, nodeID, domain.LogTypeSystem, 0, "自动触发失败: "+err.Error())
		}
	} else {
		err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			return tx.Model(&model.FlowRunNode{}).Where("id = ?", nodeID).Update("status", domain.NodeStatusFailed).Error
		})
		if err != nil {
			return dto.RunNodeDetailResponse{}, err
		}
		_ = s.appendLog(ctx, nodeID, domain.LogTypeReject, actor.PersonID, req.Comment)
	}

	return s.Detail(ctx, nodeID, actor)
}
```

> 注意：`appendLog` 需要新的重载或者直接写一个私有 helper。

- [ ] **Step 4: 添加 TakeoverNode 方法**

```go
func (s *RunNodeService) TakeoverNode(ctx context.Context, nodeID uint64, req dto.TakeoverNodeRequest, actor domain.Actor) (dto.RunNodeDetailResponse, error) {
	node, err := s.runNodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return dto.RunNodeDetailResponse{}, response.NotFound("节点不存在")
	}
	if node.Status != domain.NodeStatusFailed && node.Status != domain.NodeStatusBlocked {
		return dto.RunNodeDetailResponse{}, response.InvalidState("节点不在可接管状态")
	}
	if !canOperateNode(*node, actor, false) {
		return dto.RunNodeDetailResponse{}, response.Forbidden("无权操作")
	}

	if req.Action == "retry" {
		if node.BoundAgentID != nil {
			if err := s.agentTaskService.CreateAndDispatch(ctx, node); err != nil {
				return dto.RunNodeDetailResponse{}, err
			}
		} else {
			return dto.RunNodeDetailResponse{}, response.Validation("该节点未绑定 Agent，无法重试")
		}
	} else if req.Action == "manual_complete" {
		err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			now := time.Now()
			updates := map[string]any{
				"status":       domain.NodeStatusDone,
				"completed_at": &now,
			}
			if len(req.ManualResult) > 0 {
				updates["output_json"] = req.ManualResult
			}
			if err := tx.Model(&model.FlowRunNode{}).Where("id = ?", nodeID).Updates(updates).Error; err != nil {
				return err
			}
			return s.runService.AdvanceAfterNodeDone(tx, node.RunID, nodeID, actor)
		})
		if err != nil {
			return dto.RunNodeDetailResponse{}, err
		}
		_ = s.appendLog(ctx, nodeID, domain.LogTypeComplete, actor.PersonID, "人工接管完成")
		if err := s.maybeTriggerAgent(ctx, node.RunID); err != nil {
			_ = s.appendLog(ctx, nodeID, domain.LogTypeSystem, 0, "自动触发失败: "+err.Error())
		}
	}

	return s.Detail(ctx, nodeID, actor)
}
```

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/run_node_service.go
git commit -m "feat(p2): add auto agent trigger, confirm and takeover logic"
```

### Task 4: Handler 和路由注册

**Files:**
- Modify: `backend/internal/handler/agent_task_handler.go`
- Modify: `backend/internal/handler/run_node_handler.go`
- Modify: `backend/internal/app/router.go`

- [ ] **Step 1: AgentTaskHandler 增加 Receipt 方法**

```go
func (h *AgentTaskHandler) Receipt(c *gin.Context) {
	taskID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req dto.AgentReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.Validation(err.Error()))
		return
	}
	if err := h.agentTaskService.ProcessReceipt(c.Request.Context(), taskID, &req); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"message": "回执已处理"})
}
```

- [ ] **Step 2: RunNodeHandler 增加 ConfirmAgentResult 和 Takeover 方法**

```go
func (h *RunNodeHandler) ConfirmAgentResult(c *gin.Context) {
	nodeID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req dto.ConfirmAgentResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.Validation(err.Error()))
		return
	}
	actor := actorFromContext(c)
	detail, err := h.runNodeService.ConfirmAgentResult(c.Request.Context(), nodeID, req, actor)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, detail)
}

func (h *RunNodeHandler) TakeoverNode(c *gin.Context) {
	nodeID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req dto.TakeoverNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.Validation(err.Error()))
		return
	}
	actor := actorFromContext(c)
	detail, err := h.runNodeService.TakeoverNode(c.Request.Context(), nodeID, req, actor)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, detail)
}
```

- [ ] **Step 3: Router 注册新路由**

在已有 `runNodes` 路由组追加（若还没有则插入）：

```go
runNodes.POST("/:id/confirm-agent-result", runNodeHandler.ConfirmAgentResult)
runNodes.POST("/:id/takeover", runNodeHandler.TakeoverNode)
```

在 `agentTasks` 路由组追加：

```go
agentTasks.POST("/:id/receipt", agentTaskHandler.Receipt)
```

在 `router.go` 中实例化 `MockAgentExecutor` 并传给 `AgentTaskService`：

```go
agentTaskService := service.NewAgentTaskService(
    database,
    agentTaskRepo,
    agentTaskReceiptRepo,
    runNodeRepo,
    runService,
    nodeLogRepo,
    agentRepo,
    executor.NewMockAgentExecutor(func(taskID uint64, receipt *dto.AgentReceiptRequest) error {
        // 直接调用 service 处理回执
        return agentTaskService.ProcessReceipt(context.Background(), taskID, receipt)
    }),
)
```

- [ ] **Step 4: 编译测试**

Run: `cd backend && go build ./cmd/api`
Expected: 编译通过

- [ ] **Step 5: Commit**

```bash
git add backend/
git commit -m "feat(p2): add receipt handler, confirm/takeover endpoints and mock executor wiring"
```

---

## Chunk 3: P3 龙虾辅助编排

### 文件结构

| 新建文件 | 职责 |
|----------|------|
| `backend/internal/executor/mock_agent_planner_executor.go` | MockAgentPlannerExecutor 实现 |
| `backend/internal/dto/draft_plan.go` | DraftPlan / DraftNode 结构体 |

| 修改文件 | 修改内容 |
|----------|----------|
| `backend/internal/service/flow_draft_service.go` | 注入 planner executor + template repo，新增 `CreateWithPlanner`、`Confirm`、`Discard` |
| `backend/internal/repository/template_repository.go` | 新增 `CreateWithDB`、`CreateNodeBatchWithDB` 方法 |
| `backend/internal/handler/flow_draft_handler.go` | 新增 `Confirm`、`Discard` 方法 |
| `backend/internal/app/router.go` | 注册 confirm/discard 路由，注入 planner executor |

### Task 1: DraftPlan DTO

**Files:**
- Create: `backend/internal/dto/draft_plan.go`

- [ ] **Step 1: 写 DraftPlan / DraftNode 结构体**

```go
package dto

import "encoding/json"

type DraftPlan struct {
	Title            string      `json:"title"`
	Description      string      `json:"description"`
	Nodes            []DraftNode `json:"nodes"`
	FinalDeliverable string      `json:"final_deliverable"`
}

type DraftNode struct {
	NodeCode            string          `json:"node_code"`
	NodeName            string          `json:"node_name"`
	NodeType            string          `json:"node_type"`
	SortOrder           int             `json:"sort_order"`
	ExecutorType        string          `json:"executor_type"`
	OwnerRule           string          `json:"owner_rule"`
	OwnerPersonID       *uint64         `json:"owner_person_id"`
	ExecutorAgentCode   string          `json:"executor_agent_code"`
	ResultOwnerRule     string          `json:"result_owner_rule"`
	ResultOwnerPersonID *uint64         `json:"result_owner_person_id"`
	TaskType            string          `json:"task_type"`
	InputSchema         json.RawMessage `json:"input_schema"`
	OutputSchema        json.RawMessage `json:"output_schema"`
	CompletionCondition string          `json:"completion_condition"`
	FailureCondition    string          `json:"failure_condition"`
	EscalationRule      string          `json:"escalation_rule"`
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/dto/draft_plan.go
git commit -m "feat(p3): add DraftPlan and DraftNode DTO structs"
```

### Task 2: MockAgentPlannerExecutor

**Files:**
- Create: `backend/internal/executor/mock_agent_planner_executor.go`

- [ ] **Step 1: 写 MockAgentPlannerExecutor**

```go
package executor

import (
	"context"
	"strings"

	"the-line/backend/internal/dto"
	"the-line/backend/internal/model"
)

type MockAgentPlannerExecutor struct{}

func NewMockAgentPlannerExecutor() *MockAgentPlannerExecutor {
	return &MockAgentPlannerExecutor{}
}

func (m *MockAgentPlannerExecutor) GenerateDraft(ctx context.Context, prompt string, agent *model.Agent) (*dto.DraftPlan, error) {
	plan := &dto.DraftPlan{
		Title:            extractTitle(prompt),
		Description:      prompt,
		FinalDeliverable: "流程执行结果报表",
	}

	order := 1

	// 默认总是生成一个查询节点，确保草案至少有 3 个节点（query + review + acceptance）
	// 满足 validateDraftPlan 的 3-8 节点要求
	plan.Nodes = append(plan.Nodes, dto.DraftNode{
		NodeCode:            "collect_data",
		NodeName:            "收集数据",
		NodeType:            "agent_execute",
		SortOrder:           order,
		ExecutorType:        "agent",
		OwnerRule:           "initiator",
		ExecutorAgentCode:   "data_query_agent",
		ResultOwnerRule:     "initiator",
		TaskType:            "query",
		InputSchema:         []byte("{}"),
		OutputSchema:        []byte(`{"fields":["records_count","records"]}`),
		CompletionCondition: "返回查询结果",
		FailureCondition:    "查询超时或无权限",
		EscalationRule:      "通知发起人人工处理",
	})
	order++

	plan.Nodes = append(plan.Nodes, dto.DraftNode{
		NodeCode:        "review_data",
		NodeName:        "审核确认数据",
		NodeType:        "human_review",
		SortOrder:       order,
		ExecutorType:    "human",
		OwnerRule:       "initiator",
		ResultOwnerRule: "initiator",
		InputSchema:     []byte("{}"),
		OutputSchema:    []byte("{}"),
	})
	order++

	if containsAny(prompt, "绑定", "操作", "批量") {
		plan.Nodes = append(plan.Nodes, dto.DraftNode{
			NodeCode:            "batch_op",
			NodeName:            "执行批量操作",
			NodeType:            "agent_execute",
			SortOrder:           order,
			ExecutorType:        "agent",
			OwnerRule:           "initiator",
			ExecutorAgentCode:   "batch_op_agent",
			ResultOwnerRule:     "initiator",
			TaskType:            "batch_operation",
			InputSchema:         []byte("{}"),
			OutputSchema:        []byte(`{"fields":["success_count","failed_count"]}`),
			CompletionCondition: "操作完成",
			FailureCondition:    "失败数超过阈值",
			EscalationRule:      "通知发起人确认异常项",
		})
		order++
	}

	if containsAny(prompt, "导出", "报表") {
		plan.Nodes = append(plan.Nodes, dto.DraftNode{
			NodeCode:            "export_result",
			NodeName:            "导出结果",
			NodeType:            "agent_export",
			SortOrder:           order,
			ExecutorType:        "agent",
			OwnerRule:           "initiator",
			ExecutorAgentCode:   "export_agent",
			ResultOwnerRule:     "initiator",
			TaskType:            "export",
			InputSchema:         []byte("{}"),
			OutputSchema:        []byte(`{"fields":["file_url","file_name"]}`),
			CompletionCondition: "文件生成成功",
			FailureCondition:    "导出失败",
			EscalationRule:      "通知发起人重试",
		})
		order++
	}

	plan.Nodes = append(plan.Nodes, dto.DraftNode{
		NodeCode:            "final_acceptance",
		NodeName:            "确认最终结果",
		NodeType:            "human_acceptance",
		SortOrder:           order,
		ExecutorType:        "human",
		OwnerRule:           "specified_person",
		ResultOwnerRule:     "specified_person",
		ResultOwnerPersonID: nil,
		InputSchema:         []byte("{}"),
		OutputSchema:        []byte("{}"),
	})

	return plan, nil
}

func extractTitle(prompt string) string {
	if len([]rune(prompt)) > 20 {
		return string([]rune(prompt)[:20]) + "..."
	}
	return prompt
}

func containsAny(s string, keywords ...string) bool {
	for _, kw := range keywords {
		if strings.Contains(s, kw) {
			return true
		}
	}
	return false
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/executor/mock_agent_planner_executor.go
git commit -m "feat(p3): add MockAgentPlannerExecutor with keyword-based draft generation"
```

### Task 3: TemplateRepository — 新增 Create 方法

**Files:**
- Modify: `backend/internal/repository/template_repository.go`

- [ ] **Step 1: 在 TemplateRepository 中新增 CreateWithDB 和 CreateNodeWithDB**

在 `template_repository.go` 末尾追加：

```go
func (r *TemplateRepository) CreateWithDB(ctx context.Context, database *gorm.DB, template *model.FlowTemplate) error {
	return database.WithContext(ctx).Create(template).Error
}

func (r *TemplateRepository) CreateNodeBatchWithDB(ctx context.Context, database *gorm.DB, nodes []model.FlowTemplateNode) error {
	if len(nodes) == 0 {
		return nil
	}
	return database.WithContext(ctx).Create(&nodes).Error
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/repository/template_repository.go
git commit -m "feat(p3): add Create methods to TemplateRepository for draft-to-template conversion"
```

### Task 4: FlowDraftService — 扩展为完整编排 service

**Files:**
- Modify: `backend/internal/service/flow_draft_service.go`

- [ ] **Step 1: 扩展 FlowDraftService 构造函数**

替换 `FlowDraftService` struct 和 constructor：

```go
type FlowDraftService struct {
	db              *gorm.DB
	repo            *repository.FlowDraftRepository
	templateRepo    *repository.TemplateRepository
	agentRepo       *repository.AgentRepository
	plannerExecutor executor.AgentPlannerExecutor
}

func NewFlowDraftService(
	db *gorm.DB,
	repo *repository.FlowDraftRepository,
	templateRepo *repository.TemplateRepository,
	agentRepo *repository.AgentRepository,
	plannerExecutor executor.AgentPlannerExecutor,
) *FlowDraftService {
	return &FlowDraftService{
		db:              db,
		repo:            repo,
		templateRepo:    templateRepo,
		agentRepo:       agentRepo,
		plannerExecutor: plannerExecutor,
	}
}
```

> 注意：P1 中的 `NewFlowDraftService` 只注入了 `repo`。此步骤更新签名。同步更新 `router.go` 中的调用。

- [ ] **Step 2: 改写 Create 方法为 CreateWithPlanner**

替换原有 `Create` 方法：

```go
func (s *FlowDraftService) Create(ctx context.Context, req dto.CreateFlowDraftRequest, creatorID uint64) (*model.FlowDraft, error) {
	var planJSON datatypes.JSON

	if len(req.StructuredPlanJSON) > 0 {
		planJSON = datatypes.JSON(req.StructuredPlanJSON)
	} else {
		var agent *model.Agent
		if req.PlannerAgentID != nil {
			var err error
			agent, err = s.agentRepo.GetByID(ctx, *req.PlannerAgentID)
			if err != nil {
				return nil, response.NotFound("编排龙虾不存在")
			}
		}

		plan, err := s.plannerExecutor.GenerateDraft(ctx, req.SourcePrompt, agent)
		if err != nil {
			return nil, response.Internal("龙虾生成草案失败: " + err.Error())
		}

		planBytes, _ := json.Marshal(plan)
		planJSON = datatypes.JSON(planBytes)

		if req.Title == "" {
			req.Title = plan.Title
		}
		if req.Description == "" {
			req.Description = plan.Description
		}
	}

	draft := &model.FlowDraft{
		Title:              req.Title,
		Description:        req.Description,
		SourcePrompt:       req.SourcePrompt,
		CreatorPersonID:    creatorID,
		PlannerAgentID:     req.PlannerAgentID,
		Status:             domain.DraftStatusDraft,
		StructuredPlanJSON: planJSON,
	}
	if err := s.repo.Create(ctx, draft); err != nil {
		return nil, err
	}
	return draft, nil
}
```

- [ ] **Step 3: 添加 Confirm 方法**

```go
func (s *FlowDraftService) Confirm(ctx context.Context, draftID uint64, personID uint64) (*model.FlowTemplate, error) {
	draft, err := s.repo.GetByID(ctx, draftID)
	if err != nil {
		return nil, response.NotFound("草案不存在")
	}
	if draft.Status != domain.DraftStatusDraft {
		return nil, response.InvalidState("草案状态不允许确认")
	}

	var plan dto.DraftPlan
	if err := json.Unmarshal(draft.StructuredPlanJSON, &plan); err != nil {
		return nil, response.Validation("草案数据格式错误")
	}

	if err := validateDraftPlan(&plan); err != nil {
		return nil, err
	}

	var template model.FlowTemplate
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		template = model.FlowTemplate{
			Name:        plan.Title,
			Code:        generateTemplateCode(),
			Version:     1,
			Category:    "ai_generated",
			Description: plan.Description,
			Status:      domain.TemplateStatusPublished,
		}
		if err := s.templateRepo.CreateWithDB(ctx, tx, &template); err != nil {
			return err
		}

		nodes := make([]model.FlowTemplateNode, 0, len(plan.Nodes))
		for _, n := range plan.Nodes {
			node := model.FlowTemplateNode{
				TemplateID:       template.ID,
				NodeCode:         n.NodeCode,
				NodeName:         n.NodeName,
				NodeType:         n.NodeType,
				SortOrder:        n.SortOrder,
				DefaultOwnerRule:    n.OwnerRule,
				DefaultAgentID:      s.resolveAgentID(ctx, n.ExecutorAgentCode),
				ResultOwnerRule:     n.ResultOwnerRule,
				ResultOwnerPersonID: n.ResultOwnerPersonID,
				InputSchemaJSON:     datatypes.JSON(n.InputSchema),
				OutputSchemaJSON:    datatypes.JSON(n.OutputSchema),
				ConfigJSON:          buildNodeConfig(n),
			}
			nodes = append(nodes, node)
		}
		if err := s.templateRepo.CreateNodeBatchWithDB(ctx, tx, nodes); err != nil {
			return err
		}

		now := time.Now()
		// 使用 tx 而非 s.repo.Update，确保 draft 状态更新在同一事务内
		return tx.WithContext(ctx).Model(&model.FlowDraft{}).Where("id = ?", draftID).Updates(map[string]any{
			"status":                domain.DraftStatusConfirmed,
			"confirmed_template_id": template.ID,
			"confirmed_at":          now,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return &template, nil
}
```

- [ ] **Step 4: 添加辅助方法 validateDraftPlan、generateTemplateCode、resolveAgentID、buildNodeConfig**

```go
func validateDraftPlan(plan *dto.DraftPlan) error {
	if len(plan.Nodes) < 3 || len(plan.Nodes) > 8 {
		return response.Validation("节点数量必须在 3-8 个之间")
	}

	hasHumanReview := false
	for _, n := range plan.Nodes {
		if n.NodeType == domain.NodeTypeHumanReview || n.NodeType == domain.NodeTypeHumanAcceptance {
			hasHumanReview = true
			break
		}
	}
	if !hasHumanReview {
		return response.Validation("必须至少包含一个人工确认节点")
	}

	last := plan.Nodes[len(plan.Nodes)-1]
	if last.NodeType != domain.NodeTypeHumanAcceptance {
		return response.Validation("最后一个节点必须是最终签收节点")
	}

	validTaskTypes := map[string]bool{
		domain.AgentTaskTypeQuery:          true,
		domain.AgentTaskTypeBatchOperation: true,
		domain.AgentTaskTypeExport:         true,
	}
	for _, n := range plan.Nodes {
		if n.ExecutorType == "agent" && !validTaskTypes[n.TaskType] {
			return response.Validation("节点 " + n.NodeCode + " 的任务类型不在允许范围内")
		}
	}

	return nil
}

func generateTemplateCode() string {
	return "ai_" + fmt.Sprintf("%d", time.Now().UnixMilli())
}

func (s *FlowDraftService) resolveAgentID(ctx context.Context, agentCode string) *uint64 {
	if agentCode == "" {
		return nil
	}
	// MVP: 查找第一个可用 agent 作为默认绑定
	// 这样草案确认后创建的模板节点能有 DefaultAgentID，P2 的自动触发才能生效
	agents, _, _ := s.agentRepo.List(ctx, repository.AgentListFilter{Limit: 1})
	if len(agents) > 0 {
		id := agents[0].ID
		return &id
	}
	return nil
}

func buildNodeConfig(n dto.DraftNode) datatypes.JSON {
	cfg := map[string]any{}
	if n.TaskType != "" {
		cfg["task_type"] = n.TaskType
	}
	if n.CompletionCondition != "" {
		cfg["completion_condition"] = n.CompletionCondition
	}
	if n.FailureCondition != "" {
		cfg["failure_condition"] = n.FailureCondition
	}
	if n.EscalationRule != "" {
		cfg["escalation_rule"] = n.EscalationRule
	}
	if n.ExecutorAgentCode != "" {
		cfg["executor_agent_code"] = n.ExecutorAgentCode
	}
	data, _ := json.Marshal(cfg)
	return datatypes.JSON(data)
}
```

- [ ] **Step 5: 添加 Discard 方法**

```go
func (s *FlowDraftService) Discard(ctx context.Context, draftID uint64) error {
	draft, err := s.repo.GetByID(ctx, draftID)
	if err != nil {
		return response.NotFound("草案不存在")
	}
	if draft.Status != domain.DraftStatusDraft {
		return response.InvalidState("草案状态不允许废弃")
	}
	return s.repo.UpdateStatus(ctx, draftID, domain.DraftStatusDiscarded)
}
```

- [ ] **Step 6: Commit**

```bash
git add backend/internal/service/flow_draft_service.go
git commit -m "feat(p3): add draft planner creation, confirm-to-template, discard logic"
```

### Task 5: FlowDraftHandler — Confirm/Discard + Router 更新

**Files:**
- Modify: `backend/internal/handler/flow_draft_handler.go`
- Modify: `backend/internal/app/router.go`

- [ ] **Step 1: FlowDraftHandler 增加 Confirm 和 Discard 方法**

```go
func (h *FlowDraftHandler) Confirm(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	actor := actorFromContext(c)
	template, err := h.service.Confirm(c.Request.Context(), id, actor.PersonID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, gin.H{
		"draft_id":    id,
		"template_id": template.ID,
		"message":     "草案已确认，模板已创建",
	})
}

func (h *FlowDraftHandler) Discard(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.service.Discard(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"message": "草案已废弃"})
}
```

- [ ] **Step 2: 更新 router.go — 注册 confirm/discard 路由**

在 `flowDrafts` 路由组追加：

```go
flowDrafts.POST("/:id/confirm", flowDraftHandler.Confirm)
flowDrafts.POST("/:id/discard", flowDraftHandler.Discard)
```

- [ ] **Step 3: 更新 router.go — 调整 FlowDraftService 初始化**

```go
plannerExecutor := executor.NewMockAgentPlannerExecutor()

flowDraftService := service.NewFlowDraftService(
	database,
	flowDraftRepo,
	templateRepo,
	agentRepo,
	plannerExecutor,
)
```

- [ ] **Step 4: 编译测试**

Run: `cd backend && go build ./cmd/api`
Expected: 编译通过

- [ ] **Step 5: Commit**

```bash
git add backend/
git commit -m "feat(p3): add confirm/discard handler, wire planner executor in router"
```

---

## Chunk 4: P4 前端集成

### 文件结构

| 新建文件 | 职责 |
|----------|------|
| `frontend/src/api/drafts.ts` | 草案 API 调用 |
| `frontend/src/api/agentTasks.ts` | AgentTask API 调用 |
| `frontend/src/hooks/useDrafts.ts` | 草案数据 hooks |
| `frontend/src/hooks/useAgentTasks.ts` | AgentTask 数据 hooks |
| `frontend/src/pages/DraftCreatePage.tsx` | 草案创建页 |
| `frontend/src/pages/DraftConfirmPage.tsx` | 草案确认页 |
| `frontend/src/components/AgentNodeCard.tsx` | Agent 节点执行态卡片组件 |

| 修改文件 | 修改内容 |
|----------|----------|
| `frontend/src/types/api.ts` | 新增 FlowDraft / AgentTask / AgentTaskReceipt 类型 |
| `frontend/src/App.tsx` | 注册新路由 |
| `frontend/src/pages/DashboardPage.tsx` | 新增"让龙虾生成流程"入口 |
| `frontend/src/pages/RunDetailPage.tsx` | 节点执行态增强 |
| `frontend/src/styles.css` | 新增草案页和 agent 节点相关样式 |

### Task 1: TypeScript 类型扩展

**Files:**
- Modify: `frontend/src/types/api.ts`

- [ ] **Step 1: 在 api.ts 末尾追加新类型**

```typescript
// === 龙虾集成类型 ===

export interface FlowDraft {
  id: number;
  title: string;
  description: string;
  source_prompt: string;
  creator_person_id: number;
  planner_agent_id?: number;
  status: "draft" | "confirmed" | "discarded";
  structured_plan_json: DraftPlan;
  confirmed_template_id?: number;
  created_at: string;
  updated_at: string;
  confirmed_at?: string;
}

export interface DraftPlan {
  title: string;
  description: string;
  nodes: DraftNode[];
  final_deliverable: string;
}

export interface DraftNode {
  node_code: string;
  node_name: string;
  node_type: string;
  sort_order: number;
  executor_type: "agent" | "human";
  executor_agent_code?: string;
  result_owner_rule: string;
  task_type?: string;
  input_schema: Record<string, unknown>;
  output_schema: Record<string, unknown>;
  completion_condition?: string;
  failure_condition?: string;
  escalation_rule?: string;
}

export interface AgentTask {
  id: number;
  run_id: number;
  run_node_id: number;
  agent_id: number;
  task_type: string;
  input_json: Record<string, unknown>;
  status: "queued" | "running" | "completed" | "needs_review" | "failed" | "cancelled";
  started_at?: string;
  finished_at?: string;
  error_message?: string;
  result_json: Record<string, unknown>;
  artifacts_json: AgentArtifact[];
  created_at: string;
  updated_at: string;
}

export interface AgentArtifact {
  name: string;
  url: string;
  type: string;
}

export interface AgentTaskReceipt {
  id: number;
  agent_task_id: number;
  run_id: number;
  run_node_id: number;
  agent_id: number;
  receipt_status: "completed" | "needs_review" | "failed" | "blocked";
  payload_json: Record<string, unknown>;
  received_at: string;
}

export interface CreateDraftInput {
  source_prompt: string;
  title?: string;
  description?: string;
  planner_agent_id?: number;
}

export interface UpdateDraftInput {
  title?: string;
  description?: string;
  structured_plan_json?: DraftPlan;
}

export interface ConfirmDraftResponse {
  draft_id: number;
  template_id: number;
  message: string;
}

export interface ConfirmAgentResultInput {
  action: "approve" | "reject";
  comment?: string;
}

export interface TakeoverNodeInput {
  action: "retry" | "manual_complete";
  manual_result?: Record<string, unknown>;
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/types/api.ts
git commit -m "feat(p4): add FlowDraft, AgentTask and related TypeScript types"
```

### Task 2: API 调用层

**Files:**
- Create: `frontend/src/api/drafts.ts`
- Create: `frontend/src/api/agentTasks.ts`

- [ ] **Step 1: 写 drafts.ts**

```typescript
import { requestJSON } from "../lib/http";
import type {
  FlowDraft,
  CreateDraftInput,
  UpdateDraftInput,
  ConfirmDraftResponse,
} from "../types/api";

export function createDraft(input: CreateDraftInput): Promise<FlowDraft> {
  return requestJSON<FlowDraft>("/api/flow-drafts", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function listDrafts(query?: {
  status?: string;
  page?: number;
  page_size?: number;
}): Promise<FlowDraft[]> {
  return requestJSON<FlowDraft[]>("/api/flow-drafts", undefined, query);
}

export function getDraft(id: number): Promise<FlowDraft> {
  return requestJSON<FlowDraft>(`/api/flow-drafts/${id}`);
}

export function updateDraft(
  id: number,
  input: UpdateDraftInput
): Promise<FlowDraft> {
  return requestJSON<FlowDraft>(`/api/flow-drafts/${id}`, {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export function confirmDraft(id: number): Promise<ConfirmDraftResponse> {
  return requestJSON<ConfirmDraftResponse>(`/api/flow-drafts/${id}/confirm`, {
    method: "POST",
  });
}

export function discardDraft(id: number): Promise<void> {
  return requestJSON<void>(`/api/flow-drafts/${id}/discard`, {
    method: "POST",
  });
}
```

- [ ] **Step 2: 写 agentTasks.ts**

```typescript
import { requestJSON } from "../lib/http";
import type {
  AgentTask,
  AgentTaskReceipt,
  ConfirmAgentResultInput,
  TakeoverNodeInput,
  RunNodeDetail,
} from "../types/api";

export function listAgentTasks(runId: number): Promise<AgentTask[]> {
  return requestJSON<AgentTask[]>(`/api/agent-tasks/run/${runId}`);
}

export function getAgentTask(id: number): Promise<AgentTask> {
  return requestJSON<AgentTask>(`/api/agent-tasks/${id}`);
}

export function getAgentTaskReceipt(
  taskId: number
): Promise<AgentTaskReceipt> {
  return requestJSON<AgentTaskReceipt>(`/api/agent-tasks/${taskId}/receipt`);
}

export function confirmAgentResult(
  nodeId: number,
  input: ConfirmAgentResultInput
): Promise<RunNodeDetail> {
  return requestJSON<RunNodeDetail>(
    `/api/run-nodes/${nodeId}/confirm-agent-result`,
    {
      method: "POST",
      body: JSON.stringify(input),
    }
  );
}

export function takeoverNode(
  nodeId: number,
  input: TakeoverNodeInput
): Promise<RunNodeDetail> {
  return requestJSON<RunNodeDetail>(`/api/run-nodes/${nodeId}/takeover`, {
    method: "POST",
    body: JSON.stringify(input),
  });
}
```

> 注意：`RunNodeDetail` 使用 `api.ts` 中已有的节点详情类型。如果现有类型命名不同（如 `RunNodeDetailResponse`），请对齐。

- [ ] **Step 3: Commit**

```bash
git add frontend/src/api/drafts.ts frontend/src/api/agentTasks.ts
git commit -m "feat(p4): add drafts and agentTasks API client modules"
```

### Task 3: 数据 Hooks

**Files:**
- Create: `frontend/src/hooks/useDrafts.ts`
- Create: `frontend/src/hooks/useAgentTasks.ts`

- [ ] **Step 1: 写 useDrafts.ts**

```typescript
import { useCallback, useEffect, useState } from "react";
import {
  createDraft,
  getDraft,
  updateDraft,
  confirmDraft,
  discardDraft,
} from "../api/drafts";
import type {
  FlowDraft,
  CreateDraftInput,
  UpdateDraftInput,
  ConfirmDraftResponse,
} from "../types/api";

export function useCreateDraft() {
  const [loading, setLoading] = useState(false);
  const run = useCallback(async (input: CreateDraftInput) => {
    setLoading(true);
    try {
      return await createDraft(input);
    } finally {
      setLoading(false);
    }
  }, []);
  return { run, loading };
}

export function useDraftDetail(id?: number) {
  const [data, setData] = useState<FlowDraft | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const fetchData = useCallback(async () => {
    if (!id) {
      setData(null);
      setError("草案 ID 不合法");
      return;
    }
    setLoading(true);
    setError("");
    try {
      const result = await getDraft(id);
      setData(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载草案失败");
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    void fetchData();
  }, [fetchData]);

  return { data, loading, error, refetch: fetchData };
}

export function useUpdateDraft() {
  const [loading, setLoading] = useState(false);
  const run = useCallback(async (id: number, input: UpdateDraftInput) => {
    setLoading(true);
    try {
      return await updateDraft(id, input);
    } finally {
      setLoading(false);
    }
  }, []);
  return { run, loading };
}

export function useConfirmDraft() {
  const [loading, setLoading] = useState(false);
  const run = useCallback(async (id: number) => {
    setLoading(true);
    try {
      return await confirmDraft(id);
    } finally {
      setLoading(false);
    }
  }, []);
  return { run, loading };
}

export function useDiscardDraft() {
  const [loading, setLoading] = useState(false);
  const run = useCallback(async (id: number) => {
    setLoading(true);
    try {
      await discardDraft(id);
    } finally {
      setLoading(false);
    }
  }, []);
  return { run, loading };
}
```

- [ ] **Step 2: 写 useAgentTasks.ts**

```typescript
import { useCallback, useEffect, useState } from "react";
import {
  listAgentTasks,
  confirmAgentResult,
  takeoverNode,
} from "../api/agentTasks";
import type {
  AgentTask,
  ConfirmAgentResultInput,
  TakeoverNodeInput,
} from "../types/api";

export function useAgentTasks(runId?: number) {
  const [data, setData] = useState<AgentTask[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const fetchData = useCallback(async () => {
    if (!runId) return;
    setLoading(true);
    setError("");
    try {
      const result = await listAgentTasks(runId);
      setData(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载龙虾任务失败");
    } finally {
      setLoading(false);
    }
  }, [runId]);

  useEffect(() => {
    void fetchData();
  }, [fetchData]);

  return { data, loading, error, refetch: fetchData };
}

export function useConfirmAgentResult() {
  const [loading, setLoading] = useState(false);
  const run = useCallback(
    async (nodeId: number, input: ConfirmAgentResultInput) => {
      setLoading(true);
      try {
        return await confirmAgentResult(nodeId, input);
      } finally {
        setLoading(false);
      }
    },
    []
  );
  return { run, loading };
}

export function useTakeoverNode() {
  const [loading, setLoading] = useState(false);
  const run = useCallback(
    async (nodeId: number, input: TakeoverNodeInput) => {
      setLoading(true);
      try {
        return await takeoverNode(nodeId, input);
      } finally {
        setLoading(false);
      }
    },
    []
  );
  return { run, loading };
}
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/hooks/useDrafts.ts frontend/src/hooks/useAgentTasks.ts
git commit -m "feat(p4): add draft and agent task data hooks"
```

### Task 4: DraftCreatePage

**Files:**
- Create: `frontend/src/pages/DraftCreatePage.tsx`

- [ ] **Step 1: 写 DraftCreatePage 组件**

```tsx
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useCreateDraft } from "../hooks/useDrafts";

export default function DraftCreatePage() {
  const navigate = useNavigate();
  const { run: createDraft, loading } = useCreateDraft();
  const [prompt, setPrompt] = useState("");
  const [error, setError] = useState("");

  const handleSubmit = async () => {
    if (!prompt.trim()) {
      setError("请输入业务目标描述");
      return;
    }
    setError("");
    try {
      const draft = await createDraft({ source_prompt: prompt.trim() });
      if (draft) {
        navigate(`/drafts/${draft.id}/confirm`);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "创建草案失败");
    }
  };

  return (
    <div className="page-container">
      <div className="page-header">
        <h1>让龙虾生成流程</h1>
        <p className="muted">
          用自然语言描述你的业务目标，龙虾将为你生成可确认的流程草案
        </p>
      </div>

      <div className="draft-create-form">
        <textarea
          className="draft-prompt-input"
          placeholder="描述你的业务目标和期望的流程步骤...&#10;&#10;例如：帮我创建一个视频绑定的工作流程，第一步收集距离开课不足2天的课程场次数据，第二步我审核确认，第三步龙虾执行绑定，第四步导出结果，最后由主管签收。"
          value={prompt}
          onChange={(e) => setPrompt(e.target.value)}
          rows={8}
          disabled={loading}
        />

        {error && <p className="error-text">{error}</p>}

        <div className="draft-create-actions">
          <button
            className="btn btn-primary"
            onClick={handleSubmit}
            disabled={loading || !prompt.trim()}
          >
            {loading ? "龙虾正在生成流程草案..." : "生成流程草案"}
          </button>
          <button className="btn" onClick={() => navigate(-1)} disabled={loading}>
            返回
          </button>
        </div>
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/pages/DraftCreatePage.tsx
git commit -m "feat(p4): add DraftCreatePage for natural language draft creation"
```

### Task 5: DraftConfirmPage

**Files:**
- Create: `frontend/src/pages/DraftConfirmPage.tsx`

- [ ] **Step 1: 写 DraftConfirmPage 组件**

```tsx
import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { getDraft, updateDraft } from "../api/drafts";
import { useConfirmDraft, useDiscardDraft } from "../hooks/useDrafts";
import type { FlowDraft, DraftNode } from "../types/api";

const NODE_TYPE_LABELS: Record<string, string> = {
  human_input: "人工输入",
  human_review: "人工审核",
  agent_execute: "自动执行",
  agent_export: "自动导出",
  human_acceptance: "最终签收",
};

const EXECUTOR_TYPE_LABELS: Record<string, string> = {
  agent: "龙虾",
  human: "人工",
};

function parseID(val: string | undefined): number | undefined {
  if (!val) return undefined;
  const n = Number(val);
  return Number.isFinite(n) && n > 0 ? n : undefined;
}

export default function DraftConfirmPage() {
  const { id: rawId } = useParams<{ id: string }>();
  const id = parseID(rawId);
  const navigate = useNavigate();

  const [draft, setDraft] = useState<FlowDraft | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [editingTitle, setEditingTitle] = useState(false);
  const [title, setTitle] = useState("");

  const { run: confirm, loading: confirming } = useConfirmDraft();
  const { run: discard, loading: discarding } = useDiscardDraft();

  useEffect(() => {
    if (!id) {
      setError("草案 ID 不合法");
      setLoading(false);
      return;
    }
    setLoading(true);
    getDraft(id)
      .then((d) => {
        setDraft(d);
        setTitle(d.title);
      })
      .catch((err) =>
        setError(err instanceof Error ? err.message : "加载草案失败")
      )
      .finally(() => setLoading(false));
  }, [id]);

  const handleConfirm = async () => {
    if (!id) return;
    try {
      const result = await confirm(id);
      if (result) {
        navigate(`/templates/${result.template_id}/start`);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "确认草案失败");
    }
  };

  const handleDiscard = async () => {
    if (!id) return;
    try {
      await discard(id);
      navigate("/");
    } catch (err) {
      setError(err instanceof Error ? err.message : "废弃草案失败");
    }
  };

  const handleTitleSave = async () => {
    if (!id || !draft) return;
    try {
      const updated = await updateDraft(id, { title });
      setDraft(updated);
      setEditingTitle(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : "更新标题失败");
    }
  };

  const handleDeleteNode = async (nodeCode: string) => {
    if (!id || !draft) return;
    const plan = draft.structured_plan_json;
    const newNodes = plan.nodes
      .filter((n) => n.node_code !== nodeCode)
      .map((n, i) => ({ ...n, sort_order: i + 1 }));
    try {
      const updated = await updateDraft(id, {
        structured_plan_json: { ...plan, nodes: newNodes },
      });
      setDraft(updated);
    } catch (err) {
      setError(err instanceof Error ? err.message : "删除节点失败");
    }
  };

  if (loading) return <div className="loading-state">加载中...</div>;
  if (error) return <div className="error-state">{error}</div>;
  if (!draft) return <div className="error-state">草案不存在</div>;

  const plan = draft.structured_plan_json;

  return (
    <div className="page-container">
      <div className="page-header">
        {editingTitle ? (
          <div className="draft-title-edit">
            <input
              className="draft-title-input"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleTitleSave()}
            />
            <button className="btn btn-primary btn-sm" onClick={handleTitleSave}>
              保存
            </button>
            <button
              className="btn btn-sm"
              onClick={() => {
                setTitle(draft.title);
                setEditingTitle(false);
              }}
            >
              取消
            </button>
          </div>
        ) : (
          <h1 onClick={() => setEditingTitle(true)} style={{ cursor: "pointer" }}>
            {draft.title} ✎
          </h1>
        )}
        {plan.description && <p className="muted">{plan.description}</p>}
      </div>

      <div className="draft-prompt-display">
        <strong>原始需求：</strong>
        {draft.source_prompt}
      </div>

      <div className="draft-nodes-list">
        {plan.nodes.map((node: DraftNode) => (
          <div key={node.node_code} className="draft-node-card">
            <div className="draft-node-header">
              <span className="draft-node-order">{node.sort_order}</span>
              <span className="draft-node-name">{node.node_name}</span>
              <span
                className={`tag tag-${node.executor_type === "agent" ? "agent" : "human"}`}
              >
                {EXECUTOR_TYPE_LABELS[node.executor_type] || node.executor_type}
              </span>
              <span className="tag tag-type">
                {NODE_TYPE_LABELS[node.node_type] || node.node_type}
              </span>
              {plan.nodes.length > 3 && (
                <button
                  className="btn btn-sm danger"
                  onClick={() => handleDeleteNode(node.node_code)}
                >
                  删除
                </button>
              )}
            </div>

            <div className="draft-node-body">
              {node.task_type && (
                <div className="draft-node-field">
                  <span className="label">任务类型：</span>
                  {node.task_type}
                </div>
              )}
              {node.result_owner_rule && (
                <div className="draft-node-field">
                  <span className="label">结果责任人：</span>
                  {node.result_owner_rule}
                </div>
              )}
              {node.completion_condition && (
                <div className="draft-node-field">
                  <span className="label">完成条件：</span>
                  {node.completion_condition}
                </div>
              )}
              {node.failure_condition && (
                <div className="draft-node-field">
                  <span className="label">失败条件：</span>
                  {node.failure_condition}
                </div>
              )}
            </div>
          </div>
        ))}
      </div>

      {error && <p className="error-text">{error}</p>}

      <div className="draft-confirm-actions">
        <button
          className="btn btn-primary"
          onClick={handleConfirm}
          disabled={confirming || discarding}
        >
          {confirming ? "正在创建模板..." : "确认创建"}
        </button>
        <button
          className="btn danger"
          onClick={handleDiscard}
          disabled={confirming || discarding}
        >
          {discarding ? "正在废弃..." : "废弃草案"}
        </button>
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/pages/DraftConfirmPage.tsx
git commit -m "feat(p4): add DraftConfirmPage with node preview, edit and confirm/discard"
```

### Task 6: AgentNodeCard 组件

**Files:**
- Create: `frontend/src/components/AgentNodeCard.tsx`

- [ ] **Step 1: 写 AgentNodeCard 组件**

此组件用于在 RunDetailPage 的节点详情中展示龙虾执行态。

```tsx
import { useState } from "react";
import type { AgentTask, ConfirmAgentResultInput, TakeoverNodeInput } from "../types/api";
import { useConfirmAgentResult, useTakeoverNode } from "../hooks/useAgentTasks";

interface AgentNodeCardProps {
  nodeId: number;
  nodeStatus: string;
  agentTask?: AgentTask;
  onRefresh: () => void;
}

export default function AgentNodeCard({
  nodeId,
  nodeStatus,
  agentTask,
  onRefresh,
}: AgentNodeCardProps) {
  const { run: confirmResult, loading: confirmLoading } = useConfirmAgentResult();
  const { run: takeover, loading: takeoverLoading } = useTakeoverNode();
  const [comment, setComment] = useState("");
  const [error, setError] = useState("");

  const handleConfirm = async (action: "approve" | "reject") => {
    setError("");
    try {
      await confirmResult(nodeId, { action, comment });
      onRefresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "操作失败");
    }
  };

  const handleTakeover = async (action: "retry" | "manual_complete") => {
    setError("");
    try {
      await takeover(nodeId, { action });
      onRefresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "操作失败");
    }
  };

  if (!agentTask) return null;

  return (
    <div className="agent-node-card">
      <div className="agent-node-status">
        <span className={`agent-status-badge status-${agentTask.status}`}>
          {agentTask.status === "running" && "执行中"}
          {agentTask.status === "completed" && "已完成"}
          {agentTask.status === "needs_review" && "待确认"}
          {agentTask.status === "failed" && "执行失败"}
          {agentTask.status === "queued" && "排队中"}
          {agentTask.status === "cancelled" && "已取消"}
        </span>
        {agentTask.task_type && (
          <span className="agent-task-type">{agentTask.task_type}</span>
        )}
        {agentTask.started_at && agentTask.finished_at && (
          <span className="agent-duration">
            耗时{" "}
            {Math.round(
              (new Date(agentTask.finished_at).getTime() -
                new Date(agentTask.started_at).getTime()) /
                1000
            )}
            秒
          </span>
        )}
      </div>

      {/* 结构化结果 */}
      {agentTask.result_json && Object.keys(agentTask.result_json).length > 0 && (
        <div className="agent-result-section">
          <h4>执行结果</h4>
          {agentTask.task_type === "query" &&
            agentTask.result_json.records_count !== undefined && (
              <p>查询到 {String(agentTask.result_json.records_count)} 条记录</p>
            )}
          {agentTask.task_type === "batch_operation" && (
            <p>
              成功 {String(agentTask.result_json.success_count ?? 0)} 条，失败{" "}
              {String(agentTask.result_json.failed_count ?? 0)} 条
            </p>
          )}
          {agentTask.task_type === "export" && agentTask.result_json.file_url && (
            <a
              href={String(agentTask.result_json.file_url)}
              className="btn btn-sm"
              download
            >
              下载 {String(agentTask.result_json.file_name ?? "导出文件")}
            </a>
          )}
          <details>
            <summary>原始结果</summary>
            <pre className="agent-result-raw">
              {JSON.stringify(agentTask.result_json, null, 2)}
            </pre>
          </details>
        </div>
      )}

      {/* 附件 */}
      {agentTask.artifacts_json && agentTask.artifacts_json.length > 0 && (
        <div className="agent-artifacts-section">
          <h4>附件</h4>
          <ul>
            {agentTask.artifacts_json.map((a, i) => (
              <li key={i}>
                <a href={a.url} download>
                  {a.name}
                </a>
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* 错误信息 */}
      {agentTask.error_message && (
        <div className="agent-error-section">
          <p className="error-text">{agentTask.error_message}</p>
        </div>
      )}

      {/* 操作区 */}
      {nodeStatus === "waiting_confirm" && (
        <div className="agent-actions">
          <textarea
            className="agent-comment-input"
            placeholder="确认意见（可选）"
            value={comment}
            onChange={(e) => setComment(e.target.value)}
            rows={2}
          />
          <div className="agent-action-buttons">
            <button
              className="btn btn-primary"
              onClick={() => handleConfirm("approve")}
              disabled={confirmLoading}
            >
              确认通过
            </button>
            <button
              className="btn danger"
              onClick={() => handleConfirm("reject")}
              disabled={confirmLoading}
            >
              驳回
            </button>
          </div>
        </div>
      )}

      {(nodeStatus === "failed" || nodeStatus === "blocked") && (
        <div className="agent-actions">
          <div className="agent-action-buttons">
            <button
              className="btn btn-primary"
              onClick={() => handleTakeover("retry")}
              disabled={takeoverLoading}
            >
              重试
            </button>
            <button
              className="btn"
              onClick={() => handleTakeover("manual_complete")}
              disabled={takeoverLoading}
            >
              人工接管
            </button>
          </div>
        </div>
      )}

      {error && <p className="error-text">{error}</p>}
    </div>
  );
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/AgentNodeCard.tsx
git commit -m "feat(p4): add AgentNodeCard component for agent execution display"
```

### Task 7: 路由注册 + Dashboard 入口

**Files:**
- Modify: `frontend/src/App.tsx`
- Modify: `frontend/src/pages/DashboardPage.tsx`

- [ ] **Step 1: 在 App.tsx 中导入新页面并注册路由**

在文件顶部已有的 `import` 区域中追加：

```tsx
import DraftCreatePage from "./pages/DraftCreatePage";
import DraftConfirmPage from "./pages/DraftConfirmPage";
```

在 `<Routes>` 内部追加（放在其他路由之后、catch-all 之前）：

```tsx
<Route path="/drafts/create" element={<DraftCreatePage />} />
<Route path="/drafts/:id/confirm" element={<DraftConfirmPage />} />
```

- [ ] **Step 2: 在 DashboardPage 中新增"让龙虾生成流程"入口**

找到 DashboardPage 中现有的 `.template-shortcuts` 区域（包含 `shortcut-card` 链接卡片）。在该 grid 内追加一个新的入口卡片：

```tsx
<Link className="shortcut-card shortcut-card-ai" to="/drafts/create">
  <span className="metric-label">AI 编排</span>
  <strong>让龙虾生成流程</strong>
  <p>用自然语言描述业务目标，龙虾自动生成流程草案。</p>
</Link>
```

> 使用现有的 `shortcut-card` class 和 `Link` 组件，确保融入已有 grid 布局。`shortcut-card-ai` 为可选的额外修饰 class。如果 DashboardPage 中需要导入 `Link`，从 `react-router-dom` 导入。

- [ ] **Step 3: 在 RunStartPage 中新增"让龙虾生成"入口**

在 `frontend/src/pages/RunStartPage.tsx` 顶部区域（模板选择之前）新增入口链接：

```tsx
<div className="run-start-toolbar">
  <Link className="btn" to="/drafts/create">
    让龙虾生成流程
  </Link>
</div>
```

> 根据 RunStartPage 的实际结构调整位置和 class 名。

- [ ] **Step 4: 在 sidebar 导航中追加 "草案" 链接（可选）**

如果 sidebar 有导航链接列表，可在其中追加：

```tsx
<NavLink to="/drafts/create">草案</NavLink>
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/App.tsx frontend/src/pages/DashboardPage.tsx
git commit -m "feat(p4): register draft routes and add AI draft entry on dashboard"
```

### Task 8: RunDetailPage 节点执行态增强

**Files:**
- Modify: `frontend/src/pages/RunDetailPage.tsx` — 添加 `useAgentTasks` 数据加载
- Modify: `frontend/src/components/RunNodeWorkbench.tsx` — 节点详情展示增强（节点渲染逻辑在此组件中）

> 注意：节点卡片渲染在 `RunNodeWorkbench.tsx` 中，而非 `RunDetailPage.tsx`。以下代码片段需要根据 `RunNodeWorkbench.tsx` 的 props 和 state 结构调整。AgentTask 数据在 `RunDetailPage.tsx` 中加载，通过 props 传递给 `RunNodeWorkbench`。

- [ ] **Step 1: 导入 AgentNodeCard 和 useAgentTasks hook**

```tsx
import AgentNodeCard from "../components/AgentNodeCard";
import { useAgentTasks } from "../hooks/useAgentTasks";
```

- [ ] **Step 2: 在 RunDetailPage 中调用 useAgentTasks**

在组件内部，运行数据加载后：

```tsx
const { data: agentTasks, refetch: refetchAgentTasks } = useAgentTasks(
  run?.id
);
```

- [ ] **Step 3: 在节点卡片渲染中集成 AgentNodeCard**

在节点列表渲染处（节点遍历的位置），找到当前的节点渲染逻辑。在每个节点的展开区域中，增加条件渲染：

```tsx
{/* 在节点详情展开区域中 */}
{(node.node_type === "agent_execute" ||
  node.node_type === "agent_export" ||
  node.node_type === "execute") && (
  <AgentNodeCard
    nodeId={node.id}
    nodeStatus={node.status}
    agentTask={agentTasks.find((t) => t.run_node_id === node.id)}
    onRefresh={() => {
      refetch();
      refetchAgentTasks();
    }}
  />
)}
```

- [ ] **Step 4: 在节点卡片头部新增执行主体/结果责任人标签**

在每个节点卡片的 header 区域追加标签：

```tsx
{node.bound_agent_id && (
  <span className="tag tag-agent">龙虾执行</span>
)}
{node.owner_person_id && (
  <span className="tag tag-human">
    责任人: {node.owner_person_id}
  </span>
)}
```

- [ ] **Step 5: 为 running 状态的 agent 节点增加执行中动画**

```tsx
{node.status === "running" &&
  (node.node_type === "agent_execute" || node.node_type === "agent_export") && (
  <div className="agent-running-indicator">
    <span className="agent-running-dot" />
    龙虾执行中...
  </div>
)}
```

- [ ] **Step 6: Commit**

```bash
git add frontend/src/pages/RunDetailPage.tsx
git commit -m "feat(p4): enhance RunDetailPage with agent execution display and controls"
```

### Task 9: CSS 样式

**Files:**
- Modify: `frontend/src/styles.css`

- [ ] **Step 1: 追加草案页和 agent 节点相关样式**

在 `styles.css` 末尾追加：

```css
/* === 龙虾集成样式 === */

/* 草案创建页 */
.draft-create-form {
  max-width: 720px;
}

.draft-prompt-input {
  width: 100%;
  min-height: 200px;
  padding: 16px;
  background: var(--panel);
  border: 1px solid var(--line);
  border-radius: var(--radius-sm);
  color: var(--text);
  font-size: 15px;
  line-height: 1.6;
  resize: vertical;
}

.draft-prompt-input:focus {
  outline: none;
  border-color: var(--accent);
}

.draft-create-actions {
  display: flex;
  gap: 12px;
  margin-top: 16px;
}

/* 草案确认页 */
.draft-prompt-display {
  padding: 12px 16px;
  background: var(--panel);
  border-radius: var(--radius-sm);
  margin-bottom: 24px;
  color: var(--muted);
  font-size: 14px;
}

.draft-title-edit {
  display: flex;
  align-items: center;
  gap: 8px;
}

.draft-title-input {
  font-size: 24px;
  font-weight: 700;
  background: transparent;
  border: 1px solid var(--line);
  border-radius: var(--radius-sm);
  color: var(--text);
  padding: 4px 8px;
  flex: 1;
}

.draft-nodes-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 24px;
}

.draft-node-card {
  background: var(--panel);
  border: 1px solid var(--line);
  border-radius: var(--radius-sm);
  padding: 16px;
}

.draft-node-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.draft-node-order {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: var(--accent);
  color: var(--bg);
  font-size: 12px;
  font-weight: 700;
  flex-shrink: 0;
}

.draft-node-name {
  font-weight: 600;
  flex: 1;
}

.draft-node-body {
  padding-left: 32px;
}

.draft-node-field {
  font-size: 13px;
  color: var(--muted);
  margin-bottom: 4px;
}

.draft-node-field .label {
  color: var(--text);
  font-weight: 500;
}

.draft-confirm-actions {
  display: flex;
  gap: 12px;
}

/* 标签 */
.tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.tag-agent {
  background: rgba(229, 191, 114, 0.2);
  color: var(--warning);
}

.tag-human {
  background: rgba(118, 181, 255, 0.2);
  color: var(--accent);
}

.tag-type {
  background: rgba(255, 255, 255, 0.08);
  color: var(--muted);
}

/* Agent 节点卡片 */
.agent-node-card {
  margin-top: 12px;
  padding: 12px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid var(--line);
  border-radius: var(--radius-sm);
}

.agent-node-status {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.agent-status-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.agent-status-badge.status-running {
  background: rgba(118, 181, 255, 0.2);
  color: var(--accent);
}

.agent-status-badge.status-completed {
  background: rgba(98, 216, 170, 0.2);
  color: var(--success);
}

.agent-status-badge.status-needs_review {
  background: rgba(229, 191, 114, 0.2);
  color: var(--warning);
}

.agent-status-badge.status-failed {
  background: rgba(239, 143, 143, 0.2);
  color: var(--danger);
}

.agent-task-type {
  font-size: 12px;
  color: var(--muted);
}

.agent-duration {
  font-size: 12px;
  color: var(--muted);
}

.agent-result-section {
  margin-top: 8px;
}

.agent-result-section h4 {
  font-size: 13px;
  margin-bottom: 4px;
}

.agent-result-raw {
  font-size: 12px;
  padding: 8px;
  background: rgba(0, 0, 0, 0.2);
  border-radius: 4px;
  overflow-x: auto;
  max-height: 200px;
}

.agent-artifacts-section {
  margin-top: 8px;
}

.agent-error-section {
  margin-top: 8px;
}

.agent-actions {
  margin-top: 12px;
}

.agent-comment-input {
  width: 100%;
  padding: 8px;
  background: var(--panel);
  border: 1px solid var(--line);
  border-radius: 4px;
  color: var(--text);
  font-size: 13px;
  resize: vertical;
  margin-bottom: 8px;
}

.agent-action-buttons {
  display: flex;
  gap: 8px;
}

/* 执行中动画 */
.agent-running-indicator {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--accent);
}

.agent-running-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--accent);
  animation: agent-pulse 1.2s ease-in-out infinite;
}

@keyframes agent-pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}

/* Dashboard AI 入口卡片 */
.shortcut-card-ai {
  border: 1px dashed var(--warning);
}

.shortcut-card-ai:hover {
  border-color: var(--accent);
  background: rgba(118, 181, 255, 0.05);
}

/* 小尺寸按钮 */
.btn-sm {
  padding: 4px 8px;
  font-size: 12px;
}
```

> CSS 变量已对齐 `styles.css` 中的定义：`--panel`（面板背景）、`--line`（边框）、`--muted`（次要文字）、`--text`、`--accent`、`--warning`、`--success`、`--danger`、`--bg`。

- [ ] **Step 2: Commit**

```bash
git add frontend/src/styles.css
git commit -m "feat(p4): add CSS styles for draft pages and agent node display"
```

### Task 10: 前端编译验证

**Files:** 无新增

- [ ] **Step 1: 安装依赖并编译**

Run: `cd frontend && npm install && npm run build`
Expected: 编译通过，无 TypeScript 错误

- [ ] **Step 2: 启动 dev server 验证页面**

Run: `cd frontend && npm run dev`

手动验证清单（需后端同时运行）：
1. 打开 http://localhost:5173
2. Dashboard 页面出现"让龙虾生成流程"入口卡片
3. 点击入口跳转到 /drafts/create
4. 输入需求文本，点击"生成流程草案"
5. 跳转到 /drafts/:id/confirm，显示草案节点列表
6. 可以编辑标题、删除节点
7. 点击"确认创建"，跳转到模板详情
8. 从新模板发起 Run，观察 agent 节点执行态

- [ ] **Step 3: Commit（如有修复）**

```bash
git add frontend/
git commit -m "fix(p4): fix frontend compilation issues"
```

---
