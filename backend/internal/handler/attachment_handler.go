package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"the-line/backend/internal/dto"
	"the-line/backend/internal/response"
	"the-line/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type AttachmentHandler struct {
	attachmentService *service.AttachmentService
}

func NewAttachmentHandler(attachmentService *service.AttachmentService) *AttachmentHandler {
	return &AttachmentHandler{attachmentService: attachmentService}
}

func (h *AttachmentHandler) List(c *gin.Context) {
	var req dto.AttachmentListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, response.Validation("附件列表查询参数不合法"))
		return
	}

	attachments, err := h.attachmentService.List(c.Request.Context(), req, actorFromContext(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, attachments)
}

func (h *AttachmentHandler) Create(c *gin.Context) {
	var req dto.CreateAttachmentRequest
	if strings.HasPrefix(c.GetHeader("Content-Type"), "multipart/form-data") {
		parsed, err := h.bindMultipart(c)
		if err != nil {
			response.Error(c, err)
			return
		}
		req = parsed
	} else if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.Validation("附件创建参数不合法"))
		return
	}

	attachment, err := h.attachmentService.Create(c.Request.Context(), req, actorFromContext(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, attachment)
}

func (h *AttachmentHandler) bindMultipart(c *gin.Context) (dto.CreateAttachmentRequest, error) {
	targetID, err := strconv.ParseUint(c.PostForm("target_id"), 10, 64)
	if err != nil {
		return dto.CreateAttachmentRequest{}, response.Validation("附件目标 ID 不合法")
	}

	file, err := c.FormFile("file")
	if err != nil {
		return dto.CreateAttachmentRequest{}, response.Validation("上传文件不能为空")
	}

	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "uploads"
	}
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return dto.CreateAttachmentRequest{}, err
	}

	originalName := filepath.Base(file.Filename)
	storedName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), sanitizeFileName(originalName))
	targetPath := filepath.Join(uploadDir, storedName)
	if err := c.SaveUploadedFile(file, targetPath); err != nil {
		return dto.CreateAttachmentRequest{}, err
	}

	fileType := file.Header.Get("Content-Type")
	if fileType == "" {
		fileType = c.PostForm("file_type")
	}

	return dto.CreateAttachmentRequest{
		TargetType: c.PostForm("target_type"),
		TargetID:   targetID,
		FileName:   originalName,
		FileURL:    "/uploads/" + storedName,
		FileSize:   file.Size,
		FileType:   fileType,
	}, nil
}

func sanitizeFileName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	if name == "" || name == "." {
		return "file"
	}
	return name
}
