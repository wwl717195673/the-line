package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PageData struct {
	Items    any   `json:"items"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, data)
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, data)
}

func Page(c *gin.Context, items any, total int64, page int, pageSize int) {
	c.JSON(http.StatusOK, PageData{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func Error(c *gin.Context, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.HTTPStatus, appErr)
		return
	}
	internal := Internal("系统内部错误")
	c.JSON(internal.HTTPStatus, internal)
}
