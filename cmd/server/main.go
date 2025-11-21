package main

import (
	"github.com/dath-251-thuanle/file-sharing-web-backend2/config"
	"github.com/dath-251-thuanle/file-sharing-web-backend2/internal/app"
	"github.com/joho/godotenv"
)

func main() {
	// Application entry point

	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file")
	}

	cfg := config.NewConfig()

	application := app.NewApplication(cfg)

	application.Run()

}