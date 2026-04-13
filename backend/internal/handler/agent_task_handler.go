package handler

import (
	"the-line/backend/internal/dto"
	"the-line/backend/internal/response"
	"the-line/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type AgentTaskHandler struct {
	agentTaskService        *service.AgentTaskService
	agentTaskReceiptService *service.AgentTaskReceiptService
}

func NewAgentTaskHandler(agentTaskService *service.AgentTaskService, agentTaskReceiptService *service.AgentTaskReceiptService) *AgentTaskHandler {
	return &AgentTaskHandler{
		agentTaskService:        agentTaskService,
		agentTaskReceiptService: agentTaskReceiptService,
	}
}

func (h *AgentTaskHandler) List(c *gin.Context) {
	var req dto.AgentTaskListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, response.Validation("龙虾任务列表查询参数不合法"))
		return
	}

	items, total, page, err := h.agentTaskService.List(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, items, total, page.Page, page.PageSize)
}

func (h *AgentTaskHandler) Detail(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	task, err := h.agentTaskService.Get(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, task)
}

func (h *AgentTaskHandler) LatestReceipt(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	receipt, err := h.agentTaskReceiptService.GetLatestByTaskID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, receipt)
}

func (h *AgentTaskHandler) Receipt(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.AgentReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.Validation("龙虾任务回执参数不合法"))
		return
	}

	if err := h.agentTaskService.ProcessReceipt(c.Request.Context(), id, req); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"success": true})
}
