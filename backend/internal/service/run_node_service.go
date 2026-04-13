package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"the-line/backend/internal/domain"
	"the-line/backend/internal/dto"
	"the-line/backend/internal/model"
	"the-line/backend/internal/repository"
	"the-line/backend/internal/response"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type RunNodeService struct {
	db             *gorm.DB
	runService     *RunService
	orchestration  *RunOrchestrationService
	runRepo        *repository.RunRepository
	runNodeRepo    *repository.RunNodeRepository
	nodeLogRepo    *repository.NodeLogRepository
	personRepo     *repository.PersonRepository
	agentRepo      *repository.AgentRepository
	commentRepo    *repository.CommentRepository
	attachmentRepo *repository.AttachmentRepository
}

func NewRunNodeService(
	database *gorm.DB,
	runService *RunService,
	runRepo *repository.RunRepository,
	runNodeRepo *repository.RunNodeRepository,
	nodeLogRepo *repository.NodeLogRepository,
	personRepo *repository.PersonRepository,
	agentRepo *repository.AgentRepository,
	commentRepo *repository.CommentRepository,
	attachmentRepo *repository.AttachmentRepository,
) *RunNodeService {
	return &RunNodeService{
		db:             database,
		runService:     runService,
		runRepo:        runRepo,
		runNodeRepo:    runNodeRepo,
		nodeLogRepo:    nodeLogRepo,
		personRepo:     personRepo,
		agentRepo:      agentRepo,
		commentRepo:    commentRepo,
		attachmentRepo: attachmentRepo,
	}
}

func (s *RunNodeService) SetOrchestrationService(orchestration *RunOrchestrationService) {
	s.orchestration = orchestration
}

func (s *RunNodeService) Detail(ctx context.Context, nodeID uint64, actor domain.Actor) (dto.RunNodeDetailResponse, error) {
	node, run, err := s.loadNodeAndRun(ctx, nodeID)
	if err != nil {
		return dto.RunNodeDetailResponse{}, err
	}

	logs, err := s.nodeLogRepo.ListByRunNodeID(ctx, node.ID)
	if err != nil {
		return dto.RunNodeDetailResponse{}, err
	}
	attachments, err := s.attachmentRepo.ListByTarget(ctx, domain.TargetTypeFlowRunNode, node.ID)
	if err != nil {
		return dto.RunNodeDetailResponse{}, err
	}
	comments, err := s.commentRepo.ListByTarget(ctx, domain.TargetTypeFlowRunNode, node.ID)
	if err != nil {
		return dto.RunNodeDetailResponse{}, err
	}

	personMap, agentMap, err := s.loadNodeRelationMaps(ctx, *run, *node)
	if err != nil {
		return dto.RunNodeDetailResponse{}, err
	}
	if err := s.mergeCommentAuthors(ctx, personMap, comments); err != nil {
		return dto.RunNodeDetailResponse{}, err
	}

	nodeResp := toRunNodeResponse(*node, run.CurrentNodeCode, personMap, agentMap)
	runResp := toRunResponse(*run, personMap)
	runResp.CurrentNode = &nodeResp

	return dto.RunNodeDetailResponse{
		RunNodeResponse:  nodeResp,
		Run:              &runResp,
		Attachments:      toAttachmentResponses(attachments),
		Comments:         toCommentResponses(comments, personMap),
		Logs:             toRunNodeLogResponses(logs),
		AvailableActions: s.availableActions(*node, *run, actor),
	}, nil
}

