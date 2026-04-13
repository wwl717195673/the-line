package model

import (
	"time"

	"gorm.io/datatypes"
)

type FlowTemplateNode struct {
	ID                   uint64         `gorm:"primaryKey" json:"id"`
	TemplateID           uint64         `gorm:"not null;uniqueIndex:idx_template_node_code;uniqueIndex:idx_template_sort_order" json:"template_id"`
	NodeCode             string         `gorm:"size:64;not null;index;uniqueIndex:idx_template_node_code" json:"node_code"`
	NodeName             string         `gorm:"size:128;not null" json:"node_name"`
	NodeType             string         `gorm:"size:32;not null;index" json:"node_type"`
	SortOrder            int            `gorm:"not null;index;uniqueIndex:idx_template_sort_order" json:"sort_order"`
	DefaultOwnerRule     string         `gorm:"size:128" json:"default_owner_rule"`
	DefaultOwnerPersonID *uint64        `gorm:"index" json:"default_owner_person_id"`
	DefaultAgentID       *uint64        `gorm:"index" json:"default_agent_id"`
	ResultOwnerRule      string         `gorm:"size:128" json:"result_owner_rule"`
	ResultOwnerPersonID  *uint64        `gorm:"index" json:"result_owner_person_id"`
	InputSchemaJSON      datatypes.JSON `gorm:"type:json" json:"input_schema_json"`
	OutputSchemaJSON     datatypes.JSON `gorm:"type:json" json:"output_schema_json"`
	ConfigJSON           datatypes.JSON `gorm:"type:json" json:"config_json"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
}

func (FlowTemplateNode) TableName() string {
	return "flow_template_nodes"
}
