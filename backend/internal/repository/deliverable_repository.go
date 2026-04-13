package repository

import (
	"context"

	"the-line/backend/internal/model"

	"gorm.io/gorm"
)

type DeliverableListFilter struct {
	ReviewStatus     string
	ReviewerPersonID uint64
	Offset           int
	Limit            int
}

type DeliverableRepository struct {
	db *gorm.DB
}

func NewDeliverableRepository(database *gorm.DB) *DeliverableRepository {
	return &DeliverableRepository{db: database}
}

func (r *DeliverableRepository) List(ctx context.Context, filter DeliverableListFilter) ([]model.Deliverable, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.Deliverable{})
	if filter.ReviewStatus != "" {
		query = query.Where("review_status = ?", filter.ReviewStatus)
	}
	if filter.ReviewerPersonID > 0 {
		query = query.Where("reviewer_person_id = ?", filter.ReviewerPersonID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var deliverables []model.Deliverable
	if err := query.Order("created_at DESC").Offset(filter.Offset).Limit(filter.Limit).Find(&deliverables).Error; err != nil {
		return nil, 0, err
	}
	return deliverables, total, nil
}

func (r *DeliverableRepository) CreateWithDB(ctx context.Context, database *gorm.DB, deliverable *model.Deliverable) error {
	return database.WithContext(ctx).Create(deliverable).Error
}

func (r *DeliverableRepository) GetByID(ctx context.Context, id uint64) (*model.Deliverable, error) {
	var deliverable model.Deliverable
	if err := r.db.WithContext(ctx).First(&deliverable, id).Error; err != nil {
		return nil, err
	}
	return &deliverable, nil
}

func (r *DeliverableRepository) GetLatestByRunID(ctx context.Context, runID uint64) (*model.Deliverable, error) {
	var deliverable model.Deliverable
	err := r.db.WithContext(ctx).
		Where("run_id = ?", runID).
		Order("created_at DESC, id DESC").
		First(&deliverable).Error
	if err != nil {
		return nil, err
	}
	return &deliverable, nil
}

func (r *DeliverableRepository) Update(ctx context.Context, id uint64, updates map[string]any) (*model.Deliverable, error) {
	if err := r.db.WithContext(ctx).Model(&model.Deliverable{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}