func (s *RunNodeService) SaveInput(ctx context.Context, nodeID uint64, req dto.SaveRunNodeInputRequest, actor domain.Actor) (dto.RunNodeDetailResponse, error) {
	inputJSON, err := normalizeJSON(req.InputJSON)
	if err != nil {
		return dto.RunNodeDetailResponse{}, response.Validation("节点输入必须是合法 JSON")
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		node, run, err := s.loadNodeAndRunWithLock(ctx, tx, nodeID)
		if err != nil {
			return err
		}
		if err := ensureFlowEditable(*run); err != nil {
			return err
		}
		if node.Status == domain.NodeStatusDone {
			return response.InvalidState("已完成节点不能暂存输入")
		}
		if !canOperateNode(*node, actor, true) {
			return response.Forbidden("当前用户不能暂存该节点输入")
		}
		if err := s.runNodeRepo.UpdateWithDB(ctx, tx, node.ID, map[string]any{
			"input_json": inputJSON,
		}); err != nil {
			return err
		}
		return s.appendNodeLog(ctx, tx, node.RunID, node.ID, domain.LogTypeSaveInput, actor, "暂存节点输入", map[string]any{})
	})
	if err != nil {
		return dto.RunNodeDetailResponse{}, err
	}

	return s.Detail(ctx, nodeID, actor)
}

func (s *RunNodeService) Submit(ctx context.Context, nodeID uint64, req dto.SubmitRunNodeRequest, actor domain.Actor) (dto.RunNodeDetailResponse, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		node, run, err := s.loadNodeAndRunWithLock(ctx, tx, nodeID)
		if err != nil {
			return err
		}
		if err := ensureFlowEditable(*run); err != nil {
			return err
		}
		if !canOwnerOperate(*node, actor) {
			return response.Forbidden("当前用户不能提交该节点")
		}
		if config, ok := fixedNodeConfig(node.NodeCode); ok {
			if config.NeedReview {
				return response.InvalidState("固定审核节点无需提交确认，请直接审核")
			}
			return response.InvalidState("固定非审核节点请使用标记完成")
		}
		if !statusIn(node.Status, domain.NodeStatusReady, domain.NodeStatusRunning, domain.NodeStatusWaitMaterial) {
			return response.InvalidState("当前节点状态不能提交确认")
		}
		if err := s.validateNodeInput(ctx, *node); err != nil {
			return err
		}

		if err := s.runNodeRepo.UpdateWithDB(ctx, tx, node.ID, map[string]any{
			"status": domain.NodeStatusWaitConfirm,
		}); err != nil {
			return err
		}
		if err := s.runRepo.UpdateWithDB(ctx, tx, run.ID, map[string]any{
			"current_status": domain.RunStatusWaiting,
		}); err != nil {
			return err
		}
		return s.appendNodeLog(ctx, tx, node.RunID, node.ID, domain.LogTypeSubmit, actor, "提交节点确认", map[string]any{"comment": strings.TrimSpace(req.Comment)})
	})
	if err != nil {
		return dto.RunNodeDetailResponse{}, err
	}

	return s.Detail(ctx, nodeID, actor)
}

func (s *RunNodeService) Approve(ctx context.Context, nodeID uint64, req dto.ApproveRunNodeRequest, actor domain.Actor) (dto.RunNodeDetailResponse, error) {
	var nextNodeID uint64
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		node, run, err := s.loadNodeAndRunWithLock(ctx, tx, nodeID)
		if err != nil {
			return err
		}
		if err := ensureFlowEditable(*run); err != nil {
			return err
		}
		if !canReviewerOperate(*node, actor) {
			return response.Forbidden("当前用户不能审核该节点")
		}
		if !canReviewInStatus(node.Status) {
			return response.InvalidState("当前节点状态不能审核通过")
		}
		outputJSON, err := s.buildApproveOutput(*node, req)
		if err != nil {
			return err
		}

		now := time.Now()
		if err := s.runNodeRepo.UpdateWithDB(ctx, tx, node.ID, map[string]any{
			"status":       domain.NodeStatusDone,
			"output_json":  outputJSON,
			"completed_at": &now,
		}); err != nil {
			return err
		}
		nextNode, err := s.runService.AdvanceAfterNodeDone(tx, node.RunID, node.ID)
		if err != nil {
			return err
		}
		if nextNode != nil {
			nextNodeID = nextNode.ID
		}
		return s.appendNodeLog(ctx, tx, node.RunID, node.ID, domain.LogTypeApprove, actor, "审核通过节点", map[string]any{
			"review_comment": strings.TrimSpace(req.ReviewComment),
			"final_plan":     strings.TrimSpace(req.FinalPlan),
		})
	})
	if err != nil {
		return dto.RunNodeDetailResponse{}, err
	}
	if nextNodeID > 0 && s.orchestration != nil {
		if err := s.orchestration.DispatchIfNeeded(ctx, nextNodeID); err != nil {
			return dto.RunNodeDetailResponse{}, err
		}
	}

	return s.Detail(ctx, nodeID, actor)
}

