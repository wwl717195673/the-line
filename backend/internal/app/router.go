package app

import (
	"context"
	"the-line/backend/internal/config"
	"the-line/backend/internal/dto"
	"the-line/backend/internal/executor"
	"the-line/backend/internal/handler"
	"the-line/backend/internal/repository"
	"the-line/backend/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func NewRouter(cfg config.Config, database *gorm.DB) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	personRepo := repository.NewPersonRepository(database)
	agentRepo := repository.NewAgentRepository(database)
	templateRepo := repository.NewTemplateRepository(database)
	runRepo := repository.NewRunRepository(database)
	runNodeRepo := repository.NewRunNodeRepository(database)
	nodeLogRepo := repository.NewNodeLogRepository(database)
	flowDraftRepo := repository.NewFlowDraftRepository(database)
	agentTaskRepo := repository.NewAgentTaskRepository(database)
	agentTaskReceiptRepo := repository.NewAgentTaskReceiptRepository(database)
	attachmentRepo := repository.NewAttachmentRepository(database)
	commentRepo := repository.NewCommentRepository(database)
	deliverableRepo := repository.NewDeliverableRepository(database)
	integrationRepo := repository.NewOpenClawIntegrationRepository(database)
	regCodeRepo := repository.NewRegistrationCodeRepository(database)

	personService := service.NewPersonService(personRepo)
	agentService := service.NewAgentService(agentRepo, personRepo)
	templateService := service.NewTemplateService(database, templateRepo, runRepo, agentRepo)
	runService := service.NewRunService(database, runRepo, runNodeRepo, nodeLogRepo, templateRepo, personRepo, agentRepo, deliverableRepo)
	flowDraftService := service.NewFlowDraftService(database, flowDraftRepo, templateRepo, personRepo, agentRepo)
	agentTaskService := service.NewAgentTaskService(database, agentTaskRepo, agentTaskReceiptRepo, runRepo, runNodeRepo, nodeLogRepo, agentRepo, runService)
	agentTaskReceiptService := service.NewAgentTaskReceiptService(agentTaskReceiptRepo)
	runNodeService := service.NewRunNodeService(database, runService, runRepo, runNodeRepo, nodeLogRepo, personRepo, agentRepo, commentRepo, attachmentRepo)
	runOrchestrationService := service.NewRunOrchestrationService(runNodeRepo, agentTaskService)
	integrationService := service.NewOpenClawIntegrationService(database, integrationRepo, regCodeRepo, agentRepo, agentTaskRepo)

	var plannerExec executor.AgentPlannerExecutor
	var agentExec executor.AgentExecutor

	receiptCallback := func(ctx context.Context, taskID uint64, req *dto.AgentReceiptRequest) error {
		return agentTaskService.ProcessReceipt(ctx, taskID, *req)
	}

	if cfg.ExecutorMode == "openclaw" {
		plannerExec = executor.NewOpenClawPlannerExecutor(integrationRepo)
		agentExec = executor.NewOpenClawTaskExecutor(integrationRepo, "http://localhost:"+cfg.AppPort)
	} else {
		plannerExec = executor.NewMockAgentPlannerExecutor()
		agentExec = executor.NewMockAgentExecutor(receiptCallback)
	}

	flowDraftService.SetPlannerExecutor(plannerExec)
	runService.SetOrchestrationService(runOrchestrationService)
	runNodeService.SetOrchestrationService(runOrchestrationService)
	agentTaskService.SetOrchestrationService(runOrchestrationService)
	agentTaskService.SetExecutor(agentExec)
	commentService := service.NewCommentService(commentRepo, runRepo, runNodeRepo, personRepo)
	attachmentService := service.NewAttachmentService(database, attachmentRepo, commentRepo, deliverableRepo, runRepo, runNodeRepo, nodeLogRepo)
	deliverableService := service.NewDeliverableService(database, deliverableRepo, runRepo, runNodeRepo, personRepo, agentRepo, attachmentRepo)
	activityService := service.NewActivityService(nodeLogRepo, runRepo, runNodeRepo, personRepo, agentRepo)

	personHandler := handler.NewPersonHandler(personService)
	agentHandler := handler.NewAgentHandler(agentService)
	templateHandler := handler.NewTemplateHandler(templateService)
	runHandler := handler.NewRunHandler(runService)
	flowDraftHandler := handler.NewFlowDraftHandler(flowDraftService)
	agentTaskHandler := handler.NewAgentTaskHandler(agentTaskService, agentTaskReceiptService)
	runNodeHandler := handler.NewRunNodeHandler(runNodeService)
	commentHandler := handler.NewCommentHandler(commentService)
	attachmentHandler := handler.NewAttachmentHandler(attachmentService)
	deliverableHandler := handler.NewDeliverableHandler(deliverableService)
	activityHandler := handler.NewActivityHandler(activityService)
	integrationHandler := handler.NewOpenClawIntegrationHandler(integrationService)
	healthHandler := handler.NewHealthHandler(database)

	router.Static("/uploads", "uploads")
	router.GET("/api/healthz", healthHandler.Healthz)

	api := router.Group("/api")
	{
		persons := api.Group("/persons")
		{
			persons.GET("", personHandler.List)
			persons.POST("", personHandler.Create)
			persons.PUT("/:id", personHandler.Update)
			persons.POST("/:id/disable", personHandler.Disable)
		}

		agents := api.Group("/agents")
		{
			agents.GET("", agentHandler.List)
			agents.POST("", agentHandler.Create)
			agents.PUT("/:id", agentHandler.Update)
			agents.POST("/:id/disable", agentHandler.Disable)
		}

		templates := api.Group("/templates")
		{
			templates.GET("", templateHandler.List)
			templates.GET("/:id", templateHandler.Detail)
			templates.DELETE("/:id", templateHandler.Delete)
		}

		flowDrafts := api.Group("/flow-drafts")
		{
			flowDrafts.GET("", flowDraftHandler.List)
			flowDrafts.POST("", flowDraftHandler.Create)
			flowDrafts.GET("/:id", flowDraftHandler.Detail)
			flowDrafts.PUT("/:id", flowDraftHandler.Update)
			flowDrafts.POST("/:id/confirm", flowDraftHandler.Confirm)
			flowDrafts.POST("/:id/discard", flowDraftHandler.Discard)
			flowDrafts.DELETE("/:id", flowDraftHandler.Delete)
		}

		runs := api.Group("/runs")
		{
			runs.POST("", runHandler.Create)
			runs.GET("", runHandler.List)
			runs.GET("/:id", runHandler.Detail)
			runs.POST("/:id/cancel", runHandler.Cancel)
		}

		agentTasks := api.Group("/agent-tasks")
		{
			agentTasks.GET("", agentTaskHandler.List)
			agentTasks.GET("/:id", agentTaskHandler.Detail)
			agentTasks.POST("/:id/receipt", agentTaskHandler.Receipt)
			agentTasks.GET("/:id/receipt", agentTaskHandler.LatestReceipt)
		}

		runNodes := api.Group("/run-nodes")
		{
			runNodes.GET("/:id", runNodeHandler.Detail)
			runNodes.PUT("/:id/input", runNodeHandler.SaveInput)
			runNodes.POST("/:id/submit", runNodeHandler.Submit)
			runNodes.POST("/:id/approve", runNodeHandler.Approve)
			runNodes.POST("/:id/reject", runNodeHandler.Reject)
			runNodes.POST("/:id/request-material", runNodeHandler.RequestMaterial)
			runNodes.POST("/:id/complete", runNodeHandler.Complete)
			runNodes.POST("/:id/fail", runNodeHandler.Fail)
			runNodes.POST("/:id/run-agent", runNodeHandler.RunAgent)
			runNodes.POST("/:id/confirm-agent-result", runNodeHandler.ConfirmAgentResult)
			runNodes.POST("/:id/takeover", runNodeHandler.Takeover)
			runNodes.GET("/:id/logs", runNodeHandler.Logs)
		}

		comments := api.Group("/comments")
		{
			comments.GET("", commentHandler.List)
			comments.POST("", commentHandler.Create)
			comments.POST("/:id/resolve", commentHandler.Resolve)
		}

		attachments := api.Group("/attachments")
		{
			attachments.GET("", attachmentHandler.List)
			attachments.POST("", attachmentHandler.Create)
		}

		deliverables := api.Group("/deliverables")
		{
			deliverables.GET("", deliverableHandler.List)
			deliverables.POST("", deliverableHandler.Create)
			deliverables.GET("/:id", deliverableHandler.Detail)
			deliverables.POST("/:id/review", deliverableHandler.Review)
		}

		activities := api.Group("/activities")
		{
			activities.GET("/recent", activityHandler.Recent)
		}

		integrations := api.Group("/integrations/openclaw")
		{
			integrations.POST("/registration-codes", integrationHandler.CreateRegistrationCode)
			integrations.GET("/registration-codes", integrationHandler.ListRegistrationCodes)
			integrations.POST("/register", integrationHandler.Register)
			integrations.POST("/heartbeat", integrationHandler.Heartbeat)
			integrations.GET("", integrationHandler.List)
			integrations.GET("/:id", integrationHandler.Detail)
			integrations.POST("/:id/test", integrationHandler.TestPing)
			integrations.POST("/:id/disable", integrationHandler.Disable)
			integrations.GET("/:id/pending-tasks", integrationHandler.PendingTasks)
			integrations.POST("/:id/claim-task/:taskId", integrationHandler.ClaimTask)
		}
	}

	return router
}
