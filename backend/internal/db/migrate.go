package db

import (
	"the-line/backend/internal/model"

	"gorm.io/gorm"
)

func AutoMigrate(database *gorm.DB) error {
	if err := database.AutoMigrate(
		&model.Person{},
		&model.Agent{},
		&model.FlowTemplate{},
		&model.FlowTemplateNode{},
		&model.FlowDraft{},
		&model.FlowRun{},
		&model.FlowRunNode{},
		&model.AgentTask{},
		&model.AgentTaskReceipt{},
		&model.FlowRunNodeLog{},
		&model.Comment{},
		&model.Attachment{},
		&model.Deliverable{},
		&model.OpenClawIntegration{},
		&model.RegistrationCode{},
	); err != nil {
		return err
	}

	return SeedTeacherClassTransferTemplate(database)
}
