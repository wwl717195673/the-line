package app

import (
	"the-line-bridge/internal/client"
	"the-line-bridge/internal/handler"
	"the-line-bridge/internal/runtime"

	"github.com/gin-gonic/gin"
)

func NewRouter(rt runtime.OpenClawRuntime, thelineClient *client.TheLineClient) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	draftHandler := handler.NewDraftHandler(rt)
	executionHandler := handler.NewExecutionHandler(rt, thelineClient)
	healthHandler := handler.NewHealthHandler(rt)
	testPingHandler := handler.NewTestPingHandler()

	bridge := router.Group("/bridge")
	{
		bridge.POST("/drafts/generate", draftHandler.Generate)
		bridge.POST("/executions", executionHandler.Execute)
		bridge.POST("/executions/:agentTaskId/cancel", executionHandler.Cancel)
		bridge.GET("/health", healthHandler.Health)
		bridge.POST("/test-ping", testPingHandler.Ping)
	}

	return router
}
