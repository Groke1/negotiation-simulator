package main

import (
	"log"

	"go.uber.org/zap"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/app"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/config"
)

func main() {
	logger, err := zap.NewDevelopment()

	if err != nil {
		log.Fatalf("can not initialize logger: %s", err)
	}

	cfg, err := config.New()

	if err != nil {
		log.Fatalf("can not initialize config: %s", err)
	}

	if err := app.Run(logger, cfg); err != nil {
		log.Fatalf("can not initialize app: %s", err)
	}
}
