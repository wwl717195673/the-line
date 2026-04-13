package repository

import (
	"context"

	"the-line/backend/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RunNodeRepository struct {
	db *gorm.DB
}

func NewRunNodeRepository(database *gorm.DB) *RunNodeRepository {
	return &RunNodeRepository{db: database}
}

func (r *RunNodeRepository) CreateBatchWithDB(ctx context.Context, database *gorm.DB, nodes []model.FlowRunNode) error {
	if len(nodes) == 0 {
		return nil
	}
	return database.WithContext(ctx).Create(&nodes).Error
}

func (r *RunNodeRepository) ListByRunID(ctx context.Context, runID uint64) ([]model.FlowRunNode, error) {
	var nodes []model.FlowRunNode
	err := r.db.WithContext(ctx).
		Where("run_id = ?", runID).
		Order("sort_order ASC").
		Find(&nodes).Error
	return nodes, err
}

func (r *RunNodeRepository) GetByID(ctx context.Context, id uint64) (*model.FlowRunNode, error) {
	var node model.FlowRunNode
	if err := r.db.WithContext(ctx).First(&node, id).Error; err != nil {
		return nil, err
	}
	return &node, nil
}

func (r *RunNodeRepository) GetByIDs(ctx context.Context, ids []uint64) ([]model.FlowRunNode, error) {
	var nodes []model.FlowRunNode
	if len(ids) == 0 {
		return nodes, nil
	}
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&nodes).Error
	return nodes, err
}

func (r *RunNodeRepository) GetByIDWithLock(ctx context.Context, database *gorm.DB, id uint64) (*model.FlowRunNode, error) {
	var node model.FlowRunNode
	err := database.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&node, id).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

func (r *RunNodeRepository) UpdateWithDB(ctx context.Context, database *gorm.DB, id uint64, updates map[string]any) error {
	return database.WithContext(ctx).Model(&model.FlowRunNode{}).Where("id = ?", id).Updates(updates).Error
}

func (r *RunNodeRepository) ListByRunIDs(ctx context.Context, runIDs []uint64) ([]model.FlowRunNode, error) {
	var nodes []model.FlowRunNode
	if len(runIDs) == 0 {
		return nodes, nil
	}
	err := r.db.WithContext(ctx).
		Where("run_id IN ?", runIDs).
		Order("run_id ASC, sort_order ASC").
		Find(&nodes).Error
	return nodes, err
}
