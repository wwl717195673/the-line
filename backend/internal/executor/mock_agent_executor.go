package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"the-line/backend/internal/domain"
	"the-line/backend/internal/dto"
	"the-line/backend/internal/model"
)

type MockAgentExecutor struct {
	receiptCallback func(ctx context.Context, taskID uint64, req *dto.AgentReceiptRequest) error
}

func NewMockAgentExecutor(receiptCallback func(ctx context.Context, taskID uint64, req *dto.AgentReceiptRequest) error) *MockAgentExecutor {
	return &MockAgentExecutor{receiptCallback: receiptCallback}
}

func (m *MockAgentExecutor) Execute(ctx context.Context, task *model.AgentTask, agent *model.Agent) error {
	if m.receiptCallback == nil {
		return nil
	}

	taskCopy := *task
	go func() {
		time.Sleep(200 * time.Millisecond)
		receipt := buildMockReceipt(&taskCopy, agent)
		_ = m.receiptCallback(context.Background(), taskCopy.ID, receipt)
	}()
	return nil
}

func buildMockReceipt(task *model.AgentTask, agent *model.Agent) *dto.AgentReceiptRequest {
	startedAt := time.Now().Add(-200 * time.Millisecond)
	finishedAt := time.Now()

	switch task.TaskType {
	case domain.AgentTaskTypeQuery:
		return &dto.AgentReceiptRequest{
			AgentID:    task.AgentID,
			Status:     domain.ReceiptStatusCompleted,
			StartedAt:  &startedAt,
			FinishedAt: &finishedAt,
			Summary:    "已完成数据查询，共找到 3 条记录",
			Result: mustRawJSON(map[string]any{
				"records_count": 3,
				"records": []map[string]any{
					{"id": 101, "name": "课程场次 A", "video_bound": false},
					{"id": 102, "name": "课程场次 B", "video_bound": true},
					{"id": 103, "name": "课程场次 C", "video_bound": false},
				},
			}),
			Artifacts: mustRawJSON([]map[string]any{}),
			Logs: []string{
				fmt.Sprintf("agent:%s 查询流程输入", agent.Code),
				"筛选距离开课不足 2 天的课程场次",
				"生成待审核数据列表",
			},
		}
	case domain.AgentTaskTypeExport:
		return &dto.AgentReceiptRequest{
			AgentID:    task.AgentID,
			Status:     domain.ReceiptStatusCompleted,
			StartedAt:  &startedAt,
			FinishedAt: &finishedAt,
			Summary:    "已导出结果文件，待核查",
			Result: mustRawJSON(map[string]any{
				"export_name": "video_binding_result.xlsx",
				"rows":        18,
			}),
			Artifacts: mustRawJSON([]map[string]any{
				{
					"name": "video_binding_result.xlsx",
					"url":  "/uploads/mock/video_binding_result.xlsx",
					"type": "file",
				},
			}),
			Logs: []string{
				"读取节点结果数据",
				"生成导出文件",
				"回传导出链接",
			},
		}
	default:
		return &dto.AgentReceiptRequest{
			AgentID:    task.AgentID,
			Status:     domain.ReceiptStatusCompleted,
			StartedAt:  &startedAt,
			FinishedAt: &finishedAt,
			Summary:    "已完成批量执行",
			Result: mustRawJSON(map[string]any{
				"success_count": 12,
				"failed_count":  0,
				"decision":      "completed",
			}),
			Artifacts: mustRawJSON([]map[string]any{}),
			Logs: []string{
				fmt.Sprintf("agent:%s 开始批量执行", agent.Code),
				"逐条处理输入记录",
				"写回执行结果",
			},
		}
	}
}

func mustRawJSON(value any) json.RawMessage {
	bytes, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return json.RawMessage(bytes)
}
