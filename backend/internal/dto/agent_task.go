package dto

import (
	"encoding/json"
	"time"
)

type AgentTaskListRequest struct {
	PageQuery
	RunID     uint64 `form:"run_id"`
	RunNodeID uint64 `form:"run_node_id"`
	Status    string `form:"status"`
}

type AgentTaskResponse struct {
	ID            uint64          `json:"id"`
	RunID         uint64          `json:"run_id"`
	RunNodeID     uint64          `json:"run_node_id"`
	AgentID       uint64          `json:"agent_id"`
	TaskType      string          `json:"task_type"`
	InputJSON     json.RawMessage `json:"input_json"`
	Status        string          `json:"status"`
	StartedAt     *time.Time      `json:"started_at,omitempty"`
	FinishedAt    *time.Time      `json:"finished_at,omitempty"`
	ErrorMessage  string          `json:"error_message"`
	ResultJSON    json.RawMessage `json:"result_json"`
	ArtifactsJSON json.RawMessage `json:"artifacts_json"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type AgentTaskReceiptResponse struct {
	ID            uint64          `json:"id"`
	AgentTaskID   uint64          `json:"agent_task_id"`
	RunID         uint64          `json:"run_id"`
	RunNodeID     uint64          `json:"run_node_id"`
	AgentID       uint64          `json:"agent_id"`
	ReceiptStatus string          `json:"receipt_status"`
	PayloadJSON   json.RawMessage `json:"payload_json"`
	ReceivedAt    time.Time       `json:"received_at"`
}

type AgentReceiptRequest struct {
	AgentID      uint64          `json:"agent_id" binding:"required"`
	Status       string          `json:"status" binding:"required"`
	StartedAt    *time.Time      `json:"started_at"`
	FinishedAt   *time.Time      `json:"finished_at"`
	Summary      string          `json:"summary"`
	Result       json.RawMessage `json:"result"`
	Artifacts    json.RawMessage `json:"artifacts"`
	Logs         []string        `json:"logs"`
	ErrorMessage string          `json:"error_message"`
}
