package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"the-line/backend/internal/domain"
	"the-line/backend/internal/dto"
	"the-line/backend/internal/executor"
	"the-line/backend/internal/model"
	"the-line/backend/internal/repository"
	"the-line/backend/internal/response"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AgentTaskService struct {
	db                   *gorm.DB
	agentTaskRepo        *repository.AgentTaskRepository
	agentTaskReceiptRepo *repository.AgentTaskReceiptRepository
	runRepo              *repository.RunRepository
	runNodeRepo          *repository.RunNodeRepository
	nodeLogRepo          *repository.NodeLogRepository
	agentRepo            *repository.AgentRepository
	runService           *RunService
	executor             executor.AgentExecutor
	orchestrationService *RunOrchestrationService
}

func NewAgentTaskService(
	database *gorm.DB,
	agentTaskRepo *repository.AgentTaskRepository,
	agentTaskReceiptRepo *repository.AgentTaskReceiptRepository,
	runRepo *repository.RunRepository,
	runNodeRepo *repository.RunNodeRepository,
	nodeLogRepo *repository.NodeLogRepository,
	agentRepo *repository.AgentRepository,
	runService *RunService,
) *AgentTaskService {
	return &AgentTaskService{
		db:                   database,
		agentTaskRepo:        agentTaskRepo,
		agentTaskReceiptRepo: agentTaskReceiptRepo,
		runRepo:              runRepo,
		runNodeRepo:          runNodeRepo,
		nodeLogRepo:          nodeLogRepo,
		agentRepo:            agentRepo,
		runService:           runService,
	}
}

func (s *AgentTaskService) SetExecutor(agentExecutor executor.AgentExecutor) {
	s.executor = agentExecutor
}

func (s *AgentTaskService) SetOrchestrationService(orchestrationService *RunOrchestrationService) {
	s.orchestrationService = orchestrationService
}

func (s *AgentTaskService) List(ctx context.Context, req dto.AgentTaskListRequest) ([]dto.AgentTaskResponse, int64, dto.PageQuery, error) {
	page := req.PageQuery.Normalize()
	items, total, err := s.agentTaskRepo.List(ctx, repository.AgentTaskListFilter{
		RunID:     req.RunID,
		RunNodeID: req.RunNodeID,
		Status:    strings.TrimSpace(req.Status),
		Offset:    page.Offset(),
		Limit:     page.PageSize,
	})
	if err != nil {
		return nil, 0, page, err
	}

	resp := make([]dto.AgentTaskResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, toAgentTaskResponse(item))
	}
	return resp, total, page, nil
}

func (s *AgentTaskService) Get(ctx context.Context, id uint64) (dto.AgentTaskResponse, error) {
	if id == 0 {
		return dto.AgentTaskResponse{}, response.Validation("任务 ID 不合法")
	}

	task, err := s.agentTaskRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.AgentTaskResponse{}, response.NotFound("龙虾任务不存在")
		}
		return dto.AgentTaskResponse{}, err
	}
	return toAgentTaskResponse(*task), nil
}

