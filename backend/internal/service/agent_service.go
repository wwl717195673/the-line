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

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AgentService struct {
	agentRepo  *repository.AgentRepository
	personRepo *repository.PersonRepository
}

func NewAgentService(agentRepo *repository.AgentRepository, personRepo *repository.PersonRepository) *AgentService {
	return &AgentService{agentRepo: agentRepo, personRepo: personRepo}
}

func (s *AgentService) List(ctx context.Context, req dto.AgentListRequest) ([]dto.AgentResponse, int64, dto.PageQuery, error) {
	page := req.PageQuery.Normalize()
	if req.Status != nil && !domain.IsValidStatus(*req.Status) {
		return nil, 0, page, response.Validation("龙虾状态只能是 0 或 1")
	}

	agents, total, err := s.agentRepo.List(ctx, repository.AgentListFilter{
		Status:  req.Status,
		Keyword: strings.TrimSpace(req.Keyword),
		Offset:  page.Offset(),
		Limit:   page.PageSize,
	})
	if err != nil {
		return nil, 0, page, err
	}

	items := make([]dto.AgentResponse, 0, len(agents))
	for _, agent := range agents {
		items = append(items, toAgentResponse(agent))
	}
	return items, total, page, nil
}

func (s *AgentService) Create(ctx context.Context, req dto.CreateAgentRequest) (dto.AgentResponse, error) {
	name := strings.TrimSpace(req.Name)
	code := strings.TrimSpace(req.Code)
	provider := strings.TrimSpace(req.Provider)
	version := strings.TrimSpace(req.Version)

	if name == "" {
		return dto.AgentResponse{}, response.Validation("龙虾名称不能为空")
	}
	if code == "" {
		return dto.AgentResponse{}, response.Validation("龙虾编码不能为空")
	}
	if provider == "" {
		return dto.AgentResponse{}, response.Validation("龙虾 provider 不能为空")
	}
	if version == "" {
		return dto.AgentResponse{}, response.Validation("龙虾版本不能为空")
	}
	if req.OwnerPersonID == 0 {
		return dto.AgentResponse{}, response.Validation("龙虾负责人不能为空")
	}
	if err := s.ensurePersonExists(ctx, req.OwnerPersonID); err != nil {
		return dto.AgentResponse{}, err
	}
	if err := s.ensureCodeAvailable(ctx, code, 0); err != nil {
		return dto.AgentResponse{}, err
	}

	configJSON, err := normalizeJSON(req.ConfigJSON)
	if err != nil {
		return dto.AgentResponse{}, response.Validation("龙虾配置必须是合法 JSON")
	}

	agent := &model.Agent{
		Name:          name,
		Code:          code,
		Provider:      provider,
		Version:       version,
		OwnerPersonID: req.OwnerPersonID,
		ConfigJSON:    configJSON,
		Status:        domain.StatusEnabled,
	}
	if err := s.agentRepo.Create(ctx, agent); err != nil {
		return dto.AgentResponse{}, err
	}
	return toAgentResponse(*agent), nil
}

func (s *AgentService) Update(ctx context.Context, id uint64, req dto.UpdateAgentRequest) (dto.AgentResponse, error) {
	if id == 0 {
		return dto.AgentResponse{}, response.Validation("龙虾 ID 不合法")
	}
	if _, err := s.agentRepo.GetByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.AgentResponse{}, response.NotFound("龙虾不存在")
		}
		return dto.AgentResponse{}, err
	}

	updates := map[string]any{}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return dto.AgentResponse{}, response.Validation("龙虾名称不能为空")
		}
		updates["name"] = name
	}
	if req.Code != nil {
		code := strings.TrimSpace(*req.Code)
		if code == "" {
			return dto.AgentResponse{}, response.Validation("龙虾编码不能为空")
		}
		if err := s.ensureCodeAvailable(ctx, code, id); err != nil {
			return dto.AgentResponse{}, err
		}
		updates["code"] = code
	}
	if req.Provider != nil {
		provider := strings.TrimSpace(*req.Provider)
		if provider == "" {
			return dto.AgentResponse{}, response.Validation("龙虾 provider 不能为空")
		}
		updates["provider"] = provider
	}
	if req.Version != nil {
		version := strings.TrimSpace(*req.Version)
		if version == "" {
			return dto.AgentResponse{}, response.Validation("龙虾版本不能为空")
		}
		updates["version"] = version
	}
	if req.OwnerPersonID != nil {
		if *req.OwnerPersonID == 0 {
			return dto.AgentResponse{}, response.Validation("龙虾负责人不能为空")
		}
		if err := s.ensurePersonExists(ctx, *req.OwnerPersonID); err != nil {
			return dto.AgentResponse{}, err
		}
		updates["owner_person_id"] = *req.OwnerPersonID
	}
	if req.ConfigJSON != nil {
		configJSON, err := normalizeJSON(*req.ConfigJSON)
		if err != nil {
			return dto.AgentResponse{}, response.Validation("龙虾配置必须是合法 JSON")
		}
		updates["config_json"] = configJSON
	}
	if req.Status != nil {
		if !domain.IsValidStatus(*req.Status) {
			return dto.AgentResponse{}, response.Validation("龙虾状态只能是 0 或 1")
		}
		updates["status"] = *req.Status
	}
	if len(updates) == 0 {
		agent, err := s.agentRepo.GetByID(ctx, id)
		if err != nil {
			return dto.AgentResponse{}, err
		}
		return toAgentResponse(*agent), nil
	}

	agent, err := s.agentRepo.Update(ctx, id, updates)
	if err != nil {
		return dto.AgentResponse{}, err
	}
	return toAgentResponse(*agent), nil
}

func (s *AgentService) Disable(ctx context.Context, id uint64) (dto.AgentResponse, error) {
	return s.Update(ctx, id, dto.UpdateAgentRequest{Status: ptr(domain.StatusDisabled)})
}

func (s *AgentService) ensurePersonExists(ctx context.Context, personID uint64) error {
	exists, err := s.personRepo.ExistsByID(ctx, personID)
	if err != nil {
		return err
	}
	if !exists {
		return response.Validation("龙虾负责人不存在")
	}
	return nil
}

func (s *AgentService) ensureCodeAvailable(ctx context.Context, code string, excludeID uint64) error {
	exists, err := s.agentRepo.ExistsByCode(ctx, code, excludeID)
	if err != nil {
		return err
	}
	if exists {
		return response.Conflict("龙虾编码已存在")
	}
	return nil
}

func normalizeJSON(raw json.RawMessage) (datatypes.JSON, error) {
	if len(raw) == 0 {
		return datatypes.JSON([]byte("{}")), nil
	}
	if !json.Valid(raw) {
		return nil, errors.New("invalid json")
	}
	return datatypes.JSON(raw), nil
}

func toAgentResponse(agent model.Agent) dto.AgentResponse {
	return dto.AgentResponse{
		ID:            agent.ID,
		Name:          agent.Name,
		Code:          agent.Code,
		Provider:      agent.Provider,
		Version:       agent.Version,
		OwnerPersonID: agent.OwnerPersonID,
		ConfigJSON:    json.RawMessage(agent.ConfigJSON),
		Status:        agent.Status,
		CreatedAt:     agent.CreatedAt,
		UpdatedAt:     agent.UpdatedAt,
	}
}
