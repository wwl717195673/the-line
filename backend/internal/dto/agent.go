package dto

import (
	"encoding/json"
	"time"
)

type AgentListRequest struct {
	PageQuery
	Status  *int8  `form:"status"`
	Keyword string `form:"keyword"`
}

type CreateAgentRequest struct {
	Name          string          `json:"name"`
	Code          string          `json:"code"`
	Provider      string          `json:"provider"`
	Version       string          `json:"version"`
	OwnerPersonID uint64          `json:"owner_person_id"`
	ConfigJSON    json.RawMessage `json:"config_json"`
}

type UpdateAgentRequest struct {
	Name          *string          `json:"name"`
	Code          *string          `json:"code"`
	Provider      *string          `json:"provider"`
	Version       *string          `json:"version"`
	OwnerPersonID *uint64          `json:"owner_person_id"`
	ConfigJSON    *json.RawMessage `json:"config_json"`
	Status        *int8            `json:"status"`
}

type AgentResponse struct {
	ID            uint64          `json:"id"`
	Name          string          `json:"name"`
	Code          string          `json:"code"`
	Provider      string          `json:"provider"`
	Version       string          `json:"version"`
	OwnerPersonID uint64          `json:"owner_person_id"`
	ConfigJSON    json.RawMessage `json:"config_json"`
	Status        int8            `json:"status"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}
