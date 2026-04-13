package repository

import (
	"context"

	"the-line/backend/internal/model"

	"gorm.io/gorm"
)

type CommentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(database *gorm.DB) *CommentRepository {
	return &CommentRepository{db: database}
}

func (r *CommentRepository) ListByTarget(ctx context.Context, targetType string, targetID uint64) ([]model.Comment, error) {
	var comments []model.Comment
	err := r.db.WithContext(ctx).
		Where("target_type = ? AND target_id = ?", targetType, targetID).
		Order("created_at ASC, id ASC").
		Find(&comments).Error
	return comments, err
}

func (r *CommentRepository) Create(ctx context.Context, comment *model.Comment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

func (r *CommentRepository) GetByID(ctx context.Context, id uint64) (*model.Comment, error) {
	var comment model.Comment
	if err := r.db.WithContext(ctx).First(&comment, id).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *CommentRepository) Update(ctx context.Context, id uint64, updates map[string]any) (*model.Comment, error) {
	if err := r.db.WithContext(ctx).Model(&model.Comment{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}
