# OpenClaw Bridge Implementation Design

## Overview

Implementation spec for connecting OpenClaw (龙虾) to the-line (虾线) via a bridge service. Covers Phase 1 (manual install) and Phase 2 (semi-auto self-onboarding).

Two codebases are modified:
- `backend/` — the-line backend, adds integration management and OpenClaw executors
- `the-line-bridge/` — new standalone Go+Gin service, runs as an independent process

## Architecture

```
      the-line backend            bridge (独立进程)           OpenClaw runtime
   ┌──────────────────┐       ┌──────────────────┐       ┌──────────────────┐
   │                  │       │                  │       │                  │
   │ Integration管理   │       │ 接收执行请求      │       │ chat.send        │
   │ 注册码生成        │       │ 转发到OpenClaw    │──HTTP──│ agent.wait       │
   │ 注册/心跳接收     │       │                  │       │ session管理       │
   │                  │──HTTP──│ /bridge/drafts   │       │                  │
   │ OpenClawPlanner  │       │ /bridge/executions│       └──────────────────┘
   │  Executor        │       │ /bridge/health   │
   │ OpenClawTask     │       │                  │
   │  Executor        │◄─HTTP─│ 注册/心跳/回执    │
   │                  │       │                  │
   │ 回执处理→流程推进  │       │ setup wizard     │
   └──────────────────┘       └──────────────────┘
```

Communication: HTTPS JSON API. Bridge registers with the-line on startup, receives execution requests from the-line, dispatches to OpenClaw runtime via its existing HTTP API, and posts receipts back to the-line.

## Key Design Decisions

1. **Async execution with callback receipts** — the-line POSTs execution requests to bridge, bridge returns `accepted`, then POSTs receipts back to the-line when done.
2. **Bridge as standalone binary** — not embedded in OpenClaw code. OpenClaw AI agent can download, configure, and start it on its own by following a setup document.
3. **OpenClawRuntime interface with mock** — bridge defines a Go interface for OpenClaw runtime calls. Mock implementation lets full flow work without real OpenClaw. Real implementation calls OpenClaw HTTP API.
4. **Phase 1 + Phase 2 scope** — registration, heartbeat, executors, draft generation, task execution, setup wizard, registration code generation.

---

## Part 1: the-line Backend Changes

### 1.1 New Data Models

#### OpenClawIntegration

Represents a registered OpenClaw instance.

```go
type OpenClawIntegration struct {
    ID                  uint64         `gorm:"primaryKey;autoIncrement"`
    DisplayName         string         `gorm:"size:200;not null"`
    Status              string         `gorm:"size:20;not null;default:pending;index"` // pending, active, degraded, disabled, revoked
    BridgeVersion       string         `gorm:"size:50;not null"`
    OpenClawVersion     string         `gorm:"size:50"`
    InstanceFingerprint string         `gorm:"size:100;uniqueIndex"`
    BoundAgentID        uint64         `gorm:"index"`
    CapabilitiesJSON    datatypes.JSON
    CallbackURL         string         `gorm:"size:500"` // bridge's base URL for the-line to call
    CallbackSecret      string         `gorm:"size:200"`
    HeartbeatInterval   int            `gorm:"default:60"` // seconds
    LastHeartbeatAt     *time.Time
    LastErrorMessage    string         `gorm:"size:1000"`
    CreatedAt           time.Time
    UpdatedAt           time.Time
}
```

#### RegistrationCode

One-time codes for bridge registration.

```go
type RegistrationCode struct {
    ID            uint64     `gorm:"primaryKey;autoIncrement"`
    Code          string     `gorm:"size:50;uniqueIndex;not null"` // e.g. TL-ABCD-1234
    Status        string     `gorm:"size:20;not null;default:active;index"` // active, used, expired, revoked
    IntegrationID *uint64    `gorm:"index"` // set when used
    ExpiresAt     time.Time  `gorm:"not null"`
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

#### AgentTask additions

Add external tracking fields to existing `AgentTask` model:

```go
ExternalRuntime    string `gorm:"size:50"`  // "openclaw"
ExternalSessionKey string `gorm:"size:200"`
ExternalRunID      string `gorm:"size:200"`
```

### 1.2 New API Endpoints

All follow existing patterns: handler -> service -> repository.

#### Registration Code Management

```
POST /api/integrations/openclaw/registration-codes
  Request:  { "expires_in_minutes": 30 }
  Response: { "code": "TL-ABCD-1234", "expires_at": "..." }