func (s *AgentTaskService) CreateAndDispatch(ctx context.Context, nodeID uint64) error {
	if nodeID == 0 {
		return response.Validation("节点 ID 不合法")
	}

	var task *model.AgentTask
	var agent *model.Agent

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		node, run, err := s.loadDispatchNodeAndRun(ctx, tx, nodeID)
		if err != nil {
			return err
		}
		if domain.IsTerminalRunStatus(run.CurrentStatus) {
			return nil
		}
		if !isAgentNodeType(node.NodeType) {
			return nil
		}
		if node.Status != domain.NodeStatusReady {
			return nil
		}
		if node.BoundAgentID == nil {
			return response.Validation("当前节点未绑定龙虾")
		}

		activeTask, err := s.agentTaskRepo.GetActiveByRunNodeIDWithDB(ctx, tx, node.ID)
		if err != nil {
			return err
		}
		if activeTask != nil {
			return nil
		}

		agent, err = s.agentRepo.GetByID(ctx, *node.BoundAgentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return response.Validation("绑定龙虾不存在")
			}
			return err
		}
		if agent.Status != domain.StatusEnabled {
			return response.Validation("绑定龙虾未启用")
		}

		now := time.Now()
		task = &model.AgentTask{
			RunID:         node.RunID,
			RunNodeID:     node.ID,
			AgentID:       *node.BoundAgentID,
			TaskType:      inferAgentTaskType(*node),
			InputJSON:     node.InputJSON,
			Status:        domain.AgentTaskStatusRunning,
			StartedAt:     &now,
			ResultJSON:    datatypes.JSON([]byte("{}")),
			ArtifactsJSON: datatypes.JSON([]byte("[]")),
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
		return s.nodeLogRepo.CreateWithDB(ctx, tx, &model.FlowRunNodeLog{
			RunID:        node.RunID,
			RunNodeID:    node.ID,
			LogType:      domain.LogTypeAgentRun,
			OperatorType: domain.OperatorTypeSystem,
			OperatorID:   0,
			Content:      fmt.Sprintf("创建龙虾任务 #%d 并准备调度", task.ID),
			ExtraJSON: mustServiceJSON(map[string]any{
				"task_id":   task.ID,
				"agent_id":  task.AgentID,
				"task_type": task.TaskType,
			}),
		})
	})
	if err != nil || task == nil {
		return err
	}
	if s.executor == nil {
		return s.markDispatchFailure(ctx, task.ID, "龙虾执行器未配置")
	}
	if agent == nil {
		return s.markDispatchFailure(ctx, task.ID, "绑定龙虾不存在")
	}
	if err := s.executor.Execute(ctx, task, agent); err != nil {
		return s.markDispatchFailure(ctx, task.ID, err.Error())
	}
	return nil
}

func (s *AgentTaskService) ProcessReceipt(ctx context.Context, taskID uint64, req dto.AgentReceiptRequest) error {
	if taskID == 0 {
		return response.Validation("任务 ID 不合法")
	}
	if req.AgentID == 0 {
		return response.Validation("agent_id 不能为空")
	}
	if !isValidReceiptStatus(req.Status) {
		return response.Validation("龙虾回执状态不合法")
	}

	resultJSON, err := normalizeJSON(req.Result)
	if err != nil {
		return response.Validation("result 必须是合法 JSON")
	}
	artifactsJSON, err := normalizeJSON(req.Artifacts)
	if err != nil {
		return response.Validation("artifacts 必须是合法 JSON")
	}

	var nextNodeID uint64
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		task, err := s.agentTaskRepo.GetByIDWithLock(ctx, tx, taskID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return response.NotFound("龙虾任务不存在")
			}
			return err
		}
		node, run, err := s.loadTaskNodeAndRun(ctx, tx, task)
		if err != nil {
			return err
		}
		if task.AgentID != req.AgentID {
			return response.Forbidden("龙虾任务与回执 agent_id 不匹配")
		}
		if task.Status != domain.AgentTaskStatusRunning && task.Status != domain.AgentTaskStatusQueued {
			return response.InvalidState("当前龙虾任务状态不能接收回执")
		}
		if domain.IsTerminalRunStatus(run.CurrentStatus) {
			return response.InvalidState("流程已结束，不能接收龙虾回执")
		}

		receipt := &model.AgentTaskReceipt{
			AgentTaskID:   task.ID,
			RunID:         task.RunID,
			RunNodeID:     task.RunNodeID,
			AgentID:       req.AgentID,
			ReceiptStatus: req.Status,
			PayloadJSON:   mustServiceJSON(req),
			ReceivedAt:    time.Now(),
		}
		if err := s.agentTaskReceiptRepo.CreateWithDB(ctx, tx, receipt); err != nil {
			return err
		}

		task.Status = mapReceiptToTaskStatus(req.Status)
		task.StartedAt = chooseTime(req.StartedAt, task.StartedAt)
		task.FinishedAt = chooseTime(req.FinishedAt, ptrTime(time.Now()))
		task.ResultJSON = resultJSON
		task.ArtifactsJSON = artifactsJSON
		task.ErrorMessage = strings.TrimSpace(req.ErrorMessage)
		if err := s.agentTaskRepo.UpdateWithDB(ctx, tx, task); err != nil {
			return err
		}

		nextNode, err := s.handleNodeTransitionWithDB(ctx, tx, run, node, resultJSON, req)
		if err != nil {
			return err
		}
		if nextNode != nil {
			nextNodeID = nextNode.ID
		}

		return s.nodeLogRepo.CreateWithDB(ctx, tx, &model.FlowRunNodeLog{
			RunID:        run.ID,
			RunNodeID:    node.ID,
			LogType:      domain.LogTypeAgentRun,
			OperatorType: domain.OperatorTypeAgent,
			OperatorID:   req.AgentID,
			Content:      "接收龙虾执行回执：" + req.Status,
			ExtraJSON: mustServiceJSON(map[string]any{
				"summary":       strings.TrimSpace(req.Summary),
				"status":        req.Status,
				"logs":          req.Logs,
				"error_message": strings.TrimSpace(req.ErrorMessage),
			}),
		})
	})
	if err != nil {
		return err
	}
	if nextNodeID > 0 && s.orchestrationService != nil {
		return s.orchestrationService.DispatchIfNeeded(ctx, nextNodeID)
	}
	return nil
}

