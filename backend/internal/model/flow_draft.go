package model

import (
	"time"

	"gorm.io/datatypes"
)

type FlowDraft struct {
	ID                  uint64         `gorm:"primaryKey" json:"id"`
	Title               string         `gorm:"size:256;not null" json:"title"`
	Description         string         `gorm:"type:text" json:"description"`
	SourcePrompt        string         `gorm:"type:text;not null" json:"source_prompt"`
	CreatorPersonID     uint64         `gorm:"not null;index" json:"creator_person_id"`
	PlannerAgentID      *uint64        `gorm:"index" json:"planner_agent_id"`
	Status              string         `gorm:"size:32;not null;index" json:"status"`
	StructuredPlanJSON  datatypes.JSON `gorm:"type:json" json:"structured_plan_json"`
	ConfirmedTemplateID *uint64        `gorm:"index" json:"confirmed_template_id"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	ConfirmedAt         *time.Time     `json:"confirmed_at"`
}

func (FlowDraft) TableName() string {
	return "flow_drafts"
}
