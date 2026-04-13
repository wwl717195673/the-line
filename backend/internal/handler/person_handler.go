package handler

import (
	"strconv"

	"the-line/backend/internal/dto"
	"the-line/backend/internal/response"
	"the-line/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type PersonHandler struct {
	personService *service.PersonService
}

func NewPersonHandler(personService *service.PersonService) *PersonHandler {
	return &PersonHandler{personService: personService}
}

func (h *PersonHandler) List(c *gin.Context) {
	var req dto.PersonListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, response.Validation("人员列表查询参数不合法"))
		return
	}

	items, total, page, err := h.personService.List(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, items, total, page.Page, page.PageSize)
}

func (h *PersonHandler) Create(c *gin.Context) {
	var req dto.CreatePersonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.Validation("人员创建参数不合法"))
		return
	}

	person, err := h.personService.Create(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, person)
}

func (h *PersonHandler) Update(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdatePersonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.Validation("人员更新参数不合法"))
		return
	}

	person, err := h.personService.Update(c.Request.Context(), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, person)
}

func (h *PersonHandler) Disable(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	person, err := h.personService.Disable(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, person)
}

func parseIDParam(c *gin.Context, name string) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, response.Validation("ID 参数不合法"))
		return 0, false
	}
	return id, true
}
