package repository

import (
	"context"

	"the-line/backend/internal/model"

	"gorm.io/gorm"
)

type AttachmentRepository struct {
	db *gorm.DB
}

func NewAttachmentRepository(database *gorm.DB) *AttachmentRepository {
	return &AttachmentRepository{db: database}
}

func (r *AttachmentRepository) Create(ctx context.Context, attachment *model.Attachment) error {
	return r.db.WithContext(ctx).Create(attachment).Error
}

func (r *AttachmentRepository) CreateBatchWithDB(ctx context.Context, database *gorm.DB, attachments []model.Attachment) error {
	if len(attachments) == 0 {
		return nil
	}
	return database.WithContext(ctx).Create(&attachments).Error
}

func (r *AttachmentRepository) GetByIDs(ctx context.Context, ids []uint64) ([]model.Attachment, error) {
	var attachments []model.Attachment
	if len(ids) == 0 {
		return attachments, nil
	}
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&attachments).Error
	return attachments, err
}

func (r *AttachmentRepository) CountByTarget(ctx context.Context, targetType string, targetID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Attachment{}).
		Where("target_type = ? AND target_id = ?", targetType, targetID).
		Count(&count).Error
	return count, err
}

func (r *AttachmentRepository) ListByTarget(ctx context.Context, targetType string, targetID uint64) ([]model.Attachment, error) {
	var attachments []model.Attachment
	err := r.db.WithContext(ctx).
		Where("target_type = ? AND target_id = ?", targetType, targetID).
		Order("created_at ASC, id ASC").
		Find(&attachments).Error
	return attachments, err
}
