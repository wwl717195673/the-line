package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"the-line/backend/internal/domain"
	"the-line/backend/internal/dto"
	"the-line/backend/internal/model"
	"the-line/backend/internal/repository"
	"the-line/backend/internal/response"

	"gorm.io/gorm"
)

type TemplateService struct {
	db           *gorm.DB
	templateRepo *repository.TemplateRepository
	runRepo      *repository.RunRepository
	agentRepo    *repository.AgentRepository
}

func NewTemplateService(database *gorm.DB, templateRepo *repository.TemplateRepository, runRepo *repository.RunRepository, agentRepo *repository.AgentRepository) *TemplateService {
	return &TemplateService{
		db:           database,
		templateRepo: templateRepo,
		runRepo:      runRepo,
		agentRepo:    agentRepo,
	}
}

func (s *TemplateService) List(ctx context.Context, req dto.TemplateListRequest) ([]dto.TemplateResponse, int64, dto.PageQuery, error) {
	page := req.PageQuery.Normalize()
	templates, total, err := s.templateRepo.List(ctx, repository.TemplateListFilter{
		Status:  domain.TemplateStatusPublished,
		Keyword: strings.TrimSpace(req.Keyword),
		Offset:  page.Offset(),
		Limit:   page.PageSize,
	})
	if err != nil {
		return nil, 0, page, err
	}

	items := make([]dto.TemplateResponse, 0, len(templates))
	for _, template := range templates {
		items = append(items, toTemplateResponse(template))
	}
	return items, total, page, nil
}

func (s *TemplateService) Detail(ctx context.Context, id uint64) (dto.TemplateDetailResponse, error) {
	if id == 0 {
		return dto.TemplateDetailResponse{}, response.Validation("模板 ID 不合法")
	}

	template, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.TemplateDetailResponse{}, response.NotFound("模板不存在")
		}
		return dto.TemplateDetailResponse{}, err
	}
	if template.Status != domain.TemplateStatusPublished {
		return dto.TemplateDetailResponse{}, response.NotFound("模板不存在")
	}

	nodes, err := s.templateRepo.ListNodesByTemplateID(ctx, template.ID)
	if err != nil {
		return dto.TemplateDetailResponse{}, err
	}

	agentMap, err := s.loadDefaultAgentMap(ctx, nodes)
	if err != nil {
		return dto.TemplateDetailResponse{}, err
	}

	detail := dto.TemplateDetailResponse{
		TemplateResponse: toTemplateResponse(*template),
		Nodes:            make([]dto.TemplateNodeResponse, 0, len(nodes)),
	}
	for _, node := range nodes {
		detail.Nodes = append(detail.Nodes, toTemplateNodeResponse(node, agentMap))
	}
	return detail, nil
}

func (s *TemplateService) Delete(ctx context.Context, id uint64) error {
	if id == 0 {
		return response.Validation("模板 ID 不合法")
	}

	template, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NotFound("模板不存在")
		}
		return err
	}

	runCount, err := s.runRepo.CountByTemplateID(ctx, template.ID)
	if err != nil {
		return err
	}
	if runCount > 0 {
		return response.InvalidState("模板已有流程实例引用，不能删除")
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.templateRepo.DeleteNodesByTemplateIDWithDB(ctx, tx, template.ID); err != nil {
			return err
		}
		return s.templateRepo.DeleteWithDB(ctx, tx, template.ID)
	})
}

func (s *TemplateService) loadDefaultAgentMap(ctx context.Context, nodes []model.FlowTemplateNode) (map[uint64]model.Agent, error) {
	seen := map[uint64]struct{}{}
	ids := make([]uint64, 0)
	for _, node := range nodes {
		if node.DefaultAgentID == nil {
			continue
		}
		if _, ok := seen[*node.DefaultAgentID]; ok {
			continue
		}
		seen[*node.DefaultAgentID] = struct{}{}
		ids = append(ids, *node.DefaultAgentID)
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

func toTemplateResponse(template model.FlowTemplate) dto.TemplateResponse {
	return dto.TemplateResponse{
		ID:          template.ID,
		Name:        template.Name,
		Code:        template.Code,
		Version:     template.Version,
		Category:    template.Category,
		Description: template.Description,
		Status:      template.Status,
		CreatedBy:   template.CreatedBy,
		CreatedAt:   template.CreatedAt,
		UpdatedAt:   template.UpdatedAt,
	}
}

func toTemplateNodeResponse(node model.FlowTemplateNode, agentMap map[uint64]model.Agent) dto.TemplateNodeResponse {
	resp := dto.TemplateNodeResponse{
		ID:                   node.ID,
		TemplateID:           node.TemplateID,
		NodeCode:             node.NodeCode,
		NodeName:             node.NodeName,
		NodeType:             node.NodeType,
		SortOrder:            node.SortOrder,
		DefaultOwnerRule:     node.DefaultOwnerRule,
		DefaultOwnerPersonID: node.DefaultOwnerPersonID,
		DefaultAgentID:       node.DefaultAgentID,
		ResultOwnerRule:      node.ResultOwnerRule,
		ResultOwnerPersonID:  node.ResultOwnerPersonID,
		InputSchemaJSON:      json.RawMessage(node.InputSchemaJSON),
		OutputSchemaJSON:     json.RawMessage(node.OutputSchemaJSON),
		ConfigJSON:           json.RawMessage(node.ConfigJSON),
		CreatedAt:            node.CreatedAt,
		UpdatedAt:            node.UpdatedAt,
	}
	if node.DefaultAgentID != nil {
		if agent, ok := agentMap[*node.DefaultAgentID]; ok {
			resp.DefaultAgent = &dto.AgentBriefResponse{
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
