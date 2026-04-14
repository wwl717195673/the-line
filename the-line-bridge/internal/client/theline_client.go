package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type TheLineClient struct {
	baseURL        string
	integrationID  uint64
	callbackSecret string
	httpClient     *http.Client
}

func NewTheLineClient(baseURL string) *TheLineClient {
	return &TheLineClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *TheLineClient) SetCredentials(integrationID uint64, callbackSecret string) {
	c.integrationID = integrationID
	c.callbackSecret = callbackSecret
}

type RegisterRequest struct {
	ProtocolVersion     int             `json:"protocol_version"`
	RegistrationCode    string          `json:"registration_code"`
	BridgeVersion       string          `json:"bridge_version"`
	OpenClawVersion     string          `json:"openclaw_version"`
	InstanceFingerprint string          `json:"instance_fingerprint"`
	DisplayName         string          `json:"display_name"`
	BoundAgentID        string          `json:"bound_agent_id"`
	CallbackURL         string          `json:"callback_url"`
	Capabilities        map[string]bool `json:"capabilities"`
	IdempotencyKey      string          `json:"idempotency_key"`
}

type RegisterResponse struct {
	IntegrationID            uint64 `json:"integration_id"`
	Status                   string `json:"status"`
	CallbackSecret           string `json:"callback_secret"`
	HeartbeatIntervalSeconds int    `json:"heartbeat_interval_seconds"`
}

func (c *TheLineClient) Register(req RegisterRequest) (*RegisterResponse, error) {
	body, _ := json.Marshal(req)
	resp, err := c.doPost("/api/integrations/openclaw/register", body)
	if err != nil {
		return nil, err
	}

	var result RegisterResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("解析注册响应失败: %w", err)
	}
	return &result, nil
}

type HeartbeatRequest struct {
	IntegrationID   uint64 `json:"integration_id"`
	BridgeVersion   string `json:"bridge_version"`
	Status          string `json:"status"`
	ActiveRunsCount int    `json:"active_runs_count"`
	LastError       string `json:"last_error"`
}

func (c *TheLineClient) Heartbeat(req HeartbeatRequest) error {
	body, _ := json.Marshal(req)
	_, err := c.doPost("/api/integrations/openclaw/heartbeat", body)
	return err
}

type ReceiptRequest struct {
	ProtocolVersion int             `json:"protocol_version"`
	IntegrationID   uint64          `json:"integration_id"`
	AgentID         uint64          `json:"agent_id"`
	Status          string          `json:"status"`
	StartedAt       *time.Time      `json:"started_at"`
	FinishedAt      *time.Time      `json:"finished_at"`
	Summary         string          `json:"summary"`
	Result          json.RawMessage `json:"result"`
	Artifacts       json.RawMessage `json:"artifacts"`
	Logs            []string        `json:"logs"`
	ErrorMessage    string          `json:"error_message"`
}

func (c *TheLineClient) PostReceipt(taskID uint64, receipt ReceiptRequest) error {
	body, _ := json.Marshal(receipt)
	path := fmt.Sprintf("/api/agent-tasks/%d/receipt", taskID)
	_, err := c.doPost(path, body)
	return err
}

func (c *TheLineClient) doPost(path string, body []byte) ([]byte, error) {
	url := c.baseURL + path
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-The-Line-Protocol-Version", "1")
	if c.integrationID > 0 {
		req.Header.Set("X-The-Line-Integration-Id", fmt.Sprintf("%d", c.integrationID))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 the-line 失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("the-line 返回 HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	return respBody, nil
}
