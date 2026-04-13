package dto

import (
	"encoding/json"
	"time"
)

type CreateRunRequest struct {
	TemplateID        uint64          `json:"template_id"`
	Title             string          `json:"title"`
	BizKey            string          `json:"biz_key"`
	InitiatorPersonID uint64          `json:"initiator_person_id"`
	InputPayloadJSON  json.RawMessage `json:"input_payload_json"`
}

type RunListRequest struct {
	PageQuery
	Status            string `form:"status"`
	OwnerPersonID     uint64 `form:"owner_person_id"`
	InitiatorPersonID uint64 `form:"initiator_person_id"`
	Scope             string `form:"scope"`
}

type CancelRunRequest struct {
	Reason string `json:"reason"`
}

type RunResponse struct {
	ID                uint64               `json:"id"`
	TemplateID        uint64               `json:"template_id"`
	TemplateVersion   int                  `json:"template_version"`
	Title             string               `json:"title"`
	BizKey            string               `json:"biz_key"`
	InitiatorPersonID uint64               `json:"initiator_person_id"`
	Initiator         *PersonBriefResponse `json:"initiator,omitempty"`
	CurrentStatus     string               `json:"current_status"`
	CurrentNodeCode   string               `json:"current_node_code"`
	CurrentNode       *RunNodeResponse     `json:"current_node,omitempty"`
	InputPayloadJSON  json.RawMessage      `json:"input_payload_json"`
	OutputPayloadJSON json.RawMessage      `json:"output_payload_json"`
	StartedAt         *time.Time           `json:"started_at"`
	CompletedAt       *time.Time           `json:"completed_at"`
	CreatedAt         time.Time            `json:"created_at"`
	UpdatedAt         time.Time            `json:"updated_at"`
}

type RunNodeResponse struct {
	ID                  uint64               `json:"id"`
	RunID               uint64               `json:"run_id"`
	TemplateNodeID      uint64               `json:"template_node_id"`
	NodeCode            string               `json:"node_code"`
	NodeName            string               `json:"node_name"`
	NodeType            string               `json:"node_type"`
	SortOrder           int                  `json:"sort_order"`
	OwnerPersonID       *uint64              `json:"owner_person_id"`
	OwnerPerson         *PersonBriefResponse `json:"owner_person,omitempty"`
	ReviewerPersonID    *uint64              `json:"reviewer_person_id"`
	ReviewerPerson      *PersonBriefResponse `json:"reviewer_person,omitempty"`
	ResultOwnerPersonID *uint64              `json:"result_owner_person_id"`
	ResultOwnerPerson   *PersonBriefResponse `json:"result_owner_person,omitempty"`
	BoundAgentID        *uint64              `json:"bound_agent_id"`
	BoundAgent          *AgentBriefResponse  `json:"bound_agent,omitempty"`
	Status              string               `json:"status"`
	InputJSON           json.RawMessage      `json:"input_json"`
	OutputJSON          json.RawMessage      `json:"output_json"`
	StartedAt           *time.Time           `json:"started_at"`
	CompletedAt         *time.Time           `json:"completed_at"`
	CreatedAt           time.Time            `json:"created_at"`
	UpdatedAt           time.Time            `json:"updated_at"`
	IsCurrent           bool                 `json:"is_current"`
}

type RunDetailResponse struct {
	RunResponse
	Template       *TemplateResponse    `json:"template,omitempty"`
	Nodes          []RunNodeResponse    `json:"nodes"`
	HasDeliverable bool                 `json:"has_deliverable"`
	DeliverableID  *uint64              `json:"deliverable_id,omitempty"`
	Logs           []RunNodeLogResponse `json:"logs,omitempty"`
}

type RunNodeLogResponse struct {
	ID           uint64          `json:"id"`
	RunID        uint64          `json:"run_id"`
	RunNodeID    uint64          `json:"run_node_id"`
	LogType      string          `json:"log_type"`
	OperatorType string          `json:"operator_type"`
	OperatorID   uint64          `json:"operator_id"`
	Content      string          `json:"content"`
	ExtraJSON    json.RawMessage `json:"extra_json"`
	CreatedAt    time.Time       `json:"created_at"`
}

type PersonBriefResponse struct {
	ID       uint64 `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	RoleType string `json:"role_type"`
	Status   int8   `json:"status"`
}
