package repository

import (
	"context"

	"the-line/backend/internal/model"

	"gorm.io/gorm"
)

type RegCodeListFilter struct {
	Status string
	Offset int
	Limit  int
}

type RegistrationCodeRepository struct {
	db *gorm.DB
}

func NewRegistrationCodeRepository(database *gorm.DB) *RegistrationCodeRepository {
	return &RegistrationCodeRepository{db: database}
}

func (r *RegistrationCodeRepository) Create(ctx context.Context, code *model.RegistrationCode) error {
	return r.db.WithContext(ctx).Create(code).Error
}

func (r *RegistrationCodeRepository) GetByCode(ctx context.Context, code string) (*model.RegistrationCode, error) {
	var item model.RegistrationCode
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *RegistrationCodeRepository) List(ctx context.Context, filter RegCodeListFilter) ([]model.RegistrationCode, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.RegistrationCode{})
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []model.RegistrationCode
	if err := query.Order("created_at DESC").Offset(filter.Offset).Limit(filter.Limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *RegistrationCodeRepository) Update(ctx context.Context, code *model.RegistrationCode) error {
	return r.db.WithContext(ctx).Save(code).Error
}
