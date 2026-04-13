package executor

import (
	"context"
	"encoding/json"
	"strings"

	"the-line/backend/internal/domain"
	"the-line/backend/internal/dto"
	"the-line/backend/internal/model"
)

type MockAgentPlannerExecutor struct{}

func NewMockAgentPlannerExecutor() *MockAgentPlannerExecutor {
	return &MockAgentPlannerExecutor{}
}

func (m *MockAgentPlannerExecutor) GenerateDraft(ctx context.Context, prompt string, agent *model.Agent) (*dto.DraftPlan, error) {
	_ = ctx
	_ = agent

	normalized := strings.TrimSpace(prompt)
	title := "AI 编排工作流"
	description := "由龙虾根据自然语言需求生成的流程草案。"
	if containsPrompt(normalized, "视频", "绑定") {
		title = "视频绑定工作流"
		description = "收集课程场次、人工确认、执行视频绑定、导出结果并最终签收。"
	}

	nodes := []dto.DraftNode{
		buildAgentDraftNode("collect_data", "收集业务数据", 1, defaultAgentCodeForPrompt(normalized, "query"), domain.AgentTaskTypeQuery, "汇总待处理业务数据"),
		buildHumanReviewNode("review_data", "审核确认数据", 2),
	}

	if containsPrompt(normalized, "绑定", "操作", "执行") {
		nodes = append(nodes, buildAgentDraftNode("execute_task", "执行批量操作", 3, defaultAgentCodeForPrompt(normalized, "execute"), domain.AgentTaskTypeBatchOperation, "完成目标对象的批量执行"))
	}
	if containsPrompt(normalized, "导出", "报表", "核查") {
		nodes = append(nodes, buildAgentExportNode("export_result", "导出处理结果", len(nodes)+1, defaultAgentCodeForPrompt(normalized, "export")))
	}

	nodes = append(nodes, buildHumanAcceptanceNode("final_acceptance", "确认最终结果", len(nodes)+1))

	plan := &dto.DraftPlan{
		Title:            title,
		Description:      description,
		Nodes:            normalizeSortOrder(nodes),
		FinalDeliverable: "结果交付与最终签收记录",
	}
	return plan, nil
}

func buildAgentDraftNode(code string, name string, order int, agentCode string, taskType string, completion string) dto.DraftNode {
	return dto.DraftNode{
		NodeCode:            code,
		NodeName:            name,
		NodeType:            domain.NodeTypeAgentExecute,
		SortOrder:           order,
		ExecutorType:        "agent",
		OwnerRule:           "initiator",
		ExecutorAgentCode:   agentCode,
		ResultOwnerRule:     "initiator",
		TaskType:            taskType,
		InputSchema:         mustPlannerJSON(map[string]any{}),
		OutputSchema:        mustPlannerJSON(map[string]any{"fields": []string{"summary", "records"}}),
		CompletionCondition: completion,
		FailureCondition:    "执行失败或返回异常",
		EscalationRule:      "通知发起人介入处理",
	}
}

func buildAgentExportNode(code string, name string, order int, agentCode string) dto.DraftNode {
	return dto.DraftNode{
		NodeCode:            code,
		NodeName:            name,
		NodeType:            domain.NodeTypeAgentExport,
		SortOrder:           order,
		ExecutorType:        "agent",
		OwnerRule:           "initiator",
		ExecutorAgentCode:   agentCode,
		ResultOwnerRule:     "initiator",
		TaskType:            domain.AgentTaskTypeExport,
		InputSchema:         mustPlannerJSON(map[string]any{}),
		OutputSchema:        mustPlannerJSON(map[string]any{"fields": []string{"file_name", "file_url"}}),
		CompletionCondition: "导出文件生成成功",
		FailureCondition:    "导出失败",
		EscalationRule:      "通知发起人重试或改为人工导出",
	}
}

func buildHumanReviewNode(code string, name string, order int) dto.DraftNode {
	return dto.DraftNode{
		NodeCode:            code,
		NodeName:            name,
		NodeType:            domain.NodeTypeHumanReview,
		SortOrder:           order,
		ExecutorType:        "human",
		OwnerRule:           "initiator",
		ResultOwnerRule:     "initiator",
		InputSchema:         mustPlannerJSON(map[string]any{}),
		OutputSchema:        mustPlannerJSON(map[string]any{}),
		CompletionCondition: "人工审核通过",
		FailureCondition:    "审核驳回或要求补充材料",
		EscalationRule:      "发起人重新处理后再次提交",
	}
}

func buildHumanAcceptanceNode(code string, name string, order int) dto.DraftNode {
	return dto.DraftNode{
		NodeCode:            code,
		NodeName:            name,
		NodeType:            domain.NodeTypeHumanAcceptance,
		SortOrder:           order,
		ExecutorType:        "human",
		OwnerRule:           "initiator",
		ResultOwnerRule:     "initiator",
		InputSchema:         mustPlannerJSON(map[string]any{}),
		OutputSchema:        mustPlannerJSON(map[string]any{}),
		CompletionCondition: "最终签收人确认结果",
		FailureCondition:    "签收人拒绝确认",
		EscalationRule:      "由流程发起人跟进修复",
	}
}

func normalizeSortOrder(nodes []dto.DraftNode) []dto.DraftNode {
	for i := range nodes {
		nodes[i].SortOrder = i + 1
	}
	return nodes
}

func containsPrompt(prompt string, keywords ...string) bool {
	for _, keyword := range keywords {
		if keyword != "" && strings.Contains(prompt, keyword) {
			return true
		}
	}
	return false
}

func defaultAgentCodeForPrompt(prompt string, stage string) string {
	switch stage {
	case "query":
		if containsPrompt(prompt, "课程", "场次", "数据") {
			return "data_query_agent"
		}
		return "query_agent"
	case "export":
		return "export_agent"
	default:
		if containsPrompt(prompt, "绑定", "录播", "资源") {
			return "batch_bind_agent"
		}
		return "operation_agent"
	}
}

func mustPlannerJSON(value any) json.RawMessage {
	bytes, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return json.RawMessage(bytes)
}
