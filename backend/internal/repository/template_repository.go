package repository

import (
	"context"

	"the-line/backend/internal/model"

	"gorm.io/gorm"
)

type TemplateListFilter struct {
	Status  string
	Keyword string
	Offset  int
	Limit   int
}

type TemplateRepository struct {
	db *gorm.DB
}

func NewTemplateRepository(database *gorm.DB) *TemplateRepository {
	return &TemplateRepository{db: database}
}

func (r *TemplateRepository) List(ctx context.Context, filter TemplateListFilter) ([]model.FlowTemplate, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.FlowTemplate{})
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Keyword != "" {
		like := "%" + filter.Keyword + "%"
		query = query.Where("name LIKE ? OR code LIKE ?", like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var templates []model.FlowTemplate
	if err := query.Order("created_at DESC").Offset(filter.Offset).Limit(filter.Limit).Find(&templates).Error; err != nil {
		return nil, 0, err
	}
	return templates, total, nil
}

func (r *TemplateRepository) GetByID(ctx context.Context, id uint64) (*model.FlowTemplate, error) {
	var template model.FlowTemplate
	if err := r.db.WithContext(ctx).First(&template, id).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *TemplateRepository) GetByIDWithDB(ctx context.Context, database *gorm.DB, id uint64) (*model.FlowTemplate, error) {
	var template model.FlowTemplate
	if err := database.WithContext(ctx).First(&template, id).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *TemplateRepository) CreateWithDB(ctx context.Context, database *gorm.DB, template *model.FlowTemplate) error {
	return database.WithContext(ctx).Create(template).Error
}

func (r *TemplateRepository) CreateNodesWithDB(ctx context.Context, database *gorm.DB, nodes []model.FlowTemplateNode) error {
	if len(nodes) == 0 {
		return nil
	}
	return database.WithContext(ctx).Create(&nodes).Error
}

func (r *TemplateRepository) ListNodesByTemplateID(ctx context.Context, templateID uint64) ([]model.FlowTemplateNode, error) {
	var nodes []model.FlowTemplateNode
	err := r.db.WithContext(ctx).
		Where("template_id = ?", templateID).
		Order("sort_order ASC").
		Find(&nodes).Error
	return nodes, err
}

func (r *TemplateRepository) ListNodesByTemplateIDWithDB(ctx context.Context, database *gorm.DB, templateID uint64) ([]model.FlowTemplateNode, error) {
	var nodes []model.FlowTemplateNode
	err := database.WithContext(ctx).
		Where("template_id = ?", templateID).
		Order("sort_order ASC").
		Find(&nodes).Error
	return nodes, err
}

func (r *TemplateRepository) DeleteNodesByTemplateIDWithDB(ctx context.Context, database *gorm.DB, templateID uint64) error {
	return database.WithContext(ctx).Where("template_id = ?", templateID).Delete(&model.FlowTemplateNode{}).Error
}

func (r *TemplateRepository) DeleteWithDB(ctx context.Context, database *gorm.DB, templateID uint64) error {
	return database.WithContext(ctx).Delete(&model.FlowTemplate{}, templateID).Error
}