GET /api/integrations/openclaw/registration-codes
  Response: paginated list of codes with status
```

#### Bridge Registration (called by bridge)

```
POST /api/integrations/openclaw/register
  Request:  {
    "protocol_version": 1,
    "registration_code": "TL-ABCD-1234",
    "bridge_version": "0.1.0",
    "openclaw_version": "2026.4.14",
    "instance_fingerprint": "ocw_abc123",
    "display_name": "Alice's OpenClaw",
    "bound_agent_id": "video-ops",
    "callback_url": "http://bridge-host:9090",
    "capabilities": { "draft_generation": true, "agent_execute": true },
    "idempotency_key": "register:ocw_abc123:TL-ABCD-1234"
  }
  Response: {
    "ok": true,
    "data": {
      "integration_id": 1,
      "status": "active",
      "callback_secret": "cbsec_xxx",
      "heartbeat_interval_seconds": 60,
      "min_supported_bridge_version": "0.1.0"
    }
  }
```

Validates registration code, creates OpenClawIntegration, marks code as used.

#### Heartbeat (called by bridge)

```
POST /api/integrations/openclaw/heartbeat
  Request:  {
    "integration_id": 1,
    "bridge_version": "0.1.0",
    "status": "healthy",
    "active_runs_count": 2
  }
  Response: { "ok": true, "data": { "accepted": true } }
