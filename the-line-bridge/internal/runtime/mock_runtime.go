package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type MockRuntime struct{}

func NewMockRuntime() *MockRuntime {
	return &MockRuntime{}
}

func (m *MockRuntime) PlanDraft(ctx context.Context, req PlanDraftRequest) (*PlanDraftResult, error) {
	time.Sleep(500 * time.Millisecond)

	nodes := []map[string]any{
		{"node_code": "collect_data", "node_name": "收集业务数据", "node_type": "agent_execute", "sort_order": 1, "executor_type": "agent", "owner_rule": "initiator", "executor_agent_code": "data_query_agent", "result_owner_rule": "initiator", "task_type": "query", "completion_condition": "汇总待处理业务数据", "failure_condition": "查询失败", "escalation_rule": "通知发起人"},
		{"node_code": "review_data", "node_name": "审核确认数据", "node_type": "human_review", "sort_order": 2, "executor_type": "human", "owner_rule": "initiator", "result_owner_rule": "initiator", "completion_condition": "人工审核通过", "failure_condition": "审核驳回", "escalation_rule": "重新处理"},
		{"node_code": "execute_task", "node_name": "执行批量操作", "node_type": "agent_execute", "sort_order": 3, "executor_type": "agent", "owner_rule": "initiator", "executor_agent_code": "operation_agent", "result_owner_rule": "initiator", "task_type": "batch_operation", "completion_condition": "完成批量执行", "failure_condition": "执行失败", "escalation_rule": "通知发起人"},
		{"node_code": "final_acceptance", "node_name": "确认最终结果", "node_type": "human_acceptance", "sort_order": 4, "executor_type": "human", "owner_rule": "initiator", "result_owner_rule": "initiator", "completion_condition": "最终签收人确认结果", "failure_condition": "签收拒绝", "escalation_rule": "跟进修复"},
	}
	nodesJSON, _ := json.Marshal(nodes)

	return &PlanDraftResult{
		Title:       "AI 编排工作流",
		Description: "由龙虾根据自然语言需求生成的流程草案",
		Nodes:       nodesJSON,
		Summary:     "已生成一条 4 节点流程草案",
	}, nil
}

func (m *MockRuntime) ExecuteTask(ctx context.Context, req ExecuteTaskRequest) (*ExecuteTaskResult, error) {
	return &ExecuteTaskResult{
		ExternalRunID: fmt.Sprintf("mock_run_%d", time.Now().UnixMilli()),
	}, nil
}

func (m *MockRuntime) WaitForResult(ctx context.Context, sessionKey string) (*TaskResult, error) {
	time.Sleep(1 * time.Second)

	result, _ := json.Marshal(map[string]any{
		"success_count": 12,
		"failed_count":  0,
		"details":       []map[string]any{},
	})

	return &TaskResult{
		Status:  "succeeded",
		Summary: "已完成批量执行，共处理 12 条记录",
		Result:  result,
		Logs: []string{
			"开始执行任务",
			"逐条处理输入记录",
			"写回执行结果",
		},
	}, nil
}

func (m *MockRuntime) CancelTask(ctx context.Context, sessionKey string) error {
	return nil
}

func (m *MockRuntime) Health(ctx context.Context) (*HealthStatus, error) {
	return &HealthStatus{Status: "healthy", Version: "mock-1.0"}, nil
}

func (m *MockRuntime) ListAgents(ctx context.Context) ([]AgentInfo, error) {
	return []AgentInfo{
		{ID: "default-agent", Name: "默认执行龙虾"},
	}, nil
}
