package model

import (
	"time"

	"gorm.io/datatypes"
)

type AgentTaskReceipt struct {
	ID            uint64         `gorm:"primaryKey" json:"id"`
	AgentTaskID   uint64         `gorm:"not null;index" json:"agent_task_id"`
	RunID         uint64         `gorm:"not null;index" json:"run_id"`
	RunNodeID     uint64         `gorm:"not null;index" json:"run_node_id"`
	AgentID       uint64         `gorm:"not null;index" json:"agent_id"`
	ReceiptStatus string         `gorm:"size:32;not null" json:"receipt_status"`
	PayloadJSON   datatypes.JSON `gorm:"type:json" json:"payload_json"`
	ReceivedAt    time.Time      `json:"received_at"`
}

func (AgentTaskReceipt) TableName() string {
	return "agent_task_receipts"
}