func (s *AgentTaskService) markDispatchFailure(ctx context.Context, taskID uint64, message string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		task, err := s.agentTaskRepo.GetByIDWithLock(ctx, tx, taskID)
		if err != nil {
			return err
		}
		node, run, err := s.loadTaskNodeAndRun(ctx, tx, task)
		if err != nil {
			return err
		}

		now := time.Now()
		task.Status = domain.AgentTaskStatusFailed
		task.FinishedAt = &now
		task.ErrorMessage = strings.TrimSpace(message)
		if task.ErrorMessage == "" {
			task.ErrorMessage = "龙虾调度失败"
		}
		if err := s.agentTaskRepo.UpdateWithDB(ctx, tx, task); err != nil {
			return err
		}
		if err := s.runNodeRepo.UpdateWithDB(ctx, tx, node.ID, map[string]any{
			"status": domain.NodeStatusBlocked,
		}); err != nil {
			return err
		}
		if err := s.runRepo.UpdateWithDB(ctx, tx, run.ID, map[string]any{
			"current_status": domain.RunStatusBlocked,
		}); err != nil {
			return err
		}
		return s.nodeLogRepo.CreateWithDB(ctx, tx, &model.FlowRunNodeLog{
			RunID:        run.ID,
			RunNodeID:    node.ID,
			LogType:      domain.LogTypeFail,
			OperatorType: domain.OperatorTypeSystem,
			OperatorID:   0,
			Content:      "龙虾调度失败：" + task.ErrorMessage,
			ExtraJSON:    mustServiceJSON(map[string]any{"task_id": task.ID}),
		})
	})
}

func (s *AgentTaskService) handleNodeTransitionWithDB(
	ctx context.Context,
	tx *gorm.DB,
	run *model.FlowRun,
	node *model.FlowRunNode,
	resultJSON datatypes.JSON,
	req dto.AgentReceiptRequest,
) (*model.FlowRunNode, error) {
	switch req.Status {
	case domain.ReceiptStatusCompleted:
		now := time.Now()
		if err := s.runNodeRepo.UpdateWithDB(ctx, tx, node.ID, map[string]any{
			"status":       domain.NodeStatusDone,
			"output_json":  resultJSON,
			"completed_at": &now,
		}); err != nil {
			return nil, err
		}
		return s.runService.AdvanceAfterNodeDone(tx, run.ID, node.ID)
	case domain.ReceiptStatusNeedsReview:
		if err := s.runNodeRepo.UpdateWithDB(ctx, tx, node.ID, map[string]any{
			"status":      domain.NodeStatusWaitConfirm,
			"output_json": resultJSON,
		}); err != nil {
			return nil, err
		}
		if err := s.runRepo.UpdateWithDB(ctx, tx, run.ID, map[string]any{
			"current_status":    domain.RunStatusWaiting,
			"current_node_code": node.NodeCode,
		}); err != nil {
			return nil, err
		}
		return nil, nil
	case domain.ReceiptStatusFailed:
		if err := s.runNodeRepo.UpdateWithDB(ctx, tx, node.ID, map[string]any{
			"status":      domain.NodeStatusFailed,
			"output_json": resultJSON,
		}); err != nil {
			return nil, err
		}
		if err := s.runRepo.UpdateWithDB(ctx, tx, run.ID, map[string]any{
			"current_status":    domain.RunStatusBlocked,
			"current_node_code": node.NodeCode,
		}); err != nil {
			return nil, err
		}
		return nil, nil
	case domain.ReceiptStatusBlocked:
		if err := s.runNodeRepo.UpdateWithDB(ctx, tx, node.ID, map[string]any{
			"status":      domain.NodeStatusBlocked,
			"output_json": resultJSON,
		}); err != nil {
			return nil, err
		}
		if err := s.runRepo.UpdateWithDB(ctx, tx, run.ID, map[string]any{
			"current_status":    domain.RunStatusBlocked,
			"current_node_code": node.NodeCode,
		}); err != nil {
			return nil, err
		}
		return nil, nil
	default:
		return nil, response.Validation("龙虾回执状态不合法")
	}
}