func (s *RunNodeService) Reject(ctx context.Context, nodeID uint64, req dto.RejectRunNodeRequest, actor domain.Actor) (dto.RunNodeDetailResponse, error) {
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		return dto.RunNodeDetailResponse{}, response.Validation("驳回原因不能为空")
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		node, run, err := s.loadNodeAndRunWithLock(ctx, tx, nodeID)
		if err != nil {
			return err
		}
		if err := ensureFlowEditable(*run); err != nil {
			return err
		}
		if !canReviewerOperate(*node, actor) {
			return response.Forbidden("当前用户不能驳回该节点")
		}
		if !canReviewInStatus(node.Status) {
			return response.InvalidState("当前节点状态不能驳回")
		}

		if err := s.runNodeRepo.UpdateWithDB(ctx, tx, node.ID, map[string]any{
			"status": domain.NodeStatusRejected,
		}); err != nil {
			return err
		}
		if err := s.runRepo.UpdateWithDB(ctx, tx, run.ID, map[string]any{
			"current_status": domain.RunStatusWaiting,
		}); err != nil {
			return err
		}
		return s.appendNodeLog(ctx, tx, node.RunID, node.ID, domain.LogTypeReject, actor, "驳回节点："+reason, map[string]any{"reason": reason})
	})
	if err != nil {
		return dto.RunNodeDetailResponse{}, err
	}

	return s.Detail(ctx, nodeID, actor)
}

func (s *RunNodeService) RequestMaterial(ctx context.Context, nodeID uint64, req dto.RequestMaterialRunNodeRequest, actor domain.Actor) (dto.RunNodeDetailResponse, error) {
	requirement := strings.TrimSpace(req.Requirement)
	if requirement == "" {
		return dto.RunNodeDetailResponse{}, response.Validation("补充材料要求不能为空")
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		node, run, err := s.loadNodeAndRunWithLock(ctx, tx, nodeID)
		if err != nil {
			return err
		}
		if err := ensureFlowEditable(*run); err != nil {
			return err
		}
		if !canReviewerOperate(*node, actor) {
			return response.Forbidden("当前用户不能要求该节点补材料")
		}
		if !canReviewInStatus(node.Status) {
			return response.InvalidState("当前节点状态不能要求补材料")
		}

		if err := s.runNodeRepo.UpdateWithDB(ctx, tx, node.ID, map[string]any{
			"status": domain.NodeStatusWaitMaterial,
		}); err != nil {
			return err
		}
		if err := s.runRepo.UpdateWithDB(ctx, tx, run.ID, map[string]any{
			"current_status": domain.RunStatusWaiting,
		}); err != nil {
			return err
		}
		return s.appendNodeLog(ctx, tx, node.RunID, node.ID, domain.LogTypeRequestMaterial, actor, "要求补充材料："+requirement, map[string]any{"requirement": requirement})
	})
	if err != nil {
		return dto.RunNodeDetailResponse{}, err
	}

	return s.Detail(ctx, nodeID, actor)
}

