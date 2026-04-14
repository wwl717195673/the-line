package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"the-line-bridge/internal/client"
	"the-line-bridge/internal/receipt"
	"the-line-bridge/internal/response"
	"the-line-bridge/internal/runtime"

	"github.com/gin-gonic/gin"
)

type ExecutionHandler struct {
	rt     runtime.OpenClawRuntime
	client *client.TheLineClient
}

func NewExecutionHandler(rt runtime.OpenClawRuntime, client *client.TheLineClient) *ExecutionHandler {
	return &ExecutionHandler{rt: rt, client: client}
}

type ExecutionRequest struct {
	ProtocolVersion int             `json:"protocol_version"`
	IntegrationID   uint64          `json:"integration_id"`
	AgentTaskID     uint64          `json:"agent_task_id"`
	RunID           uint64          `json:"run_id"`
	RunNodeID       uint64          `json:"run_node_id"`
	AgentID         uint64          `json:"agent_id"`
	AgentCode       string          `json:"agent_code"`
	NodeType        string          `json:"node_type"`
	SessionKey      string          `json:"session_key"`
	Objective       string          `json:"objective"`
	InputJSON       json.RawMessage `json:"input_json"`
	Callback        *CallbackInfo   `json:"callback"`
	IdempotencyKey  string          `json:"idempotency_key"`
}

type CallbackInfo struct {
	URL               string `json:"url"`
	AuthType          string `json:"auth_type"`
	CallbackSecretRef string `json:"callback_secret_ref"`
}

type CancelRequest struct {
	ProtocolVersion int    `json:"protocol_version"`
	IntegrationID   uint64 `json:"integration_id"`
	AgentTaskID     uint64 `json:"agent_task_id"`
	Reason          string `json:"reason"`
}

func (h *ExecutionHandler) Execute(c *gin.Context) {
	var req ExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "INVALID_REQUEST", "执行请求参数不合法", false)
		return
	}

	sessionKey := req.SessionKey
	if sessionKey == "" {
		sessionKey = fmt.Sprintf("theline:run:%d:node:%d", req.RunID, req.RunNodeID)
	}

	result, err := h.rt.ExecuteTask(c.Request.Context(), runtime.ExecuteTaskRequest{
		SessionKey: sessionKey,
		AgentCode:  req.AgentCode,
		Objective:  req.Objective,
		InputJSON:  req.InputJSON,
	})
	if err != nil {
		response.Error(c, "EXECUTION_FAILED", err.Error(), true)
		return
	}

	// Respond immediately with accepted
	response.OK(c, gin.H{
		"accepted":             true,
		"agent_task_id":        req.AgentTaskID,
		"external_session_key": sessionKey,
		"external_run_id":      result.ExternalRunID,
		"status":               "running",
	})

	// Background: wait for result and post receipt
	go func() {
		startedAt := time.Now()
		taskResult, err := h.rt.WaitForResult(context.Background(), sessionKey)
		if err != nil {
			log.Printf("WaitForResult failed for task %d: %v", req.AgentTaskID, err)
			taskResult = &runtime.TaskResult{
				Status:       "failed",
				Summary:      "执行等待失败",
				ErrorMessage: err.Error(),
			}
		}

		receiptReq := receipt.MapToReceipt(req.IntegrationID, req.AgentID, taskResult, startedAt)
		if err := h.client.PostReceipt(req.AgentTaskID, receiptReq); err != nil {
			log.Printf("PostReceipt failed for task %d: %v", req.AgentTaskID, err)
		}
	}()
}

func (h *ExecutionHandler) Cancel(c *gin.Context) {
	var req CancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "INVALID_REQUEST", "取消请求参数不合法", false)
		return
	}

	sessionKey := fmt.Sprintf("agent_task:%d", req.AgentTaskID)
	if err := h.rt.CancelTask(c.Request.Context(), sessionKey); err != nil {
		response.Error(c, "CANCEL_FAILED", err.Error(), true)
		return
	}

	response.OK(c, gin.H{
		"accepted":      true,
		"agent_task_id": req.AgentTaskID,
		"status":        "cancelling",
	})
}
