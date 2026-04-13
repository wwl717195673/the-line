package repository

import (
	"context"

	"the-line/backend/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RunListFilter struct {
	Status            string
	OwnerPersonID     uint64
	InitiatorPersonID uint64
	Scope             string
	ActorPersonID     uint64
	Offset            int
	Limit             int
}

type RunRepository struct {
	db *gorm.DB
}

func NewRunRepository(database *gorm.DB) *RunRepository {
	return &RunRepository{db: database}
}

func (r *RunRepository) List(ctx context.Context, filter RunListFilter) ([]model.FlowRun, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.FlowRun{})
	if filter.Status != "" {
		query = query.Where("current_status = ?", filter.Status)
	}
	if filter.InitiatorPersonID > 0 {
		query = query.Where("initiator_person_id = ?", filter.InitiatorPersonID)
	}
	if filter.Scope == "initiated_by_me" && filter.ActorPersonID > 0 {
		query = query.Where("initiator_person_id = ?", filter.ActorPersonID)
	}
	if filter.OwnerPersonID > 0 || (filter.Scope == "todo" && filter.ActorPersonID > 0) {
		ownerID := filter.OwnerPersonID
		if filter.Scope == "todo" && filter.ActorPersonID > 0 {
			ownerID = filter.ActorPersonID
		}
		query = query.Where("EXISTS (?)",
			r.db.Model(&model.FlowRunNode{}).
				Select("1").
				Where("flow_run_nodes.run_id = flow_runs.id").
				Where("flow_run_nodes.node_code = flow_runs.current_node_code").
				Where("flow_run_nodes.owner_person_id = ?", ownerID),
		)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var runs []model.FlowRun
	if err := query.Order("updated_at DESC").Offset(filter.Offset).Limit(filter.Limit).Find(&runs).Error; err != nil {
		return nil, 0, err
	}
	return runs, total, nil
}

func (r *RunRepository) CreateWithDB(ctx context.Context, database *gorm.DB, run *model.FlowRun) error {
	return database.WithContext(ctx).Create(run).Error
}

func (r *RunRepository) GetByID(ctx context.Context, id uint64) (*model.FlowRun, error) {
	var run model.FlowRun
	if err := r.db.WithContext(ctx).First(&run, id).Error; err != nil {
		return nil, err
	}
	return &run, nil
}

func (r *RunRepository) GetByIDs(ctx context.Context, ids []uint64) ([]model.FlowRun, error) {
	var runs []model.FlowRun
	if len(ids) == 0 {
		return runs, nil
	}
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&runs).Error
	return runs, err
}

func (r *RunRepository) GetByIDWithLock(ctx context.Context, database *gorm.DB, id uint64) (*model.FlowRun, error) {
	var run model.FlowRun
	err := database.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&run, id).Error
	if err != nil {
		return nil, err
	}
	return &run, nil
}

func (r *RunRepository) UpdateWithDB(ctx context.Context, database *gorm.DB, id uint64, updates map[string]any) error {
	return database.WithContext(ctx).Model(&model.FlowRun{}).Where("id = ?", id).Updates(updates).Error
}

func (r *RunRepository) CountByTemplateID(ctx context.Context, templateID uint64) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&model.FlowRun{}).Where("template_id = ?", templateID).Count(&total).Error
	return total, err
}
