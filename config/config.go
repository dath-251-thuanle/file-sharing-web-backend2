package config

import (
	"fmt"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/pkg/utils"
)

type SystemPolicy struct {
	MaxFileSizeMB            int
	MinValidityHours         int
	MaxValidityDays          int
	DefaultValidityDays      int
	RequirePasswordMinLength int
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type Config struct {
	DB            DatabaseConfig
	ServerAddress string
	Policy        *SystemPolicy
}

func NewConfig() *Config {
	return &Config{
		ServerAddress: fmt.Sprintf(":%s", utils.GetEnv("SERVER_PORT", "8080")),
		DB: DatabaseConfig{
			Host:     utils.GetEnv("DB_HOST", "localhost"),
			Port:     utils.GetEnv("DB_PORT", "5432"),
			User:     utils.GetEnv("DB_USER", "postgres"),
			Password: utils.GetEnv("DB_PASSWORD", "postgres"),
			DBName:   utils.GetEnv("DB_NAME", "myapp"),
			SSLMode:  utils.GetEnv("DB_SSLMODE", "disable"),
		},
		Policy: &SystemPolicy{
			MaxFileSizeMB:            50,
			MinValidityHours:         1,
			MaxValidityDays:          30,
			DefaultValidityDays:      7,
			RequirePasswordMinLength: 6,
		},
	}
}

func (c *Config) DNS() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", c.DB.Host, c.DB.Port, c.DB.User, c.DB.Password, c.DB.DBName, c.DB.SSLMode)
}
