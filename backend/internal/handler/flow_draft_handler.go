package handler

import (
	"the-line/backend/internal/dto"
	"the-line/backend/internal/response"
	"the-line/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type FlowDraftHandler struct {
	flowDraftService *service.FlowDraftService
}

func NewFlowDraftHandler(flowDraftService *service.FlowDraftService) *FlowDraftHandler {
	return &FlowDraftHandler{flowDraftService: flowDraftService}
}

func (h *FlowDraftHandler) List(c *gin.Context) {
	var req dto.FlowDraftListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, response.Validation("草案列表查询参数不合法"))
		return
	}

	items, total, page, err := h.flowDraftService.List(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, items, total, page.Page, page.PageSize)
}

func (h *FlowDraftHandler) Create(c *gin.Context) {
	var req dto.CreateFlowDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.Validation("草案创建参数不合法"))
		return
	}

	draft, err := h.flowDraftService.Create(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, draft)
}

func (h *FlowDraftHandler) Detail(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	draft, err := h.flowDraftService.Get(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, draft)
}

func (h *FlowDraftHandler) Update(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateFlowDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.Validation("草案更新参数不合法"))
		return
	}

	draft, err := h.flowDraftService.Update(c.Request.Context(), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, draft)
}

func (h *FlowDraftHandler) Confirm(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.ConfirmFlowDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.Validation("草案确认参数不合法"))
		return
	}

	result, err := h.flowDraftService.Confirm(c.Request.Context(), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *FlowDraftHandler) Discard(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.DiscardFlowDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.Validation("草案废弃参数不合法"))
		return
	}

	draft, err := h.flowDraftService.Discard(c.Request.Context(), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, draft)
}

func (h *FlowDraftHandler) Delete(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.flowDraftService.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"success": true})
}
