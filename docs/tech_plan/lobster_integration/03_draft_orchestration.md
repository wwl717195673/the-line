# P3：龙虾辅助编排

## 目标

实现自然语言生成流程草案、草案编辑确认、草案转化为模板的完整链路。

---

## 1. 编排流程总览

```
用户输入自然语言需求
  → POST /api/flow-drafts
  → FlowDraftService 调用 AgentPlannerExecutor.GenerateDraft(prompt)
  → 龙虾返回结构化草案
  → 创建 FlowDraft(status=draft)
  → 前端展示草案 → 用户编辑调整
  → POST /api/flow-drafts/:id/confirm
  → 创建 FlowTemplate + FlowTemplateNodes
  → 用户从新模板发起 Run
```

---

## 2. AgentPlannerExecutor 接口

```go
type AgentPlannerExecutor interface {
    GenerateDraft(ctx context.Context, prompt string, agent *model.Agent) (*dto.DraftPlan, error)
}
```

### Mock 实现

根据 prompt 中的关键词返回预定义草案，例如：

- 包含"查询"/"收集" → 插入 agent_execute 节点（task_type=query）
- 包含"导出"/"报表" → 插入 agent_export 节点（task_type=export）
- 包含"绑定"/"操作" → 插入 agent_execute 节点（task_type=batch_operation）
- 默认追加 human_review 和 human_acceptance 节点

---

## 3. 草案结构化输出格式

`FlowDraft.StructuredPlanJSON` 和 `DraftPlan DTO` 结构一致：

```go
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
    ExecutorType        string          `json:"executor_type"`          // "agent" | "human"
    OwnerRule           string          `json:"owner_rule"`             // 谁执行该节点
    OwnerPersonID       *uint64         `json:"owner_person_id"`        // specified_person 时使用
    ExecutorAgentCode   string          `json:"executor_agent_code"`    // 仅 agent 类型
    ResultOwnerRule     string          `json:"result_owner_rule"`      // 结果责任人规则
    ResultOwnerPersonID *uint64         `json:"result_owner_person_id"` // specified_person 时使用
    TaskType            string          `json:"task_type"`              // 仅 agent 类型: query / batch_operation / export
    InputSchema         json.RawMessage `json:"input_schema"`
    OutputSchema        json.RawMessage `json:"output_schema"`
    CompletionCondition string          `json:"completion_condition"`
    FailureCondition    string          `json:"failure_condition"`
    EscalationRule      string          `json:"escalation_rule"`
}
```

**JSON 示例：**

```json
{
  "title": "视频绑定工作流",
  "description": "收集未绑课程场次、审核、执行绑定、导出结果并签收",
  "nodes": [
    {
      "node_code": "collect_data",
      "node_name": "收集距离开课不足2天的课程场次数据",
      "node_type": "agent_execute",
      "sort_order": 1,
      "executor_type": "agent",
      "owner_rule": "initiator",
      "executor_agent_code": "data_query_agent",
      "result_owner_rule": "initiator",
      "task_type": "query",
      "input_schema": {},
      "output_schema": {
        "fields": ["session_count", "session_ids", "start_times", "video_bound_status"]
      },
      "completion_condition": "返回场次列表且数量 > 0",
      "failure_condition": "查询超时或无权限",
      "escalation_rule": "通知发起人人工处理"
    },
    {
      "node_code": "review_data",
      "node_name": "审核确认数据",
      "node_type": "human_review",
      "sort_order": 2,
      "executor_type": "human",
      "owner_rule": "initiator",
      "result_owner_rule": "initiator",
      "input_schema": {},
      "output_schema": {}
    },
    {
      "node_code": "bind_videos",
      "node_name": "执行录播课资源绑定",
      "node_type": "agent_execute",
      "sort_order": 3,
      "executor_type": "agent",
      "owner_rule": "initiator",
      "executor_agent_code": "batch_bind_agent",
      "result_owner_rule": "initiator",
      "task_type": "batch_operation",
      "input_schema": {},
      "output_schema": {
        "fields": ["success_count", "failed_count", "failed_ids"]
      },
      "completion_condition": "所有绑定成功或失败数 < 阈值",
      "failure_condition": "绑定失败数 > 阈值",
      "escalation_rule": "通知发起人确认异常项"
    },
    {
      "node_code": "export_result",
      "node_name": "导出绑定情况",
      "node_type": "agent_export",
      "sort_order": 4,
      "executor_type": "agent",
      "owner_rule": "initiator",
      "executor_agent_code": "export_agent",
      "result_owner_rule": "initiator",
      "task_type": "export",
      "input_schema": {},
      "output_schema": {
        "fields": ["file_url", "file_name"]
      },
      "completion_condition": "文件生成成功",
      "failure_condition": "导出失败",
      "escalation_rule": "通知发起人重试"
    },
    {
      "node_code": "final_acceptance",
      "node_name": "确认最终结果",
      "node_type": "human_acceptance",
      "sort_order": 5,
      "executor_type": "human",
      "owner_rule": "specified_person",
      "owner_person_id": 2001,
      "result_owner_rule": "specified_person",
      "result_owner_person_id": 2001,
      "input_schema": {},
      "output_schema": {}
    }
  ],
  "final_deliverable": "视频绑定结果导出报表"
}
```

