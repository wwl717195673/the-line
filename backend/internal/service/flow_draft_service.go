package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
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

var nonCodeCharPattern = regexp.MustCompile(`[^a-z0-9]+`)

type FlowDraftService struct {
	db              *gorm.DB
	flowDraftRepo   *repository.FlowDraftRepository
	templateRepo    *repository.TemplateRepository
	personRepo      *repository.PersonRepository
	agentRepo       *repository.AgentRepository
	plannerExecutor executor.AgentPlannerExecutor
}

func NewFlowDraftService(
	database *gorm.DB,
	flowDraftRepo *repository.FlowDraftRepository,
	templateRepo *repository.TemplateRepository,
	personRepo *repository.PersonRepository,
	agentRepo *repository.AgentRepository,
) *FlowDraftService {
	return &FlowDraftService{
		db:            database,
		flowDraftRepo: flowDraftRepo,
		templateRepo:  templateRepo,
		personRepo:    personRepo,
		agentRepo:     agentRepo,
	}
}

func (s *FlowDraftService) SetPlannerExecutor(plannerExecutor executor.AgentPlannerExecutor) {
	s.plannerExecutor = plannerExecutor
}

func (s *FlowDraftService) List(ctx context.Context, req dto.FlowDraftListRequest) ([]dto.FlowDraftResponse, int64, dto.PageQuery, error) {
	page := req.PageQuery.Normalize()
	status := strings.TrimSpace(req.Status)
	if status != "" && !isValidDraftStatus(status) {
		return nil, 0, page, response.Validation("草案状态不合法")
	}

	items, total, err := s.flowDraftRepo.List(ctx, repository.FlowDraftListFilter{
		CreatorPersonID: req.CreatorPersonID,
		Status:          status,
		Offset:          page.Offset(),
		Limit:           page.PageSize,
	})
	if err != nil {
		return nil, 0, page, err
	}

	resp := make([]dto.FlowDraftResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, toFlowDraftResponse(item))
	}
	return resp, total, page, nil
}

func (s *FlowDraftService) Get(ctx context.Context, id uint64) (dto.FlowDraftResponse, error) {
	if id == 0 {
		return dto.FlowDraftResponse{}, response.Validation("草案 ID 不合法")
	}

	draft, err := s.flowDraftRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.FlowDraftResponse{}, response.NotFound("草案不存在")
		}
		return dto.FlowDraftResponse{}, err
	}
	return toFlowDraftResponse(*draft), nil
}

func (s *FlowDraftService) Create(ctx context.Context, req dto.CreateFlowDraftRequest) (dto.FlowDraftResponse, error) {
	sourcePrompt := strings.TrimSpace(req.SourcePrompt)
	if sourcePrompt == "" {
		return dto.FlowDraftResponse{}, response.Validation("草案原始需求不能为空")
	}
	if req.CreatorPersonID == 0 {
		return dto.FlowDraftResponse{}, response.Validation("草案创建人不能为空")
	}
	if err := s.ensurePersonExists(ctx, req.CreatorPersonID); err != nil {
		return dto.FlowDraftResponse{}, err
	}

	var plan *dto.DraftPlan
	var structuredPlanJSON datatypes.JSON
	if len(req.StructuredPlanJSON) > 0 {
		if err := json.Unmarshal(req.StructuredPlanJSON, &plan); err != nil {
			return dto.FlowDraftResponse{}, response.Validation("草案结构化计划必须是合法 JSON")
		}
	} else {
		if req.PlannerAgentID == nil {
			return dto.FlowDraftResponse{}, response.Validation("缺少编排龙虾或结构化计划")
		}
		plannerAgent, err := s.ensureAgentExists(ctx, *req.PlannerAgentID)
		if err != nil {
			return dto.FlowDraftResponse{}, err
		}
		if s.plannerExecutor == nil {
			return dto.FlowDraftResponse{}, response.Internal("草案编排器未配置")
		}
		plan, err = s.plannerExecutor.GenerateDraft(ctx, sourcePrompt, plannerAgent)
		if err != nil {
			return dto.FlowDraftResponse{}, err
		}
	}
	if plan == nil {
		return dto.FlowDraftResponse{}, response.Validation("草案结构化计划不能为空")
	}
	if err := validateDraftPlan(plan); err != nil {
		return dto.FlowDraftResponse{}, err
	}
	structuredPlanJSON = mustServiceJSON(plan)

	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = strings.TrimSpace(plan.Title)
	}
	if title == "" {
		return dto.FlowDraftResponse{}, response.Validation("草案标题不能为空")
	}
	description := strings.TrimSpace(req.Description)
	if description == "" {
		description = strings.TrimSpace(plan.Description)
	}

	draft := &model.FlowDraft{
		Title:              title,
		Description:        description,
		SourcePrompt:       sourcePrompt,
		CreatorPersonID:    req.CreatorPersonID,
		PlannerAgentID:     req.PlannerAgentID,
		Status:             domain.DraftStatusDraft,
		StructuredPlanJSON: structuredPlanJSON,
	}
	if err := s.flowDraftRepo.Create(ctx, draft); err != nil {
		return dto.FlowDraftResponse{}, err
	}
	return toFlowDraftResponse(*draft), nil
}