func (s *RunNodeService) Complete(ctx context.Context, nodeID uint64, req dto.CompleteRunNodeRequest, actor domain.Actor) (dto.RunNodeDetailResponse, error) {
	var nextNodeID uint64
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		node, run, err := s.loadNodeAndRunWithLock(ctx, tx, nodeID)
		if err != nil {
			return err
		}
		if err := ensureFlowEditable(*run); err != nil {
			return err
		}
		if !canOwnerOperate(*node, actor) {
			return response.Forbidden("当前用户不能完成该节点")
		}
		if !statusIn(node.Status, domain.NodeStatusReady, domain.NodeStatusRunning, domain.NodeStatusWaitMaterial) {
			return response.InvalidState("当前节点状态不能标记完成")
		}
		if isAgentNodeType(node.NodeType) {
			return response.InvalidState("自动节点不能手动标记完成，请等待龙虾回执或使用人工接管")
		}
		if nodeRequiresReview(*node) {
			return response.InvalidState("审核节点必须通过审核完成")
		}
		if err := s.validateNodeInput(ctx, *node); err != nil {
			return err
		}
		outputJSON, err := normalizeJSON(req.OutputJSON)
		if err != nil {
			return response.Validation("节点输出必须是合法 JSON")
		}

		now := time.Now()
		if err := s.runNodeRepo.UpdateWithDB(ctx, tx, node.ID, map[string]any{
			"status":       domain.NodeStatusDone,
			"output_json":  outputJSON,
			"completed_at": &now,
		}); err != nil {
			return err
		}
		nextNode, err := s.runService.AdvanceAfterNodeDone(tx, node.RunID, node.ID)
		if err != nil {
			return err
		}
		if nextNode != nil {
			nextNodeID = nextNode.ID
		}
		return s.appendNodeLog(ctx, tx, node.RunID, node.ID, domain.LogTypeComplete, actor, "标记节点完成", map[string]any{"comment": strings.TrimSpace(req.Comment)})
	})
	if err != nil {
		return dto.RunNodeDetailResponse{}, err
	}
	if nextNodeID > 0 && s.orchestration != nil {
		if err := s.orchestration.DispatchIfNeeded(ctx, nextNodeID); err != nil {
			return dto.RunNodeDetailResponse{}, err
		}
	}

	return s.Detail(ctx, nodeID, actor)
}

func (s *RunNodeService) Fail(ctx context.Context, nodeID uint64, req dto.FailRunNodeRequest, actor domain.Actor) (dto.RunNodeDetailResponse, error) {
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		return dto.RunNodeDetailResponse{}, response.Validation("异常原因不能为空")
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		node, run, err := s.loadNodeAndRunWithLock(ctx, tx, nodeID)
		if err != nil {
			return err
		}
		if err := ensureFlowEditable(*run); err != nil {
			return err
		}
		if node.Status == domain.NodeStatusDone {
			return response.InvalidState("已完成节点不能标记异常")
		}
		if !canOperateNode(*node, actor, true) {
			return response.Forbidden("当前用户不能标记该节点异常")
		}

		if err := s.runNodeRepo.UpdateWithDB(ctx, tx, node.ID, map[string]any{
			"status": domain.NodeStatusFailed,
		}); err != nil {
			return err
		}
		if err := s.runRepo.UpdateWithDB(ctx, tx, run.ID, map[string]any{
			"current_status": domain.RunStatusBlocked,
		}); err != nil {
			return err
		}
		return s.appendNodeLog(ctx, tx, node.RunID, node.ID, domain.LogTypeFail, actor, "标记节点异常："+reason, map[string]any{"reason": reason})
	})
	if err != nil {
		return dto.RunNodeDetailResponse{}, err
	}

	return s.Detail(ctx, nodeID, actor)
}

