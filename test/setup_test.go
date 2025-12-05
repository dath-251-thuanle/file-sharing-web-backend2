package test

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/config"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/app"
)

var TestApp *app.Application

func TestMain(m *testing.M) {

	os.Setenv("SERVER_PORT", "9999")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5435")
	os.Setenv("DB_NAME", "file-sharing")
	os.Setenv("DB_USER", "haixon")
	os.Setenv("DB_PASSWORD", "123456")
	os.Setenv("DB_SSLMODE", "disable")
	os.Setenv("JWT_SECRET_KEY", "TEST_SECRET_123")

	_, file, _, _ := runtime.Caller(0)
	root := filepath.Dir(filepath.Dir(file))

	initPath := filepath.Join(
		root,
		"internal",
		"infrastructure",
		"database",
		"init.sql",
	)

	cfg := config.NewConfig()
	TestApp = app.NewApplication(cfg)

	if TestApp == nil {
		log.Fatal("Cannot initialize Application")
	}

	sqlDB := TestApp.DB()

	if _, err := sqlDB.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`); err != nil {
		log.Fatal("Failed to enable uuid-ossp:", err)
	}

	initSQL, err := os.ReadFile(initPath)
	if err != nil {
		log.Fatal("Failed to read init.sql:", err)
	}

	if _, err = sqlDB.Exec(string(initSQL)); err != nil {
		log.Fatal("Migration failed:", err)
	}

	os.Exit(m.Run())
}
