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
	"the-line/backend/internal/model"
	"the-line/backend/internal/repository"
	"the-line/backend/internal/response"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type RunService struct {
	db              *gorm.DB
	runRepo         *repository.RunRepository
	runNodeRepo     *repository.RunNodeRepository
	nodeLogRepo     *repository.NodeLogRepository
	templateRepo    *repository.TemplateRepository
	personRepo      *repository.PersonRepository
	agentRepo       *repository.AgentRepository
	deliverableRepo *repository.DeliverableRepository
	orchestration   *RunOrchestrationService
}

func NewRunService(
	database *gorm.DB,
	runRepo *repository.RunRepository,
	runNodeRepo *repository.RunNodeRepository,
	nodeLogRepo *repository.NodeLogRepository,
	templateRepo *repository.TemplateRepository,
	personRepo *repository.PersonRepository,
	agentRepo *repository.AgentRepository,
	deliverableRepo *repository.DeliverableRepository,
) *RunService {
	return &RunService{
		db:              database,
		runRepo:         runRepo,
		runNodeRepo:     runNodeRepo,
		nodeLogRepo:     nodeLogRepo,
		templateRepo:    templateRepo,
		personRepo:      personRepo,
		agentRepo:       agentRepo,
		deliverableRepo: deliverableRepo,
	}
}

func (s *RunService) SetOrchestrationService(orchestration *RunOrchestrationService) {
	s.orchestration = orchestration
}

func (s *RunService) CreateRun(ctx context.Context, req dto.CreateRunRequest, actor domain.Actor) (dto.RunDetailResponse, error) {
	title := strings.TrimSpace(req.Title)
	if req.TemplateID == 0 {
		return dto.RunDetailResponse{}, response.Validation("模板 ID 不能为空")
	}
	if title == "" {
		return dto.RunDetailResponse{}, response.Validation("流程标题不能为空")
	}

	initiatorID := req.InitiatorPersonID
	if initiatorID == 0 {
		initiatorID = actor.PersonID
	}
	if initiatorID == 0 {
		return dto.RunDetailResponse{}, response.Validation("发起人不能为空")
	}
	if err := s.ensurePersonExists(ctx, initiatorID); err != nil {
		return dto.RunDetailResponse{}, err
	}

	inputPayload, err := normalizeJSON(req.InputPayloadJSON)
	if err != nil {
		return dto.RunDetailResponse{}, response.Validation("流程输入必须是合法 JSON")
	}

	var createdRunID uint64
	var firstNodeID uint64
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		template, err := s.templateRepo.GetByIDWithDB(ctx, tx, req.TemplateID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return response.NotFound("模板不存在")
			}
			return err
		}
		if template.Status != domain.TemplateStatusPublished {
			return response.NotFound("模板不存在")
		}

		templateNodes, err := s.templateRepo.ListNodesByTemplateIDWithDB(ctx, tx, template.ID)
		if err != nil {
			return err
		}
		if len(templateNodes) == 0 {
			return response.Validation("模板节点不能为空")
		}
		if err := validateRequiredFields(inputPayload, templateNodes[0].ConfigJSON); err != nil {
			return err
		}

		now := time.Now()
		run := &model.FlowRun{
			TemplateID:        template.ID,
			TemplateVersion:   template.Version,
			Title:             title,
			BizKey:            strings.TrimSpace(req.BizKey),
			InitiatorPersonID: initiatorID,
			CurrentStatus:     domain.RunStatusRunning,
			CurrentNodeCode:   templateNodes[0].NodeCode,
			InputPayloadJSON:  inputPayload,
			OutputPayloadJSON: datatypes.JSON([]byte("{}")),
			StartedAt:         &now,
		}
		if err := s.runRepo.CreateWithDB(ctx, tx, run); err != nil {
			return err
		}

		runNodes := make([]model.FlowRunNode, 0, len(templateNodes))
		var currentOwnerID *uint64
		for index, templateNode := range templateNodes {
			status := domain.NodeStatusNotStarted
			var startedAt *time.Time
			inputJSON := datatypes.JSON([]byte("{}"))
			if index == 0 {
				status = domain.NodeStatusReady
				startedAt = &now
				inputJSON = inputPayload
			}
			ownerPersonID, err := s.resolveOwnerPersonIDWithDB(
				ctx,
				tx,
				templateNode.DefaultOwnerRule,
				templateNode.DefaultOwnerPersonID,
				initiatorID,
				currentOwnerID,
			)
			if err != nil {
				return err
			}
			reviewerPersonID := resolveReviewerPersonID(isReviewNodeType(templateNode.NodeType), ownerPersonID)
			resultOwnerPersonID, err := s.resolveResultOwnerPersonIDWithDB(
				ctx,
				tx,
				templateNode.ResultOwnerRule,
				templateNode.ResultOwnerPersonID,
				initiatorID,
				ownerPersonID,
				currentOwnerID,
			)
			if err != nil {
				return err
			}

			runNodes = append(runNodes, model.FlowRunNode{
				RunID:               run.ID,
				TemplateNodeID:      templateNode.ID,
				NodeCode:            templateNode.NodeCode,
				NodeName:            templateNode.NodeName,
				NodeType:            templateNode.NodeType,
				SortOrder:           templateNode.SortOrder,
				OwnerPersonID:       ownerPersonID,
				ReviewerPersonID:    reviewerPersonID,
				ResultOwnerPersonID: resultOwnerPersonID,
				BoundAgentID:        templateNode.DefaultAgentID,
				Status:              status,
				InputJSON:           inputJSON,
				OutputJSON:          datatypes.JSON([]byte("{}")),
				StartedAt:           startedAt,
			})
			if ownerPersonID != nil {
				currentOwnerID = copyUint64Ptr(ownerPersonID)
			}
		}

		if err := s.runNodeRepo.CreateBatchWithDB(ctx, tx, runNodes); err != nil {
			return err
		}

		if len(runNodes) > 0 {
			firstNodeID = runNodes[0].ID
		}
		if err := s.nodeLogRepo.CreateWithDB(ctx, tx, &model.FlowRunNodeLog{
			RunID:        run.ID,
			RunNodeID:    firstNodeID,
			LogType:      domain.LogTypeRunCreated,
			OperatorType: domain.OperatorTypePerson,
			OperatorID:   initiatorID,
			Content:      "创建流程实例",
			ExtraJSON:    datatypes.JSON([]byte("{}")),
		}); err != nil {
			return err
		}

		createdRunID = run.ID
		return nil
	})
	if err != nil {
		return dto.RunDetailResponse{}, err
	}
	if firstNodeID > 0 && s.orchestration != nil {
		if err := s.orchestration.DispatchIfNeeded(ctx, firstNodeID); err != nil {
			return dto.RunDetailResponse{}, err
		}
	}

	return s.Detail(ctx, createdRunID, actor)
}

