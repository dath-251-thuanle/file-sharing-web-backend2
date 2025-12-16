package config

import (
	"fmt"
	"strings"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/pkg/utils"
)

type SystemPolicy struct {
	MaxFileSizeMB            int
	MinValidityHours         int
	MaxValidityDays          int
	DefaultValidityDays      int
	RequirePasswordMinLength int
}

type CORSConfig struct {
	AllowedOrigins []string
}

type Config struct {
	ServerAddress string
	DatabaseURL   string
	Policy        *SystemPolicy
	CORS          CORSConfig
}

func NewConfig() *Config {
	dbURL := utils.GetEnv("DATABASE_URL", "")
	if dbURL == "" {
		panic("DATABASE_URL is required")
	}

	return &Config{
		ServerAddress: fmt.Sprintf(":%s", utils.GetEnv("SERVER_PORT", "8080")),
		DatabaseURL:   dbURL,
		CORS:          loadCORSConfig(),
		Policy: &SystemPolicy{
			MaxFileSizeMB:            50,
			MinValidityHours:         1,
			MaxValidityDays:          30,
			DefaultValidityDays:      7,
			RequirePasswordMinLength: 6,
		},
	}
}

func (c *Config) DSN() string {
	return c.DatabaseURL
}

func loadCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: splitAndTrim(
			utils.GetEnv("CORS_ALLOWED_ORIGINS", "*"),
		),
	}
}

func splitAndTrim(s string) []string {
	if s == "" {
		return nil
	}

	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
