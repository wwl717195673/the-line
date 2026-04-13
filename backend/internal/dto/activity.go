package dto

import "time"

type RecentActivityRequest struct {
	Limit int `form:"limit"`
}

type RecentActivityResponse struct {
	ID           uint64    `json:"id"`
	RunID        uint64    `json:"run_id"`
	RunTitle     string    `json:"run_title"`
	RunNodeID    uint64    `json:"run_node_id"`
	NodeName     string    `json:"node_name"`
	LogType      string    `json:"log_type"`
	OperatorType string    `json:"operator_type"`
	OperatorID   uint64    `json:"operator_id"`
	OperatorName string    `json:"operator_name"`
	Content      string    `json:"content"`
	CreatedAt    time.Time `json:"created_at"`
}
