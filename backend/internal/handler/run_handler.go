package handler

import (
	"strconv"

	"the-line/backend/internal/domain"
	"the-line/backend/internal/dto"
	"the-line/backend/internal/response"
	"the-line/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type RunHandler struct {
	runService *service.RunService
}

func NewRunHandler(runService *service.RunService) *RunHandler {
	return &RunHandler{runService: runService}
}

func (h *RunHandler) Create(c *gin.Context) {
	var req dto.CreateRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.Validation("流程创建参数不合法"))
		return
	}

	detail, err := h.runService.CreateRun(c.Request.Context(), req, actorFromContext(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, detail)
}

func (h *RunHandler) List(c *gin.Context) {
	var req dto.RunListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, response.Validation("流程列表查询参数不合法"))
		return
	}

	items, total, page, err := h.runService.List(c.Request.Context(), req, actorFromContext(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, items, total, page.Page, page.PageSize)
}

func (h *RunHandler) Detail(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	detail, err := h.runService.Detail(c.Request.Context(), id, actorFromContext(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, detail)
}

func (h *RunHandler) Cancel(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.CancelRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.Validation("流程取消参数不合法"))
		return
	}

	detail, err := h.runService.CancelRun(c.Request.Context(), id, req.Reason, actorFromContext(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, detail)
}

func actorFromContext(c *gin.Context) domain.Actor {
	personID, _ := strconv.ParseUint(c.GetHeader("X-Person-ID"), 10, 64)
	return domain.Actor{
		PersonID: personID,
		RoleType: c.GetHeader("X-Role-Type"),
	}
}
