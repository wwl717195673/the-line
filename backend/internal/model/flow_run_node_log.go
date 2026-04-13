package model

import (
	"time"

	"gorm.io/datatypes"
)

type FlowRunNodeLog struct {
	ID           uint64         `gorm:"primaryKey" json:"id"`
	RunID        uint64         `gorm:"not null;index" json:"run_id"`
	RunNodeID    uint64         `gorm:"not null;index" json:"run_node_id"`
	LogType      string         `gorm:"size:32;not null" json:"log_type"`
	OperatorType string         `gorm:"size:32;not null" json:"operator_type"`
	OperatorID   uint64         `gorm:"index" json:"operator_id"`
	Content      string         `gorm:"type:text;not null" json:"content"`
	ExtraJSON    datatypes.JSON `gorm:"type:json" json:"extra_json"`
	CreatedAt    time.Time      `json:"created_at"`
}

func (FlowRunNodeLog) TableName() string {
	return "flow_run_node_logs"
}
