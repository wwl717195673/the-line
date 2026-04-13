package domain

const (
	AgentTaskStatusQueued      = "queued"
	AgentTaskStatusRunning     = "running"
	AgentTaskStatusCompleted   = "completed"
	AgentTaskStatusNeedsReview = "needs_review"
	AgentTaskStatusFailed      = "failed"
	AgentTaskStatusBlocked     = "blocked"
	AgentTaskStatusCancelled   = "cancelled"

	AgentTaskTypeQuery          = "query"
	AgentTaskTypeBatchOperation = "batch_operation"
	AgentTaskTypeExport         = "export"

	ReceiptStatusCompleted   = "completed"
	ReceiptStatusNeedsReview = "needs_review"
	ReceiptStatusFailed      = "failed"
	ReceiptStatusBlocked     = "blocked"
)