func (s *RunNodeService) RunAgent(ctx context.Context, nodeID uint64, actor domain.Actor) (dto.RunNodeDetailResponse, error) {
	node, run, err := s.loadNodeAndRun(ctx, nodeID)
	if err != nil {
		return dto.RunNodeDetailResponse{}, err
	}
	if err := ensureFlowEditable(*run); err != nil {
		return dto.RunNodeDetailResponse{}, err
	}
	if !canOwnerOperate(*node, actor) {
		return dto.RunNodeDetailResponse{}, response.Forbidden("当前用户不能运行该节点龙虾")
	}
	if !statusIn(node.Status, domain.NodeStatusReady, domain.NodeStatusRunning, domain.NodeStatusWaitMaterial) {
		return dto.RunNodeDetailResponse{}, response.InvalidState("当前节点状态不能运行龙虾")
	}
	if !isAgentNodeType(node.NodeType) {
		return dto.RunNodeDetailResponse{}, response.InvalidState("当前节点不是龙虾自动节点")
	}
	if node.BoundAgentID == nil {
		return dto.RunNodeDetailResponse{}, response.Validation("当前节点未绑定龙虾")
	}
	if s.orchestration == nil {
		return dto.RunNodeDetailResponse{}, response.Internal("流程调度器未配置")
	}
	if err := s.orchestration.DispatchIfNeeded(ctx, node.ID); err != nil {
		return dto.RunNodeDetailResponse{}, err
	}

	return s.Detail(ctx, nodeID, actor)
}

func (s *RunNodeService) ConfirmAgentResult(ctx context.Context, nodeID uint64, req dto.ConfirmAgentResultRequest, actor domain.Actor) (dto.RunNodeDetailResponse, error) {
	action := strings.TrimSpace(req.Action)
	if action != "approve" && action != "reject" {
		return dto.RunNodeDetailResponse{}, response.Validation("确认动作只能是 approve 或 reject")
	}

	var nextNodeID uint64
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		node, run, err := s.loadNodeAndRunWithLock(ctx, tx, nodeID)
		if err != nil {
			return err
		}
		if err := ensureFlowEditable(*run); err != nil {
			return err
		}
		if node.Status != domain.NodeStatusWaitConfirm {
			return response.InvalidState("当前节点状态不能确认龙虾结果")
		}
		if !isAgentNodeType(node.NodeType) {
			return response.InvalidState("当前节点不是龙虾自动节点")
		}
		if !canResultOwnerOperate(*node, actor) {
			return response.Forbidden("当前用户不能确认龙虾结果")
		}

		if action == "approve" {
			now := time.Now()
			if err := s.runNodeRepo.UpdateWithDB(ctx, tx, node.ID, map[string]any{
				"status":       domain.NodeStatusDone,
				"completed_at": &now,
			}); err != nil {
				return err
			}
			nextNode, err := s.runService.AdvanceAfterNodeDone(tx, node.RunID, node.ID)
			if err != nil {
				return err
			}
			if nextNode != nil {
				nextNodeID = nextNode.ID
			}
			return s.appendNodeLog(ctx, tx, node.RunID, node.ID, domain.LogTypeApprove, actor, "确认龙虾结果通过", map[string]any{
				"comment": strings.TrimSpace(req.Comment),
			})
		}

		if err := s.runNodeRepo.UpdateWithDB(ctx, tx, node.ID, map[string]any{
			"status": domain.NodeStatusBlocked,
		}); err != nil {
			return err
		}
		if err := s.runRepo.UpdateWithDB(ctx, tx, run.ID, map[string]any{
			"current_status":    domain.RunStatusBlocked,
			"current_node_code": node.NodeCode,
		}); err != nil {
			return err
		}
		return s.appendNodeLog(ctx, tx, node.RunID, node.ID, domain.LogTypeReject, actor, "驳回龙虾执行结果", map[string]any{
			"comment": strings.TrimSpace(req.Comment),
		})
	})
	if err != nil {
		return dto.RunNodeDetailResponse{}, err
	}
	if nextNodeID > 0 && s.orchestration != nil {
		if err := s.orchestration.DispatchIfNeeded(ctx, nextNodeID); err != nil {
			return dto.RunNodeDetailResponse{}, err
		}
	}
	return s.Detail(ctx, nodeID, actor)
}

