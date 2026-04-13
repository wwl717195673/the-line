# P2：龙虾真实执行

## 目标

实现自动节点触发龙虾真实执行、标准回执协议、基于回执的节点流转和人工接管机制。

---

## 1. 设计原则

1. 龙虾只能提交回执，不能直接修改流程状态
2. 外部执行必须发生在数据库事务提交之后
3. 同一 `run_node` 同时只能存在一个活跃 `AgentTask`
4. 自动节点的结果既要写入 `AgentTask`，也要能回流到 `FlowRunNode.OutputJSON`

---

## 2. Agent Executor 接口

在 `internal/executor/` 下定义执行器接口：

```go
type AgentExecutor interface {
    Execute(ctx context.Context, task *model.AgentTask, agent *model.Agent) error
}

type AgentPlannerExecutor interface {
    GenerateDraft(ctx context.Context, prompt string, agent *model.Agent) (*dto.DraftPlan, error)
}
```

---

## 3. Mock 实现

### 3.1 MockAgentExecutor

```go
type MockAgentExecutor struct {
    receiptCallback func(receipt *dto.AgentReceiptRequest) error
}

func (m *MockAgentExecutor) Execute(ctx context.Context, task *model.AgentTask, agent *model.Agent) error {
    go func() {
        time.Sleep(2*time.Second + time.Duration(rand.Intn(1000))*time.Millisecond)
        receipt := buildMockReceipt(task)
        _ = m.receiptCallback(receipt)
    }()
    return nil
}
```

Mock 回执根据 `task_type` 返回不同结果：

* `query` → 模拟记录列表
* `batch_operation` → `success_count` / `failed_count`
* `export` → 模拟导出文件 URL

### 3.2 MockAgentPlannerExecutor

返回预定义草案 JSON，供 P3 联调。

---

## 4. 自动节点触发机制

### 4.1 触发点

不建议新增一个脱离现有基座的“AdvanceNode”服务，而是复用当前链路：

1. `RunService.CreateRun`
   * 创建流程后，如果首节点是自动节点，则在事务提交后调度
2. `RunService.AdvanceAfterNodeDone`
   * 将下一节点推进到 `ready` 后，返回 `nextNode`
   * 由调用方在事务提交后异步调度

注意：

* 不要在数据库事务内部直接触发外部龙虾执行
* 事务内部只负责把节点推进到 `ready`
* 事务提交后再调用 `AgentTaskService.CreateAndDispatch`

### 4.2 调度入口

建议新增一个轻量 orchestration 层负责“提交后调度”：

```go
func (s *RunOrchestrationService) DispatchIfNeeded(ctx context.Context, nodeID uint64) error {
    node, err := s.runNodeRepo.GetByID(ctx, nodeID)
    if err != nil {
        return err
    }
    if !isAgentNode(node.NodeType) || node.BoundAgentID == nil {
        return nil
    }
    return s.agentTaskService.CreateAndDispatch(ctx, node.ID)
}
```

### 4.3 单飞执行

`CreateAndDispatch` 必须保证：

* 对节点加锁
* 节点状态仍为 `ready`
* 不存在活跃任务（`queued` / `running`）
* 才允许创建新任务

```go
func (s *AgentTaskService) CreateAndDispatch(ctx context.Context, nodeID uint64) error {
    var task *model.AgentTask
    var agent *model.Agent

    err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        node, err := s.runNodeRepo.GetByIDWithLock(ctx, tx, nodeID)
        if err != nil {
            return err
        }
        if node.Status != domain.NodeStatusReady {
            return nil
        }
        if node.BoundAgentID == nil {
            return response.Validation("当前节点未绑定龙虾")
        }

        activeTask, err := s.agentTaskRepo.GetActiveByRunNodeID(ctx, node.ID)
        if err != nil {
            return err
        }
        if activeTask != nil {
            return nil
        }

        now := time.Now()
        task = &model.AgentTask{
            RunID:     node.RunID,
            RunNodeID: node.ID,
            AgentID:   *node.BoundAgentID,
            TaskType:  inferTaskType(node),
            InputJSON: node.InputJSON,
            Status:    domain.AgentTaskStatusRunning,
            StartedAt: &now,
        }
        if err := s.agentTaskRepo.CreateWithDB(ctx, tx, task); err != nil {
            return err
        }

        if err := s.runNodeRepo.UpdateWithDB(ctx, tx, node.ID, map[string]any{
            "status":     domain.NodeStatusRunning,
            "started_at": &now,
        }); err != nil {
            return err
        }

        if err := s.nodeLogRepo.CreateWithDB(ctx, tx, buildAgentRunCreatedLog(node, task)); err != nil {
            return err
        }

        agent, err = s.agentRepo.GetByID(ctx, *node.BoundAgentID)
        return err
    })
    if err != nil || task == nil || agent == nil {
        return err
    }

    return s.executor.Execute(ctx, task, agent)
}
```