func (s *RunService) List(ctx context.Context, req dto.RunListRequest, actor domain.Actor) ([]dto.RunResponse, int64, dto.PageQuery, error) {
	page := req.PageQuery.Normalize()
	scope := strings.TrimSpace(req.Scope)
	if scope == "" {
		scope = "all"
	}
	if scope != "all" && scope != "initiated_by_me" && scope != "todo" {
		return nil, 0, page, response.Validation("流程列表 scope 不合法")
	}
	if (scope == "initiated_by_me" || scope == "todo") && actor.PersonID == 0 {
		return nil, 0, page, response.Validation("当前用户不能为空")
	}

	runs, total, err := s.runRepo.List(ctx, repository.RunListFilter{
		Status:            strings.TrimSpace(req.Status),
		OwnerPersonID:     req.OwnerPersonID,
		InitiatorPersonID: req.InitiatorPersonID,
		Scope:             scope,
		ActorPersonID:     actor.PersonID,
		Offset:            page.Offset(),
		Limit:             page.PageSize,
	})
	if err != nil {
		return nil, 0, page, err
	}

	runIDs := make([]uint64, 0, len(runs))
	for _, run := range runs {
		runIDs = append(runIDs, run.ID)
	}
	nodes, err := s.runNodeRepo.ListByRunIDs(ctx, runIDs)
	if err != nil {
		return nil, 0, page, err
	}

	currentNodeMap := make(map[uint64]model.FlowRunNode, len(runs))
	for _, node := range nodes {
		if _, ok := currentNodeMap[node.RunID]; ok {
			continue
		}
		for _, run := range runs {
			if run.ID == node.RunID && run.CurrentNodeCode == node.NodeCode {
				currentNodeMap[run.ID] = node
				break
			}
		}
	}

	personMap, agentMap, err := s.loadRunRelationMaps(ctx, runs, nodes)
	if err != nil {
		return nil, 0, page, err
	}

	items := make([]dto.RunResponse, 0, len(runs))
	for _, run := range runs {
		var currentNode *dto.RunNodeResponse
		if node, ok := currentNodeMap[run.ID]; ok {
			nodeResp := toRunNodeResponse(node, run.CurrentNodeCode, personMap, agentMap)
			currentNode = &nodeResp
		}
		runResp := toRunResponse(run, personMap)
		runResp.CurrentNode = currentNode
		items = append(items, runResp)
	}
	return items, total, page, nil
}

