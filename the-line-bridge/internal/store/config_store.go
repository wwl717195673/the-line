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
