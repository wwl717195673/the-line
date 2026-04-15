package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"the-line/backend/internal/domain"
	"the-line/backend/internal/dto"
	"the-line/backend/internal/model"
	"the-line/backend/internal/repository"
	"the-line/backend/internal/response"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type OpenClawIntegrationService struct {
	integrationRepo *repository.OpenClawIntegrationRepository
	regCodeRepo     *repository.RegistrationCodeRepository
	agentRepo       *repository.AgentRepository
	agentTaskRepo   *repository.AgentTaskRepository
	httpClient      *http.Client
}

func NewOpenClawIntegrationService(
	integrationRepo *repository.OpenClawIntegrationRepository,
	regCodeRepo *repository.RegistrationCodeRepository,
	agentRepo *repository.AgentRepository,
	agentTaskRepo *repository.AgentTaskRepository,
) *OpenClawIntegrationService {
	return &OpenClawIntegrationService{
		integrationRepo: integrationRepo,
		regCodeRepo:     regCodeRepo,
		agentRepo:       agentRepo,
		agentTaskRepo:   agentTaskRepo,
		httpClient:      &http.Client{Timeout: 10 * time.Second},
	}
}

// --- Registration Code ---

func (s *OpenClawIntegrationService) CreateRegistrationCode(ctx context.Context, req dto.CreateRegistrationCodeRequest) (dto.RegistrationCodeResponse, error) {
	minutes := req.ExpiresInMinutes
	if minutes <= 0 {
		minutes = 30
	}

	code := generateRegCode()
	regCode := &model.RegistrationCode{
		Code:      code,
		Status:    domain.RegCodeStatusActive,
		ExpiresAt: time.Now().Add(time.Duration(minutes) * time.Minute),
	}
	if err := s.regCodeRepo.Create(ctx, regCode); err != nil {
		return dto.RegistrationCodeResponse{}, err
	}
	return toRegCodeResponse(*regCode), nil
}

func (s *OpenClawIntegrationService) ListRegistrationCodes(ctx context.Context, req dto.RegistrationCodeListRequest) ([]dto.RegistrationCodeResponse, int64, dto.PageQuery, error) {
	page := req.PageQuery.Normalize()
	items, total, err := s.regCodeRepo.List(ctx, repository.RegCodeListFilter{
		Status: req.Status,
		Offset: page.Offset(),
		Limit:  page.PageSize,
	})
	if err != nil {
		return nil, 0, page, err
	}

	result := make([]dto.RegistrationCodeResponse, 0, len(items))
	for _, item := range items {
		result = append(result, toRegCodeResponse(item))
	}
	return result, total, page, nil
}

// --- Registration ---

func (s *OpenClawIntegrationService) Register(ctx context.Context, req dto.BridgeRegisterRequest) (dto.BridgeRegisterResponse, error) {
	if req.ProtocolVersion != 1 {
		return dto.BridgeRegisterResponse{}, response.Validation("不支持的协议版本")
	}

	// Check if already registered (idempotent)
	existing, err := s.integrationRepo.GetByFingerprint(ctx, req.InstanceFingerprint)
	if err == nil && existing != nil {
		return dto.BridgeRegisterResponse{
			IntegrationID:             existing.ID,
			Status:                    existing.Status,
			CallbackSecret:            existing.CallbackSecret,
			HeartbeatIntervalSeconds:  existing.HeartbeatInterval,
			MinSupportedBridgeVersion: "0.1.0",
		}, nil
	}

	// Validate registration code
	regCode, err := s.regCodeRepo.GetByCode(ctx, req.RegistrationCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.BridgeRegisterResponse{}, response.Validation("注册码无效")
		}
		return dto.BridgeRegisterResponse{}, err
	}
	if regCode.Status != domain.RegCodeStatusActive {
		return dto.BridgeRegisterResponse{}, response.Validation("注册码已使用或已过期")
	}
	if time.Now().After(regCode.ExpiresAt) {
		return dto.BridgeRegisterResponse{}, response.Validation("注册码已过期")
	}

	// Generate callback secret
	secret := generateSecret()

	// Resolve bound agent ID
	var boundAgentID uint64
	if req.BoundAgentID != "" {
		agent, err := s.agentRepo.GetByCode(ctx, req.BoundAgentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return dto.BridgeRegisterResponse{}, response.Validation("绑定的龙虾编码不存在")
			}
			return dto.BridgeRegisterResponse{}, err
		}
		boundAgentID = agent.ID
	}

	displayName := strings.TrimSpace(req.DisplayName)
	if displayName == "" {
		displayName = "OpenClaw Instance"
	}

	capJSON := datatypes.JSON([]byte("{}"))
	if len(req.Capabilities) > 0 {
		capJSON = datatypes.JSON(req.Capabilities)
	}

	integration := &model.OpenClawIntegration{
		DisplayName:         displayName,
		Status:              domain.IntegrationStatusActive,
		BridgeVersion:       req.BridgeVersion,
		OpenClawVersion:     req.OpenClawVersion,
		InstanceFingerprint: req.InstanceFingerprint,
		BoundAgentID:        boundAgentID,
		CapabilitiesJSON:    capJSON,
		CallbackURL:         strings.TrimRight(req.CallbackURL, "/"),
		CallbackSecret:      secret,
		HeartbeatInterval:   60,
	}
	if err := s.integrationRepo.Create(ctx, integration); err != nil {
		return dto.BridgeRegisterResponse{}, err
	}

	// Mark registration code as used
	regCode.Status = domain.RegCodeStatusUsed
	regCode.IntegrationID = &integration.ID
	_ = s.regCodeRepo.Update(ctx, regCode)

	return dto.BridgeRegisterResponse{
		IntegrationID:             integration.ID,
		Status:                    integration.Status,
		CallbackSecret:            secret,
		HeartbeatIntervalSeconds:  integration.HeartbeatInterval,
		MinSupportedBridgeVersion: "0.1.0",
	}, nil
}

