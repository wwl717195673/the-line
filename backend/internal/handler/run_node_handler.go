package handler

import (
	"errors"
	"io"

	"the-line/backend/internal/dto"
	"the-line/backend/internal/response"
	"the-line/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type RunNodeHandler struct {
	runNodeService *service.RunNodeService
}

func NewRunNodeHandler(runNodeService *service.RunNodeService) *RunNodeHandler {
	return &RunNodeHandler{runNodeService: runNodeService}
}

func (h *RunNodeHandler) Detail(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	detail, err := h.runNodeService.Detail(c.Request.Context(), id, actorFromContext(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, detail)
}

func (h *RunNodeHandler) SaveInput(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.SaveRunNodeInputRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.Validation("节点输入参数不合法"))
		return
	}

	detail, err := h.runNodeService.SaveInput(c.Request.Context(), id, req, actorFromContext(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, detail)
}

func (h *RunNodeHandler) Submit(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.SubmitRunNodeRequest
	if err := bindOptionalJSON(c, &req); err != nil {
		response.Error(c, response.Validation("节点提交参数不合法"))
		return
	}

	detail, err := h.runNodeService.Submit(c.Request.Context(), id, req, actorFromContext(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, detail)
}

func (h *RunNodeHandler) Approve(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.ApproveRunNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.Validation("节点审核参数不合法"))
		return
	}

	detail, err := h.runNodeService.Approve(c.Request.Context(), id, req, actorFromContext(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, detail)
}

func (h *RunNodeHandler) Reject(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.RejectRunNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.Validation("节点驳回参数不合法"))
		return
	}

	detail, err := h.runNodeService.Reject(c.Request.Context(), id, req, actorFromContext(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, detail)
}

func (h *RunNodeHandler) RequestMaterial(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.RequestMaterialRunNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.Validation("节点补材料参数不合法"))
		return
	}

	detail, err := h.runNodeService.RequestMaterial(c.Request.Context(), id, req, actorFromContext(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, detail)
}

func (h *RunNodeHandler) Complete(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.CompleteRunNodeRequest
	if err := bindOptionalJSON(c, &req); err != nil {
		response.Error(c, response.Validation("节点完成参数不合法"))
		return
	}

	detail, err := h.runNodeService.Complete(c.Request.Context(), id, req, actorFromContext(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, detail)
}

func bindOptionalJSON(c *gin.Context, req any) error {
	if err := c.ShouldBindJSON(req); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}
	return nil
}

func (h *RunNodeHandler) Fail(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.FailRunNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.Validation("节点异常参数不合法"))
		return
	}

	detail, err := h.runNodeService.Fail(c.Request.Context(), id, req, actorFromContext(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, detail)
}

func (h *RunNodeHandler) RunAgent(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	detail, err := h.runNodeService.RunAgent(c.Request.Context(), id, actorFromContext(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, detail)
}

func (h *RunNodeHandler) ConfirmAgentResult(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.ConfirmAgentResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.Validation("龙虾结果确认参数不合法"))
		return
	}

	detail, err := h.runNodeService.ConfirmAgentResult(c.Request.Context(), id, req, actorFromContext(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, detail)
}

func (h *RunNodeHandler) Takeover(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req dto.TakeoverRunNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.Validation("人工接管参数不合法"))
		return
	}

	detail, err := h.runNodeService.Takeover(c.Request.Context(), id, req, actorFromContext(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, detail)
}

func (h *RunNodeHandler) Logs(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	logs, err := h.runNodeService.Logs(c.Request.Context(), id, actorFromContext(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, logs)
}
