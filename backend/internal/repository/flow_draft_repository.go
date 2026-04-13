package repository

import (
	"context"

	"the-line/backend/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FlowDraftListFilter struct {
	CreatorPersonID uint64
	Status          string
	Offset          int
	Limit           int
}

type FlowDraftRepository struct {
	db *gorm.DB
}

func NewFlowDraftRepository(database *gorm.DB) *FlowDraftRepository {
	return &FlowDraftRepository{db: database}
}

func (r *FlowDraftRepository) Create(ctx context.Context, draft *model.FlowDraft) error {
	return r.db.WithContext(ctx).Create(draft).Error
}

func (r *FlowDraftRepository) CreateWithDB(ctx context.Context, database *gorm.DB, draft *model.FlowDraft) error {
	return database.WithContext(ctx).Create(draft).Error
}

func (r *FlowDraftRepository) GetByID(ctx context.Context, id uint64) (*model.FlowDraft, error) {
	var draft model.FlowDraft
	if err := r.db.WithContext(ctx).First(&draft, id).Error; err != nil {
		return nil, err
	}
	return &draft, nil
}

func (r *FlowDraftRepository) GetByIDWithLock(ctx context.Context, database *gorm.DB, id uint64) (*model.FlowDraft, error) {
	var draft model.FlowDraft
	if err := database.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&draft, id).Error; err != nil {
		return nil, err
	}
	return &draft, nil
}

func (r *FlowDraftRepository) List(ctx context.Context, filter FlowDraftListFilter) ([]model.FlowDraft, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.FlowDraft{})
	if filter.CreatorPersonID > 0 {
		query = query.Where("creator_person_id = ?", filter.CreatorPersonID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []model.FlowDraft
	if err := query.Order("created_at DESC").Offset(filter.Offset).Limit(filter.Limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *FlowDraftRepository) Update(ctx context.Context, draft *model.FlowDraft) error {
	return r.db.WithContext(ctx).Save(draft).Error
}

func (r *FlowDraftRepository) UpdateWithDB(ctx context.Context, database *gorm.DB, draft *model.FlowDraft) error {
	return database.WithContext(ctx).Save(draft).Error
}

func (r *FlowDraftRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.FlowDraft{}, id).Error
}
