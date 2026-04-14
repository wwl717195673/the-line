package service

import (
	"context"
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
	health, err := s.rt.Health(context.Background())
	if err != nil {
		log.Printf("警告: OpenClaw 运行时不可用: %v", err)
	} else {
		log.Printf("OpenClaw 运行时健康: %s (版本 %s)", health.Status, health.Version)
	}

	// 3. List agents and pick one
	agentID := opts.BoundAgentID
	if agentID == "" {
		agents, err := s.rt.ListAgents(context.Background())
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