### 4.4 与现有 `run-agent` 入口兼容

当前系统已有手动 `run-agent` 入口。

MVP 建议：

* 对自动节点，不再暴露旧 `run-agent` 按钮
* 如果需要“重试”，统一走 `CreateAndDispatch`
* 禁止手动入口和自动触发并存地创建多个活跃任务

---

## 5. 回执接口

### 5.1 API 定义

```
POST /api/agent-tasks/:taskId/receipt
```

### 5.2 请求体

```json
{
  "agent_id": 1,
  "status": "completed",
  "started_at": "2026-04-13T10:00:00Z",
  "finished_at": "2026-04-13T10:00:05Z",
  "summary": "已完成数据查询，共找到 12 条记录",
  "result": {
    "records_count": 12,
    "records": [
      {"id": 101, "name": "场次A", "start_time": "2026-04-15T09:00:00Z", "video_bound": false}
    ]
  },
  "artifacts": [
    {"name": "export.xlsx", "url": "/uploads/xxx.xlsx", "type": "file"}
  ],
  "logs": [
    "step1: 查询数据库",
    "step2: 过滤未绑定视频的场次",
    "step3: 输出结果"
  ],
  "error_message": null
}
```

### 5.3 DTO

```go
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
```

---

## 6. 回执处理流程

### 6.1 事务处理

回执处理必须在单个事务里完成：

1. 锁定 `AgentTask`
2. 锁定对应 `FlowRunNode` / `FlowRun`
3. 写入 `AgentTaskReceipt`
4. 更新 `AgentTask`
5. 更新节点和流程状态
6. 写节点日志

```go
func (s *AgentTaskService) ProcessReceipt(ctx context.Context, taskID uint64, req *dto.AgentReceiptRequest) error {
    return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        task, err := s.agentTaskRepo.GetByIDWithLock(ctx, tx, taskID)
        if err != nil {
            return err
        }
        node, err := s.runNodeRepo.GetByIDWithLock(ctx, tx, task.RunNodeID)
        if err != nil {
            return err
        }
        run, err := s.runRepo.GetByIDWithLock(ctx, tx, task.RunID)
        if err != nil {
            return err
        }

        if task.AgentID != req.AgentID {
            return errors.New("agent_id mismatch")
        }
        if task.Status != domain.AgentTaskStatusRunning {
            return errors.New("task not in running state")
        }

        receipt := &model.AgentTaskReceipt{
            AgentTaskID:   task.ID,
            RunID:         task.RunID,
            RunNodeID:     task.RunNodeID,
            AgentID:       req.AgentID,
            ReceiptStatus: req.Status,
            PayloadJSON:   marshalPayload(req),
            ReceivedAt:    time.Now(),
        }
        if err := s.receiptRepo.CreateWithDB(ctx, tx, receipt); err != nil {
            return err
        }

        task.Status = mapReceiptToTaskStatus(req.Status)
        task.FinishedAt = req.FinishedAt
        task.ResultJSON = req.Result
        task.ArtifactsJSON = req.Artifacts
        task.ErrorMessage = req.ErrorMessage
        if err := s.agentTaskRepo.UpdateWithDB(ctx, tx, task); err != nil {
            return err
        }

        if err := s.handleNodeTransitionWithDB(ctx, tx, run, node, task, req); err != nil {
            return err
        }

        return s.nodeLogRepo.CreateWithDB(ctx, tx, buildAgentReceiptLog(run, node, req))
    })
}
```

### 6.2 节点流转规则