func (s *RunNodeService) Takeover(ctx context.Context, nodeID uint64, req dto.TakeoverRunNodeRequest, actor domain.Actor) (dto.RunNodeDetailResponse, error) {
	action := strings.TrimSpace(req.Action)
	if action != "retry" && action != "manual_complete" {
		return dto.RunNodeDetailResponse{}, response.Validation("接管动作只能是 retry 或 manual_complete")
	}

	var dispatchNodeID uint64
	var nextNodeID uint64
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		node, run, err := s.loadNodeAndRunWithLock(ctx, tx, nodeID)
		if err != nil {
			return err
		}
		if err := ensureFlowEditable(*run); err != nil {
			return err
		}
		if !statusIn(node.Status, domain.NodeStatusBlocked, domain.NodeStatusFailed) {
			return response.InvalidState("当前节点状态不能人工接管")
		}
		if !canResultOwnerOperate(*node, actor) {
			return response.Forbidden("当前用户不能接管该节点")
		}

		if action == "retry" {
			if !isAgentNodeType(node.NodeType) || node.BoundAgentID == nil {
				return response.InvalidState("当前节点不能重试龙虾执行")
			}
			now := time.Now()
			if err := s.runNodeRepo.UpdateWithDB(ctx, tx, node.ID, map[string]any{
				"status":       domain.NodeStatusReady,
				"completed_at": nil,
				"started_at":   &now,
			}); err != nil {
				return err
			}
			if err := s.runRepo.UpdateWithDB(ctx, tx, run.ID, map[string]any{
				"current_status":    domain.RunStatusRunning,
				"current_node_code": node.NodeCode,
			}); err != nil {
				return err
			}
			dispatchNodeID = node.ID
			return s.appendNodeLog(ctx, tx, node.RunID, node.ID, domain.LogTypeAgentRun, actor, "人工接管后重试龙虾执行", map[string]any{
				"comment": strings.TrimSpace(req.Comment),
			})
		}

		manualResult, err := normalizeJSON(req.ManualResult)
		if err != nil {
			return response.Validation("manual_result 必须是合法 JSON")
		}
		now := time.Now()
		if err := s.runNodeRepo.UpdateWithDB(ctx, tx, node.ID, map[string]any{
			"status":       domain.NodeStatusDone,
			"output_json":  manualResult,
			"completed_at": &now,
		}); err != nil {
			return err
		}
		nextNode, err := s.runService.AdvanceAfterNodeDone(tx, node.RunID, node.ID)
		if err != nil {
			return err
		}
		if nextNode != nil {
			nextNodeID = nextNode.ID
		}
		return s.appendNodeLog(ctx, tx, node.RunID, node.ID, domain.LogTypeComplete, actor, "人工接管并手动完成节点", map[string]any{
			"comment": strings.TrimSpace(req.Comment),
		})
	})
	if err != nil {
		return dto.RunNodeDetailResponse{}, err
	}
	if dispatchNodeID > 0 && s.orchestration != nil {
		if err := s.orchestration.DispatchIfNeeded(ctx, dispatchNodeID); err != nil {
			return dto.RunNodeDetailResponse{}, err
		}
	}
	if nextNodeID > 0 && s.orchestration != nil {
		if err := s.orchestration.DispatchIfNeeded(ctx, nextNodeID); err != nil {
			return dto.RunNodeDetailResponse{}, err
		}
	}
	return s.Detail(ctx, nodeID, actor)
}

func (s *RunNodeService) Logs(ctx context.Context, nodeID uint64, actor domain.Actor) ([]dto.RunNodeLogResponse, error) {
	if _, _, err := s.loadNodeAndRun(ctx, nodeID); err != nil {
		return nil, err
	}
	logs, err := s.nodeLogRepo.ListByRunNodeID(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	return toRunNodeLogResponses(logs), nil
}

func (s *RunNodeService) loadNodeAndRun(ctx context.Context, nodeID uint64) (*model.FlowRunNode, *model.FlowRun, error) {
	if nodeID == 0 {
		return nil, nil, response.Validation("节点 ID 不合法")
	}
	node, err := s.runNodeRepo.GetByID(ctx, nodeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, response.NotFound("节点不存在")
		}
		return nil, nil, err
	}
	run, err := s.runRepo.GetByID(ctx, node.RunID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, response.NotFound("流程不存在")
		}
		return nil, nil, err
	}
	return node, run, nil
}

