package receipt

import (
	"the-line-bridge/internal/client"
	"the-line-bridge/internal/runtime"
	"time"
)

func MapToReceipt(integrationID uint64, agentID uint64, result *runtime.TaskResult, startedAt time.Time) client.ReceiptRequest {
	finishedAt := time.Now()

	status := mapStatus(result.Status)

	return client.ReceiptRequest{
		ProtocolVersion: 1,
		IntegrationID:   integrationID,
		AgentID:         agentID,
		Status:          status,
		StartedAt:       &startedAt,
		FinishedAt:      &finishedAt,
		Summary:         result.Summary,
		Result:          result.Result,
		Artifacts:       result.Artifacts,
		Logs:            result.Logs,
		ErrorMessage:    result.ErrorMessage,
	}
}

func mapStatus(openclawStatus string) string {
	switch openclawStatus {
	case "succeeded":
		return "completed"
	case "blocked":
		return "blocked"
	case "review_needed":
		return "needs_review"
	case "failed":
		return "failed"
	case "timed_out":
		return "failed"
	case "cancelled":
		return "cancelled"
	default:
		return "failed"
	}
}
