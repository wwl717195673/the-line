package handler

import (
	"context"
	"time"

	"the-line/backend/internal/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db *gorm.DB
}

func NewHealthHandler(database *gorm.DB) *HealthHandler {
	return &HealthHandler{db: database}
}

func (h *HealthHandler) Healthz(c *gin.Context) {
	sqlDB, err := h.db.DB()
	if err != nil {
		response.Error(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, gin.H{
		"status":   "ok",
		"database": "ok",
		"time":     time.Now(),
	})
}
