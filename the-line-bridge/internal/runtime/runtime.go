package runtime

import (
	"context"
	"encoding/json"
)

type PlanDraftRequest struct {
	SessionKey   string          `json:"session_key"`
	AgentID      string          `json:"agent_id"`
	SourcePrompt string          `json:"source_prompt"`
	Constraints  json.RawMessage `json:"constraints"`
}

type PlanDraftResult struct {
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Nodes       json.RawMessage `json:"nodes"`
	Summary     string          `json:"summary"`
}

type ExecuteTaskRequest struct {
	SessionKey string          `json:"session_key"`
	AgentCode  string          `json:"agent_code"`
	Objective  string          `json:"objective"`
	InputJSON  json.RawMessage `json:"input_json"`
}

type ExecuteTaskResult struct {
	ExternalRunID string `json:"external_run_id"`
}

type TaskResult struct {
	Status       string          `json:"status"`
	Summary      string          `json:"summary"`
	Result       json.RawMessage `json:"result"`
	Artifacts    json.RawMessage `json:"artifacts"`
	Logs         []string        `json:"logs"`
	ErrorMessage string          `json:"error_message"`
}

type HealthStatus struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

type AgentInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type OpenClawRuntime interface {
	PlanDraft(ctx context.Context, req PlanDraftRequest) (*PlanDraftResult, error)
	ExecuteTask(ctx context.Context, req ExecuteTaskRequest) (*ExecuteTaskResult, error)
	WaitForResult(ctx context.Context, sessionKey string) (*TaskResult, error)
	CancelTask(ctx context.Context, sessionKey string) error
	Health(ctx context.Context) (*HealthStatus, error)
	ListAgents(ctx context.Context) ([]AgentInfo, error)
}