func (s *FlowDraftService) Update(ctx context.Context, id uint64, req dto.UpdateFlowDraftRequest) (dto.FlowDraftResponse, error) {
	if id == 0 {
		return dto.FlowDraftResponse{}, response.Validation("草案 ID 不合法")
	}

	draft, err := s.flowDraftRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.FlowDraftResponse{}, response.NotFound("草案不存在")
		}
		return dto.FlowDraftResponse{}, err
	}
	if draft.Status != domain.DraftStatusDraft {
		return dto.FlowDraftResponse{}, response.InvalidState("当前草案状态不能编辑")
	}

	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if title == "" {
			return dto.FlowDraftResponse{}, response.Validation("草案标题不能为空")
		}
		draft.Title = title
	}
	if req.Description != nil {
		draft.Description = strings.TrimSpace(*req.Description)
	}
	if req.PlannerAgentID != nil {
		if _, err := s.ensureAgentExists(ctx, *req.PlannerAgentID); err != nil {
			return dto.FlowDraftResponse{}, err
		}
		draft.PlannerAgentID = req.PlannerAgentID
	}
	if req.StructuredPlanJSON != nil {
		var plan dto.DraftPlan
		if err := json.Unmarshal(*req.StructuredPlanJSON, &plan); err != nil {
			return dto.FlowDraftResponse{}, response.Validation("草案结构化计划必须是合法 JSON")
		}
		if err := validateDraftPlan(&plan); err != nil {
			return dto.FlowDraftResponse{}, err
		}
		draft.StructuredPlanJSON = mustServiceJSON(plan)
		if strings.TrimSpace(draft.Title) == "" {
			draft.Title = strings.TrimSpace(plan.Title)
		}
		if strings.TrimSpace(draft.Description) == "" {
			draft.Description = strings.TrimSpace(plan.Description)
		}
	}

	if err := s.flowDraftRepo.Update(ctx, draft); err != nil {
		return dto.FlowDraftResponse{}, err
	}
	return toFlowDraftResponse(*draft), nil
}

