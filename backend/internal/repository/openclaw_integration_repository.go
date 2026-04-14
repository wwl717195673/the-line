package repository

import (
	"context"

	"the-line/backend/internal/model"

	"gorm.io/gorm"
)

type IntegrationListFilter struct {
	Status string
	Offset int
	Limit  int
}

type OpenClawIntegrationRepository struct {
	db *gorm.DB
}

func NewOpenClawIntegrationRepository(database *gorm.DB) *OpenClawIntegrationRepository {
	return &OpenClawIntegrationRepository{db: database}
}

func (r *OpenClawIntegrationRepository) Create(ctx context.Context, integration *model.OpenClawIntegration) error {
	return r.db.WithContext(ctx).Create(integration).Error
}

func (r *OpenClawIntegrationRepository) GetByID(ctx context.Context, id uint64) (*model.OpenClawIntegration, error) {
	var item model.OpenClawIntegration
	if err := r.db.WithContext(ctx).First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *OpenClawIntegrationRepository) GetByFingerprint(ctx context.Context, fingerprint string) (*model.OpenClawIntegration, error) {
	var item model.OpenClawIntegration
	if err := r.db.WithContext(ctx).Where("instance_fingerprint = ?", fingerprint).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *OpenClawIntegrationRepository) GetActiveByAgentID(ctx context.Context, agentID uint64) (*model.OpenClawIntegration, error) {
	var item model.OpenClawIntegration
	if err := r.db.WithContext(ctx).Where("bound_agent_id = ? AND status = ?", agentID, "active").First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *OpenClawIntegrationRepository) List(ctx context.Context, filter IntegrationListFilter) ([]model.OpenClawIntegration, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.OpenClawIntegration{})
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []model.OpenClawIntegration
	if err := query.Order("created_at DESC").Offset(filter.Offset).Limit(filter.Limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *OpenClawIntegrationRepository) Update(ctx context.Context, integration *model.OpenClawIntegration) error {
	return r.db.WithContext(ctx).Save(integration).Error
}

func (r *OpenClawIntegrationRepository) UpdateFields(ctx context.Context, id uint64, updates map[string]any) error {
	return r.db.WithContext(ctx).Model(&model.OpenClawIntegration{}).Where("id = ?", id).Updates(updates).Error
}
