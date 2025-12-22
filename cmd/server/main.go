package main

import (
	"log"
	"os"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/config"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/app"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("DATABASE_URL =", os.Getenv("DATABASE_URL"))

	// load .env only in local
	if os.Getenv("RAILWAY_ENVIRONMENT") == "" {
		_ = godotenv.Load()
	}

	cfg := config.NewConfig()
	application := app.NewApplication(cfg)
	application.Run()
}
