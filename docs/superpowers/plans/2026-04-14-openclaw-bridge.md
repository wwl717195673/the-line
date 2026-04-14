# OpenClaw Bridge Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Connect OpenClaw (龙虾) to the-line (虾线) via a standalone bridge service, enabling real agent execution and draft generation through a structured protocol.

**Architecture:** Two codebases. `backend/` gets new models (OpenClawIntegration, RegistrationCode), API endpoints (register, heartbeat, integration management), and real executors (OpenClawPlannerExecutor, OpenClawTaskExecutor) that call the bridge via HTTP. `the-line-bridge/` is a new Go+Gin standalone service that receives execution requests from the-line, dispatches to OpenClaw runtime (via interface with mock), and posts receipts back.

**Tech Stack:** Go 1.22, Gin, GORM, MySQL 8, standard lib net/http for HTTP clients

**Spec:** `docs/superpowers/specs/2026-04-14-openclaw-bridge-design.md`

**Existing patterns to follow:**
- Module: `the-line/backend` (Go module name)
- Handler pattern: constructor DI, `c.ShouldBindJSON/Query`, `response.OK/Created/Error/Page`
- Repository pattern: constructor takes `*gorm.DB`, methods take `context.Context`, `WithDB` suffix for transactional variants
- Service pattern: constructor takes repos + other services, returns `*SomeService`
- DTO pattern: request structs with `json`/`form` tags, response structs, `PageQuery` for pagination
- Error pattern: `response.Validation/NotFound/Forbidden/Conflict/InvalidState/Internal`
- Config: env-based with `getEnv/getBoolEnv` helpers
- Domain constants: in `internal/domain/` package

---

## Chunk 1: Backend Data Models and Migration

### Task 1: OpenClawIntegration model

**Files:**
- Create: `backend/internal/model/openclaw_integration.go`
- Create: `backend/internal/domain/openclaw_integration.go`

- [ ] **Step 1: Create domain constants**

Create `backend/internal/domain/openclaw_integration.go`:

```go
package domain

const (
	IntegrationStatusPending  = "pending"
	IntegrationStatusActive   = "active"
	IntegrationStatusDegraded = "degraded"
	IntegrationStatusDisabled = "disabled"
	IntegrationStatusRevoked  = "revoked"
)
```

- [ ] **Step 2: Create OpenClawIntegration model**

Create `backend/internal/model/openclaw_integration.go`:

```go
package model

import (
	"time"

	"gorm.io/datatypes"
)

type OpenClawIntegration struct {
	ID                  uint64         `gorm:"primaryKey" json:"id"`
	DisplayName         string         `gorm:"size:200;not null" json:"display_name"`
	Status              string         `gorm:"size:20;not null;default:pending;index" json:"status"`
	BridgeVersion       string         `gorm:"size:50;not null" json:"bridge_version"`
	OpenClawVersion     string         `gorm:"size:50" json:"openclaw_version"`
	InstanceFingerprint string         `gorm:"size:100;uniqueIndex" json:"instance_fingerprint"`
	BoundAgentID        uint64         `gorm:"index" json:"bound_agent_id"`
	CapabilitiesJSON    datatypes.JSON `gorm:"type:json" json:"capabilities_json"`
	CallbackURL         string         `gorm:"size:500" json:"callback_url"`
	CallbackSecret      string         `gorm:"size:200" json:"-"`
	HeartbeatInterval   int            `gorm:"default:60" json:"heartbeat_interval"`
	LastHeartbeatAt     *time.Time     `json:"last_heartbeat_at"`
	LastErrorMessage    string         `gorm:"size:1000" json:"last_error_message"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
}

func (OpenClawIntegration) TableName() string {
	return "openclaw_integrations"
}
```

- [ ] **Step 3: Verify it compiles**

Run: `cd backend && go build ./internal/model/`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add backend/internal/model/openclaw_integration.go backend/internal/domain/openclaw_integration.go
git commit -m "feat: add OpenClawIntegration model and domain constants"
```

---

### Task 2: RegistrationCode model

**Files:**
- Create: `backend/internal/model/registration_code.go`
- Create: `backend/internal/domain/registration_code.go`

- [ ] **Step 1: Create domain constants**

Create `backend/internal/domain/registration_code.go`:

```go
package domain

const (
	RegCodeStatusActive  = "active"
	RegCodeStatusUsed    = "used"
	RegCodeStatusExpired = "expired"
	RegCodeStatusRevoked = "revoked"
)
```

- [ ] **Step 2: Create RegistrationCode model**

Create `backend/internal/model/registration_code.go`:

```go
package model

import "time"

type RegistrationCode struct {
	ID            uint64    `gorm:"primaryKey" json:"id"`
	Code          string    `gorm:"size:50;uniqueIndex;not null" json:"code"`
	Status        string    `gorm:"size:20;not null;default:active;index" json:"status"`
	IntegrationID *uint64   `gorm:"index" json:"integration_id"`
	ExpiresAt     time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (RegistrationCode) TableName() string {
	return "registration_codes"
}
```

- [ ] **Step 3: Verify it compiles**

Run: `cd backend && go build ./internal/model/`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add backend/internal/model/registration_code.go backend/internal/domain/registration_code.go
git commit -m "feat: add RegistrationCode model and domain constants"
```

---

### Task 3: Add external tracking fields to AgentTask

**Files:**
- Modify: `backend/internal/model/agent_task.go`

- [ ] **Step 1: Add fields to AgentTask**

Add three fields after `ArtifactsJSON` in `backend/internal/model/agent_task.go`:

```go
ExternalRuntime    string `gorm:"size:50" json:"external_runtime"`
ExternalSessionKey string `gorm:"size:200" json:"external_session_key"`
ExternalRunID      string `gorm:"size:200" json:"external_run_id"`
```

- [ ] **Step 2: Verify it compiles**

Run: `cd backend && go build ./internal/model/`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add backend/internal/model/agent_task.go
git commit -m "feat: add external tracking fields to AgentTask for OpenClaw bridge"
```

---

### Task 4: Update migration to include new models

**Files:**
- Modify: `backend/internal/db/migrate.go`

- [ ] **Step 1: Add new models to AutoMigrate**

In `backend/internal/db/migrate.go`, add `&model.OpenClawIntegration{}` and `&model.RegistrationCode{}` to the `AutoMigrate` call:

```go
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
```

- [ ] **Step 2: Verify it compiles**

Run: `cd backend && go build ./...`
Expected: no errors

- [ ] **Step 3: Test migration runs**

Run: `cd backend && go run ./cmd/api` (then Ctrl+C after startup)
Expected: tables `openclaw_integrations` and `registration_codes` are created. `agent_tasks` gets new columns.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/db/migrate.go
git commit -m "feat: add OpenClawIntegration and RegistrationCode to auto-migration"
```

---

### Task 5: Add ExecutorMode to config

**Files:**
- Modify: `backend/internal/config/config.go`

- [ ] **Step 1: Add ExecutorMode field**

Add to `Config` struct and `Load()` in `backend/internal/config/config.go`:

```go
type Config struct {
	AppPort      string
	GinMode      string
	MySQLDSN     string
	AutoMigrate  bool
	ExecutorMode string // "mock" or "openclaw"
}

