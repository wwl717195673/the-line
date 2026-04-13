package service

import (
	"context"
	"encoding/json"
	"errors"
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

type DeliverableService struct {
	db              *gorm.DB
	deliverableRepo *repository.DeliverableRepository
	runRepo         *repository.RunRepository
	runNodeRepo     *repository.RunNodeRepository
	personRepo      *repository.PersonRepository
	agentRepo       *repository.AgentRepository
	attachmentRepo  *repository.AttachmentRepository
}

func NewDeliverableService(
	database *gorm.DB,
	deliverableRepo *repository.DeliverableRepository,
	runRepo *repository.RunRepository,
	runNodeRepo *repository.RunNodeRepository,
	personRepo *repository.PersonRepository,
	agentRepo *repository.AgentRepository,
	attachmentRepo *repository.AttachmentRepository,
) *DeliverableService {
	return &DeliverableService{
		db:              database,
		deliverableRepo: deliverableRepo,
		runRepo:         runRepo,
		runNodeRepo:     runNodeRepo,
		personRepo:      personRepo,
		agentRepo:       agentRepo,
		attachmentRepo:  attachmentRepo,
	}
}

func (s *DeliverableService) Create(ctx context.Context, req dto.CreateDeliverableRequest, actor domain.Actor) (dto.DeliverableDetailResponse, error) {
	title := strings.TrimSpace(req.Title)
	summary := strings.TrimSpace(req.Summary)
	if req.RunID == 0 {
		return dto.DeliverableDetailResponse{}, response.Validation("流程 ID 不能为空")
	}
	if title == "" {
		return dto.DeliverableDetailResponse{}, response.Validation("交付标题不能为空")
	}
	if summary == "" {
		return dto.DeliverableDetailResponse{}, response.Validation("交付摘要不能为空")
	}
	if req.ReviewerPersonID == 0 {
		return dto.DeliverableDetailResponse{}, response.Validation("验收人不能为空")
	}
	resultJSON, err := normalizeJSON(req.ResultJSON)
	if err != nil {
		return dto.DeliverableDetailResponse{}, response.Validation("交付结果必须是合法 JSON")
	}

	reviewer, err := s.personRepo.GetByID(ctx, req.ReviewerPersonID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.DeliverableDetailResponse{}, response.Validation("验收人不存在")
		}
		return dto.DeliverableDetailResponse{}, err
	}
	if reviewer.Status != domain.StatusEnabled {
		return dto.DeliverableDetailResponse{}, response.Validation("验收人未启用")
	}

	var deliverableID uint64
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		run, err := s.runRepo.GetByIDWithLock(ctx, tx, req.RunID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return response.NotFound("流程不存在")
			}
			return err
		}
		if run.CurrentStatus != domain.RunStatusCompleted {
			return response.InvalidState("只有已完成流程可以生成交付物")
		}

		nodes, err := s.runNodeRepo.ListByRunID(ctx, run.ID)
		if err != nil {
			return err
		}
		if !canCreateDeliverable(*run, nodes, actor) {
			return response.Forbidden("当前用户不能为该流程生成交付物")
		}

		resultJSON, err := mergeDeliverableResult(resultJSON, nodes, "")
		if err != nil {
			return err
		}
		deliverable := &model.Deliverable{
			RunID:            run.ID,
			Title:            title,
			Summary:          summary,
			ResultJSON:       resultJSON,
			ReviewerPersonID: req.ReviewerPersonID,
			ReviewStatus:     domain.DeliverableReviewStatusPending,
		}
		if err := s.deliverableRepo.CreateWithDB(ctx, tx, deliverable); err != nil {
			return err
		}

		if err := s.copyAttachmentsToDeliverable(ctx, tx, req.AttachmentIDs, deliverable.ID, actor.PersonID); err != nil {
			return err
		}

		deliverableID = deliverable.ID
		return nil
	})
	if err != nil {
		return dto.DeliverableDetailResponse{}, err
	}

	return s.Detail(ctx, deliverableID, actor)
}