func (s *RunService) Detail(ctx context.Context, runID uint64, actor domain.Actor) (dto.RunDetailResponse, error) {
	if runID == 0 {
		return dto.RunDetailResponse{}, response.Validation("流程 ID 不合法")
	}

	run, err := s.runRepo.GetByID(ctx, runID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.RunDetailResponse{}, response.NotFound("流程不存在")
		}
		return dto.RunDetailResponse{}, err
	}

	template, err := s.templateRepo.GetByID(ctx, run.TemplateID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.RunDetailResponse{}, err
	}

	nodes, err := s.runNodeRepo.ListByRunID(ctx, run.ID)
	if err != nil {
		return dto.RunDetailResponse{}, err
	}
	logs, err := s.nodeLogRepo.ListByRunID(ctx, run.ID)
	if err != nil {
		return dto.RunDetailResponse{}, err
	}

	personMap, agentMap, err := s.loadRunRelationMaps(ctx, []model.FlowRun{*run}, nodes)
	if err != nil {
		return dto.RunDetailResponse{}, err
	}

	runResp := toRunResponse(*run, personMap)
	nodeResponses := make([]dto.RunNodeResponse, 0, len(nodes))
	for _, node := range nodes {
		nodeResp := toRunNodeResponse(node, run.CurrentNodeCode, personMap, agentMap)
		if nodeResp.IsCurrent {
			current := nodeResp
			runResp.CurrentNode = &current
		}
		nodeResponses = append(nodeResponses, nodeResp)
	}

	detail := dto.RunDetailResponse{
		RunResponse: runResp,
		Nodes:       nodeResponses,
		Logs:        toRunNodeLogResponses(logs),
	}
	if s.deliverableRepo != nil {
		deliverable, err := s.deliverableRepo.GetLatestByRunID(ctx, run.ID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.RunDetailResponse{}, err
		}
		if deliverable != nil {
			detail.HasDeliverable = true
			detail.DeliverableID = &deliverable.ID
		}
	}
	if template != nil {
		templateResp := toTemplateResponse(*template)
		detail.Template = &templateResp
	}
	return detail, nil
}

func (s *RunService) CancelRun(ctx context.Context, runID uint64, reason string, actor domain.Actor) (dto.RunDetailResponse, error) {
	reason = strings.TrimSpace(reason)
	if runID == 0 {
		return dto.RunDetailResponse{}, response.Validation("流程 ID 不合法")
	}
	if reason == "" {
		return dto.RunDetailResponse{}, response.Validation("取消原因不能为空")
	}
	if actor.PersonID == 0 {
		return dto.RunDetailResponse{}, response.Validation("当前用户不能为空")
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		run, err := s.runRepo.GetByIDWithLock(ctx, tx, runID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return response.NotFound("流程不存在")
			}
			return err
		}
		if run.CurrentStatus == domain.RunStatusCompleted {
			return response.Conflict("已完成流程不能取消")
		}
		if run.CurrentStatus == domain.RunStatusCancelled {
			return response.Conflict("流程已取消，不能重复取消")
		}
		if run.InitiatorPersonID != actor.PersonID && !actor.IsAdmin() {
			return response.NewAppError(403, "FORBIDDEN", "只有发起人或管理员可以取消流程")
		}

		now := time.Now()
		if err := s.runRepo.UpdateWithDB(ctx, tx, run.ID, map[string]any{
			"current_status": domain.RunStatusCancelled,
			"completed_at":   &now,
		}); err != nil {
			return err
		}

		var currentNode model.FlowRunNode
		err = tx.WithContext(ctx).
			Where("run_id = ? AND node_code = ?", run.ID, run.CurrentNodeCode).
			First(&currentNode).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		return s.nodeLogRepo.CreateWithDB(ctx, tx, &model.FlowRunNodeLog{
			RunID:        run.ID,
			RunNodeID:    currentNode.ID,
			LogType:      domain.LogTypeRunCancel,
			OperatorType: domain.OperatorTypePerson,
			OperatorID:   actor.PersonID,
			Content:      "取消流程：" + reason,
			ExtraJSON:    mustServiceJSON(map[string]any{"reason": reason}),
		})
	})
	if err != nil {
		return dto.RunDetailResponse{}, err
	}

	return s.Detail(ctx, runID, actor)
}