func Load() Config {
	return Config{
		AppPort:      getEnv("APP_PORT", "8080"),
		GinMode:      getEnv("GIN_MODE", "debug"),
		MySQLDSN:     getEnv("MYSQL_DSN", "root:root@tcp(127.0.0.1:3306)/the_line?charset=utf8mb4&parseTime=True&loc=Local"),
		AutoMigrate:  getBoolEnv("AUTO_MIGRATE", true),
		ExecutorMode: getEnv("EXECUTOR_MODE", "mock"),
	}
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd backend && go build ./...`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add backend/internal/config/config.go
git commit -m "feat: add ExecutorMode config for mock/openclaw executor selection"
```

---

## Chunk 2: Backend DTOs, Repositories, and Service

### Task 6: OpenClaw integration DTOs

**Files:**
- Create: `backend/internal/dto/openclaw_integration.go`

- [ ] **Step 1: Create DTO file**

Create `backend/internal/dto/openclaw_integration.go`:

```go
package dto

import (
	"encoding/json"
	"time"
)

// --- Registration Code ---

type CreateRegistrationCodeRequest struct {
	ExpiresInMinutes int `json:"expires_in_minutes"`
}

type RegistrationCodeResponse struct {
	ID        uint64    `json:"id"`
	Code      string    `json:"code"`
	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type RegistrationCodeListRequest struct {
	PageQuery
	Status string `form:"status"`
}

// --- Bridge Registration ---

type BridgeRegisterRequest struct {
	ProtocolVersion     int             `json:"protocol_version" binding:"required"`
	RegistrationCode    string          `json:"registration_code" binding:"required"`
	BridgeVersion       string          `json:"bridge_version" binding:"required"`
	OpenClawVersion     string          `json:"openclaw_version"`
	InstanceFingerprint string          `json:"instance_fingerprint" binding:"required"`
	DisplayName         string          `json:"display_name"`
	BoundAgentID        string          `json:"bound_agent_id"`
	CallbackURL         string          `json:"callback_url" binding:"required"`
	Capabilities        json.RawMessage `json:"capabilities"`
	IdempotencyKey      string          `json:"idempotency_key"`
}

type BridgeRegisterResponse struct {
	IntegrationID             uint64 `json:"integration_id"`
	Status                    string `json:"status"`
	CallbackSecret            string `json:"callback_secret"`
	HeartbeatIntervalSeconds  int    `json:"heartbeat_interval_seconds"`
	MinSupportedBridgeVersion string `json:"min_supported_bridge_version"`
}

// --- Heartbeat ---

type BridgeHeartbeatRequest struct {
	IntegrationID   uint64 `json:"integration_id" binding:"required"`
	BridgeVersion   string `json:"bridge_version"`
	Status          string `json:"status"`
	ActiveRunsCount int    `json:"active_runs_count"`
	LastError       string `json:"last_error"`
}

type BridgeHeartbeatResponse struct {
	Accepted                  bool   `json:"accepted"`
	IntegrationStatus         string `json:"integration_status"`
	MinSupportedBridgeVersion string `json:"min_supported_bridge_version"`
}

// --- Integration Management ---

type IntegrationListRequest struct {
	PageQuery
	Status string `form:"status"`
}

type IntegrationResponse struct {
	ID                  uint64          `json:"id"`
	DisplayName         string          `json:"display_name"`
	Status              string          `json:"status"`
	BridgeVersion       string          `json:"bridge_version"`
	OpenClawVersion     string          `json:"openclaw_version"`
	InstanceFingerprint string          `json:"instance_fingerprint"`
	BoundAgentID        uint64          `json:"bound_agent_id"`
	CapabilitiesJSON    json.RawMessage `json:"capabilities_json"`
	CallbackURL         string          `json:"callback_url"`
	HeartbeatInterval   int             `json:"heartbeat_interval"`
	LastHeartbeatAt     *time.Time      `json:"last_heartbeat_at"`
	LastErrorMessage    string          `json:"last_error_message"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd backend && go build ./internal/dto/`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add backend/internal/dto/openclaw_integration.go
git commit -m "feat: add OpenClaw integration DTOs"
```

---

### Task 7: OpenClawIntegration repository

**Files:**
- Create: `backend/internal/repository/openclaw_integration_repository.go`

- [ ] **Step 1: Create repository**

Create `backend/internal/repository/openclaw_integration_repository.go`:

```go
package repository

import (
	"context"

	"the-line/backend/internal/model"

	"gorm.io/gorm"
)

type IntegrationListFilter struct {
	Status string
	Offset int
	Limit  int
}

type OpenClawIntegrationRepository struct {
	db *gorm.DB
}

func NewOpenClawIntegrationRepository(database *gorm.DB) *OpenClawIntegrationRepository {
	return &OpenClawIntegrationRepository{db: database}
}

func (r *OpenClawIntegrationRepository) Create(ctx context.Context, integration *model.OpenClawIntegration) error {
	return r.db.WithContext(ctx).Create(integration).Error
}

func (r *OpenClawIntegrationRepository) GetByID(ctx context.Context, id uint64) (*model.OpenClawIntegration, error) {
	var item model.OpenClawIntegration
	if err := r.db.WithContext(ctx).First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *OpenClawIntegrationRepository) GetByFingerprint(ctx context.Context, fingerprint string) (*model.OpenClawIntegration, error) {
	var item model.OpenClawIntegration
	if err := r.db.WithContext(ctx).Where("instance_fingerprint = ?", fingerprint).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *OpenClawIntegrationRepository) GetActiveByAgentID(ctx context.Context, agentID uint64) (*model.OpenClawIntegration, error) {
	var item model.OpenClawIntegration
	if err := r.db.WithContext(ctx).Where("bound_agent_id = ? AND status = ?", agentID, "active").First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *OpenClawIntegrationRepository) List(ctx context.Context, filter IntegrationListFilter) ([]model.OpenClawIntegration, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.OpenClawIntegration{})
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []model.OpenClawIntegration
	if err := query.Order("created_at DESC").Offset(filter.Offset).Limit(filter.Limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *OpenClawIntegrationRepository) Update(ctx context.Context, integration *model.OpenClawIntegration) error {
	return r.db.WithContext(ctx).Save(integration).Error
}

func (r *OpenClawIntegrationRepository) UpdateFields(ctx context.Context, id uint64, updates map[string]any) error {
	return r.db.WithContext(ctx).Model(&model.OpenClawIntegration{}).Where("id = ?", id).Updates(updates).Error
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd backend && go build ./internal/repository/`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add backend/internal/repository/openclaw_integration_repository.go
git commit -m "feat: add OpenClawIntegration repository"
```

---

### Task 8: RegistrationCode repository

**Files:**
- Create: `backend/internal/repository/registration_code_repository.go`

- [ ] **Step 1: Create repository**

Create `backend/internal/repository/registration_code_repository.go`:

```go
package repository

import (
	"context"

	"the-line/backend/internal/model"

	"gorm.io/gorm"
)

type RegCodeListFilter struct {
	Status string
	Offset int
	Limit  int
}

type RegistrationCodeRepository struct {
	db *gorm.DB
}

func NewRegistrationCodeRepository(database *gorm.DB) *RegistrationCodeRepository {
	return &RegistrationCodeRepository{db: database}
}

func (r *RegistrationCodeRepository) Create(ctx context.Context, code *model.RegistrationCode) error {
	return r.db.WithContext(ctx).Create(code).Error
}

func (r *RegistrationCodeRepository) GetByCode(ctx context.Context, code string) (*model.RegistrationCode, error) {
	var item model.RegistrationCode
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *RegistrationCodeRepository) List(ctx context.Context, filter RegCodeListFilter) ([]model.RegistrationCode, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.RegistrationCode{})
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []model.RegistrationCode
	if err := query.Order("created_at DESC").Offset(filter.Offset).Limit(filter.Limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *RegistrationCodeRepository) Update(ctx context.Context, code *model.RegistrationCode) error {
	return r.db.WithContext(ctx).Save(code).Error
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd backend && go build ./internal/repository/`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add backend/internal/repository/registration_code_repository.go
git commit -m "feat: add RegistrationCode repository"
```

---

### Task 9: OpenClaw integration service

**Files:**
- Create: `backend/internal/service/openclaw_integration_service.go`

- [ ] **Step 1: Create service**

Create `backend/internal/service/openclaw_integration_service.go`:

```go
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
	httpClient      *http.Client
}

func NewOpenClawIntegrationService(
	integrationRepo *repository.OpenClawIntegrationRepository,
	regCodeRepo *repository.RegistrationCodeRepository,
	agentRepo *repository.AgentRepository,
) *OpenClawIntegrationService {
	return &OpenClawIntegrationService{
		integrationRepo: integrationRepo,
		regCodeRepo:     regCodeRepo,
		agentRepo:       agentRepo,
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
```

- [ ] **Step 2: Verify it compiles**

Run: `cd backend && go build ./internal/service/`

Note: This file uses `s.agentRepo.GetByCode(ctx, code)`. If `AgentRepository` doesn't have a `GetByCode` method, add it:

```go
func (r *AgentRepository) GetByCode(ctx context.Context, code string) (*model.Agent, error) {
	var agent model.Agent
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&agent).Error; err != nil {
		return nil, err
	}
	return &agent, nil
}
```

Add this to `backend/internal/repository/agent_repository.go` if missing.

Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add backend/internal/service/openclaw_integration_service.go backend/internal/repository/agent_repository.go
git commit -m "feat: add OpenClaw integration service with registration, heartbeat, and management"
```

---

### Task 10: OpenClaw integration handler

**Files:**
- Create: `backend/internal/handler/openclaw_integration_handler.go`

- [ ] **Step 1: Create handler**

Create `backend/internal/handler/openclaw_integration_handler.go`:

```go
package handler

import (
	"the-line/backend/internal/dto"
	"the-line/backend/internal/response"
	"the-line/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type OpenClawIntegrationHandler struct {
	integrationService *service.OpenClawIntegrationService
}

func NewOpenClawIntegrationHandler(integrationService *service.OpenClawIntegrationService) *OpenClawIntegrationHandler {
	return &OpenClawIntegrationHandler{integrationService: integrationService}
}

func (h *OpenClawIntegrationHandler) CreateRegistrationCode(c *gin.Context) {
	var req dto.CreateRegistrationCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.Validation("注册码参数不合法"))
		return
	}

	code, err := h.integrationService.CreateRegistrationCode(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, code)
}

