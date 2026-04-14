package handler

import (
	"the-line-bridge/internal/response"
	"the-line-bridge/internal/runtime"

	"github.com/gin-gonic/gin"
)

const bridgeVersion = "0.1.0"

type HealthHandler struct {
	rt runtime.OpenClawRuntime
}

func NewHealthHandler(rt runtime.OpenClawRuntime) *HealthHandler {
	return &HealthHandler{rt: rt}
}

func (h *HealthHandler) Health(c *gin.Context) {
	health, err := h.rt.Health(c.Request.Context())
	status := "healthy"
	ocVersion := ""
	if err != nil {
		status = "degraded"
	} else {
		status = health.Status
		ocVersion = health.Version
	}

	response.OK(c, gin.H{
		"status":                    status,
		"bridge_version":            bridgeVersion,
		"openclaw_version":          ocVersion,
		"supports_protocol_version": 1,
	})
}
