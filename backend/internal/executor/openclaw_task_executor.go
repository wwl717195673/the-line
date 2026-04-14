package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"the-line/backend/internal/model"
	"the-line/backend/internal/repository"
	"the-line/backend/internal/response"
)

type OpenClawTaskExecutor struct {
	integrationRepo *repository.OpenClawIntegrationRepository
	httpClient      *http.Client
	receiptURL      string // the-line's base URL for receipt callbacks
}

func NewOpenClawTaskExecutor(integrationRepo *repository.OpenClawIntegrationRepository, receiptBaseURL string) *OpenClawTaskExecutor {
	return &OpenClawTaskExecutor{
		integrationRepo: integrationRepo,
		httpClient:      &http.Client{Timeout: 30 * time.Second},
		receiptURL:      receiptBaseURL,
	}
}

func (e *OpenClawTaskExecutor) Execute(ctx context.Context, task *model.AgentTask, agent *model.Agent) error {
	integration, err := e.integrationRepo.GetActiveByAgentID(ctx, agent.ID)
	if err != nil {
		return response.Validation("未找到可用的 OpenClaw 集成实例")
	}

	sessionKey := fmt.Sprintf("theline:run:%d:node:%d", task.RunID, task.RunNodeID)
	callbackURL := fmt.Sprintf("%s/api/agent-tasks/%d/receipt", e.receiptURL, task.ID)

	reqBody := map[string]any{
		"protocol_version": 1,
		"integration_id":   integration.ID,
		"agent_task_id":    task.ID,
		"run_id":           task.RunID,
		"run_node_id":      task.RunNodeID,
		"agent_code":       agent.Code,
		"node_type":        "agent_execute",
		"session_key":      sessionKey,
		"objective":        "",
		"input_json":       json.RawMessage(task.InputJSON),
		"callback": map[string]any{
			"url":                 callbackURL,
			"auth_type":           "signature",
			"callback_secret_ref": integration.CallbackSecret,
		},
		"idempotency_key": fmt.Sprintf("agent_task:%d", task.ID),
	}
	bodyBytes, _ := json.Marshal(reqBody)

	url := integration.CallbackURL + "/bridge/executions"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := e.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("调用 bridge 执行失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bridge 执行返回 HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var bridgeResp struct {
		OK   bool `json:"ok"`
		Data struct {
			ExternalSessionKey string `json:"external_session_key"`
			ExternalRunID      string `json:"external_run_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &bridgeResp); err != nil {
		return fmt.Errorf("解析 bridge 执行响应失败: %w", err)
	}

	// Store external references on task (caller should save)
	task.ExternalRuntime = "openclaw"
	task.ExternalSessionKey = bridgeResp.Data.ExternalSessionKey
	task.ExternalRunID = bridgeResp.Data.ExternalRunID

	return nil
}
