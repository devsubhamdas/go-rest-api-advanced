package main

import (
	"log"

	"github.com/Subham-Das-98/go-rest-api-advanced/internal/application"
	"github.com/Subham-Das-98/go-rest-api-advanced/internal/platform/config"
)

func main() {
	cfg := config.MustLoad()

	app, err := application.New(cfg)

	if err != nil {
		log.Fatalf("failed to init app: %v", err)
	}

	if err := app.Run(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