```

#### Integration Management

```
GET    /api/integrations/openclaw              # list integrations
GET    /api/integrations/openclaw/:id          # detail
POST   /api/integrations/openclaw/:id/test     # trigger test-ping to bridge
POST   /api/integrations/openclaw/:id/disable  # disable integration
```

### 1.3 Executor Implementations

#### OpenClawPlannerExecutor

Implements `executor.AgentPlannerExecutor`. Called by `FlowDraftService` when generating drafts.

```go
func (e *OpenClawPlannerExecutor) GenerateDraft(ctx context.Context, prompt string, agent *model.Agent) (*dto.DraftPlan, error) {
    // 1. Find active integration bound to this agent
    // 2. HTTP POST to bridge: {callback_url}/bridge/drafts/generate
    // 3. Parse structured plan from response
    // 4. Return dto.DraftPlan
}
```

#### OpenClawTaskExecutor

Implements `executor.AgentExecutor`. Called by `AgentTaskService` when dispatching auto-node tasks.

```go
func (e *OpenClawTaskExecutor) Execute(ctx context.Context, task *model.AgentTask, agent *model.Agent) error {
    // 1. Find active integration bound to this agent
    // 2. Build execution request with callback URL pointing to the-line's receipt endpoint
    // 3. HTTP POST to bridge: {callback_url}/bridge/executions
    // 4. Update task with external_runtime, external_session_key, external_run_id
    // 5. Return nil (actual completion arrives via receipt callback)
}
```

#### Executor Selection

Add config: `EXECUTOR_MODE=mock|openclaw` (default: `mock`).

In `router.go`, select executor based on config:
```go
if cfg.ExecutorMode == "openclaw" {
    plannerExec = executor.NewOpenClawPlannerExecutor(integrationRepo, httpClient)
    taskExec = executor.NewOpenClawTaskExecutor(integrationRepo, httpClient)
} else {
    plannerExec = executor.NewMockAgentPlannerExecutor()
    taskExec = executor.NewMockAgentExecutor(receiptCallback)
}
```

### 1.4 New Files in backend/

```
internal/model/openclaw_integration.go
internal/model/registration_code.go
internal/repository/openclaw_integration_repository.go
internal/repository/registration_code_repository.go
internal/service/openclaw_integration_service.go
internal/handler/openclaw_integration_handler.go
internal/dto/openclaw_integration.go
internal/executor/openclaw_planner_executor.go
internal/executor/openclaw_task_executor.go
internal/config/config.go  (add ExecutorMode field)
internal/db/migration.go   (add new models to AutoMigrate)
internal/app/router.go     (add new routes and executor selection)
```

---

## Part 2: the-line-bridge New Project

### 2.1 Project Structure

```
the-line-bridge/
├── cmd/bridge/main.go                  # Entry point
├── internal/
│   ├── config/
│   │   └── config.go                   # Env-based configuration
│   ├── app/
│   │   ├── router.go                   # Route registration
│   │   └── server.go                   # Server struct
│   ├── handler/
│   │   ├── draft_handler.go            # POST /bridge/drafts/generate
│   │   ├── execution_handler.go        # POST /bridge/executions, cancel
│   │   ├── health_handler.go           # GET /bridge/health
│   │   └── test_ping_handler.go        # POST /bridge/test-ping
│   ├── service/
│   │   ├── draft_service.go            # Draft generation orchestration
│   │   ├── execution_service.go        # Task execution orchestration
│   │   ├── heartbeat_service.go        # Periodic heartbeat to the-line
│   │   └── setup_service.go            # Setup wizard logic (Phase 2)
│   ├── client/
│   │   └── theline_client.go           # HTTP client for the-line API
│   ├── runtime/
│   │   ├── runtime.go                  # OpenClawRuntime interface
│   │   └── mock_runtime.go             # Mock implementation
│   ├── receipt/
│   │   └── mapper.go                   # OpenClaw result -> receipt mapping
│   └── store/
│       └── config_store.go             # Local config persistence (JSON file)
├── go.mod
└── go.sum
```

### 2.2 Configuration

```go
type Config struct {
    Port             string // default: "9090"
    PlatformURL      string // the-line platform URL
    RegistrationCode string // one-time registration code (for setup)
    OpenClawAPIURL   string // OpenClaw runtime HTTP API URL
    IntegrationID    uint64 // set after registration
    CallbackSecret   string // set after registration
    DataDir          string // local config/state persistence directory
    MockMode         bool   // use mock OpenClaw runtime
}
```

### 2.3 OpenClawRuntime Interface

```go
type OpenClawRuntime interface {
    // PlanDraft sends a planning request and returns structured draft
    PlanDraft(ctx context.Context, req PlanDraftRequest) (*PlanDraftResult, error)

    // ExecuteTask starts a task execution, returns immediately with run reference
    ExecuteTask(ctx context.Context, req ExecuteTaskRequest) (*ExecuteTaskResult, error)

    // WaitForResult blocks until task completes or times out
    WaitForResult(ctx context.Context, sessionKey string) (*TaskResult, error)

    // CancelTask requests cancellation of a running task
    CancelTask(ctx context.Context, sessionKey string) error

    // Health returns OpenClaw runtime health status
    Health(ctx context.Context) (*HealthStatus, error)

    // ListAgents returns available agents/profiles
    ListAgents(ctx context.Context) ([]AgentInfo, error)
}
```

Mock implementation returns realistic Chinese test data with configurable delays.

### 2.4 Bridge HTTP Endpoints

#### POST /bridge/drafts/generate

Receives draft generation request from the-line, dispatches to OpenClaw planner session.

```
Request:  { integration_id, draft_id, planner_agent_id, session_key, source_prompt, constraints }
Response: { ok: true, data: { draft_id, plan: { title, description, nodes [...] }, summary } }
```

Flow:
1. Validate integration_id and signature
2. Call `runtime.PlanDraft()`
3. Format result into protocol-compliant response

#### POST /bridge/executions

Receives execution request from the-line, starts async execution.

```
Request:  { integration_id, agent_task_id, run_id, run_node_id, agent_code, objective, input_json, callback }
Response: { ok: true, data: { accepted: true, external_session_key, external_run_id, status: "running" } }
```

Flow:
1. Validate request
2. Call `runtime.ExecuteTask()` to start execution
3. Return accepted immediately
4. In background goroutine: `runtime.WaitForResult()`, then POST receipt to the-line callback URL

#### POST /bridge/executions/:agentTaskId/cancel

```
Request:  { integration_id, agent_task_id, reason }
Response: { ok: true, data: { accepted: true, status: "cancelling" } }
```

#### GET /bridge/health

```
Response: { ok: true, data: { status: "healthy", bridge_version, supports_protocol_version: 1 } }
```

#### POST /bridge/test-ping

```
Request:  { integration_id, ping_id, kind: "handshake_validation" }
Response: { ok: true, data: { pong: true, ping_id, bridge_version } }
```

### 2.5 the-line API Client

HTTP client for calling the-line endpoints:

```go
type TheLineClient struct {
    baseURL        string
    integrationID  uint64
    callbackSecret string
    httpClient     *http.Client
}

