package config

import (
	"os"
	"strconv"
)

type Config struct {
	AppPort      string
	GinMode      string
	MySQLDSN     string
	AutoMigrate  bool
	ExecutorMode string
}

func Load() Config {
	return Config{
		AppPort:     getEnv("APP_PORT", "8080"),
		GinMode:     getEnv("GIN_MODE", "debug"),
		MySQLDSN:    getEnv("MYSQL_DSN", "root:root@tcp(127.0.0.1:3306)/the_line?charset=utf8mb4&parseTime=True&loc=Local"),
		AutoMigrate:  getBoolEnv("AUTO_MIGRATE", true),
		ExecutorMode: getEnv("EXECUTOR_MODE", "mock"),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getBoolEnv(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
