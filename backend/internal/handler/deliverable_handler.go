package handler

import (
	"the-line/backend/internal/dto"
	"the-line/backend/internal/response"
	"the-line/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type DeliverableHandler struct {
	deliverableService *service.DeliverableService
}

func NewDeliverableHandler(deliverableService *service.DeliverableService) *DeliverableHandler {
	return &DeliverableHandler{deliverableService: deliverableService}
}

func (h *DeliverableHandler) List(c *gin.Context) {
	var req dto.DeliverableListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, response.Validation("交付物列表查询参数不合法"))
		return
	}

	items, total, page, err := h.deliverableService.List(c.Request.Context(), req, actorFromContext(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, items, total, page.Page, page.PageSize)
}

func (h *DeliverableHandler) Create(c *gin.Context) {
	var req dto.CreateDeliverableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.Validation("交付物创建参数不合法"))
		return
	}

	detail, err := h.deliverableService.Create(c.Request.Context(), req, actorFromContext(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, detail)
}

func (h *DeliverableHandler) Detail(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	detail, err := h.deliverableService.Detail(c.Request.Context(), id, actorFromContext(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, detail)
}

func (h *DeliverableHandler) Review(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.ReviewDeliverableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.Validation("交付物验收参数不合法"))
		return
	}

	detail, err := h.deliverableService.Review(c.Request.Context(), id, req, actorFromContext(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, detail)
}
