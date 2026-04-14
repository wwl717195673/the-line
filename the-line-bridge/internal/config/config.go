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
