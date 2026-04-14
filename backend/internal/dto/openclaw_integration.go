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
