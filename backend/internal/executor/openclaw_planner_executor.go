package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"the-line/backend/internal/dto"
	"the-line/backend/internal/model"
	"the-line/backend/internal/repository"
	"the-line/backend/internal/response"
)

type OpenClawPlannerExecutor struct {
	integrationRepo *repository.OpenClawIntegrationRepository
	httpClient      *http.Client
}

func NewOpenClawPlannerExecutor(integrationRepo *repository.OpenClawIntegrationRepository) *OpenClawPlannerExecutor {
	return &OpenClawPlannerExecutor{
		integrationRepo: integrationRepo,
		httpClient:      &http.Client{Timeout: 120 * time.Second},
	}
}

func (e *OpenClawPlannerExecutor) GenerateDraft(ctx context.Context, prompt string, agent *model.Agent) (*dto.DraftPlan, error) {
	integration, err := e.integrationRepo.GetActiveByAgentID(ctx, agent.ID)
	if err != nil {
		return nil, response.Validation("未找到可用的 OpenClaw 集成实例")
	}

	reqBody := map[string]any{
		"protocol_version": 1,
		"integration_id":   integration.ID,
		"planner_agent_id": agent.Code,
		"source_prompt":    prompt,
		"constraints": map[string]any{
			"must_end_with_human_acceptance": true,
			"allowed_node_types":            []string{"human_input", "human_review", "agent_execute", "agent_export", "human_acceptance"},
		},
		"output_schema_version": "v1",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	url := integration.CallbackURL + "/bridge/drafts/generate"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := e.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("调用 bridge 草案生成失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bridge 草案生成返回 HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var bridgeResp struct {
		OK   bool `json:"ok"`
		Data struct {
			Plan dto.DraftPlan `json:"plan"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &bridgeResp); err != nil {
		return nil, fmt.Errorf("解析 bridge 草案响应失败: %w", err)
	}
	if !bridgeResp.OK {
		return nil, fmt.Errorf("bridge 草案生成失败")
	}

	return &bridgeResp.Data.Plan, nil
}
