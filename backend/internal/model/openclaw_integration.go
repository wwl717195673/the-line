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
