package service

import (
	"context"
	"strings"

	"the-line/backend/internal/domain"
	"the-line/backend/internal/model"
	"the-line/backend/internal/repository"
)

type RunOrchestrationService struct {
	runNodeRepo      *repository.RunNodeRepository
	agentTaskService *AgentTaskService
}

func NewRunOrchestrationService(runNodeRepo *repository.RunNodeRepository, agentTaskService *AgentTaskService) *RunOrchestrationService {
	return &RunOrchestrationService{
		runNodeRepo:      runNodeRepo,
		agentTaskService: agentTaskService,
	}
}

func (s *RunOrchestrationService) DispatchIfNeeded(ctx context.Context, nodeID uint64) error {
	if nodeID == 0 || s == nil || s.agentTaskService == nil {
		return nil
	}

	node, err := s.runNodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		return err
	}
	if !isAgentNodeType(node.NodeType) || node.BoundAgentID == nil {
		return nil
	}
	return s.agentTaskService.CreateAndDispatch(ctx, node.ID)
}

func isAgentNodeType(nodeType string) bool {
	switch nodeType {
	case
		domain.NodeTypeExecute,
		domain.NodeTypeAgentExecute,
		domain.NodeTypeAgentExport:
		return true
	default:
		return false
	}
}

func inferAgentTaskType(node model.FlowRunNode) string {
	if node.NodeType == domain.NodeTypeAgentExport {
		return domain.AgentTaskTypeExport
	}
	if containsAny(node.NodeCode, "query", "collect", "list", "scan", "fetch") ||
		containsAny(node.NodeName, "查询", "收集", "汇总", "拉取") {
		return domain.AgentTaskTypeQuery
	}
	if containsAny(node.NodeCode, "export", "archive", "report") ||
		containsAny(node.NodeName, "导出", "归档", "报表") {
		return domain.AgentTaskTypeExport
	}
	return domain.AgentTaskTypeBatchOperation
}

func containsAny(value string, keywords ...string) bool {
	for _, keyword := range keywords {
		if keyword != "" && strings.Contains(strings.ToLower(value), strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}
