package repository

import (
	"context"

	"the-line/backend/internal/model"

	"gorm.io/gorm"
)

type AgentTaskReceiptRepository struct {
	db *gorm.DB
}

func NewAgentTaskReceiptRepository(database *gorm.DB) *AgentTaskReceiptRepository {
	return &AgentTaskReceiptRepository{db: database}
}

func (r *AgentTaskReceiptRepository) Create(ctx context.Context, receipt *model.AgentTaskReceipt) error {
	return r.db.WithContext(ctx).Create(receipt).Error
}

func (r *AgentTaskReceiptRepository) CreateWithDB(ctx context.Context, database *gorm.DB, receipt *model.AgentTaskReceipt) error {
	return database.WithContext(ctx).Create(receipt).Error
}

func (r *AgentTaskReceiptRepository) GetLatestByTaskID(ctx context.Context, taskID uint64) (*model.AgentTaskReceipt, error) {
	var receipt model.AgentTaskReceipt
	if err := r.db.WithContext(ctx).Where("agent_task_id = ?", taskID).Order("received_at DESC, id DESC").First(&receipt).Error; err != nil {
		return nil, err
	}
	return &receipt, nil
}

func (r *AgentTaskReceiptRepository) ListByRunID(ctx context.Context, runID uint64) ([]model.AgentTaskReceipt, error) {
	var receipts []model.AgentTaskReceipt
	if err := r.db.WithContext(ctx).Where("run_id = ?", runID).Order("received_at DESC, id DESC").Find(&receipts).Error; err != nil {
		return nil, err
	}
	return receipts, nil
}