// --- Heartbeat ---

func (s *OpenClawIntegrationService) Heartbeat(ctx context.Context, req dto.BridgeHeartbeatRequest) (dto.BridgeHeartbeatResponse, error) {
	integration, err := s.integrationRepo.GetByID(ctx, req.IntegrationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.BridgeHeartbeatResponse{}, response.NotFound("集成实例不存在")
		}
		return dto.BridgeHeartbeatResponse{}, err
	}

	now := time.Now()
	updates := map[string]any{
		"last_heartbeat_at": &now,
	}
	if req.BridgeVersion != "" {
		updates["bridge_version"] = req.BridgeVersion
	}
	if req.Status == "healthy" && integration.Status == domain.IntegrationStatusDegraded {
		updates["status"] = domain.IntegrationStatusActive
		updates["last_error_message"] = ""
	}
	if req.Status == "degraded" || req.Status == "unavailable" {
		updates["status"] = domain.IntegrationStatusDegraded
	}
	if req.LastError != "" {
		updates["last_error_message"] = req.LastError
	}

	_ = s.integrationRepo.UpdateFields(ctx, integration.ID, updates)

	return dto.BridgeHeartbeatResponse{
		Accepted:                  true,
		IntegrationStatus:         integration.Status,
		MinSupportedBridgeVersion: "0.1.0",
	}, nil
}

// --- Management ---

func (s *OpenClawIntegrationService) List(ctx context.Context, req dto.IntegrationListRequest) ([]dto.IntegrationResponse, int64, dto.PageQuery, error) {
	page := req.PageQuery.Normalize()
	items, total, err := s.integrationRepo.List(ctx, repository.IntegrationListFilter{
		Status: req.Status,
		Offset: page.Offset(),
		Limit:  page.PageSize,
	})
	if err != nil {
		return nil, 0, page, err
	}

	result := make([]dto.IntegrationResponse, 0, len(items))
	for _, item := range items {
		result = append(result, toIntegrationResponse(item))
	}
	return result, total, page, nil
}

func (s *OpenClawIntegrationService) Get(ctx context.Context, id uint64) (dto.IntegrationResponse, error) {
	item, err := s.integrationRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.IntegrationResponse{}, response.NotFound("集成实例不存在")
		}
		return dto.IntegrationResponse{}, err
	}
	return toIntegrationResponse(*item), nil
}

func (s *OpenClawIntegrationService) Disable(ctx context.Context, id uint64) (dto.IntegrationResponse, error) {
	item, err := s.integrationRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.IntegrationResponse{}, response.NotFound("集成实例不存在")
		}
		return dto.IntegrationResponse{}, err
	}
	item.Status = domain.IntegrationStatusDisabled
	if err := s.integrationRepo.Update(ctx, item); err != nil {
		return dto.IntegrationResponse{}, err
	}
	return toIntegrationResponse(*item), nil
}

func (s *OpenClawIntegrationService) TestPing(ctx context.Context, id uint64) (map[string]any, error) {
	item, err := s.integrationRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NotFound("集成实例不存在")
		}
		return nil, err
	}
	if item.CallbackURL == "" {
		return nil, response.Validation("集成实例未配置回调地址")
	}

	url := item.CallbackURL + "/bridge/test-ping"
	reqBody := fmt.Sprintf(`{"protocol_version":1,"integration_id":%d,"ping_id":"ping_%d","kind":"handshake_validation"}`, item.ID, time.Now().UnixMilli())

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return map[string]any{"success": false, "error": err.Error()}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return map[string]any{"success": false, "error": fmt.Sprintf("HTTP %d", resp.StatusCode)}, nil
	}

	return map[string]any{"success": true}, nil
}

