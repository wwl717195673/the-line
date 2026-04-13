package dto

import "time"

type PersonListRequest struct {
	PageQuery
	Status  *int8  `form:"status"`
	Keyword string `form:"keyword"`
}

type CreatePersonRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	RoleType string `json:"role_type"`
}

type UpdatePersonRequest struct {
	Name     *string `json:"name"`
	Email    *string `json:"email"`
	RoleType *string `json:"role_type"`
	Status   *int8   `json:"status"`
}

type PersonResponse struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	RoleType  string    `json:"role_type"`
	Status    int8      `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