func (s *RunNodeService) loadNodeAndRunWithLock(ctx context.Context, tx *gorm.DB, nodeID uint64) (*model.FlowRunNode, *model.FlowRun, error) {
	if nodeID == 0 {
		return nil, nil, response.Validation("节点 ID 不合法")
	}
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

func (s *RunNodeService) validateNodeInput(ctx context.Context, node model.FlowRunNode) error {
	config, ok := fixedNodeConfig(node.NodeCode)
	if !ok {
		return nil
	}
	if err := validateRequiredFields(node.InputJSON, mustServiceJSON(map[string]any{"required_fields": config.RequiredFields})); err != nil {
		return err
	}
	if config.RequireAttachment {
		count, err := s.attachmentRepo.CountByTarget(ctx, domain.TargetTypeFlowRunNode, node.ID)
		if err != nil {
			return err
		}
		if count == 0 {
			return response.Validation(fmt.Sprintf("节点 %s 至少需要 1 个附件", node.NodeName))
		}
	}
	return nil
}

func (s *RunNodeService) buildApproveOutput(node model.FlowRunNode, req dto.ApproveRunNodeRequest) (datatypes.JSON, error) {
	if node.NodeCode == "middle_office_review" && strings.TrimSpace(req.ReviewComment) == "" {
		return nil, response.Validation("审核意见不能为空")
	}
	if node.NodeCode == "operation_confirm_plan" && strings.TrimSpace(req.FinalPlan) == "" {
		return nil, response.Validation("最终甩班方案不能为空")
	}

	if len(req.OutputJSON) > 0 {
		outputJSON, err := normalizeJSON(req.OutputJSON)
		if err != nil {
			return nil, response.Validation("节点输出必须是合法 JSON")
		}
		return outputJSON, nil
	}

	return mustServiceJSON(map[string]any{
		"review_comment": strings.TrimSpace(req.ReviewComment),
		"final_plan":     strings.TrimSpace(req.FinalPlan),
		"decision":       "approved",
	}), nil
}

func (s *RunNodeService) appendNodeLog(ctx context.Context, tx *gorm.DB, runID uint64, nodeID uint64, logType string, actor domain.Actor, content string, extra map[string]any) error {
	return s.nodeLogRepo.CreateWithDB(ctx, tx, &model.FlowRunNodeLog{
		RunID:        runID,
		RunNodeID:    nodeID,
		LogType:      logType,
		OperatorType: domain.OperatorTypePerson,
		OperatorID:   actor.PersonID,
		Content:      content,
		ExtraJSON:    mustServiceJSON(extra),
	})
}

func (s *RunNodeService) availableActions(node model.FlowRunNode, run model.FlowRun, actor domain.Actor) []string {
	if domain.IsTerminalRunStatus(run.CurrentStatus) || node.Status == domain.NodeStatusDone {
		return []string{}
	}

	actions := make([]string, 0)
	if canOperateNode(node, actor, true) {
		actions = append(actions, "save_input", "fail")
	}
	if canOwnerOperate(node, actor) && statusIn(node.Status, domain.NodeStatusReady, domain.NodeStatusRunning, domain.NodeStatusWaitMaterial) {
		if !nodeRequiresReview(node) && !isAgentNodeType(node.NodeType) {
			actions = append(actions, "complete")
		}
	}
	if canReviewerOperate(node, actor) && nodeRequiresReview(node) && canReviewInStatus(node.Status) {
		actions = append(actions, "approve", "reject", "request_material")
	}
	if node.Status == domain.NodeStatusWaitConfirm && canResultOwnerOperate(node, actor) {
		actions = append(actions, "confirm_agent_result")
	}
	if statusIn(node.Status, domain.NodeStatusBlocked, domain.NodeStatusFailed) && canResultOwnerOperate(node, actor) {
		actions = append(actions, "takeover")
	}
	return actions
}

func ensureFlowEditable(run model.FlowRun) error {
	switch run.CurrentStatus {
	case domain.RunStatusCancelled:
		return response.InvalidState("已取消流程不能处理节点")
	case domain.RunStatusCompleted:
		return response.InvalidState("已完成流程不能处理节点")
	default:
		return nil
	}
}

func canOperateNode(node model.FlowRunNode, actor domain.Actor, allowReviewer bool) bool {
	if actor.IsAdmin() {
		return true
	}
	if actor.PersonID == 0 {
		return false
	}
	if node.OwnerPersonID != nil && *node.OwnerPersonID == actor.PersonID {
		return true
	}
	return allowReviewer && canReviewerOperate(node, actor)
}

func canOwnerOperate(node model.FlowRunNode, actor domain.Actor) bool {
	if actor.IsAdmin() {
		return true
	}
	return actor.PersonID != 0 && node.OwnerPersonID != nil && *node.OwnerPersonID == actor.PersonID
}

func canReviewerOperate(node model.FlowRunNode, actor domain.Actor) bool {
	if actor.IsAdmin() {
		return true
	}
	if actor.PersonID == 0 {
		return false
	}
	if node.ReviewerPersonID != nil && *node.ReviewerPersonID == actor.PersonID {
		return true
	}
	return node.ReviewerPersonID == nil && node.OwnerPersonID != nil && *node.OwnerPersonID == actor.PersonID
}

func canResultOwnerOperate(node model.FlowRunNode, actor domain.Actor) bool {
	if actor.IsAdmin() {
		return true
	}
	if actor.PersonID == 0 {
		return false
	}
	if node.ResultOwnerPersonID != nil {
		return *node.ResultOwnerPersonID == actor.PersonID
	}
	if node.OwnerPersonID != nil {
		return *node.OwnerPersonID == actor.PersonID
	}
	return false
}

func statusIn(status string, allowed ...string) bool {
	for _, item := range allowed {
		if status == item {
			return true
		}
	}
	return false
}

func canReviewInStatus(status string) bool {
	return statusIn(status, domain.NodeStatusReady, domain.NodeStatusRunning, domain.NodeStatusWaitConfirm, domain.NodeStatusWaitMaterial)
}

func nodeRequiresReview(node model.FlowRunNode) bool {
	if isReviewNodeType(node.NodeType) {
		return true
	}
	config, ok := fixedNodeConfig(node.NodeCode)
	return ok && config.NeedReview
}

func fixedNodeConfig(nodeCode string) (domain.FixedNodeConfig, bool) {
	for _, node := range domain.TeacherClassTransferTemplate().Nodes {
		if node.NodeCode == nodeCode {
			return node, true
		}
	}
	return domain.FixedNodeConfig{}, false
}

func (s *RunNodeService) loadNodeRelationMaps(ctx context.Context, run model.FlowRun, node model.FlowRunNode) (map[uint64]model.Person, map[uint64]model.Agent, error) {
	return s.runService.loadRunRelationMaps(ctx, []model.FlowRun{run}, []model.FlowRunNode{node})
}

func (s *RunNodeService) mergeCommentAuthors(ctx context.Context, personMap map[uint64]model.Person, comments []model.Comment) error {
	missingIDs := make([]uint64, 0)
	seen := map[uint64]struct{}{}
	for _, comment := range comments {
		if _, ok := personMap[comment.AuthorPersonID]; ok {
			continue
		}
		if _, ok := seen[comment.AuthorPersonID]; ok {
			continue
		}
		seen[comment.AuthorPersonID] = struct{}{}
		missingIDs = append(missingIDs, comment.AuthorPersonID)
	}
	if len(missingIDs) == 0 {
		return nil
	}

	persons, err := s.personRepo.GetByIDs(ctx, missingIDs)
	if err != nil {
		return err
	}
	for _, person := range persons {
		personMap[person.ID] = person
	}
	return nil
}

func toAttachmentResponses(attachments []model.Attachment) []dto.AttachmentResponse {
	responses := make([]dto.AttachmentResponse, 0, len(attachments))
	for _, attachment := range attachments {
		responses = append(responses, toAttachmentResponse(attachment))
	}
	return responses
}
