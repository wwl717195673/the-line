package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"ok": true, "data": data})
}

func Error(c *gin.Context, code string, message string, retryable bool) {
	c.JSON(http.StatusBadRequest, gin.H{
		"ok": false,
		"error": gin.H{
			"code":      code,
			"message":   message,
			"retryable": retryable,
		},
	})
}

func ErrorWithStatus(c *gin.Context, httpStatus int, code string, message string, retryable bool) {
	c.JSON(httpStatus, gin.H{
		"ok": false,
		"error": gin.H{
			"code":      code,
			"message":   message,
			"retryable": retryable,
		},
	})
}
