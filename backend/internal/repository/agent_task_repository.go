package repository

import (
	"context"
	"errors"

	"the-line/backend/internal/domain"
	"the-line/backend/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AgentTaskListFilter struct {
	RunID     uint64
	RunNodeID uint64
	Status    string
	Offset    int
	Limit     int
}

type AgentTaskRepository struct {
	db *gorm.DB
}

func NewAgentTaskRepository(database *gorm.DB) *AgentTaskRepository {
	return &AgentTaskRepository{db: database}
}

func (r *AgentTaskRepository) Create(ctx context.Context, task *model.AgentTask) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *AgentTaskRepository) CreateWithDB(ctx context.Context, database *gorm.DB, task *model.AgentTask) error {
	return database.WithContext(ctx).Create(task).Error
}

func (r *AgentTaskRepository) GetByID(ctx context.Context, id uint64) (*model.AgentTask, error) {
	var task model.AgentTask
	if err := r.db.WithContext(ctx).First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *AgentTaskRepository) GetByIDWithLock(ctx context.Context, database *gorm.DB, id uint64) (*model.AgentTask, error) {
	var task model.AgentTask
	if err := database.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *AgentTaskRepository) GetByRunNodeID(ctx context.Context, runNodeID uint64) ([]model.AgentTask, error) {
	var tasks []model.AgentTask
	if err := r.db.WithContext(ctx).Where("run_node_id = ?", runNodeID).Order("created_at DESC").Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *AgentTaskRepository) GetActiveByRunNodeID(ctx context.Context, runNodeID uint64) (*model.AgentTask, error) {
	return r.GetActiveByRunNodeIDWithDB(ctx, r.db, runNodeID)
}

func (r *AgentTaskRepository) GetActiveByRunNodeIDWithDB(ctx context.Context, database *gorm.DB, runNodeID uint64) (*model.AgentTask, error) {
	var task model.AgentTask
	err := database.WithContext(ctx).
		Where("run_node_id = ? AND status IN ?", runNodeID, []string{
			domain.AgentTaskStatusQueued,
			domain.AgentTaskStatusRunning,
		}).
		Order("created_at DESC").
		First(&task).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &task, nil
}

func (r *AgentTaskRepository) List(ctx context.Context, filter AgentTaskListFilter) ([]model.AgentTask, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.AgentTask{})
	if filter.RunID > 0 {
		query = query.Where("run_id = ?", filter.RunID)
	}
	if filter.RunNodeID > 0 {
		query = query.Where("run_node_id = ?", filter.RunNodeID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []model.AgentTask
	if err := query.Order("created_at DESC").Offset(filter.Offset).Limit(filter.Limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *AgentTaskRepository) Update(ctx context.Context, task *model.AgentTask) error {
	return r.db.WithContext(ctx).Save(task).Error
}

func (r *AgentTaskRepository) UpdateWithDB(ctx context.Context, database *gorm.DB, task *model.AgentTask) error {
	return database.WithContext(ctx).Save(task).Error
}