func (s *DeliverableService) List(ctx context.Context, req dto.DeliverableListRequest, actor domain.Actor) ([]dto.DeliverableResponse, int64, dto.PageQuery, error) {
	page := req.PageQuery.Normalize()
	reviewStatus := strings.TrimSpace(req.ReviewStatus)
	if reviewStatus != "" && !domain.IsDeliverableReviewStatus(reviewStatus) {
		return nil, 0, page, response.Validation("交付物验收状态不合法")
	}

	deliverables, total, err := s.deliverableRepo.List(ctx, repository.DeliverableListFilter{
		ReviewStatus:     reviewStatus,
		ReviewerPersonID: req.ReviewerPersonID,
		Offset:           page.Offset(),
		Limit:            page.PageSize,
	})
	if err != nil {
		return nil, 0, page, err
	}

	runMap, personMap, err := s.loadDeliverableRelationMaps(ctx, deliverables, nil)
	if err != nil {
		return nil, 0, page, err
	}

	items := make([]dto.DeliverableResponse, 0, len(deliverables))
	for _, deliverable := range deliverables {
		items = append(items, toDeliverableResponse(deliverable, runMap, personMap))
	}
	return items, total, page, nil
}

func (s *DeliverableService) Detail(ctx context.Context, deliverableID uint64, actor domain.Actor) (dto.DeliverableDetailResponse, error) {
	if deliverableID == 0 {
		return dto.DeliverableDetailResponse{}, response.Validation("交付物 ID 不合法")
	}

	deliverable, err := s.deliverableRepo.GetByID(ctx, deliverableID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.DeliverableDetailResponse{}, response.NotFound("交付物不存在")
		}
		return dto.DeliverableDetailResponse{}, err
	}
	nodes, err := s.runNodeRepo.ListByRunID(ctx, deliverable.RunID)
	if err != nil {
		return dto.DeliverableDetailResponse{}, err
	}
	attachments, err := s.attachmentRepo.ListByTarget(ctx, domain.TargetTypeDeliverable, deliverable.ID)
	if err != nil {
		return dto.DeliverableDetailResponse{}, err
	}
	runMap, personMap, err := s.loadDeliverableRelationMaps(ctx, []model.Deliverable{*deliverable}, nodes)
	if err != nil {
		return dto.DeliverableDetailResponse{}, err
	}
	agentMap, err := s.loadNodeAgentMap(ctx, nodes)
	if err != nil {
		return dto.DeliverableDetailResponse{}, err
	}

	detail := dto.DeliverableDetailResponse{
		DeliverableResponse: toDeliverableResponse(*deliverable, runMap, personMap),
		Nodes:               make([]dto.RunNodeResponse, 0, len(nodes)),
		Attachments:         toAttachmentResponses(attachments),
	}
	if run, ok := runMap[deliverable.RunID]; ok {
		for _, node := range nodes {
			detail.Nodes = append(detail.Nodes, toRunNodeResponse(node, run.CurrentNodeCode, personMap, agentMap))
		}
	}
	return detail, nil
}

func (s *DeliverableService) Review(ctx context.Context, deliverableID uint64, req dto.ReviewDeliverableRequest, actor domain.Actor) (dto.DeliverableDetailResponse, error) {
	reviewStatus := strings.TrimSpace(req.ReviewStatus)
	if !domain.IsDeliverableReviewDecision(reviewStatus) {
		return dto.DeliverableDetailResponse{}, response.Validation("验收状态只能是 approved 或 rejected")
	}
	if actor.PersonID == 0 {
		return dto.DeliverableDetailResponse{}, response.Validation("当前用户不能为空")
	}

	deliverable, err := s.deliverableRepo.GetByID(ctx, deliverableID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.DeliverableDetailResponse{}, response.NotFound("交付物不存在")
		}
		return dto.DeliverableDetailResponse{}, err
	}
	if deliverable.ReviewerPersonID != actor.PersonID && !actor.IsAdmin() {
		return dto.DeliverableDetailResponse{}, response.Forbidden("当前用户不能验收该交付物")
	}

	resultJSON, err := mergeReviewComment(deliverable.ResultJSON, strings.TrimSpace(req.ReviewComment))
	if err != nil {
		return dto.DeliverableDetailResponse{}, err
	}
	now := time.Now()
	if _, err := s.deliverableRepo.Update(ctx, deliverable.ID, map[string]any{
		"review_status": reviewStatus,
		"reviewed_at":   &now,
		"result_json":   resultJSON,
	}); err != nil {
		return dto.DeliverableDetailResponse{}, err
	}

	return s.Detail(ctx, deliverable.ID, actor)
}

