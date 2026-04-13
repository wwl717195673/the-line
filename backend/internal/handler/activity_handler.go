package handler

import (
	"the-line/backend/internal/dto"
	"the-line/backend/internal/response"
	"the-line/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type ActivityHandler struct {
	activityService *service.ActivityService
}

func NewActivityHandler(activityService *service.ActivityService) *ActivityHandler {
	return &ActivityHandler{activityService: activityService}
}

func (h *ActivityHandler) Recent(c *gin.Context) {
	var req dto.RecentActivityRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, response.Validation("最近动态查询参数不合法"))
		return
	}

	items, err := h.activityService.Recent(c.Request.Context(), req, actorFromContext(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, items)
}
