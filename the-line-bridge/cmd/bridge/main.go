package main

import (
	"fmt"
	"log"
	"os"
	"strings"

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

func parseFlag(arg, prefix string) (string, bool) {
	full := prefix + "="
	if strings.HasPrefix(arg, full) {
		return arg[len(full):], true
	}
	return "", false
}
