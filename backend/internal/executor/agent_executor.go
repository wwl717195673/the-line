package executor

import (
	"context"

	"the-line/backend/internal/dto"
	"the-line/backend/internal/model"
)

type AgentExecutor interface {
	Execute(ctx context.Context, task *model.AgentTask, agent *model.Agent) error
}

type AgentPlannerExecutor interface {
	GenerateDraft(ctx context.Context, prompt string, agent *model.Agent) (*dto.DraftPlan, error)
}
