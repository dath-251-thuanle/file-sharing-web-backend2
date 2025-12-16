package database

import (
	"log"

	"github.com/golang-migrate/migrate/v4"
)

func RunMigrations(dsn string) {
	m, err := migrate.New(
		"file://internal/infrastructure/database/migrations",
		dsn,
	)
	if err != nil {
		log.Fatal("migration init failed:", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("migration failed:", err)
	}

	log.Println("database migration completed")
}
