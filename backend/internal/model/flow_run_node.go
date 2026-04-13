package model

import (
	"time"

	"gorm.io/datatypes"
)

type FlowRunNode struct {
	ID                  uint64         `gorm:"primaryKey" json:"id"`
	RunID               uint64         `gorm:"not null;index" json:"run_id"`
	TemplateNodeID      uint64         `gorm:"not null;index" json:"template_node_id"`
	NodeCode            string         `gorm:"size:64;not null;index" json:"node_code"`
	NodeName            string         `gorm:"size:128;not null" json:"node_name"`
	NodeType            string         `gorm:"size:32;not null;index" json:"node_type"`
	SortOrder           int            `gorm:"not null;index" json:"sort_order"`
	OwnerPersonID       *uint64        `gorm:"index" json:"owner_person_id"`
	ReviewerPersonID    *uint64        `gorm:"index" json:"reviewer_person_id"`
	ResultOwnerPersonID *uint64        `gorm:"index" json:"result_owner_person_id"`
	BoundAgentID        *uint64        `gorm:"index" json:"bound_agent_id"`
	Status              string         `gorm:"size:32;not null;index" json:"status"`
	InputJSON           datatypes.JSON `gorm:"type:json" json:"input_json"`
	OutputJSON          datatypes.JSON `gorm:"type:json" json:"output_json"`
	StartedAt           *time.Time     `json:"started_at"`
	CompletedAt         *time.Time     `json:"completed_at"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
}

func (FlowRunNode) TableName() string {
	return "flow_run_nodes"
}