func (s *RunService) AdvanceAfterNodeDone(tx *gorm.DB, runID uint64, nodeID uint64) (*model.FlowRunNode, error) {
	var doneNode model.FlowRunNode
	if err := tx.First(&doneNode, nodeID).Error; err != nil {
		return nil, err
	}

	var nextNode model.FlowRunNode
	err := tx.Where("run_id = ? AND sort_order > ?", runID, doneNode.SortOrder).
		Order("sort_order ASC").
		First(&nextNode).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		now := time.Now()
		if err := tx.Model(&model.FlowRun{}).Where("id = ?", runID).Updates(map[string]any{
			"current_status":    domain.RunStatusCompleted,
			"current_node_code": "",
			"completed_at":      &now,
		}).Error; err != nil {
			return nil, err
		}
		return nil, nil
	}

	now := time.Now()
	if err := tx.Model(&model.FlowRunNode{}).Where("id = ?", nextNode.ID).Updates(map[string]any{
		"status":     domain.NodeStatusReady,
		"started_at": &now,
	}).Error; err != nil {
		return nil, err
	}

	if err := tx.Model(&model.FlowRun{}).Where("id = ?", runID).Updates(map[string]any{
		"current_status":    domain.RunStatusRunning,
		"current_node_code": nextNode.NodeCode,
	}).Error; err != nil {
		return nil, err
	}
	nextNode.Status = domain.NodeStatusReady
	nextNode.StartedAt = &now
	return &nextNode, nil
}

func (s *RunService) ensurePersonExists(ctx context.Context, personID uint64) error {
	exists, err := s.personRepo.ExistsByID(ctx, personID)
	if err != nil {
		return err
	}
	if !exists {
		return response.Validation("人员不存在")
	}
	return nil
}

func (s *RunService) loadRunRelationMaps(ctx context.Context, runs []model.FlowRun, nodes []model.FlowRunNode) (map[uint64]model.Person, map[uint64]model.Agent, error) {
	personIDs := make([]uint64, 0)
	agentIDs := make([]uint64, 0)
	seenPersons := map[uint64]struct{}{}
	seenAgents := map[uint64]struct{}{}

	addPerson := func(id uint64) {
		if id == 0 {
			return
		}
		if _, ok := seenPersons[id]; ok {
			return
		}
		seenPersons[id] = struct{}{}
		personIDs = append(personIDs, id)
	}
	addAgent := func(id uint64) {
		if id == 0 {
			return
		}
		if _, ok := seenAgents[id]; ok {
			return
		}
		seenAgents[id] = struct{}{}
		agentIDs = append(agentIDs, id)
	}

	for _, run := range runs {
		addPerson(run.InitiatorPersonID)
	}
	for _, node := range nodes {
		if node.OwnerPersonID != nil {
			addPerson(*node.OwnerPersonID)
		}
		if node.ReviewerPersonID != nil {
			addPerson(*node.ReviewerPersonID)
		}
		if node.ResultOwnerPersonID != nil {
			addPerson(*node.ResultOwnerPersonID)
		}
		if node.BoundAgentID != nil {
			addAgent(*node.BoundAgentID)
		}
	}

	persons, err := s.personRepo.GetByIDs(ctx, personIDs)
	if err != nil {
		return nil, nil, err
	}
	agents, err := s.agentRepo.GetByIDs(ctx, agentIDs)
	if err != nil {
		return nil, nil, err
	}

	personMap := make(map[uint64]model.Person, len(persons))
	for _, person := range persons {
		personMap[person.ID] = person
	}
	agentMap := make(map[uint64]model.Agent, len(agents))
	for _, agent := range agents {
		agentMap[agent.ID] = agent
	}
	return personMap, agentMap, nil
}