func (s *DeliverableService) copyAttachmentsToDeliverable(ctx context.Context, tx *gorm.DB, attachmentIDs []uint64, deliverableID uint64, actorPersonID uint64) error {
	if len(attachmentIDs) == 0 {
		return nil
	}
	attachments, err := s.attachmentRepo.GetByIDs(ctx, attachmentIDs)
	if err != nil {
		return err
	}
	if len(attachments) != len(uniqueUint64s(attachmentIDs)) {
		return response.Validation("存在无效的关键附件 ID")
	}

	copies := make([]model.Attachment, 0, len(attachments))
	for _, attachment := range attachments {
		copies = append(copies, model.Attachment{
			TargetType: domain.TargetTypeDeliverable,
			TargetID:   deliverableID,
			FileName:   attachment.FileName,
			FileURL:    attachment.FileURL,
			FileSize:   attachment.FileSize,
			FileType:   attachment.FileType,
			UploadedBy: actorPersonID,
		})
	}
	return s.attachmentRepo.CreateBatchWithDB(ctx, tx, copies)
}

func (s *DeliverableService) loadDeliverableRelationMaps(ctx context.Context, deliverables []model.Deliverable, nodes []model.FlowRunNode) (map[uint64]model.FlowRun, map[uint64]model.Person, error) {
	runIDs := make([]uint64, 0, len(deliverables))
	personIDs := make([]uint64, 0)
	seenRunIDs := map[uint64]struct{}{}
	seenPersonIDs := map[uint64]struct{}{}

	addPersonID := func(id uint64) {
		if id == 0 {
			return
		}
		if _, ok := seenPersonIDs[id]; ok {
			return
		}
		seenPersonIDs[id] = struct{}{}
		personIDs = append(personIDs, id)
	}

	for _, deliverable := range deliverables {
		if _, ok := seenRunIDs[deliverable.RunID]; !ok {
			seenRunIDs[deliverable.RunID] = struct{}{}
			runIDs = append(runIDs, deliverable.RunID)
		}
		addPersonID(deliverable.ReviewerPersonID)
	}

	runs, err := s.listRunsByIDs(ctx, runIDs)
	if err != nil {
		return nil, nil, err
	}
	for _, run := range runs {
		addPersonID(run.InitiatorPersonID)
	}
	for _, node := range nodes {
		if node.OwnerPersonID != nil {
			addPersonID(*node.OwnerPersonID)
		}
		if node.ReviewerPersonID != nil {
			addPersonID(*node.ReviewerPersonID)
		}
	}

	persons, err := s.personRepo.GetByIDs(ctx, personIDs)
	if err != nil {
		return nil, nil, err
	}
	runMap := make(map[uint64]model.FlowRun, len(runs))
	for _, run := range runs {
		runMap[run.ID] = run
	}
	personMap := make(map[uint64]model.Person, len(persons))
	for _, person := range persons {
		personMap[person.ID] = person
	}
	return runMap, personMap, nil
}

func (s *DeliverableService) listRunsByIDs(ctx context.Context, ids []uint64) ([]model.FlowRun, error) {
	var runs []model.FlowRun
	if len(ids) == 0 {
		return runs, nil
	}
	err := s.db.WithContext(ctx).Where("id IN ?", ids).Find(&runs).Error
	return runs, err
}

