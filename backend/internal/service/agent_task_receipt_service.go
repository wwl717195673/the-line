package service

import (
	"context"
	"encoding/json"
	"errors"

	"the-line/backend/internal/dto"
	"the-line/backend/internal/model"
	"the-line/backend/internal/repository"
	"the-line/backend/internal/response"

	"gorm.io/gorm"
)

type AgentTaskReceiptService struct {
	agentTaskReceiptRepo *repository.AgentTaskReceiptRepository
}

func NewAgentTaskReceiptService(agentTaskReceiptRepo *repository.AgentTaskReceiptRepository) *AgentTaskReceiptService {
	return &AgentTaskReceiptService{agentTaskReceiptRepo: agentTaskReceiptRepo}
}

func (s *AgentTaskReceiptService) GetLatestByTaskID(ctx context.Context, taskID uint64) (dto.AgentTaskReceiptResponse, error) {
	if taskID == 0 {
		return dto.AgentTaskReceiptResponse{}, response.Validation("任务 ID 不合法")
	}

	receipt, err := s.agentTaskReceiptRepo.GetLatestByTaskID(ctx, taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.AgentTaskReceiptResponse{}, response.NotFound("龙虾任务回执不存在")
		}
		return dto.AgentTaskReceiptResponse{}, err
	}
	return toAgentTaskReceiptResponse(*receipt), nil
}

func toAgentTaskReceiptResponse(receipt model.AgentTaskReceipt) dto.AgentTaskReceiptResponse {
	return dto.AgentTaskReceiptResponse{
		ID:            receipt.ID,
		AgentTaskID:   receipt.AgentTaskID,
		RunID:         receipt.RunID,
		RunNodeID:     receipt.RunNodeID,
		AgentID:       receipt.AgentID,
		ReceiptStatus: receipt.ReceiptStatus,
		PayloadJSON:   json.RawMessage(receipt.PayloadJSON),
		ReceivedAt:    receipt.ReceivedAt,
	}
}