func (h *OpenClawIntegrationHandler) ListRegistrationCodes(c *gin.Context) {
	var req dto.RegistrationCodeListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, response.Validation("查询参数不合法"))
		return
	}

	items, total, page, err := h.integrationService.ListRegistrationCodes(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, items, total, page.Page, page.PageSize)
}

func (h *OpenClawIntegrationHandler) Register(c *gin.Context) {
	var req dto.BridgeRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.Validation("注册参数不合法"))
		return
	}

	result, err := h.integrationService.Register(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *OpenClawIntegrationHandler) Heartbeat(c *gin.Context) {
	var req dto.BridgeHeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.Validation("心跳参数不合法"))
		return
	}

	result, err := h.integrationService.Heartbeat(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *OpenClawIntegrationHandler) List(c *gin.Context) {
	var req dto.IntegrationListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, response.Validation("查询参数不合法"))
		return
	}

	items, total, page, err := h.integrationService.List(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Page(c, items, total, page.Page, page.PageSize)
}

func (h *OpenClawIntegrationHandler) Detail(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	result, err := h.integrationService.Get(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *OpenClawIntegrationHandler) Disable(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	result, err := h.integrationService.Disable(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *OpenClawIntegrationHandler) TestPing(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	result, err := h.integrationService.TestPing(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd backend && go build ./internal/handler/`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add backend/internal/handler/openclaw_integration_handler.go
git commit -m "feat: add OpenClaw integration handler"
```

---

### Task 11: Wire up new routes in router.go

**Files:**
- Modify: `backend/internal/app/router.go`

- [ ] **Step 1: Add repositories, service, handler, and routes**

In `backend/internal/app/router.go`, add the new repo/service/handler instantiation after the existing ones, and add a route group. Add to the import the `config` package.

After `agentTaskReceiptRepo` line, add:
```go
integrationRepo := repository.NewOpenClawIntegrationRepository(database)
regCodeRepo := repository.NewRegistrationCodeRepository(database)
```

After `activityService` line, add:
```go
integrationService := service.NewOpenClawIntegrationService(integrationRepo, regCodeRepo, agentRepo)
```

After `healthHandler` line, add:
```go
integrationHandler := handler.NewOpenClawIntegrationHandler(integrationService)
```

Before the `return router` line, inside the `api` group, add:
```go
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
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd backend && go build ./...`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add backend/internal/app/router.go
git commit -m "feat: wire OpenClaw integration routes into router"
```

---

### Task 12: OpenClaw executors (PlannerExecutor + TaskExecutor)

**Files:**
- Create: `backend/internal/executor/openclaw_planner_executor.go`
- Create: `backend/internal/executor/openclaw_task_executor.go`

- [ ] **Step 1: Create OpenClawPlannerExecutor**

Create `backend/internal/executor/openclaw_planner_executor.go`:

```go
package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"the-line/backend/internal/dto"
	"the-line/backend/internal/model"
	"the-line/backend/internal/repository"
	"the-line/backend/internal/response"
)

type OpenClawPlannerExecutor struct {
	integrationRepo *repository.OpenClawIntegrationRepository
	httpClient      *http.Client
}

func NewOpenClawPlannerExecutor(integrationRepo *repository.OpenClawIntegrationRepository) *OpenClawPlannerExecutor {
	return &OpenClawPlannerExecutor{
		integrationRepo: integrationRepo,
		httpClient:      &http.Client{Timeout: 120 * time.Second},
	}
}

func (e *OpenClawPlannerExecutor) GenerateDraft(ctx context.Context, prompt string, agent *model.Agent) (*dto.DraftPlan, error) {
	integration, err := e.integrationRepo.GetActiveByAgentID(ctx, agent.ID)
	if err != nil {
		return nil, response.Validation("未找到可用的 OpenClaw 集成实例")
	}

	reqBody := map[string]any{
		"protocol_version": 1,
		"integration_id":   integration.ID,
		"planner_agent_id": agent.Code,
		"source_prompt":    prompt,
		"constraints": map[string]any{
			"must_end_with_human_acceptance": true,
			"allowed_node_types":            []string{"human_input", "human_review", "agent_execute", "agent_export", "human_acceptance"},
		},
		"output_schema_version": "v1",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	url := integration.CallbackURL + "/bridge/drafts/generate"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := e.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("调用 bridge 草案生成失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bridge 草案生成返回 HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var bridgeResp struct {
		OK   bool `json:"ok"`
		Data struct {
			Plan dto.DraftPlan `json:"plan"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &bridgeResp); err != nil {
		return nil, fmt.Errorf("解析 bridge 草案响应失败: %w", err)
	}
	if !bridgeResp.OK {
		return nil, fmt.Errorf("bridge 草案生成失败")
	}

	return &bridgeResp.Data.Plan, nil
}
```

- [ ] **Step 2: Create OpenClawTaskExecutor**

Create `backend/internal/executor/openclaw_task_executor.go`:

```go
package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"the-line/backend/internal/model"
	"the-line/backend/internal/repository"
	"the-line/backend/internal/response"
)

type OpenClawTaskExecutor struct {
	integrationRepo *repository.OpenClawIntegrationRepository
	httpClient      *http.Client
	receiptURL      string // the-line's base URL for receipt callbacks
}

func NewOpenClawTaskExecutor(integrationRepo *repository.OpenClawIntegrationRepository, receiptBaseURL string) *OpenClawTaskExecutor {
	return &OpenClawTaskExecutor{
		integrationRepo: integrationRepo,
		httpClient:      &http.Client{Timeout: 30 * time.Second},
		receiptURL:      receiptBaseURL,
	}
}

func (e *OpenClawTaskExecutor) Execute(ctx context.Context, task *model.AgentTask, agent *model.Agent) error {
	integration, err := e.integrationRepo.GetActiveByAgentID(ctx, agent.ID)
	if err != nil {
		return response.Validation("未找到可用的 OpenClaw 集成实例")
	}

	sessionKey := fmt.Sprintf("theline:run:%d:node:%d", task.RunID, task.RunNodeID)
	callbackURL := fmt.Sprintf("%s/api/agent-tasks/%d/receipt", e.receiptURL, task.ID)

	reqBody := map[string]any{
		"protocol_version": 1,
		"integration_id":   integration.ID,
		"agent_task_id":    task.ID,
		"run_id":           task.RunID,
		"run_node_id":      task.RunNodeID,
		"agent_code":       agent.Code,
		"node_type":        "agent_execute",
		"session_key":      sessionKey,
		"objective":        "",
		"input_json":       json.RawMessage(task.InputJSON),
		"callback": map[string]any{
			"url":                callbackURL,
			"auth_type":          "signature",
			"callback_secret_ref": integration.CallbackSecret,
		},
		"idempotency_key": fmt.Sprintf("agent_task:%d", task.ID),
	}
	bodyBytes, _ := json.Marshal(reqBody)

	url := integration.CallbackURL + "/bridge/executions"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := e.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("调用 bridge 执行失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bridge 执行返回 HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var bridgeResp struct {
		OK   bool `json:"ok"`
		Data struct {
			ExternalSessionKey string `json:"external_session_key"`
			ExternalRunID      string `json:"external_run_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &bridgeResp); err != nil {
		return fmt.Errorf("解析 bridge 执行响应失败: %w", err)
	}

	// Store external references on task (caller should save)
	task.ExternalRuntime = "openclaw"
	task.ExternalSessionKey = bridgeResp.Data.ExternalSessionKey
	task.ExternalRunID = bridgeResp.Data.ExternalRunID

	return nil
}
```

- [ ] **Step 3: Verify it compiles**

Run: `cd backend && go build ./internal/executor/`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add backend/internal/executor/openclaw_planner_executor.go backend/internal/executor/openclaw_task_executor.go
git commit -m "feat: add OpenClaw planner and task executors"
```

---

### Task 13: Add executor selection to router.go

**Files:**
- Modify: `backend/internal/app/router.go`

- [ ] **Step 1: Add config parameter and executor switching**

Change `NewRouter` signature to accept config:
```go
func NewRouter(cfg config.Config, database *gorm.DB) *gin.Engine {
```

Replace the mock executor lines with conditional logic:

```go
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
agentTaskService.SetExecutor(agentExec)
```

Update `server.go` to pass config:
```go
func NewServer(cfg config.Config, database *gorm.DB) *Server {
	gin.SetMode(cfg.GinMode)
	return &Server{
		cfg:    cfg,
		engine: NewRouter(cfg, database),
	}
}
```

Add `"the-line/backend/internal/config"` to imports of `router.go`.

- [ ] **Step 2: Verify it compiles**

Run: `cd backend && go build ./...`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add backend/internal/app/router.go backend/internal/app/server.go
git commit -m "feat: add config-based executor selection (mock/openclaw)"
```

---

## Chunk 3: Bridge Project Scaffold

### Task 14: Initialize the-line-bridge Go project

**Files:**
- Create: `the-line-bridge/go.mod`
- Create: `the-line-bridge/cmd/bridge/main.go`
- Create: `the-line-bridge/internal/config/config.go`

- [ ] **Step 1: Initialize Go module**

Run:
```bash
mkdir -p the-line-bridge/cmd/bridge
cd the-line-bridge && go mod init the-line-bridge
```

- [ ] **Step 2: Create config**

Create `the-line-bridge/internal/config/config.go`:

```go
package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port           string
	PlatformURL    string
	OpenClawAPIURL string
	DataDir        string
	MockMode       bool
}

func Load() Config {
	return Config{
		Port:           getEnv("BRIDGE_PORT", "9090"),
		PlatformURL:    getEnv("PLATFORM_URL", "http://localhost:8080"),
		OpenClawAPIURL: getEnv("OPENCLAW_API_URL", "http://localhost:8081"),
		DataDir:        getEnv("DATA_DIR", "data"),
		MockMode:       getBoolEnv("MOCK_MODE", true),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getBoolEnv(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return parsed
}
```

- [ ] **Step 3: Create main.go**

Create `the-line-bridge/cmd/bridge/main.go`:

```go
package main

import (
	"fmt"
	"log"
	"os"

	"the-line-bridge/internal/config"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: the-line-bridge <setup|serve>")
		os.Exit(1)
	}

	cfg := config.Load()

	switch os.Args[1] {
	case "setup":
		log.Println("setup wizard not yet implemented")
	case "serve":
		log.Printf("the-line-bridge %s starting on :%s (mock=%v)", version, cfg.Port, cfg.MockMode)
		log.Println("serve not yet implemented")
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
```

- [ ] **Step 4: Verify it compiles**

Run: `cd the-line-bridge && go build ./cmd/bridge/`
Expected: produces `bridge` binary with no errors

- [ ] **Step 5: Commit**

```bash
git add the-line-bridge/
git commit -m "feat: scaffold the-line-bridge Go project"
```

---

### Task 15: Bridge response helpers and config store

**Files:**
- Create: `the-line-bridge/internal/response/response.go`
- Create: `the-line-bridge/internal/store/config_store.go`

- [ ] **Step 1: Create response helpers**

Create `the-line-bridge/internal/response/response.go`:

```go
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
```

- [ ] **Step 2: Create config store**

Create `the-line-bridge/internal/store/config_store.go`:

```go
package store

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type BridgeConfig struct {
	PlatformURL    string `json:"platform_url"`
	IntegrationID  uint64 `json:"integration_id"`
	CallbackSecret string `json:"callback_secret"`
	BoundAgentID   string `json:"bound_agent_id"`
	OpenClawAPIURL string `json:"openclaw_api_url"`
	BridgeVersion  string `json:"bridge_version"`
	HeartbeatSec   int    `json:"heartbeat_interval_seconds"`
}

type ConfigStore struct {
	dir string
}

func NewConfigStore(dir string) *ConfigStore {
	return &ConfigStore{dir: dir}
}

func (s *ConfigStore) Save(cfg BridgeConfig) error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.dir, "bridge-config.json"), data, 0o644)
}

func (s *ConfigStore) Load() (*BridgeConfig, error) {
	data, err := os.ReadFile(filepath.Join(s.dir, "bridge-config.json"))
	if err != nil {
		return nil, err
	}
	var cfg BridgeConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (s *ConfigStore) Exists() bool {
	_, err := os.Stat(filepath.Join(s.dir, "bridge-config.json"))
	return err == nil
}
```

- [ ] **Step 3: Add gin dependency**

Run: `cd the-line-bridge && go get github.com/gin-gonic/gin`

- [ ] **Step 4: Verify it compiles**

Run: `cd the-line-bridge && go build ./...`
Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add the-line-bridge/
git commit -m "feat: add bridge response helpers and config store"
```

---

### Task 16: OpenClawRuntime interface and mock implementation

**Files:**
- Create: `the-line-bridge/internal/runtime/runtime.go`
- Create: `the-line-bridge/internal/runtime/mock_runtime.go`

- [ ] **Step 1: Create interface**

Create `the-line-bridge/internal/runtime/runtime.go`:

```go
package runtime

import (
	"context"
	"encoding/json"
)

type PlanDraftRequest struct {
	SessionKey   string          `json:"session_key"`
	AgentID      string          `json:"agent_id"`
	SourcePrompt string          `json:"source_prompt"`
	Constraints  json.RawMessage `json:"constraints"`
}

type PlanDraftResult struct {
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Nodes       json.RawMessage `json:"nodes"`
	Summary     string          `json:"summary"`
}

type ExecuteTaskRequest struct {
	SessionKey string          `json:"session_key"`
	AgentCode  string          `json:"agent_code"`
	Objective  string          `json:"objective"`
	InputJSON  json.RawMessage `json:"input_json"`
}

type ExecuteTaskResult struct {
	ExternalRunID string `json:"external_run_id"`
}

type TaskResult struct {
	Status       string          `json:"status"` // succeeded, failed, timed_out, cancelled, blocked
	Summary      string          `json:"summary"`
	Result       json.RawMessage `json:"result"`
	Artifacts    json.RawMessage `json:"artifacts"`
	Logs         []string        `json:"logs"`
	ErrorMessage string          `json:"error_message"`
}

type HealthStatus struct {
	Status  string `json:"status"` // healthy, degraded, unavailable
	Version string `json:"version"`
}

type AgentInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type OpenClawRuntime interface {
	PlanDraft(ctx context.Context, req PlanDraftRequest) (*PlanDraftResult, error)
	ExecuteTask(ctx context.Context, req ExecuteTaskRequest) (*ExecuteTaskResult, error)
	WaitForResult(ctx context.Context, sessionKey string) (*TaskResult, error)
	CancelTask(ctx context.Context, sessionKey string) error
	Health(ctx context.Context) (*HealthStatus, error)
	ListAgents(ctx context.Context) ([]AgentInfo, error)
}
```

- [ ] **Step 2: Create mock implementation**

Create `the-line-bridge/internal/runtime/mock_runtime.go`:

```go
package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type MockRuntime struct{}

func NewMockRuntime() *MockRuntime {
	return &MockRuntime{}
}

func (m *MockRuntime) PlanDraft(ctx context.Context, req PlanDraftRequest) (*PlanDraftResult, error) {
	time.Sleep(500 * time.Millisecond)

	nodes := []map[string]any{
		{"node_code": "collect_data", "node_name": "收集业务数据", "node_type": "agent_execute", "sort_order": 1, "executor_type": "agent", "owner_rule": "initiator", "executor_agent_code": "data_query_agent", "result_owner_rule": "initiator", "task_type": "query", "completion_condition": "汇总待处理业务数据", "failure_condition": "查询失败", "escalation_rule": "通知发起人"},
		{"node_code": "review_data", "node_name": "审核确认数据", "node_type": "human_review", "sort_order": 2, "executor_type": "human", "owner_rule": "initiator", "result_owner_rule": "initiator", "completion_condition": "人工审核通过", "failure_condition": "审核驳回", "escalation_rule": "重新处理"},
		{"node_code": "execute_task", "node_name": "执行批量操作", "node_type": "agent_execute", "sort_order": 3, "executor_type": "agent", "owner_rule": "initiator", "executor_agent_code": "operation_agent", "result_owner_rule": "initiator", "task_type": "batch_operation", "completion_condition": "完成批量执行", "failure_condition": "执行失败", "escalation_rule": "通知发起人"},
		{"node_code": "final_acceptance", "node_name": "确认最终结果", "node_type": "human_acceptance", "sort_order": 4, "executor_type": "human", "owner_rule": "initiator", "result_owner_rule": "initiator", "completion_condition": "最终签收人确认结果", "failure_condition": "签收拒绝", "escalation_rule": "跟进修复"},
	}
	nodesJSON, _ := json.Marshal(nodes)

	return &PlanDraftResult{
		Title:       "AI 编排工作流",
		Description: "由龙虾根据自然语言需求生成的流程草案",
		Nodes:       nodesJSON,
		Summary:     "已生成一条 4 节点流程草案",
	}, nil
}

func (m *MockRuntime) ExecuteTask(ctx context.Context, req ExecuteTaskRequest) (*ExecuteTaskResult, error) {
	return &ExecuteTaskResult{
		ExternalRunID: fmt.Sprintf("mock_run_%d", time.Now().UnixMilli()),
	}, nil
}

func (m *MockRuntime) WaitForResult(ctx context.Context, sessionKey string) (*TaskResult, error) {
	time.Sleep(1 * time.Second)

	result, _ := json.Marshal(map[string]any{
		"success_count": 12,
		"failed_count":  0,
		"details":       []map[string]any{},
	})

	return &TaskResult{
		Status:  "succeeded",
		Summary: "已完成批量执行，共处理 12 条记录",
		Result:  result,
		Logs: []string{
			"开始执行任务",
			"逐条处理输入记录",
			"写回执行结果",
		},
	}, nil
}

func (m *MockRuntime) CancelTask(ctx context.Context, sessionKey string) error {
	return nil
}

func (m *MockRuntime) Health(ctx context.Context) (*HealthStatus, error) {
	return &HealthStatus{Status: "healthy", Version: "mock-1.0"}, nil
}

func (m *MockRuntime) ListAgents(ctx context.Context) ([]AgentInfo, error) {
	return []AgentInfo{
		{ID: "default-agent", Name: "默认执行龙虾"},
	}, nil
}
```

- [ ] **Step 3: Verify it compiles**

Run: `cd the-line-bridge && go build ./internal/runtime/`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add the-line-bridge/internal/runtime/
git commit -m "feat: add OpenClawRuntime interface and mock implementation"
```

---

### Task 17: the-line API client

**Files:**
- Create: `the-line-bridge/internal/client/theline_client.go`

- [ ] **Step 1: Create the-line client**

Create `the-line-bridge/internal/client/theline_client.go`:

```go
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type TheLineClient struct {
	baseURL        string
	integrationID  uint64
	callbackSecret string
	httpClient     *http.Client
}

func NewTheLineClient(baseURL string) *TheLineClient {
	return &TheLineClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *TheLineClient) SetCredentials(integrationID uint64, callbackSecret string) {
	c.integrationID = integrationID
	c.callbackSecret = callbackSecret
}

type RegisterRequest struct {
	ProtocolVersion     int             `json:"protocol_version"`
	RegistrationCode    string          `json:"registration_code"`
	BridgeVersion       string          `json:"bridge_version"`
	OpenClawVersion     string          `json:"openclaw_version"`
	InstanceFingerprint string          `json:"instance_fingerprint"`
	DisplayName         string          `json:"display_name"`
	BoundAgentID        string          `json:"bound_agent_id"`
	CallbackURL         string          `json:"callback_url"`
	Capabilities        map[string]bool `json:"capabilities"`
	IdempotencyKey      string          `json:"idempotency_key"`
}

type RegisterResponse struct {
	IntegrationID            uint64 `json:"integration_id"`
	Status                   string `json:"status"`
	CallbackSecret           string `json:"callback_secret"`
	HeartbeatIntervalSeconds int    `json:"heartbeat_interval_seconds"`
}

func (c *TheLineClient) Register(req RegisterRequest) (*RegisterResponse, error) {
	body, _ := json.Marshal(req)
	resp, err := c.doPost("/api/integrations/openclaw/register", body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data RegisterResponse `json:"data"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("解析注册响应失败: %w", err)
	}
	return &result.Data, nil
}

type HeartbeatRequest struct {
	IntegrationID   uint64 `json:"integration_id"`
	BridgeVersion   string `json:"bridge_version"`
	Status          string `json:"status"`
	ActiveRunsCount int    `json:"active_runs_count"`
	LastError       string `json:"last_error"`
}

func (c *TheLineClient) Heartbeat(req HeartbeatRequest) error {
	body, _ := json.Marshal(req)
	_, err := c.doPost("/api/integrations/openclaw/heartbeat", body)
	return err
}

type ReceiptRequest struct {
	ProtocolVersion int             `json:"protocol_version"`
	IntegrationID   uint64          `json:"integration_id"`
	AgentID         uint64          `json:"agent_id"`
	Status          string          `json:"status"`
	StartedAt       *time.Time      `json:"started_at"`
	FinishedAt      *time.Time      `json:"finished_at"`
	Summary         string          `json:"summary"`
	Result          json.RawMessage `json:"result"`
	Artifacts       json.RawMessage `json:"artifacts"`
	Logs            []string        `json:"logs"`
	ErrorMessage    string          `json:"error_message"`
}

func (c *TheLineClient) PostReceipt(taskID uint64, receipt ReceiptRequest) error {
	body, _ := json.Marshal(receipt)
	path := fmt.Sprintf("/api/agent-tasks/%d/receipt", taskID)
	_, err := c.doPost(path, body)
	return err
}

func (c *TheLineClient) doPost(path string, body []byte) ([]byte, error) {
	url := c.baseURL + path
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-The-Line-Protocol-Version", "1")
	if c.integrationID > 0 {
		req.Header.Set("X-The-Line-Integration-Id", fmt.Sprintf("%d", c.integrationID))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 the-line 失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("the-line 返回 HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	return respBody, nil
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd the-line-bridge && go build ./internal/client/`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add the-line-bridge/internal/client/
git commit -m "feat: add the-line API client for bridge"
```

---

### Task 18: Receipt mapper

**Files:**
- Create: `the-line-bridge/internal/receipt/mapper.go`

- [ ] **Step 1: Create receipt mapper**

Create `the-line-bridge/internal/receipt/mapper.go`:

```go
package receipt

import (
	"the-line-bridge/internal/client"
	"the-line-bridge/internal/runtime"
	"time"
)

func MapToReceipt(integrationID uint64, agentID uint64, result *runtime.TaskResult, startedAt time.Time) client.ReceiptRequest {
	finishedAt := time.Now()

	status := mapStatus(result.Status)

	return client.ReceiptRequest{
		ProtocolVersion: 1,
		IntegrationID:   integrationID,
		AgentID:         agentID,
		Status:          status,
		StartedAt:       &startedAt,
		FinishedAt:      &finishedAt,
		Summary:         result.Summary,
		Result:          result.Result,
		Artifacts:       result.Artifacts,
		Logs:            result.Logs,
		ErrorMessage:    result.ErrorMessage,
	}
}

func mapStatus(openclawStatus string) string {
	switch openclawStatus {
	case "succeeded":
		return "completed"
	case "blocked":
		return "blocked"
	case "review_needed":
		return "needs_review"
	case "failed":
		return "failed"
	case "timed_out":
		return "failed"
	case "cancelled":
		return "cancelled"
	default:
		return "failed"
	}
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd the-line-bridge && go build ./internal/receipt/`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add the-line-bridge/internal/receipt/
git commit -m "feat: add receipt mapper for OpenClaw to the-line status mapping"
```

---

## Chunk 4: Bridge Handlers, Services, and Server

### Task 19: Bridge handlers

**Files:**
- Create: `the-line-bridge/internal/handler/draft_handler.go`
- Create: `the-line-bridge/internal/handler/execution_handler.go`
- Create: `the-line-bridge/internal/handler/health_handler.go`
- Create: `the-line-bridge/internal/handler/test_ping_handler.go`

- [ ] **Step 1: Create draft handler**

Create `the-line-bridge/internal/handler/draft_handler.go`:

```go
package handler

import (
	"encoding/json"

	"the-line-bridge/internal/response"
	"the-line-bridge/internal/runtime"

	"github.com/gin-gonic/gin"
)

type DraftHandler struct {
	rt runtime.OpenClawRuntime
}

func NewDraftHandler(rt runtime.OpenClawRuntime) *DraftHandler {
	return &DraftHandler{rt: rt}
}

type DraftGenerateRequest struct {
	ProtocolVersion int             `json:"protocol_version"`
	IntegrationID   uint64          `json:"integration_id"`
	DraftID         uint64          `json:"draft_id"`
	PlannerAgentID  string          `json:"planner_agent_id"`
	SessionKey      string          `json:"session_key"`
	SourcePrompt    string          `json:"source_prompt"`
	Constraints     json.RawMessage `json:"constraints"`
	IdempotencyKey  string          `json:"idempotency_key"`
}

func (h *DraftHandler) Generate(c *gin.Context) {
	var req DraftGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "INVALID_REQUEST", "草案生成请求参数不合法", false)
		return
	}

	result, err := h.rt.PlanDraft(c.Request.Context(), runtime.PlanDraftRequest{
		SessionKey:   req.SessionKey,
		AgentID:      req.PlannerAgentID,
		SourcePrompt: req.SourcePrompt,
		Constraints:  req.Constraints,
	})
	if err != nil {
		response.Error(c, "PLANNER_EXECUTION_FAILED", err.Error(), true)
		return
	}

	var nodes json.RawMessage
	if result.Nodes != nil {
		nodes = result.Nodes
	} else {
		nodes = json.RawMessage("[]")
	}

	response.OK(c, gin.H{
		"draft_id": req.DraftID,
		"plan": gin.H{
			"title":       result.Title,
			"description": result.Description,
			"nodes":       json.RawMessage(nodes),
		},
		"summary": result.Summary,
	})
}
```

- [ ] **Step 2: Create execution handler**

Create `the-line-bridge/internal/handler/execution_handler.go`:

```go
package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"the-line-bridge/internal/client"
	"the-line-bridge/internal/receipt"
	"the-line-bridge/internal/response"
	"the-line-bridge/internal/runtime"

	"github.com/gin-gonic/gin"
)

type ExecutionHandler struct {
	rt     runtime.OpenClawRuntime
	client *client.TheLineClient
}

func NewExecutionHandler(rt runtime.OpenClawRuntime, client *client.TheLineClient) *ExecutionHandler {
	return &ExecutionHandler{rt: rt, client: client}
}

type ExecutionRequest struct {
	ProtocolVersion int             `json:"protocol_version"`
	IntegrationID   uint64          `json:"integration_id"`
	AgentTaskID     uint64          `json:"agent_task_id"`
	RunID           uint64          `json:"run_id"`
	RunNodeID       uint64          `json:"run_node_id"`
	AgentID         uint64          `json:"agent_id"`
	AgentCode       string          `json:"agent_code"`
	NodeType        string          `json:"node_type"`
	SessionKey      string          `json:"session_key"`
	Objective       string          `json:"objective"`
	InputJSON       json.RawMessage `json:"input_json"`
	Callback        *CallbackInfo   `json:"callback"`
	IdempotencyKey  string          `json:"idempotency_key"`
}

type CallbackInfo struct {
	URL              string `json:"url"`
	AuthType         string `json:"auth_type"`
	CallbackSecretRef string `json:"callback_secret_ref"`
}

type CancelRequest struct {
	ProtocolVersion int    `json:"protocol_version"`
	IntegrationID   uint64 `json:"integration_id"`
	AgentTaskID     uint64 `json:"agent_task_id"`
	Reason          string `json:"reason"`
}

func (h *ExecutionHandler) Execute(c *gin.Context) {
	var req ExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "INVALID_REQUEST", "执行请求参数不合法", false)
		return
	}

	sessionKey := req.SessionKey
	if sessionKey == "" {
		sessionKey = fmt.Sprintf("theline:run:%d:node:%d", req.RunID, req.RunNodeID)
	}

	result, err := h.rt.ExecuteTask(c.Request.Context(), runtime.ExecuteTaskRequest{
		SessionKey: sessionKey,
		AgentCode:  req.AgentCode,
		Objective:  req.Objective,
		InputJSON:  req.InputJSON,
	})
	if err != nil {
		response.Error(c, "EXECUTION_FAILED", err.Error(), true)
		return
	}

	// Respond immediately with accepted
	response.OK(c, gin.H{
		"accepted":             true,
		"agent_task_id":        req.AgentTaskID,
		"external_session_key": sessionKey,
		"external_run_id":      result.ExternalRunID,
		"status":               "running",
	})

	// Background: wait for result and post receipt
	go func() {
		startedAt := time.Now()
		taskResult, err := h.rt.WaitForResult(context.Background(), sessionKey)
		if err != nil {
			log.Printf("WaitForResult failed for task %d: %v", req.AgentTaskID, err)
			taskResult = &runtime.TaskResult{
				Status:       "failed",
				Summary:      "执行等待失败",
				ErrorMessage: err.Error(),
			}
		}

		receiptReq := receipt.MapToReceipt(req.IntegrationID, req.AgentID, taskResult, startedAt)
		if err := h.client.PostReceipt(req.AgentTaskID, receiptReq); err != nil {
			log.Printf("PostReceipt failed for task %d: %v", req.AgentTaskID, err)
		}
	}()
}

func (h *ExecutionHandler) Cancel(c *gin.Context) {
	var req CancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "INVALID_REQUEST", "取消请求参数不合法", false)
		return
	}

	sessionKey := fmt.Sprintf("agent_task:%d", req.AgentTaskID)
	if err := h.rt.CancelTask(c.Request.Context(), sessionKey); err != nil {
		response.Error(c, "CANCEL_FAILED", err.Error(), true)
		return
	}

	response.OK(c, gin.H{
		"accepted":      true,
		"agent_task_id": req.AgentTaskID,
		"status":        "cancelling",
	})
}
```

- [ ] **Step 3: Create health handler**

Create `the-line-bridge/internal/handler/health_handler.go`:

```go
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
		"status":                     status,
		"bridge_version":             bridgeVersion,
		"openclaw_version":           ocVersion,
		"supports_protocol_version":  1,
	})
}
```

- [ ] **Step 4: Create test ping handler**

Create `the-line-bridge/internal/handler/test_ping_handler.go`:

```go
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
```

- [ ] **Step 5: Verify it compiles**

Run: `cd the-line-bridge && go build ./internal/handler/`
Expected: no errors

- [ ] **Step 6: Commit**

```bash
git add the-line-bridge/internal/handler/
git commit -m "feat: add bridge HTTP handlers (draft, execution, health, test-ping)"
```

---

### Task 20: Bridge heartbeat service

**Files:**
- Create: `the-line-bridge/internal/service/heartbeat_service.go`

- [ ] **Step 1: Create heartbeat service**

Create `the-line-bridge/internal/service/heartbeat_service.go`:

```go
package service

import (
	"context"
	"log"
	"sync"
	"time"

	"the-line-bridge/internal/client"
	"the-line-bridge/internal/runtime"
)

type HeartbeatService struct {
	client        *client.TheLineClient
	rt            runtime.OpenClawRuntime
	integrationID uint64
	interval      time.Duration
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

func NewHeartbeatService(c *client.TheLineClient, rt runtime.OpenClawRuntime, integrationID uint64, intervalSec int) *HeartbeatService {
	if intervalSec <= 0 {
		intervalSec = 60
	}
	return &HeartbeatService{
		client:        c,
		rt:            rt,
		integrationID: integrationID,
		interval:      time.Duration(intervalSec) * time.Second,
	}
}

func (s *HeartbeatService) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		s.sendHeartbeat()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.sendHeartbeat()
			}
		}
	}()
}

func (s *HeartbeatService) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
}

func (s *HeartbeatService) sendHeartbeat() {
	health, err := s.rt.Health(context.Background())
	status := "healthy"
	lastError := ""
	if err != nil {
		status = "degraded"
		lastError = err.Error()
	} else {
		status = health.Status
	}

	err = s.client.Heartbeat(client.HeartbeatRequest{
		IntegrationID: s.integrationID,
		BridgeVersion: "0.1.0",
		Status:        status,
		LastError:     lastError,
	})
	if err != nil {
		log.Printf("heartbeat failed: %v", err)
	}
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd the-line-bridge && go build ./internal/service/`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add the-line-bridge/internal/service/
git commit -m "feat: add bridge heartbeat service"
```

---

### Task 21: Bridge router and server

**Files:**
- Create: `the-line-bridge/internal/app/router.go`
- Create: `the-line-bridge/internal/app/server.go`

- [ ] **Step 1: Create router**

Create `the-line-bridge/internal/app/router.go`:

```go
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
```

- [ ] **Step 2: Create server**

Create `the-line-bridge/internal/app/server.go`:

```go
package app

import (
	"the-line-bridge/internal/client"
	"the-line-bridge/internal/config"
	"the-line-bridge/internal/runtime"

	"github.com/gin-gonic/gin"
)

type Server struct {
	cfg    config.Config
	engine *gin.Engine
}

func NewServer(cfg config.Config, rt runtime.OpenClawRuntime, thelineClient *client.TheLineClient) *Server {
	return &Server{
		cfg:    cfg,
		engine: NewRouter(rt, thelineClient),
	}
}

func (s *Server) Run() error {
	return s.engine.Run(":" + s.cfg.Port)
}
```

- [ ] **Step 3: Verify it compiles**

Run: `cd the-line-bridge && go build ./internal/app/`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add the-line-bridge/internal/app/
git commit -m "feat: add bridge router and server"
```

---

### Task 22: Wire up main.go serve command

**Files:**
- Modify: `the-line-bridge/cmd/bridge/main.go`

- [ ] **Step 1: Implement serve command**

Replace `the-line-bridge/cmd/bridge/main.go` with:

```go
package main

import (
	"fmt"
	"log"
	"os"

	"the-line-bridge/internal/app"
	"the-line-bridge/internal/client"
	"the-line-bridge/internal/config"
	"the-line-bridge/internal/runtime"
	"the-line-bridge/internal/service"
	"the-line-bridge/internal/store"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: the-line-bridge <setup|serve>")
		os.Exit(1)
	}

	cfg := config.Load()

	switch os.Args[1] {
	case "setup":
		runSetup(cfg)
	case "serve":
		runServe(cfg)
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func runSetup(cfg config.Config) {
	log.Println("setup wizard not yet implemented")
}

func runServe(cfg config.Config) {
	configStore := store.NewConfigStore(cfg.DataDir)

	var rt runtime.OpenClawRuntime
	if cfg.MockMode {
		rt = runtime.NewMockRuntime()
		log.Println("Using mock OpenClaw runtime")
	} else {
		log.Fatal("Real OpenClaw runtime not yet implemented. Set MOCK_MODE=true.")
	}

	thelineClient := client.NewTheLineClient(cfg.PlatformURL)

	// Load saved config if exists (from prior setup)
	if configStore.Exists() {
		savedCfg, err := configStore.Load()
		if err != nil {
			log.Fatalf("Failed to load bridge config: %v", err)
		}
		thelineClient.SetCredentials(savedCfg.IntegrationID, savedCfg.CallbackSecret)

		// Start heartbeat
		heartbeatSvc := service.NewHeartbeatService(thelineClient, rt, savedCfg.IntegrationID, savedCfg.HeartbeatSec)
		heartbeatSvc.Start()
		defer heartbeatSvc.Stop()

		log.Printf("Registered as integration %d, heartbeat every %ds", savedCfg.IntegrationID, savedCfg.HeartbeatSec)
	} else {
		log.Println("No bridge config found. Run 'setup' first, or bridge will run without registration.")
	}

	server := app.NewServer(cfg, rt, thelineClient)
	log.Printf("the-line-bridge %s starting on :%s", version, cfg.Port)
	if err := server.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd the-line-bridge && go build ./cmd/bridge/`
Expected: no errors

- [ ] **Step 3: Test it starts**

Run: `cd the-line-bridge && MOCK_MODE=true ./bridge serve` (Ctrl+C after startup)
Expected: prints startup log, starts HTTP server on :9090

- [ ] **Step 4: Commit**

```bash
git add the-line-bridge/cmd/bridge/main.go
git commit -m "feat: implement bridge serve command with mock runtime and heartbeat"
```

---

## Chunk 5: Setup Wizard and Integration Test

### Task 23: Setup wizard (Phase 2)

**Files:**
- Create: `the-line-bridge/internal/service/setup_service.go`
- Modify: `the-line-bridge/cmd/bridge/main.go`

- [ ] **Step 1: Create setup service**

Create `the-line-bridge/internal/service/setup_service.go`:

```go
package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"time"

	"the-line-bridge/internal/client"
	"the-line-bridge/internal/config"
	"the-line-bridge/internal/runtime"
	"the-line-bridge/internal/store"
)

type SetupService struct {
	cfg   config.Config
	rt    runtime.OpenClawRuntime
	store *store.ConfigStore
}

func NewSetupService(cfg config.Config, rt runtime.OpenClawRuntime, configStore *store.ConfigStore) *SetupService {
	return &SetupService{cfg: cfg, rt: rt, store: configStore}
}

type SetupOptions struct {
	PlatformURL      string
	RegistrationCode string
	OpenClawAPIURL   string
	BoundAgentID     string
	DisplayName      string
}

func (s *SetupService) Run(opts SetupOptions) error {
	log.Println("=== the-line bridge setup wizard ===")

	// 1. Validate platform URL
	log.Printf("检查平台地址: %s", opts.PlatformURL)
	if err := checkURL(opts.PlatformURL + "/api/healthz"); err != nil {
		return fmt.Errorf("无法访问平台: %w", err)
	}
	log.Println("平台可达")

	// 2. Check OpenClaw runtime health
	log.Println("检查 OpenClaw 运行时...")
	health, err := s.rt.Health(nil)
	if err != nil {
		log.Printf("警告: OpenClaw 运行时不可用: %v", err)
	} else {
		log.Printf("OpenClaw 运行时健康: %s (版本 %s)", health.Status, health.Version)
	}

	// 3. List agents and pick one
	agentID := opts.BoundAgentID
	if agentID == "" {
		agents, err := s.rt.ListAgents(nil)
		if err != nil || len(agents) == 0 {
			agentID = "default-agent"
			log.Printf("使用默认 agent: %s", agentID)
		} else if len(agents) == 1 {
			agentID = agents[0].ID
			log.Printf("自动选择唯一 agent: %s (%s)", agentID, agents[0].Name)
		} else {
			agentID = agents[0].ID
			log.Printf("自动选择第一个 agent: %s (%s)", agentID, agents[0].Name)
		}
	}

	// 4. Generate fingerprint
	fingerprint := generateFingerprint()

	// 5. Display name
	displayName := opts.DisplayName
	if displayName == "" {
		displayName = "OpenClaw Bridge Instance"
	}

	// 6. Register with the-line
	log.Println("向平台注册...")
	thelineClient := client.NewTheLineClient(opts.PlatformURL)

	bridgeCallbackURL := fmt.Sprintf("http://localhost:%s", s.cfg.Port)

	regResp, err := thelineClient.Register(client.RegisterRequest{
		ProtocolVersion:     1,
		RegistrationCode:    opts.RegistrationCode,
		BridgeVersion:       "0.1.0",
		InstanceFingerprint: fingerprint,
		DisplayName:         displayName,
		BoundAgentID:        agentID,
		CallbackURL:         bridgeCallbackURL,
		Capabilities:        map[string]bool{"draft_generation": true, "agent_execute": true},
		IdempotencyKey:      fmt.Sprintf("register:%s:%s", fingerprint, opts.RegistrationCode),
	})
	if err != nil {
		return fmt.Errorf("注册失败: %w", err)
	}
	log.Printf("注册成功! Integration ID: %d", regResp.IntegrationID)

	// 7. Save config
	bridgeCfg := store.BridgeConfig{
		PlatformURL:    opts.PlatformURL,
		IntegrationID:  regResp.IntegrationID,
		CallbackSecret: regResp.CallbackSecret,
		BoundAgentID:   agentID,
		OpenClawAPIURL: opts.OpenClawAPIURL,
		BridgeVersion:  "0.1.0",
		HeartbeatSec:   regResp.HeartbeatIntervalSeconds,
	}
	if err := s.store.Save(bridgeCfg); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}
	log.Println("配置已保存")

	// 8. Summary
	log.Println("=== 接入完成 ===")
	log.Printf("  平台地址:     %s", opts.PlatformURL)
	log.Printf("  Integration:  %d", regResp.IntegrationID)
	log.Printf("  绑定 Agent:   %s", agentID)
	log.Printf("  Bridge 版本:  0.1.0")
	log.Println("")
	log.Println("运行 'the-line-bridge serve' 启动服务")

	return nil
}

func checkURL(url string) error {
	c := &http.Client{Timeout: 10 * time.Second}
	resp, err := c.Get(url)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func generateFingerprint() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "ocw_" + hex.EncodeToString(b)
}
```

- [ ] **Step 2: Wire setup into main.go**

Replace `runSetup` function in `the-line-bridge/cmd/bridge/main.go`:

```go
func runSetup(cfg config.Config) {
	var rt runtime.OpenClawRuntime
	if cfg.MockMode {
		rt = runtime.NewMockRuntime()
	} else {
		log.Fatal("Real OpenClaw runtime not yet implemented. Set MOCK_MODE=true.")
	}

	configStore := store.NewConfigStore(cfg.DataDir)
	setupSvc := service.NewSetupService(cfg, rt, configStore)

	opts := service.SetupOptions{
		PlatformURL:    cfg.PlatformURL,
		OpenClawAPIURL: cfg.OpenClawAPIURL,
	}

	// Parse CLI flags
	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]
		if val, ok := parseFlag(arg, "--platform-url"); ok {
			opts.PlatformURL = val
		}
		if val, ok := parseFlag(arg, "--registration-code"); ok {
			opts.RegistrationCode = val
		}
		if val, ok := parseFlag(arg, "--openclaw-api"); ok {
			opts.OpenClawAPIURL = val
		}
		if val, ok := parseFlag(arg, "--agent-id"); ok {
			opts.BoundAgentID = val
		}
		if val, ok := parseFlag(arg, "--display-name"); ok {
			opts.DisplayName = val
		}
	}

	if opts.PlatformURL == "" {
		log.Fatal("必须提供 --platform-url")
	}
	if opts.RegistrationCode == "" {
		log.Fatal("必须提供 --registration-code")
	}

	if err := setupSvc.Run(opts); err != nil {
		log.Fatalf("Setup failed: %v", err)
	}
}

func parseFlag(arg, prefix string) (string, bool) {
	full := prefix + "="
	if len(arg) > len(full) && arg[:len(full)] == full {
		return arg[len(full):], true
	}
	return "", false
}
```

Add `"the-line-bridge/internal/runtime"` and `"the-line-bridge/internal/service"` and `"the-line-bridge/internal/store"` to imports if not already present.

- [ ] **Step 3: Verify it compiles**

Run: `cd the-line-bridge && go build ./cmd/bridge/`
Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add the-line-bridge/internal/service/setup_service.go the-line-bridge/cmd/bridge/main.go
git commit -m "feat: implement bridge setup wizard with CLI flags"
```

---

### Task 24: End-to-end smoke test

This is a manual integration test to verify the full flow works.

- [ ] **Step 1: Start the-line backend**

Run: `cd backend && go run ./cmd/api`
Expected: server starts on :8080

- [ ] **Step 2: Create a registration code**

Run:
```bash
curl -s -X POST http://localhost:8080/api/integrations/openclaw/registration-codes \
  -H 'Content-Type: application/json' \
  -d '{"expires_in_minutes": 30}'
```
Expected: returns JSON with `code` like `TL-XXXX-XXXX`

- [ ] **Step 3: Run bridge setup**

Run (replace TL-XXXX-XXXX with actual code):
```bash
cd the-line-bridge && MOCK_MODE=true go run ./cmd/bridge setup \
  --platform-url=http://localhost:8080 \
  --registration-code=TL-XXXX-XXXX
```
Expected: prints "注册成功" and "接入完成", creates `data/bridge-config.json`

- [ ] **Step 4: Start bridge**

Run: `cd the-line-bridge && MOCK_MODE=true go run ./cmd/bridge serve`
Expected: server starts on :9090, heartbeat starts

- [ ] **Step 5: Test health endpoint**

Run: `curl -s http://localhost:9090/bridge/health | python3 -m json.tool`
Expected: returns `{"ok": true, "data": {"status": "healthy", "bridge_version": "0.1.0", ...}}`

- [ ] **Step 6: Test draft generation**

Run:
```bash
curl -s -X POST http://localhost:9090/bridge/drafts/generate \
  -H 'Content-Type: application/json' \
  -d '{"protocol_version":1,"integration_id":1,"draft_id":1,"source_prompt":"创建视频绑定流程"}' | python3 -m json.tool
```
Expected: returns structured draft plan with nodes

- [ ] **Step 7: Check integration list on the-line**

Run: `curl -s http://localhost:8080/api/integrations/openclaw | python3 -m json.tool`
Expected: shows the registered integration with `status: "active"` and recent heartbeat

- [ ] **Step 8: Commit all remaining changes**

```bash
git add -A
git commit -m "feat: complete OpenClaw bridge Phase 1+2 implementation"
```
