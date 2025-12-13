package test

import (
	"log"
	"os"
	"testing"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/config"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/app"
)

var TestApp *app.Application

func setEnvIfEmpty(key, value string) {
	if os.Getenv(key) == "" {
		os.Setenv(key, value)
	}
}

func TestMain(m *testing.M) {

	// ===== FIX ENV FOR TESTING =====
	setEnvIfEmpty("SERVER_PORT", "9999")
	setEnvIfEmpty("DB_NAME", "testdb")
	setEnvIfEmpty("DB_USER", "ci")
	setEnvIfEmpty("DB_PASSWORD", "ci")
	setEnvIfEmpty("DB_HOST", "localhost")
	setEnvIfEmpty("DB_PORT", "5435")
	setEnvIfEmpty("DB_SSLMODE", "disable")
	setEnvIfEmpty("JWT_SECRET_KEY", "ci_secret")

	cfg := config.NewConfig()

	TestApp = app.NewApplication(cfg)
	if TestApp == nil {
		log.Fatal("Cannot initialize Application")
	}

	exitCode := m.Run()
	os.Exit(exitCode)
}
