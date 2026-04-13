package handler

import (
	"the-line/backend/internal/dto"
	"the-line/backend/internal/response"
	"the-line/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type TemplateHandler struct {
	templateService *service.TemplateService
}

func NewTemplateHandler(templateService *service.TemplateService) *TemplateHandler {
	return &TemplateHandler{templateService: templateService}
}

func (h *TemplateHandler) List(c *gin.Context) {
	var req dto.TemplateListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, response.Validation("模板列表查询参数不合法"))
		return
	}

	items, total, page, err := h.templateService.List(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, items, total, page.Page, page.PageSize)
}

func (h *TemplateHandler) Detail(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	detail, err := h.templateService.Detail(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, detail)
}

func (h *TemplateHandler) Delete(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.templateService.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"success": true})
}
