package handler

import (
	"encoding/json"

	"the-line-bridge/internal/response"
	"the-line-bridge/internal/runtime"

	"github.com/gin-gonic/gin"
)

type DraftHandler struct {
	rt runtime.OpenClawRuntime
}

func NewDraftHandler(rt runtime.OpenClawRuntime) *DraftHandler {
	return &DraftHandler{rt: rt}
}

type DraftGenerateRequest struct {
	ProtocolVersion int             `json:"protocol_version"`
	IntegrationID   uint64          `json:"integration_id"`
	DraftID         uint64          `json:"draft_id"`
	PlannerAgentID  string          `json:"planner_agent_id"`
	SessionKey      string          `json:"session_key"`
	SourcePrompt    string          `json:"source_prompt"`
	Constraints     json.RawMessage `json:"constraints"`
	IdempotencyKey  string          `json:"idempotency_key"`
}

func (h *DraftHandler) Generate(c *gin.Context) {
	var req DraftGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "INVALID_REQUEST", "草案生成请求参数不合法", false)
		return
	}

	result, err := h.rt.PlanDraft(c.Request.Context(), runtime.PlanDraftRequest{
		SessionKey:   req.SessionKey,
		AgentID:      req.PlannerAgentID,
		SourcePrompt: req.SourcePrompt,
		Constraints:  req.Constraints,
	})
	if err != nil {
		response.Error(c, "PLANNER_EXECUTION_FAILED", err.Error(), true)
		return
	}

	var nodes json.RawMessage
	if result.Nodes != nil {
		nodes = result.Nodes
	} else {
		nodes = json.RawMessage("[]")
	}

	response.OK(c, gin.H{
		"draft_id": req.DraftID,
		"plan": gin.H{
			"title":       result.Title,
			"description": result.Description,
			"nodes":       json.RawMessage(nodes),
		},
		"summary": result.Summary,
	})
}
