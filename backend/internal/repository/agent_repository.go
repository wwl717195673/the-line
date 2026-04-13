package repository

import (
	"context"

	"the-line/backend/internal/model"

	"gorm.io/gorm"
)

type AgentListFilter struct {
	Status  *int8
	Keyword string
	Offset  int
	Limit   int
}

type AgentRepository struct {
	db *gorm.DB
}

func NewAgentRepository(database *gorm.DB) *AgentRepository {
	return &AgentRepository{db: database}
}

func (r *AgentRepository) List(ctx context.Context, filter AgentListFilter) ([]model.Agent, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.Agent{})
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.Keyword != "" {
		like := "%" + filter.Keyword + "%"
		query = query.Where("name LIKE ? OR code LIKE ?", like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var agents []model.Agent
	if err := query.Order("created_at DESC").Offset(filter.Offset).Limit(filter.Limit).Find(&agents).Error; err != nil {
		return nil, 0, err
	}
	return agents, total, nil
}

func (r *AgentRepository) Create(ctx context.Context, agent *model.Agent) error {
	return r.db.WithContext(ctx).Create(agent).Error
}

func (r *AgentRepository) GetByID(ctx context.Context, id uint64) (*model.Agent, error) {
	var agent model.Agent
	if err := r.db.WithContext(ctx).First(&agent, id).Error; err != nil {
		return nil, err
	}
	return &agent, nil
}

func (r *AgentRepository) GetByCode(ctx context.Context, code string) (*model.Agent, error) {
	var agent model.Agent
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&agent).Error; err != nil {
		return nil, err
	}
	return &agent, nil
}

func (r *AgentRepository) GetByIDs(ctx context.Context, ids []uint64) ([]model.Agent, error) {
	var agents []model.Agent
	if len(ids) == 0 {
		return agents, nil
	}
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&agents).Error
	return agents, err
}

func (r *AgentRepository) ExistsByCode(ctx context.Context, code string, excludeID uint64) (bool, error) {
	query := r.db.WithContext(ctx).Model(&model.Agent{}).Where("code = ?", code)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *AgentRepository) Update(ctx context.Context, id uint64, updates map[string]any) (*model.Agent, error) {
	if err := r.db.WithContext(ctx).Model(&model.Agent{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}
