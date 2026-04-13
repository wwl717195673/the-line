package dto

import (
	"encoding/json"
	"time"
)

type TemplateListRequest struct {
	PageQuery
	Keyword string `form:"keyword"`
}

type TemplateResponse struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Version     int       `json:"version"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedBy   uint64    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TemplateNodeResponse struct {
	ID                   uint64              `json:"id"`
	TemplateID           uint64              `json:"template_id"`
	NodeCode             string              `json:"node_code"`
	NodeName             string              `json:"node_name"`
	NodeType             string              `json:"node_type"`
	SortOrder            int                 `json:"sort_order"`
	DefaultOwnerRule     string              `json:"default_owner_rule"`
	DefaultOwnerPersonID *uint64             `json:"default_owner_person_id"`
	DefaultAgentID       *uint64             `json:"default_agent_id"`
	ResultOwnerRule      string              `json:"result_owner_rule"`
	ResultOwnerPersonID  *uint64             `json:"result_owner_person_id"`
	DefaultAgent         *AgentBriefResponse `json:"default_agent,omitempty"`
	InputSchemaJSON      json.RawMessage     `json:"input_schema_json"`
	OutputSchemaJSON     json.RawMessage     `json:"output_schema_json"`
	ConfigJSON           json.RawMessage     `json:"config_json"`
	CreatedAt            time.Time           `json:"created_at"`
	UpdatedAt            time.Time           `json:"updated_at"`
}

type AgentBriefResponse struct {
	ID       uint64 `json:"id"`
	Name     string `json:"name"`
	Code     string `json:"code"`
	Provider string `json:"provider"`
	Version  string `json:"version"`
	Status   int8   `json:"status"`
}

type TemplateDetailResponse struct {
	TemplateResponse
	Nodes []TemplateNodeResponse `json:"nodes"`
}
