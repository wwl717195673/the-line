package service

import (
	"context"
	"errors"
	"net/mail"
	"strings"

	"the-line/backend/internal/domain"
	"the-line/backend/internal/dto"
	"the-line/backend/internal/model"
	"the-line/backend/internal/repository"
	"the-line/backend/internal/response"

	"gorm.io/gorm"
)

type PersonService struct {
	personRepo *repository.PersonRepository
}

func NewPersonService(personRepo *repository.PersonRepository) *PersonService {
	return &PersonService{personRepo: personRepo}
}

func (s *PersonService) List(ctx context.Context, req dto.PersonListRequest) ([]dto.PersonResponse, int64, dto.PageQuery, error) {
	page := req.PageQuery.Normalize()
	if req.Status != nil && !domain.IsValidStatus(*req.Status) {
		return nil, 0, page, response.Validation("人员状态只能是 0 或 1")
	}

	persons, total, err := s.personRepo.List(ctx, repository.PersonListFilter{
		Status:  req.Status,
		Keyword: strings.TrimSpace(req.Keyword),
		Offset:  page.Offset(),
		Limit:   page.PageSize,
	})
	if err != nil {
		return nil, 0, page, err
	}

	items := make([]dto.PersonResponse, 0, len(persons))
	for _, person := range persons {
		items = append(items, toPersonResponse(person))
	}
	return items, total, page, nil
}

func (s *PersonService) Create(ctx context.Context, req dto.CreatePersonRequest) (dto.PersonResponse, error) {
	name := strings.TrimSpace(req.Name)
	email := strings.TrimSpace(req.Email)
	roleType := strings.TrimSpace(req.RoleType)

	if name == "" {
		return dto.PersonResponse{}, response.Validation("人员姓名不能为空")
	}
	if email == "" {
		return dto.PersonResponse{}, response.Validation("人员邮箱不能为空")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return dto.PersonResponse{}, response.Validation("人员邮箱格式不合法")
	}
	if roleType == "" {
		return dto.PersonResponse{}, response.Validation("人员角色不能为空")
	}

	person := &model.Person{
		Name:     name,
		Email:    email,
		RoleType: roleType,
		Status:   domain.StatusEnabled,
	}
	if err := s.personRepo.Create(ctx, person); err != nil {
		return dto.PersonResponse{}, err
	}
	return toPersonResponse(*person), nil
}

func (s *PersonService) Update(ctx context.Context, id uint64, req dto.UpdatePersonRequest) (dto.PersonResponse, error) {
	if id == 0 {
		return dto.PersonResponse{}, response.Validation("人员 ID 不合法")
	}
	if _, err := s.personRepo.GetByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.PersonResponse{}, response.NotFound("人员不存在")
		}
		return dto.PersonResponse{}, err
	}

	updates := map[string]any{}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return dto.PersonResponse{}, response.Validation("人员姓名不能为空")
		}
		updates["name"] = name
	}
	if req.Email != nil {
		email := strings.TrimSpace(*req.Email)
		if email == "" {
			return dto.PersonResponse{}, response.Validation("人员邮箱不能为空")
		}
		if _, err := mail.ParseAddress(email); err != nil {
			return dto.PersonResponse{}, response.Validation("人员邮箱格式不合法")
		}
		updates["email"] = email
	}
	if req.RoleType != nil {
		roleType := strings.TrimSpace(*req.RoleType)
		if roleType == "" {
			return dto.PersonResponse{}, response.Validation("人员角色不能为空")
		}
		updates["role_type"] = roleType
	}
	if req.Status != nil {
		if !domain.IsValidStatus(*req.Status) {
			return dto.PersonResponse{}, response.Validation("人员状态只能是 0 或 1")
		}
		updates["status"] = *req.Status
	}
	if len(updates) == 0 {
		person, err := s.personRepo.GetByID(ctx, id)
		if err != nil {
			return dto.PersonResponse{}, err
		}
		return toPersonResponse(*person), nil
	}

	person, err := s.personRepo.Update(ctx, id, updates)
	if err != nil {
		return dto.PersonResponse{}, err
	}
	return toPersonResponse(*person), nil
}

func (s *PersonService) Disable(ctx context.Context, id uint64) (dto.PersonResponse, error) {
	return s.Update(ctx, id, dto.UpdatePersonRequest{Status: ptr(domain.StatusDisabled)})
}

func toPersonResponse(person model.Person) dto.PersonResponse {
	return dto.PersonResponse{
		ID:        person.ID,
		Name:      person.Name,
		Email:     person.Email,
		RoleType:  person.RoleType,
		Status:    person.Status,
		CreatedAt: person.CreatedAt,
		UpdatedAt: person.UpdatedAt,
	}
}

func ptr[T any](value T) *T {
	return &value
}
