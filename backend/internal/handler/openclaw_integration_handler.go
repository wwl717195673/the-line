package handler

import (
	"the-line/backend/internal/dto"
	"the-line/backend/internal/response"
	"the-line/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type OpenClawIntegrationHandler struct {
	integrationService *service.OpenClawIntegrationService
}

func NewOpenClawIntegrationHandler(integrationService *service.OpenClawIntegrationService) *OpenClawIntegrationHandler {
	return &OpenClawIntegrationHandler{integrationService: integrationService}
}

func (h *OpenClawIntegrationHandler) CreateRegistrationCode(c *gin.Context) {
	var req dto.CreateRegistrationCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.Validation("注册码参数不合法"))
		return
	}

	code, err := h.integrationService.CreateRegistrationCode(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, code)
}

func (h *OpenClawIntegrationHandler) ListRegistrationCodes(c *gin.Context) {
	var req dto.RegistrationCodeListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, response.Validation("查询参数不合法"))
		return
	}

	items, total, page, err := h.integrationService.ListRegistrationCodes(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, items, total, page.Page, page.PageSize)
}

func (h *OpenClawIntegrationHandler) Register(c *gin.Context) {
	var req dto.BridgeRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.Validation("注册参数不合法"))
		return
	}

	result, err := h.integrationService.Register(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *OpenClawIntegrationHandler) Heartbeat(c *gin.Context) {
	var req dto.BridgeHeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.Validation("心跳参数不合法"))
		return
	}

	result, err := h.integrationService.Heartbeat(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *OpenClawIntegrationHandler) List(c *gin.Context) {
	var req dto.IntegrationListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, response.Validation("查询参数不合法"))
		return
	}

	items, total, page, err := h.integrationService.List(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, items, total, page.Page, page.PageSize)
}

func (h *OpenClawIntegrationHandler) Detail(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	result, err := h.integrationService.Get(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *OpenClawIntegrationHandler) Disable(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	result, err := h.integrationService.Disable(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *OpenClawIntegrationHandler) TestPing(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	result, err := h.integrationService.TestPing(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}