func (s *DeliverableService) loadNodeAgentMap(ctx context.Context, nodes []model.FlowRunNode) (map[uint64]model.Agent, error) {
	ids := make([]uint64, 0)
	seen := map[uint64]struct{}{}
	for _, node := range nodes {
		if node.BoundAgentID == nil {
			continue
		}
		if _, ok := seen[*node.BoundAgentID]; ok {
			continue
		}
		seen[*node.BoundAgentID] = struct{}{}
		ids = append(ids, *node.BoundAgentID)
	}
	agents, err := s.agentRepo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	agentMap := make(map[uint64]model.Agent, len(agents))
	for _, agent := range agents {
		agentMap[agent.ID] = agent
	}
	return agentMap, nil
}

func canCreateDeliverable(run model.FlowRun, nodes []model.FlowRunNode, actor domain.Actor) bool {
	if actor.IsAdmin() {
		return true
	}
	if actor.PersonID == 0 {
		return false
	}
	if run.InitiatorPersonID == actor.PersonID {
		return true
	}
	if len(nodes) == 0 {
		return false
	}
	lastNode := nodes[len(nodes)-1]
	return lastNode.OwnerPersonID != nil && *lastNode.OwnerPersonID == actor.PersonID
}

func mergeDeliverableResult(resultJSON datatypes.JSON, nodes []model.FlowRunNode, reviewComment string) (datatypes.JSON, error) {
	result := map[string]any{}
	if len(resultJSON) > 0 {
		if err := json.Unmarshal(resultJSON, &result); err != nil {
			return nil, response.Validation("交付结果必须是合法 JSON")
		}
	}
	result["node_summary"] = summarizeRunNodes(nodes)
	if reviewComment != "" {
		result["review_comment"] = reviewComment
	}
	bytes, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(bytes), nil
}

func mergeReviewComment(resultJSON datatypes.JSON, reviewComment string) (datatypes.JSON, error) {
	result := map[string]any{}
	if len(resultJSON) > 0 {
		if err := json.Unmarshal(resultJSON, &result); err != nil {
			return nil, response.Validation("交付结果不是合法 JSON")
		}
	}
	result["review_comment"] = reviewComment
	bytes, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(bytes), nil
}

func summarizeRunNodes(nodes []model.FlowRunNode) []map[string]any {
	summary := make([]map[string]any, 0, len(nodes))
	for _, node := range nodes {
		summary = append(summary, map[string]any{
			"id":           node.ID,
			"node_code":    node.NodeCode,
			"node_name":    node.NodeName,
			"node_type":    node.NodeType,
			"sort_order":   node.SortOrder,
			"status":       node.Status,
			"completed_at": node.CompletedAt,
		})
	}
	return summary
}

func toDeliverableResponse(deliverable model.Deliverable, runMap map[uint64]model.FlowRun, personMap map[uint64]model.Person) dto.DeliverableResponse {
	resp := dto.DeliverableResponse{
		ID:               deliverable.ID,
		RunID:            deliverable.RunID,
		Title:            deliverable.Title,
		Summary:          deliverable.Summary,
		ResultJSON:       json.RawMessage(deliverable.ResultJSON),
		ReviewerPersonID: deliverable.ReviewerPersonID,
		ReviewStatus:     deliverable.ReviewStatus,
		ReviewedAt:       deliverable.ReviewedAt,
		CreatedAt:        deliverable.CreatedAt,
		UpdatedAt:        deliverable.UpdatedAt,
	}
	if run, ok := runMap[deliverable.RunID]; ok {
		runResp := toRunResponse(run, personMap)
		resp.Run = &runResp
	}
	if reviewer, ok := personMap[deliverable.ReviewerPersonID]; ok {
		resp.Reviewer = toPersonBriefResponsePtr(reviewer)
	}
	return resp
}

func uniqueUint64s(ids []uint64) []uint64 {
	seen := map[uint64]struct{}{}
	unique := make([]uint64, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	return unique
}
