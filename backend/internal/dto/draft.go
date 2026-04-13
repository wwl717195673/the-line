package dto

import (
	"encoding/json"
	"time"
)

type FlowDraftListRequest struct {
	PageQuery
	CreatorPersonID uint64 `form:"creator_person_id"`
	Status          string `form:"status"`
}

type CreateFlowDraftRequest struct {
	Title              string          `json:"title"`
	Description        string          `json:"description"`
	SourcePrompt       string          `json:"source_prompt"`
	CreatorPersonID    uint64          `json:"creator_person_id"`
	PlannerAgentID     *uint64         `json:"planner_agent_id"`
	StructuredPlanJSON json.RawMessage `json:"structured_plan_json"`
}

type UpdateFlowDraftRequest struct {
	Title              *string          `json:"title"`
	Description        *string          `json:"description"`
	PlannerAgentID     *uint64          `json:"planner_agent_id"`
	StructuredPlanJSON *json.RawMessage `json:"structured_plan_json"`
}

type FlowDraftResponse struct {
	ID                  uint64          `json:"id"`
	Title               string          `json:"title"`
	Description         string          `json:"description"`
	SourcePrompt        string          `json:"source_prompt"`
	CreatorPersonID     uint64          `json:"creator_person_id"`
	PlannerAgentID      *uint64         `json:"planner_agent_id"`
	Status              string          `json:"status"`
	StructuredPlanJSON  json.RawMessage `json:"structured_plan_json"`
	ConfirmedTemplateID *uint64         `json:"confirmed_template_id,omitempty"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
	ConfirmedAt         *time.Time      `json:"confirmed_at,omitempty"`
}

type DraftPlan struct {
	Title            string      `json:"title"`
	Description      string      `json:"description"`
	Nodes            []DraftNode `json:"nodes"`
	FinalDeliverable string      `json:"final_deliverable"`
}

type DraftNode struct {
	NodeCode            string          `json:"node_code"`
	NodeName            string          `json:"node_name"`
	NodeType            string          `json:"node_type"`
	SortOrder           int             `json:"sort_order"`
	ExecutorType        string          `json:"executor_type"`
	OwnerRule           string          `json:"owner_rule"`
	OwnerPersonID       *uint64         `json:"owner_person_id"`
	ExecutorAgentCode   string          `json:"executor_agent_code"`
	ResultOwnerRule     string          `json:"result_owner_rule"`
	ResultOwnerPersonID *uint64         `json:"result_owner_person_id"`
	TaskType            string          `json:"task_type"`
	InputSchema         json.RawMessage `json:"input_schema"`
	OutputSchema        json.RawMessage `json:"output_schema"`
	CompletionCondition string          `json:"completion_condition"`
	FailureCondition    string          `json:"failure_condition"`
	EscalationRule      string          `json:"escalation_rule"`
}

type ConfirmFlowDraftRequest struct {
	ConfirmedBy uint64 `json:"confirmed_by"`
}

type ConfirmFlowDraftResponse struct {
	DraftID    uint64 `json:"draft_id"`
	TemplateID uint64 `json:"template_id"`
	Message    string `json:"message"`
}

type DiscardFlowDraftRequest struct {
	DiscardedBy uint64 `json:"discarded_by"`
	Reason      string `json:"reason"`
}
