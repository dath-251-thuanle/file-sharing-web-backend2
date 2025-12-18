package test

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/config"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/app"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

var TestApp *app.Application
var TestDB *sql.DB

func TestMain(m *testing.M) {
	setupEnv()

	dbURL := os.Getenv("DATABASE_URL")

	var err error
	TestDB, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Cannot connect to DB: %v", err)
	}

	// verify connection
	if err := TestDB.Ping(); err != nil {
		log.Fatalf("Cannot ping DB: %v", err)
	}

	setupTestSchema(TestDB)
	runMigration(TestDB)

	cfg := config.NewConfig()
	TestApp = app.NewApplication(cfg)
	if TestApp == nil {
		log.Fatal("Cannot initialize TestApp")
	}

	exitCode := m.Run()
	TestDB.Close()
	os.Exit(exitCode)
}

func setupEnv() {
	if os.Getenv("DATABASE_URL") == "" {
		//FOR LOCAL USE: you need to change the username (postgres) and password of your psql
		os.Setenv("DATABASE_URL",
			"postgres://postgres:password@localhost:5432/file_sharing_test?sslmode=disable&options=-csearch_path=test_schema")
	}
	if os.Getenv("SERVER_PORT") == "" {
		os.Setenv("SERVER_PORT", "9999")
	}

	if os.Getenv("JWT_SECRET_KEY") == "" {
		os.Setenv("JWT_SECRET_KEY", "test_secret_key")
	}
}

func setupTestSchema(db *sql.DB) {
	if _, err := db.Exec(`CREATE SCHEMA IF NOT EXISTS test_schema`); err != nil {
		log.Fatalf("Create schema failed: %v", err)
	}
}

func runMigration(db *sql.DB) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Create migrate driver failed: %v", err)
	}

	path, err := filepath.Abs("../internal/infrastructure/database")
	if err != nil {
		log.Fatal(err)
	}

	migrationPath := "file://" + filepath.ToSlash(path)

	m, err := migrate.NewWithDatabaseInstance(
		migrationPath,
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatalf("Migration init error: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration failed: %v", err)
	}
}