func (s *AgentTaskService) loadDispatchNodeAndRun(ctx context.Context, tx *gorm.DB, nodeID uint64) (*model.FlowRunNode, *model.FlowRun, error) {
	node, err := s.runNodeRepo.GetByIDWithLock(ctx, tx, nodeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, response.NotFound("节点不存在")
		}
		return nil, nil, err
	}
	run, err := s.runRepo.GetByIDWithLock(ctx, tx, node.RunID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, response.NotFound("流程不存在")
		}
		return nil, nil, err
	}
	return node, run, nil
}

func (s *AgentTaskService) loadTaskNodeAndRun(ctx context.Context, tx *gorm.DB, task *model.AgentTask) (*model.FlowRunNode, *model.FlowRun, error) {
	node, err := s.runNodeRepo.GetByIDWithLock(ctx, tx, task.RunNodeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, response.NotFound("节点不存在")
		}
		return nil, nil, err
	}
	run, err := s.runRepo.GetByIDWithLock(ctx, tx, task.RunID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, response.NotFound("流程不存在")
		}
		return nil, nil, err
	}
	return node, run, nil
}

func isValidReceiptStatus(status string) bool {
	switch status {
	case domain.ReceiptStatusCompleted, domain.ReceiptStatusNeedsReview, domain.ReceiptStatusFailed, domain.ReceiptStatusBlocked:
		return true
	default:
		return false
	}
}

func mapReceiptToTaskStatus(status string) string {
	switch status {
	case domain.ReceiptStatusCompleted:
		return domain.AgentTaskStatusCompleted
	case domain.ReceiptStatusNeedsReview:
		return domain.AgentTaskStatusNeedsReview
	case domain.ReceiptStatusFailed:
		return domain.AgentTaskStatusFailed
	case domain.ReceiptStatusBlocked:
		return domain.AgentTaskStatusBlocked
	default:
		return domain.AgentTaskStatusFailed
	}
}

func chooseTime(value *time.Time, fallback *time.Time) *time.Time {
	if value != nil {
		return value
	}
	return fallback
}

func ptrTime(value time.Time) *time.Time {
	return &value
}

func toAgentTaskResponse(task model.AgentTask) dto.AgentTaskResponse {
	return dto.AgentTaskResponse{
		ID:            task.ID,
		RunID:         task.RunID,
		RunNodeID:     task.RunNodeID,
		AgentID:       task.AgentID,
		TaskType:      task.TaskType,
		InputJSON:     json.RawMessage(task.InputJSON),
		Status:        task.Status,
		StartedAt:     task.StartedAt,
		FinishedAt:    task.FinishedAt,
		ErrorMessage:  task.ErrorMessage,
		ResultJSON:    json.RawMessage(task.ResultJSON),
		ArtifactsJSON: json.RawMessage(task.ArtifactsJSON),
		CreatedAt:     task.CreatedAt,
		UpdatedAt:     task.UpdatedAt,
	}
}
