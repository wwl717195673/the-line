package handler

import (
	"the-line-bridge/internal/response"

	"github.com/gin-gonic/gin"
)

type TestPingHandler struct{}

func NewTestPingHandler() *TestPingHandler {
	return &TestPingHandler{}
}

type TestPingRequest struct {
	ProtocolVersion int    `json:"protocol_version"`
	IntegrationID   uint64 `json:"integration_id"`
	PingID          string `json:"ping_id"`
	Kind            string `json:"kind"`
}

func (h *TestPingHandler) Ping(c *gin.Context) {
	var req TestPingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "INVALID_REQUEST", "test-ping 参数不合法", false)
		return
	}

	response.OK(c, gin.H{
		"pong":           true,
		"ping_id":        req.PingID,
		"bridge_version": bridgeVersion,
	})
}