func (s *FlowDraftService) Confirm(ctx context.Context, id uint64, req dto.ConfirmFlowDraftRequest) (dto.ConfirmFlowDraftResponse, error) {
	if id == 0 {
		return dto.ConfirmFlowDraftResponse{}, response.Validation("草案 ID 不合法")
	}
	if req.ConfirmedBy == 0 {
		return dto.ConfirmFlowDraftResponse{}, response.Validation("确认人不能为空")
	}
	if err := s.ensurePersonExists(ctx, req.ConfirmedBy); err != nil {
		return dto.ConfirmFlowDraftResponse{}, err
	}

	var createdTemplateID uint64
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		draft, err := s.flowDraftRepo.GetByIDWithLock(ctx, tx, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return response.NotFound("草案不存在")
			}
			return err
		}
		if draft.Status != domain.DraftStatusDraft {
			return response.InvalidState("当前草案状态不能确认")
		}

		var plan dto.DraftPlan
		if err := json.Unmarshal(draft.StructuredPlanJSON, &plan); err != nil {
			return response.Validation("草案结构化计划格式不合法")
		}
		if err := validateDraftPlan(&plan); err != nil {
			return err
		}

		template := &model.FlowTemplate{
			Name:        chooseDraftField(strings.TrimSpace(plan.Title), draft.Title),
			Code:        generateTemplateCode(chooseDraftField(strings.TrimSpace(plan.Title), draft.Title)),
			Version:     1,
			Category:    "ai_generated",
			Description: chooseDraftField(strings.TrimSpace(plan.Description), draft.Description),
			Status:      domain.TemplateStatusPublished,
			CreatedBy:   req.ConfirmedBy,
		}
		if err := s.templateRepo.CreateWithDB(ctx, tx, template); err != nil {
			return err
		}

		nodes := make([]model.FlowTemplateNode, 0, len(plan.Nodes))
		for _, node := range plan.Nodes {
			defaultAgentID, err := s.resolveExecutorAgentID(ctx, node)
			if err != nil {
				return err
			}
			if err := s.validateSpecifiedPeople(ctx, node); err != nil {
				return err
			}

			nodes = append(nodes, model.FlowTemplateNode{
				TemplateID:           template.ID,
				NodeCode:             strings.TrimSpace(node.NodeCode),
				NodeName:             strings.TrimSpace(node.NodeName),
				NodeType:             strings.TrimSpace(node.NodeType),
				SortOrder:            node.SortOrder,
				DefaultOwnerRule:     strings.TrimSpace(node.OwnerRule),
				DefaultOwnerPersonID: node.OwnerPersonID,
				DefaultAgentID:       defaultAgentID,
				ResultOwnerRule:      strings.TrimSpace(node.ResultOwnerRule),
				ResultOwnerPersonID:  node.ResultOwnerPersonID,
				InputSchemaJSON:      normalizeOrEmptyJSON(node.InputSchema),
				OutputSchemaJSON:     normalizeOrEmptyJSON(node.OutputSchema),
				ConfigJSON:           buildDraftNodeConfig(node),
			})
		}
		if err := s.templateRepo.CreateNodesWithDB(ctx, tx, nodes); err != nil {
			return err
		}

		now := time.Now()
		draft.Status = domain.DraftStatusConfirmed
		draft.ConfirmedTemplateID = &template.ID
		draft.ConfirmedAt = &now
		if err := s.flowDraftRepo.UpdateWithDB(ctx, tx, draft); err != nil {
			return err
		}

		createdTemplateID = template.ID
		return nil
	})
	if err != nil {
		return dto.ConfirmFlowDraftResponse{}, err
	}

	return dto.ConfirmFlowDraftResponse{
		DraftID:    id,
		TemplateID: createdTemplateID,
		Message:    "草案已确认，模板已创建",
	}, nil
}

func (s *FlowDraftService) Discard(ctx context.Context, id uint64, req dto.DiscardFlowDraftRequest) (dto.FlowDraftResponse, error) {
	if id == 0 {
		return dto.FlowDraftResponse{}, response.Validation("草案 ID 不合法")
	}
	if req.DiscardedBy == 0 {
		return dto.FlowDraftResponse{}, response.Validation("废弃人不能为空")
	}
	if err := s.ensurePersonExists(ctx, req.DiscardedBy); err != nil {
		return dto.FlowDraftResponse{}, err
	}

	draft, err := s.flowDraftRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.FlowDraftResponse{}, response.NotFound("草案不存在")
		}
		return dto.FlowDraftResponse{}, err
	}
	if draft.Status != domain.DraftStatusDraft {
		return dto.FlowDraftResponse{}, response.InvalidState("当前草案状态不能废弃")
	}
	draft.Status = domain.DraftStatusDiscarded
	if err := s.flowDraftRepo.Update(ctx, draft); err != nil {
		return dto.FlowDraftResponse{}, err
	}
	return toFlowDraftResponse(*draft), nil
}

func (s *FlowDraftService) Delete(ctx context.Context, id uint64) error {
	if id == 0 {
		return response.Validation("草案 ID 不合法")
	}

	draft, err := s.flowDraftRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NotFound("草案不存在")
		}
		return err
	}
	if draft.Status == domain.DraftStatusConfirmed || draft.ConfirmedTemplateID != nil {
		return response.InvalidState("已确认草案不能删除")
	}
	return s.flowDraftRepo.Delete(ctx, id)
}

func (s *FlowDraftService) ensurePersonExists(ctx context.Context, personID uint64) error {
	exists, err := s.personRepo.ExistsByID(ctx, personID)
	if err != nil {
		return err
	}
	if !exists {
		return response.Validation("草案创建人不存在")
	}
	return nil
}

