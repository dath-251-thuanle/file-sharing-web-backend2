package test

import (
	"log"
	"os"
	"testing"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/config"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/app"
)

var TestApp *app.Application

func TestMain(m *testing.M) {

	// ==== ENV for TEST ====
	os.Setenv("SERVER_PORT", "9999")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5435")
	os.Setenv("DB_NAME", "file-sharing")
	os.Setenv("DB_USER", "haixon")
	os.Setenv("DB_PASSWORD", "123456")
	os.Setenv("DB_SSLMODE", "disable")
	os.Setenv("JWT_SECRET_KEY", "TEST_SECRET_123")

	cfg := config.NewConfig()

	TestApp = app.NewApplication(cfg)

	if TestApp == nil {
		log.Fatal("Cannot initialize Application")
	}

	code := m.Run()
	os.Exit(code)
}
