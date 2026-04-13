package repository

import (
	"context"

	"the-line/backend/internal/model"

	"gorm.io/gorm"
)

type NodeLogRepository struct {
	db *gorm.DB
}

func NewNodeLogRepository(database *gorm.DB) *NodeLogRepository {
	return &NodeLogRepository{db: database}
}

func (r *NodeLogRepository) CreateWithDB(ctx context.Context, database *gorm.DB, log *model.FlowRunNodeLog) error {
	return database.WithContext(ctx).Create(log).Error
}

func (r *NodeLogRepository) ListByRunID(ctx context.Context, runID uint64) ([]model.FlowRunNodeLog, error) {
	var logs []model.FlowRunNodeLog
	err := r.db.WithContext(ctx).
		Where("run_id = ?", runID).
		Order("created_at ASC, id ASC").
		Find(&logs).Error
	return logs, err
}

func (r *NodeLogRepository) ListByRunNodeID(ctx context.Context, runNodeID uint64) ([]model.FlowRunNodeLog, error) {
	var logs []model.FlowRunNodeLog
	err := r.db.WithContext(ctx).
		Where("run_node_id = ?", runNodeID).
		Order("created_at ASC, id ASC").
		Find(&logs).Error
	return logs, err
}

func (r *NodeLogRepository) ListRecent(ctx context.Context, limit int) ([]model.FlowRunNodeLog, error) {
	var logs []model.FlowRunNodeLog
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	err := r.db.WithContext(ctx).
		Order("created_at DESC, id DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}