func (c *TheLineClient) Register(req RegisterRequest) (*RegisterResponse, error)
func (c *TheLineClient) Heartbeat(req HeartbeatRequest) error
func (c *TheLineClient) PostReceipt(taskID uint64, receipt ReceiptRequest) error
```

All requests include:
- `X-The-Line-Integration-Id` header
- `X-The-Line-Signature` header (HMAC-SHA256 of timestamp + path + body)
- `X-The-Line-Protocol-Version: 1`

### 2.6 Receipt Mapper

Maps OpenClaw execution results to the-line receipt format:

| OpenClaw Result | Receipt Status |
|----------------|---------------|
| succeeded | completed |
| succeeded + blocked | blocked |
| succeeded + review_needed | needs_review |
| failed | failed |
| timed_out | failed |
| cancelled | cancelled |

### 2.7 Heartbeat Service

Background goroutine that POSTs heartbeat to the-line at configured interval (default 60s). Reports:
- bridge version
- OpenClaw runtime health
- active execution count
- last error

### 2.8 Setup Wizard (Phase 2)

CLI-driven setup flow:

```bash
./the-line-bridge setup \
  --platform-url=https://the-line.example.com \
  --registration-code=TL-ABCD-1234 \
  --openclaw-api=http://localhost:8080
```

Steps:
1. Validate platform URL is reachable
2. Validate OpenClaw API is reachable (or skip in mock mode)
3. List available agents from OpenClaw (or default in mock mode)
4. Prompt user to select agent (or auto-select if only one)
5. Call the-line register endpoint
6. Save integration_id, callback_secret, config to local JSON file
7. Execute test-ping
8. Print success summary

After setup, the bridge can be started with:
```bash
./the-line-bridge serve
```

It reads saved config from `data/bridge-config.json` and starts the HTTP server + heartbeat.

### 2.9 Startup Flow

```
main.go
  ├── "setup" subcommand → run setup wizard, exit
  └── "serve" subcommand → load config, start server
        ├── Load bridge-config.json
        ├── Initialize OpenClawRuntime (mock or real)
        ├── Initialize TheLineClient
        ├── Start HTTP server (Gin)
        ├── Start heartbeat goroutine
        └── Wait for shutdown signal
```

---

## Part 3: Protocol Headers and Auth

### Request Authentication

Bridge -> the-line requests include:
- `X-The-Line-Protocol-Version: 1`
- `X-The-Line-Integration-Id: {integration_id}`
- `X-The-Line-Signature: {hmac_sha256(callback_secret, timestamp + path + body)}`
- `X-The-Line-Timestamp: {unix_timestamp}`

the-line -> bridge requests include:
- `X-The-Line-Protocol-Version: 1`
- `X-The-Line-Integration-Id: {integration_id}`

### Error Response Format

```json
{
  "ok": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "human readable message",
    "retryable": false
  }
}
```

### Success Response Format

```json
{
  "ok": true,
  "data": { ... }
}
```

---

## Part 4: Setup Document for OpenClaw AI Agent

The-line platform generates a setup instruction document that the OpenClaw AI agent can follow:

```markdown
# 接入虾线平台

## 接入信息
- 平台地址: https://the-line.example.com
- 注册码: TL-ABCD-1234 (30分钟内有效)

## 安装步骤
1. 下载 the-line-bridge:
   git clone https://github.com/xxx/the-line-bridge.git
   cd the-line-bridge && go build -o the-line-bridge ./cmd/bridge

2. 运行 setup:
   ./the-line-bridge setup \
     --platform-url=https://the-line.example.com \
     --registration-code=TL-ABCD-1234 \
     --openclaw-api=http://localhost:8080

3. 启动服务:
   ./the-line-bridge serve

4. 验证接入状态:
   curl http://localhost:9090/bridge/health
```

---

## Implementation Order

1. **Backend models and migration** — OpenClawIntegration, RegistrationCode, AgentTask field additions
2. **Backend registration code endpoints** — generate and list codes
3. **Backend register/heartbeat endpoints** — bridge calls these
4. **Backend integration management** — list, detail, disable
5. **Bridge project scaffold** — Go module, config, server, router
6. **Bridge config store** — local JSON persistence
7. **Bridge the-line client** — register, heartbeat, receipt posting
8. **Bridge OpenClawRuntime interface + mock** — interface and mock implementation
9. **Bridge handlers** — drafts, executions, cancel, health, test-ping
10. **Bridge services** — draft, execution, heartbeat
11. **Bridge receipt mapper** — result to receipt conversion
12. **Backend OpenClaw executors** — PlannerExecutor, TaskExecutor
13. **Backend executor selection** — config-based switching
14. **Bridge setup wizard** — CLI setup flow (Phase 2)
15. **Bridge serve command** — full startup with heartbeat
16. **Integration test** — end-to-end flow with mock runtime