---

## 4. 草案确认 → 创建模板

### 4.1 转换逻辑

```go
func (s *FlowDraftService) Confirm(ctx context.Context, draftID uint64, personID uint64) (*model.FlowTemplate, error) {
    var createdTemplate *model.FlowTemplate

    err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        draft := s.repo.GetByIDWithLock(ctx, tx, draftID)

        // 1. 校验草案状态
        if draft.Status != domain.DraftStatusDraft {
            return errors.New("草案状态不允许确认")
        }

        // 2. 解析 StructuredPlanJSON
        var plan dto.DraftPlan
        if err := json.Unmarshal(draft.StructuredPlanJSON, &plan); err != nil {
            return err
        }

        // 3. MVP 校验
        if err := validateDraftPlan(&plan); err != nil {
            return err
        }

        // 4. 创建 FlowTemplate
        template := &model.FlowTemplate{
            Name:        plan.Title,
            Code:        generateTemplateCode(plan.Title),
            Version:     1,
            Category:    "ai_generated",
            Description: plan.Description,
            Status:      domain.TemplateStatusPublished,
        }
        if err := s.templateRepo.CreateWithDB(ctx, tx, template); err != nil {
            return err
        }

        // 5. 创建 FlowTemplateNodes
        for _, node := range plan.Nodes {
            templateNode := &model.FlowTemplateNode{
                TemplateID:          template.ID,
                NodeCode:            node.NodeCode,
                NodeName:            node.NodeName,
                NodeType:            node.NodeType,
                SortOrder:           node.SortOrder,
                DefaultOwnerRule:    node.OwnerRule,
                DefaultAgentID:      resolveAgentID(node.ExecutorAgentCode),
                ResultOwnerRule:     node.ResultOwnerRule,
                ResultOwnerPersonID: node.ResultOwnerPersonID,
                InputSchemaJSON:     node.InputSchema,
                OutputSchemaJSON:    node.OutputSchema,
                ConfigJSON:          buildNodeConfig(node),
            }
            if err := s.templateNodeRepo.CreateWithDB(ctx, tx, templateNode); err != nil {
                return err
            }
        }

        // 6. 更新草案状态
        now := time.Now()
        draft.Status = domain.DraftStatusConfirmed
        draft.ConfirmedTemplateID = &template.ID
        draft.ConfirmedAt = &now
        if err := s.repo.UpdateWithDB(ctx, tx, draft); err != nil {
            return err
        }

        createdTemplate = template
        return nil
    })
    if err != nil {
        return nil, err
    }

    return createdTemplate, nil
}
```

### 4.2 MVP 校验规则