func (s *RunService) resolveOwnerPersonIDWithDB(
	ctx context.Context,
	tx *gorm.DB,
	rule string,
	specifiedPersonID *uint64,
	initiatorID uint64,
	currentOwnerID *uint64,
) (*uint64, error) {
	switch strings.TrimSpace(rule) {
	case "initiator":
		return &initiatorID, nil
	case "specified_person":
		if specifiedPersonID == nil || *specifiedPersonID == 0 {
			return nil, response.Validation("模板节点执行责任人不能为空")
		}
		if err := s.ensurePersonExists(ctx, *specifiedPersonID); err != nil {
			return nil, err
		}
		return copyUint64Ptr(specifiedPersonID), nil
	case "middle_office", "operation":
		person, err := s.personRepo.GetFirstEnabledByRoleWithDB(ctx, tx, rule)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil
			}
			return nil, err
		}
		return &person.ID, nil
	case "current_owner":
		return copyUint64Ptr(currentOwnerID), nil
	default:
		return nil, nil
	}
}

func resolveReviewerPersonID(needReview bool, ownerPersonID *uint64) *uint64 {
	if !needReview || ownerPersonID == nil {
		return nil
	}
	return copyUint64Ptr(ownerPersonID)
}

func isReviewNodeType(nodeType string) bool {
	switch nodeType {
	case domain.NodeTypeReview, domain.NodeTypeHumanReview, domain.NodeTypeHumanAcceptance:
		return true
	default:
		return false
	}
}

func (s *RunService) resolveResultOwnerPersonIDWithDB(
	ctx context.Context,
	tx *gorm.DB,
	rule string,
	specifiedPersonID *uint64,
	initiatorID uint64,
	ownerPersonID *uint64,
	currentOwnerID *uint64,
) (*uint64, error) {
	switch strings.TrimSpace(rule) {
	case "", "none":
		return copyUint64Ptr(specifiedPersonID), nil
	case "specified_person":
		if specifiedPersonID == nil || *specifiedPersonID == 0 {
			return nil, response.Validation("模板节点结果责任人不能为空")
		}
		if err := s.ensurePersonExists(ctx, *specifiedPersonID); err != nil {
			return nil, err
		}
		return copyUint64Ptr(specifiedPersonID), nil
	case "initiator":
		return &initiatorID, nil
	case "node_owner":
		return copyUint64Ptr(ownerPersonID), nil
	case "current_owner":
		return copyUint64Ptr(currentOwnerID), nil
	case "middle_office", "operation":
		person, err := s.personRepo.GetFirstEnabledByRoleWithDB(ctx, tx, rule)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil
			}
			return nil, err
		}
		return &person.ID, nil
	default:
		return nil, response.Validation("模板节点结果责任人规则不合法")
	}
}

func copyUint64Ptr(value *uint64) *uint64 {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

func validateRequiredFields(input datatypes.JSON, config datatypes.JSON) error {
	var configValue struct {
		RequiredFields []string `json:"required_fields"`
	}
	if len(config) > 0 {
		if err := json.Unmarshal(config, &configValue); err != nil {
			return response.Validation("模板节点配置不是合法 JSON")
		}
	}
	if len(configValue.RequiredFields) == 0 {
		return nil
	}

	var payload map[string]any
	if len(input) > 0 {
		if err := json.Unmarshal(input, &payload); err != nil {
			return response.Validation("流程输入必须是合法 JSON")
		}
	}
	formData := payload
	if nested, ok := payload["form_data"].(map[string]any); ok {
		formData = nested
	}

	for _, field := range configValue.RequiredFields {
		value, ok := formData[field]
		if !ok || isEmptyJSONValue(value) {
			return response.Validation(fmt.Sprintf("发起表单字段 %s 不能为空", field))
		}
	}
	return nil
}

func isEmptyJSONValue(value any) bool {
	if value == nil {
		return true
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed) == ""
	case []any:
		return len(typed) == 0
	case map[string]any:
		return len(typed) == 0
	default:
		return false
	}
}

