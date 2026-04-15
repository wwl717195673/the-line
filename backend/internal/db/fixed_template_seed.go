package db

import (
	"encoding/json"
	"errors"

	"the-line/backend/internal/domain"
	"the-line/backend/internal/model"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func SeedTeacherClassTransferTemplate(database *gorm.DB) error {
	templateConfig := domain.TeacherClassTransferTemplate()

	var template model.FlowTemplate
	if err := database.Where("code = ?", templateConfig.Code).First(&template).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		template = model.FlowTemplate{
			Name:        templateConfig.Name,
			Code:        templateConfig.Code,
			Version:     templateConfig.Version,
			Category:    templateConfig.Category,
			Description: templateConfig.Description,
			Status:      templateConfig.Status,
		}
		if err := database.Create(&template).Error; err != nil {
			return err
		}
	}

	for _, nodeConfig := range templateConfig.Nodes {
		var node model.FlowTemplateNode
		queryErr := database.Where("template_id = ? AND node_code = ?", template.ID, nodeConfig.NodeCode).First(&node).Error
		if queryErr != nil && !errors.Is(queryErr, gorm.ErrRecordNotFound) {
			return queryErr
		}

		defaultAgentID, err := findDefaultAgentID(database, nodeConfig.DefaultAgentCode)
		if err != nil {
			return err
		}

		payload := model.FlowTemplateNode{
			TemplateID:       template.ID,
			NodeCode:         nodeConfig.NodeCode,
			NodeName:         nodeConfig.NodeName,
			NodeType:         nodeConfig.NodeType,
			SortOrder:        nodeConfig.SortOrder,
			DefaultOwnerRule: nodeConfig.DefaultOwnerRule,
			DefaultAgentID:   defaultAgentID,
			InputSchemaJSON:  mustJSON(buildInputSchema(nodeConfig)),
			OutputSchemaJSON: mustJSON(buildOutputSchema(nodeConfig)),
			ConfigJSON:       mustJSON(buildNodeConfig(nodeConfig)),
		}

		if errors.Is(queryErr, gorm.ErrRecordNotFound) {
			if err := database.Create(&payload).Error; err != nil {
				return err
			}
			continue
		}

		updates := map[string]any{
			"node_name":          payload.NodeName,
			"node_type":          payload.NodeType,
			"sort_order":         payload.SortOrder,
			"default_owner_rule": payload.DefaultOwnerRule,
			"default_agent_id":   payload.DefaultAgentID,
			"input_schema_json":  payload.InputSchemaJSON,
			"output_schema_json": payload.OutputSchemaJSON,
			"config_json":        payload.ConfigJSON,
		}
		if err := database.Model(&model.FlowTemplateNode{}).Where("id = ?", node.ID).Updates(updates).Error; err != nil {
			return err
		}
	}

	return nil
}

func findDefaultAgentID(database *gorm.DB, agentCode string) (*uint64, error) {
	if agentCode == "" {
		return nil, nil
	}
	var agent model.Agent
	if err := database.Where("code = ?", agentCode).First(&agent).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &agent.ID, nil
}

func buildInputSchema(node domain.FixedNodeConfig) map[string]any {
	return map[string]any{
		"type":            "object",
		"required_fields": node.RequiredFields,
	}
}

func buildOutputSchema(node domain.FixedNodeConfig) map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"summary":         "string",
			"structured_data": "object",
			"decision":        "string",
		},
	}
}

func buildNodeConfig(node domain.FixedNodeConfig) map[string]any {
	return map[string]any{
		"need_review":        node.NeedReview,
		"required_fields":    node.RequiredFields,
		"require_attachment": node.RequireAttachment,
		"default_agent_code": node.DefaultAgentCode,
	}
}

func mustJSON(value any) datatypes.JSON {
	bytes, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return datatypes.JSON(bytes)
}
