package model

import (
	"time"

	"gorm.io/datatypes"
)

type FlowRun struct {
	ID                uint64         `gorm:"primaryKey" json:"id"`
	TemplateID        uint64         `gorm:"not null;index" json:"template_id"`
	TemplateVersion   int            `gorm:"not null" json:"template_version"`
	Title             string         `gorm:"size:256;not null" json:"title"`
	BizKey            string         `gorm:"size:128;index" json:"biz_key"`
	InitiatorPersonID uint64         `gorm:"not null;index" json:"initiator_person_id"`
	CurrentStatus     string         `gorm:"size:32;not null;index" json:"current_status"`
	CurrentNodeCode   string         `gorm:"size:64;index" json:"current_node_code"`
	InputPayloadJSON  datatypes.JSON `gorm:"type:json" json:"input_payload_json"`
	OutputPayloadJSON datatypes.JSON `gorm:"type:json" json:"output_payload_json"`
	StartedAt         *time.Time     `json:"started_at"`
	CompletedAt       *time.Time     `json:"completed_at"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

func (FlowRun) TableName() string {
	return "flow_runs"
}
