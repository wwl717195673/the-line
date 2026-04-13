package repository

import (
	"context"

	"the-line/backend/internal/model"

	"gorm.io/gorm"
)

type PersonListFilter struct {
	Status  *int8
	Keyword string
	Offset  int
	Limit   int
}

type PersonRepository struct {
	db *gorm.DB
}

func NewPersonRepository(database *gorm.DB) *PersonRepository {
	return &PersonRepository{db: database}
}

func (r *PersonRepository) List(ctx context.Context, filter PersonListFilter) ([]model.Person, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.Person{})
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.Keyword != "" {
		like := "%" + filter.Keyword + "%"
		query = query.Where("name LIKE ? OR email LIKE ?", like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var persons []model.Person
	if err := query.Order("created_at DESC").Offset(filter.Offset).Limit(filter.Limit).Find(&persons).Error; err != nil {
		return nil, 0, err
	}
	return persons, total, nil
}

func (r *PersonRepository) Create(ctx context.Context, person *model.Person) error {
	return r.db.WithContext(ctx).Create(person).Error
}

func (r *PersonRepository) GetByID(ctx context.Context, id uint64) (*model.Person, error) {
	var person model.Person
	if err := r.db.WithContext(ctx).First(&person, id).Error; err != nil {
		return nil, err
	}
	return &person, nil
}

func (r *PersonRepository) GetFirstEnabledByRoleWithDB(ctx context.Context, db *gorm.DB, roleType string) (*model.Person, error) {
	var person model.Person
	err := db.WithContext(ctx).
		Where("role_type = ? AND status = ?", roleType, 1).
		Order("id ASC").
		First(&person).Error
	if err != nil {
		return nil, err
	}
	return &person, nil
}

func (r *PersonRepository) GetByIDs(ctx context.Context, ids []uint64) ([]model.Person, error) {
	var persons []model.Person
	if len(ids) == 0 {
		return persons, nil
	}
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&persons).Error
	return persons, err
}

func (r *PersonRepository) ExistsByID(ctx context.Context, id uint64) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Person{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *PersonRepository) Update(ctx context.Context, id uint64, updates map[string]any) (*model.Person, error) {
	if err := r.db.WithContext(ctx).Model(&model.Person{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}
