package handler

import (
	"the-line/backend/internal/dto"
	"the-line/backend/internal/response"
	"the-line/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type AgentHandler struct {
	agentService *service.AgentService
}

func NewAgentHandler(agentService *service.AgentService) *AgentHandler {
	return &AgentHandler{agentService: agentService}
}

func (h *AgentHandler) List(c *gin.Context) {
	var req dto.AgentListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, response.Validation("龙虾列表查询参数不合法"))
		return
	}

	items, total, page, err := h.agentService.List(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, items, total, page.Page, page.PageSize)
}

func (h *AgentHandler) Create(c *gin.Context) {
	var req dto.CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.Validation("龙虾创建参数不合法"))
		return
	}

	agent, err := h.agentService.Create(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, agent)
}

func (h *AgentHandler) Update(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.Validation("龙虾更新参数不合法"))
		return
	}

	agent, err := h.agentService.Update(c.Request.Context(), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, agent)
}

func (h *AgentHandler) Disable(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	agent, err := h.agentService.Disable(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, agent)
}
