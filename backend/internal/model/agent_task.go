package model

import (
	"time"

	"gorm.io/datatypes"
)

type AgentTask struct {
	ID            uint64         `gorm:"primaryKey" json:"id"`
	RunID         uint64         `gorm:"not null;index" json:"run_id"`
	RunNodeID     uint64         `gorm:"not null;index" json:"run_node_id"`
	AgentID       uint64         `gorm:"not null;index" json:"agent_id"`
	TaskType      string         `gorm:"size:64;not null" json:"task_type"`
	InputJSON     datatypes.JSON `gorm:"type:json" json:"input_json"`
	Status        string         `gorm:"size:32;not null;index" json:"status"`
	StartedAt     *time.Time     `json:"started_at"`
	FinishedAt    *time.Time     `json:"finished_at"`
	ErrorMessage  string         `gorm:"type:text" json:"error_message"`
	ResultJSON    datatypes.JSON `gorm:"type:json" json:"result_json"`
	ArtifactsJSON datatypes.JSON `gorm:"type:json" json:"artifacts_json"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

func (AgentTask) TableName() string {
	return "agent_tasks"
}
