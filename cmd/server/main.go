package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/dath-251-thuanle/file-sharing-web-backend2/config"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/app"
	"github.com/joho/godotenv"
)

func main() {
	// Application entry point

	cwd, err := os.Getwd()

	if err != nil {
		log.Fatal(" Unable to get working dir:", err)
	}

	envPath := filepath.Join(cwd, ".env")

	if err := godotenv.Load(envPath); err != nil {
		panic("Error loading .env file")
	}

	cfg := config.NewConfig()

	application := app.NewApplication(cfg)

	application.Run()

}