func (s *FlowDraftService) ensureAgentExists(ctx context.Context, agentID uint64) (*model.Agent, error) {
	agent, err := s.agentRepo.GetByID(ctx, agentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.Validation("编排龙虾不存在")
		}
		return nil, err
	}
	if agent.Status != domain.StatusEnabled {
		return nil, response.Validation("编排龙虾未启用")
	}
	return agent, nil
}

func (s *FlowDraftService) resolveExecutorAgentID(ctx context.Context, node dto.DraftNode) (*uint64, error) {
	if strings.TrimSpace(node.ExecutorType) != "agent" {
		return nil, nil
	}
	agentCode := strings.TrimSpace(node.ExecutorAgentCode)
	if agentCode == "" {
		return nil, response.Validation(fmt.Sprintf("节点 %s 缺少 executor_agent_code", node.NodeCode))
	}
	agent, err := s.agentRepo.GetByCode(ctx, agentCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.Validation(fmt.Sprintf("节点 %s 绑定的龙虾 %s 不存在", node.NodeCode, agentCode))
		}
		return nil, err
	}
	if agent.Status != domain.StatusEnabled {
		return nil, response.Validation(fmt.Sprintf("节点 %s 绑定的龙虾未启用", node.NodeCode))
	}
	return &agent.ID, nil
}

func (s *FlowDraftService) validateSpecifiedPeople(ctx context.Context, node dto.DraftNode) error {
	if strings.TrimSpace(node.OwnerRule) == "specified_person" {
		if node.OwnerPersonID == nil || *node.OwnerPersonID == 0 {
			return response.Validation(fmt.Sprintf("节点 %s 缺少 owner_person_id", node.NodeCode))
		}
		if err := s.ensurePersonExists(ctx, *node.OwnerPersonID); err != nil {
			return err
		}
	}
	if strings.TrimSpace(node.ResultOwnerRule) == "specified_person" {
		if node.ResultOwnerPersonID == nil || *node.ResultOwnerPersonID == 0 {
			return response.Validation(fmt.Sprintf("节点 %s 缺少 result_owner_person_id", node.NodeCode))
		}
		if err := s.ensurePersonExists(ctx, *node.ResultOwnerPersonID); err != nil {
			return err
		}
	}
	return nil
}

func isValidDraftStatus(status string) bool {
	switch status {
	case domain.DraftStatusDraft, domain.DraftStatusConfirmed, domain.DraftStatusDiscarded:
		return true
	default:
		return false
	}
}

func validateDraftPlan(plan *dto.DraftPlan) error {
	if plan == nil {
		return response.Validation("草案结构化计划不能为空")
	}
	if strings.TrimSpace(plan.Title) == "" {
		return response.Validation("草案标题不能为空")
	}
	if len(plan.Nodes) < 3 || len(plan.Nodes) > 8 {
		return response.Validation("节点数量必须在 3-8 个之间")
	}

	nodes := append([]dto.DraftNode(nil), plan.Nodes...)
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].SortOrder < nodes[j].SortOrder
	})

	seenCodes := map[string]struct{}{}
	validTaskTypes := map[string]bool{
		domain.AgentTaskTypeQuery:          true,
		domain.AgentTaskTypeBatchOperation: true,
		domain.AgentTaskTypeExport:         true,
	}
	for index, node := range nodes {
		if strings.TrimSpace(node.NodeCode) == "" {
			return response.Validation("节点编码不能为空")
		}
		if _, exists := seenCodes[node.NodeCode]; exists {
			return response.Validation(fmt.Sprintf("节点编码 %s 重复", node.NodeCode))
		}
		seenCodes[node.NodeCode] = struct{}{}
		if node.SortOrder != index+1 {
			return response.Validation("节点排序必须从 1 开始连续递增")
		}
		if strings.TrimSpace(node.NodeName) == "" {
			return response.Validation(fmt.Sprintf("节点 %s 名称不能为空", node.NodeCode))
		}
		if !isSupportedDraftNodeType(node.NodeType) {
			return response.Validation(fmt.Sprintf("节点 %s 的类型不支持", node.NodeCode))
		}

		switch strings.TrimSpace(node.ExecutorType) {
		case "agent":
			if !isAgentNodeType(node.NodeType) {
				return response.Validation(fmt.Sprintf("节点 %s 的执行主体与节点类型不匹配", node.NodeCode))
			}
			if !validTaskTypes[node.TaskType] {
				return response.Validation(fmt.Sprintf("节点 %s 的任务类型 %s 不在允许范围内", node.NodeCode, node.TaskType))
			}
			if strings.TrimSpace(node.ExecutorAgentCode) == "" {
				return response.Validation(fmt.Sprintf("节点 %s 缺少 executor_agent_code", node.NodeCode))
			}
		case "human":
			if isAgentNodeType(node.NodeType) {
				return response.Validation(fmt.Sprintf("节点 %s 的执行主体与节点类型不匹配", node.NodeCode))
			}
		default:
			return response.Validation(fmt.Sprintf("节点 %s 的 executor_type 不合法", node.NodeCode))
		}

		if strings.TrimSpace(node.OwnerRule) == "specified_person" && (node.OwnerPersonID == nil || *node.OwnerPersonID == 0) {
			return response.Validation(fmt.Sprintf("节点 %s 缺少 owner_person_id", node.NodeCode))
		}
		if strings.TrimSpace(node.ResultOwnerRule) == "specified_person" && (node.ResultOwnerPersonID == nil || *node.ResultOwnerPersonID == 0) {
			return response.Validation(fmt.Sprintf("节点 %s 缺少 result_owner_person_id", node.NodeCode))
		}
	}
	return nil
}