func toRunResponse(run model.FlowRun, personMap map[uint64]model.Person) dto.RunResponse {
	resp := dto.RunResponse{
		ID:                run.ID,
		TemplateID:        run.TemplateID,
		TemplateVersion:   run.TemplateVersion,
		Title:             run.Title,
		BizKey:            run.BizKey,
		InitiatorPersonID: run.InitiatorPersonID,
		CurrentStatus:     run.CurrentStatus,
		CurrentNodeCode:   run.CurrentNodeCode,
		InputPayloadJSON:  json.RawMessage(run.InputPayloadJSON),
		OutputPayloadJSON: json.RawMessage(run.OutputPayloadJSON),
		StartedAt:         run.StartedAt,
		CompletedAt:       run.CompletedAt,
		CreatedAt:         run.CreatedAt,
		UpdatedAt:         run.UpdatedAt,
	}
	if person, ok := personMap[run.InitiatorPersonID]; ok {
		resp.Initiator = toPersonBriefResponsePtr(person)
	}
	return resp
}

func toRunNodeResponse(node model.FlowRunNode, currentNodeCode string, personMap map[uint64]model.Person, agentMap map[uint64]model.Agent) dto.RunNodeResponse {
	resp := dto.RunNodeResponse{
		ID:                  node.ID,
		RunID:               node.RunID,
		TemplateNodeID:      node.TemplateNodeID,
		NodeCode:            node.NodeCode,
		NodeName:            node.NodeName,
		NodeType:            node.NodeType,
		SortOrder:           node.SortOrder,
		OwnerPersonID:       node.OwnerPersonID,
		ReviewerPersonID:    node.ReviewerPersonID,
		ResultOwnerPersonID: node.ResultOwnerPersonID,
		BoundAgentID:        node.BoundAgentID,
		Status:              node.Status,
		InputJSON:           json.RawMessage(node.InputJSON),
		OutputJSON:          json.RawMessage(node.OutputJSON),
		StartedAt:           node.StartedAt,
		CompletedAt:         node.CompletedAt,
		CreatedAt:           node.CreatedAt,
		UpdatedAt:           node.UpdatedAt,
		IsCurrent:           node.NodeCode == currentNodeCode,
	}
	if node.OwnerPersonID != nil {
		if person, ok := personMap[*node.OwnerPersonID]; ok {
			resp.OwnerPerson = toPersonBriefResponsePtr(person)
		}
	}
	if node.ReviewerPersonID != nil {
		if person, ok := personMap[*node.ReviewerPersonID]; ok {
			resp.ReviewerPerson = toPersonBriefResponsePtr(person)
		}
	}
	if node.ResultOwnerPersonID != nil {
		if person, ok := personMap[*node.ResultOwnerPersonID]; ok {
			resp.ResultOwnerPerson = toPersonBriefResponsePtr(person)
		}
	}
	if node.BoundAgentID != nil {
		if agent, ok := agentMap[*node.BoundAgentID]; ok {
			resp.BoundAgent = &dto.AgentBriefResponse{
				ID:       agent.ID,
				Name:     agent.Name,
				Code:     agent.Code,
				Provider: agent.Provider,
				Version:  agent.Version,
				Status:   agent.Status,
			}
		}
	}
	return resp
}

func toPersonBriefResponsePtr(person model.Person) *dto.PersonBriefResponse {
	return &dto.PersonBriefResponse{
		ID:       person.ID,
		Name:     person.Name,
		Email:    person.Email,
		RoleType: person.RoleType,
		Status:   person.Status,
	}
}

func toRunNodeLogResponses(logs []model.FlowRunNodeLog) []dto.RunNodeLogResponse {
	responses := make([]dto.RunNodeLogResponse, 0, len(logs))
	for _, log := range logs {
		responses = append(responses, dto.RunNodeLogResponse{
			ID:           log.ID,
			RunID:        log.RunID,
			RunNodeID:    log.RunNodeID,
			LogType:      log.LogType,
			OperatorType: log.OperatorType,
			OperatorID:   log.OperatorID,
			Content:      log.Content,
			ExtraJSON:    json.RawMessage(log.ExtraJSON),
			CreatedAt:    log.CreatedAt,
		})
	}
	return responses
}

func mustServiceJSON(value any) datatypes.JSON {
	bytes, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return datatypes.JSON(bytes)
}