// --- Pull Mode (pending tasks) ---

func (s *OpenClawIntegrationService) PendingTasks(ctx context.Context, integrationID uint64) ([]dto.AgentTaskResponse, error) {
	integration, err := s.integrationRepo.GetByID(ctx, integrationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NotFound("集成实例不存在")
		}
		return nil, err
	}
	if integration.Status != domain.IntegrationStatusActive {
		return nil, response.InvalidState("集成实例未激活")
	}

	tasks, _, err := s.agentTaskRepo.List(ctx, repository.AgentTaskListFilter{
		Status: domain.AgentTaskStatusQueued,
		Limit:  20,
	})
	if err != nil {
		return nil, err
	}

	// Filter tasks bound to this integration's agent
	result := make([]dto.AgentTaskResponse, 0)
	for _, t := range tasks {
		if t.AgentID == integration.BoundAgentID {
			result = append(result, dto.AgentTaskResponse{
				ID:            t.ID,
				RunID:         t.RunID,
				RunNodeID:     t.RunNodeID,
				AgentID:       t.AgentID,
				TaskType:      t.TaskType,
				InputJSON:     json.RawMessage(t.InputJSON),
				Status:        t.Status,
				CreatedAt:     t.CreatedAt,
				UpdatedAt:     t.UpdatedAt,
			})
		}
	}
	return result, nil
}

func (s *OpenClawIntegrationService) ClaimTask(ctx context.Context, integrationID uint64, taskID uint64) (dto.AgentTaskResponse, error) {
	integration, err := s.integrationRepo.GetByID(ctx, integrationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.AgentTaskResponse{}, response.NotFound("集成实例不存在")
		}
		return dto.AgentTaskResponse{}, err
	}
	if integration.Status != domain.IntegrationStatusActive {
		return dto.AgentTaskResponse{}, response.InvalidState("集成实例未激活")
	}

	task, err := s.agentTaskRepo.GetByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.AgentTaskResponse{}, response.NotFound("任务不存在")
		}
		return dto.AgentTaskResponse{}, err
	}
	if task.AgentID != integration.BoundAgentID {
		return dto.AgentTaskResponse{}, response.Forbidden("该任务不属于当前集成实例绑定的龙虾")
	}
	if task.Status != domain.AgentTaskStatusQueued {
		return dto.AgentTaskResponse{}, response.InvalidState("任务状态不是待执行")
	}

	now := time.Now()
	task.Status = domain.AgentTaskStatusRunning
	task.StartedAt = &now
	task.ExternalRuntime = "openclaw"
	if err := s.agentTaskRepo.Update(ctx, task); err != nil {
		return dto.AgentTaskResponse{}, err
	}

	return dto.AgentTaskResponse{
		ID:            task.ID,
		RunID:         task.RunID,
		RunNodeID:     task.RunNodeID,
		AgentID:       task.AgentID,
		TaskType:      task.TaskType,
		InputJSON:     json.RawMessage(task.InputJSON),
		Status:        task.Status,
		StartedAt:     task.StartedAt,
		CreatedAt:     task.CreatedAt,
		UpdatedAt:     task.UpdatedAt,
	}, nil
}

// --- Helpers ---

func generateRegCode() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("TL-%s-%s", strings.ToUpper(hex.EncodeToString(b[:2])), strings.ToUpper(hex.EncodeToString(b[2:])))
}

func generateSecret() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return "cbsec_" + hex.EncodeToString(b)
}

func toRegCodeResponse(c model.RegistrationCode) dto.RegistrationCodeResponse {
	return dto.RegistrationCodeResponse{
		ID:        c.ID,
		Code:      c.Code,
		Status:    c.Status,
		ExpiresAt: c.ExpiresAt,
		CreatedAt: c.CreatedAt,
	}
}

func toIntegrationResponse(i model.OpenClawIntegration) dto.IntegrationResponse {
	return dto.IntegrationResponse{
		ID:                  i.ID,
		DisplayName:         i.DisplayName,
		Status:              i.Status,
		BridgeVersion:       i.BridgeVersion,
		OpenClawVersion:     i.OpenClawVersion,
		InstanceFingerprint: i.InstanceFingerprint,
		BoundAgentID:        i.BoundAgentID,
		CapabilitiesJSON:    json.RawMessage(i.CapabilitiesJSON),
		CallbackURL:         i.CallbackURL,
		HeartbeatInterval:   i.HeartbeatInterval,
		LastHeartbeatAt:     i.LastHeartbeatAt,
		LastErrorMessage:    i.LastErrorMessage,
		CreatedAt:           i.CreatedAt,
		UpdatedAt:           i.UpdatedAt,
	}
}
