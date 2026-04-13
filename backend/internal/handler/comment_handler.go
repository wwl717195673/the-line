package handler

import (
	"the-line/backend/internal/dto"
	"the-line/backend/internal/response"
	"the-line/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type CommentHandler struct {
	commentService *service.CommentService
}

func NewCommentHandler(commentService *service.CommentService) *CommentHandler {
	return &CommentHandler{commentService: commentService}
}

func (h *CommentHandler) List(c *gin.Context) {
	var req dto.CommentListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, response.Validation("评论列表查询参数不合法"))
		return
	}

	comments, err := h.commentService.List(c.Request.Context(), req, actorFromContext(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, comments)
}

func (h *CommentHandler) Create(c *gin.Context) {
	var req dto.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.Validation("评论创建参数不合法"))
		return
	}

	comment, err := h.commentService.Create(c.Request.Context(), req, actorFromContext(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, comment)
}

func (h *CommentHandler) Resolve(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	comment, err := h.commentService.Resolve(c.Request.Context(), id, actorFromContext(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, comment)
}