func isSupportedDraftNodeType(nodeType string) bool {
	switch strings.TrimSpace(nodeType) {
	case domain.NodeTypeHumanInput, domain.NodeTypeHumanReview, domain.NodeTypeHumanAcceptance, domain.NodeTypeAgentExecute, domain.NodeTypeAgentExport:
		return true
	default:
		return false
	}
}

func buildDraftNodeConfig(node dto.DraftNode) datatypes.JSON {
	return mustServiceJSON(map[string]any{
		"task_type":            strings.TrimSpace(node.TaskType),
		"completion_condition": strings.TrimSpace(node.CompletionCondition),
		"failure_condition":    strings.TrimSpace(node.FailureCondition),
		"escalation_rule":      strings.TrimSpace(node.EscalationRule),
		"executor_type":        strings.TrimSpace(node.ExecutorType),
		"executor_agent_code":  strings.TrimSpace(node.ExecutorAgentCode),
		"input_schema":         rawJSONToAny(node.InputSchema),
		"output_schema":        rawJSONToAny(node.OutputSchema),
		"required_fields":      []string{},
		"final_deliverable":    "",
	})
}

func rawJSONToAny(raw json.RawMessage) any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return map[string]any{}
	}
	return value
}

func normalizeOrEmptyJSON(raw json.RawMessage) datatypes.JSON {
	if len(raw) == 0 {
		return datatypes.JSON([]byte("{}"))
	}
	if !json.Valid(raw) {
		return datatypes.JSON([]byte("{}"))
	}
	return datatypes.JSON(raw)
}

func chooseDraftField(primary string, fallback string) string {
	if strings.TrimSpace(primary) != "" {
		return strings.TrimSpace(primary)
	}
	return strings.TrimSpace(fallback)
}

func generateTemplateCode(title string) string {
	base := strings.ToLower(strings.TrimSpace(title))
	base = nonCodeCharPattern.ReplaceAllString(base, "-")
	base = strings.Trim(base, "-")
	if base == "" {
		base = "ai-generated"
	}
	return fmt.Sprintf("%s-%d", base, time.Now().UnixNano())
}

func toFlowDraftResponse(draft model.FlowDraft) dto.FlowDraftResponse {
	return dto.FlowDraftResponse{
		ID:                  draft.ID,
		Title:               draft.Title,
		Description:         draft.Description,
		SourcePrompt:        draft.SourcePrompt,
		CreatorPersonID:     draft.CreatorPersonID,
		PlannerAgentID:      draft.PlannerAgentID,
		Status:              draft.Status,
		StructuredPlanJSON:  json.RawMessage(draft.StructuredPlanJSON),
		ConfirmedTemplateID: draft.ConfirmedTemplateID,
		CreatedAt:           draft.CreatedAt,
		UpdatedAt:           draft.UpdatedAt,
		ConfirmedAt:         draft.ConfirmedAt,
	}
}