```go
func validateDraftPlan(plan *dto.DraftPlan) error {
    // 串行流程（当前架构就是串行，无需额外检查）

    // 3-8 个节点
    if len(plan.Nodes) < 3 || len(plan.Nodes) > 8 {
        return errors.New("节点数量必须在 3-8 个之间")
    }

    // 至少 1 个人工确认节点
    hasHumanReview := false
    for _, n := range plan.Nodes {
        if n.NodeType == domain.NodeTypeHumanReview || n.NodeType == domain.NodeTypeHumanAcceptance {
            hasHumanReview = true
            break
        }
    }
    if !hasHumanReview {
        return errors.New("必须至少包含一个人工确认节点")
    }

    // 最后一个节点必须是 human_acceptance
    last := plan.Nodes[len(plan.Nodes)-1]
    if last.NodeType != domain.NodeTypeHumanAcceptance {
        return errors.New("最后一个节点必须是最终签收节点")
    }

    // 自动节点 task_type 校验
    validTaskTypes := map[string]bool{
        domain.AgentTaskTypeQuery:          true,
        domain.AgentTaskTypeBatchOperation: true,
        domain.AgentTaskTypeExport:         true,
    }
    for _, n := range plan.Nodes {
        if n.ExecutorType == "agent" && !validTaskTypes[n.TaskType] {
            return fmt.Errorf("节点 %s 的任务类型 %s 不在允许范围内", n.NodeCode, n.TaskType)
        }
    }

    // specified_person 场景必须带 person_id
    for _, n := range plan.Nodes {
        if n.OwnerRule == "specified_person" && n.OwnerPersonID == nil {
            return fmt.Errorf("节点 %s 缺少 owner_person_id", n.NodeCode)
        }
        if n.ResultOwnerRule == "specified_person" && n.ResultOwnerPersonID == nil {
            return fmt.Errorf("节点 %s 缺少 result_owner_person_id", n.NodeCode)
        }
    }

    return nil
}
```

---

## 5. FlowDraft API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/flow-drafts` | 创建草案（触发龙虾生成） |
| GET | `/api/flow-drafts` | 草案列表（按 creator 过滤） |
| GET | `/api/flow-drafts/:id` | 草案详情 |
| PUT | `/api/flow-drafts/:id` | 更新草案（用户编辑节点） |
| POST | `/api/flow-drafts/:id/confirm` | 确认草案 → 创建模板 |
| POST | `/api/flow-drafts/:id/discard` | 废弃草案 |

### 创建草案请求

```json
POST /api/flow-drafts
{
  "source_prompt": "帮我创建一个视频绑定的工作流程...",
  "planner_agent_id": 1
}
```

### 创建草案响应

```json
{
  "id": 1,
  "title": "视频绑定工作流",
  "status": "draft",
  "source_prompt": "帮我创建一个视频绑定的工作流程...",
  "structured_plan_json": { ... },
  "created_at": "2026-04-13T10:00:00Z"
}
```

### 更新草案请求

```json
PUT /api/flow-drafts/:id
{
  "title": "视频绑定工作流（修改版）",
  "structured_plan_json": {
    "title": "...",
    "nodes": [ ... ]
  }
}
```

### 确认草案响应

```json
POST /api/flow-drafts/:id/confirm
→ {
  "draft_id": 1,
  "template_id": 5,
  "message": "草案已确认，模板已创建"
}
```

---

## 6. 交付清单

- [ ] `internal/executor/agent_planner_executor.go` — 接口定义
- [ ] `internal/executor/mock_agent_planner_executor.go` — Mock 实现
- [ ] `internal/service/flow_draft_service.go` — Create（调用 Planner）、Confirm（转模板）、Discard
- [ ] `internal/handler/flow_draft_handler.go` — 6 个 API
- [ ] `internal/dto/flow_draft_dto.go` — DraftPlan / DraftNode / 请求响应 DTO
- [ ] `internal/app/router.go` — 路由注册
- [ ] MVP 校验规则实现