```go
func (s *AgentTaskService) handleNodeTransitionWithDB(ctx context.Context, tx *gorm.DB, run *model.FlowRun, node *model.FlowRunNode, task *model.AgentTask, req *dto.AgentReceiptRequest) error {
    switch req.Status {
    case domain.ReceiptStatusCompleted:
        now := time.Now()
        if err := s.runNodeRepo.UpdateWithDB(ctx, tx, node.ID, map[string]any{
            "status":       domain.NodeStatusDone,
            "output_json":  req.Result,
            "completed_at": &now,
        }); err != nil {
            return err
        }
        return s.runService.AdvanceAfterNodeDone(tx, run.ID, node.ID, domain.Actor{})

    case domain.ReceiptStatusNeedsReview:
        if err := s.runNodeRepo.UpdateWithDB(ctx, tx, node.ID, map[string]any{
            "status":      domain.NodeStatusWaitConfirm,
            "output_json": req.Result,
        }); err != nil {
            return err
        }
        return s.runRepo.UpdateWithDB(ctx, tx, run.ID, map[string]any{
            "current_status": domain.RunStatusWaiting,
        })

    case domain.ReceiptStatusFailed:
        if err := s.runNodeRepo.UpdateWithDB(ctx, tx, node.ID, map[string]any{
            "status":      domain.NodeStatusFailed,
            "output_json": req.Result,
        }); err != nil {
            return err
        }
        return s.runRepo.UpdateWithDB(ctx, tx, run.ID, map[string]any{
            "current_status": domain.RunStatusBlocked,
        })

    case domain.ReceiptStatusBlocked:
        if err := s.runNodeRepo.UpdateWithDB(ctx, tx, node.ID, map[string]any{
            "status":      domain.NodeStatusBlocked,
            "output_json": req.Result,
        }); err != nil {
            return err
        }
        return s.runRepo.UpdateWithDB(ctx, tx, run.ID, map[string]any{
            "current_status": domain.RunStatusBlocked,
        })
    }
    return nil
}
```

关键点：

* `AgentTask.ResultJSON` 是任务级原始结果
* `FlowRunNode.OutputJSON` 是节点级承接结果
* `needs_review` 时，流程应切到 `waiting`
* `failed` / `blocked` 时，流程应切到 `blocked`

---

## 7. 人工确认与异常接管

### 7.1 确认 `waiting_confirm` 节点

当龙虾回执为 `needs_review` 时，结果责任人确认后节点才能继续。

```
POST /api/run-nodes/:nodeId/confirm-agent-result
{
  "action": "approve" | "reject",
  "comment": "确认结果无误"
}
```

规则：

* `approve` → 节点 `done`，流程推进到下一节点
* `reject` → 节点 `blocked`，等待人工处理
* 确认权限优先基于 `ResultOwnerPersonID`

### 7.2 异常接管

当节点处于 `blocked` / `failed` 状态时，责任人可以：

```
POST /api/run-nodes/:nodeId/takeover
{
  "action": "retry" | "manual_complete",
  "manual_result": { ... }
}
```

规则：

* `retry` → 重新创建 AgentTask，再次触发
* `manual_complete` → 人工补齐结果，节点 `done`，流程推进

---

## 8. 完整 API 列表

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/agent-tasks` | AgentTask 列表（按 run_id 过滤） |
| GET | `/api/agent-tasks/:id` | AgentTask 详情 |
| POST | `/api/agent-tasks/:id/receipt` | 接收龙虾回执 |
| GET | `/api/agent-tasks/:id/receipt` | 查看回执详情 |
| POST | `/api/run-nodes/:nodeId/confirm-agent-result` | 确认龙虾执行结果 |
| POST | `/api/run-nodes/:nodeId/takeover` | 人工接管异常节点 |

---

## 9. 交付清单

- [ ] `internal/executor/agent_executor.go` — 接口定义
- [ ] `internal/executor/mock_agent_executor.go` — Mock 实现
- [ ] `internal/service/agent_task_service.go` — CreateAndDispatch + ProcessReceipt + handleNodeTransition
- [ ] `internal/service/run_orchestration_service.go` — 事务提交后调度自动节点
- [ ] `internal/handler/agent_task_handler.go` — 回执接口 + 查询接口
- [ ] `internal/handler/run_node_handler.go` — 新增 confirm-agent-result + takeover
- [ ] `internal/service/run_node_service.go` — 自动节点确认与接管
- [ ] `internal/service/run_service.go` — 返回 nextNode / 对接自动调度
- [ ] `internal/dto/agent_task_dto.go` — 请求/响应 DTO
- [ ] `internal/app/router.go` — 路由注册
