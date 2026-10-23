package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/tpu-orch/avia/internal/app"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/dev.yaml"
	}

	application, err := app.New(configPath)
	if err != nil {
		log.Fatalf("failed to init app: %v", err)
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	log.Println("starting app...")

	if err := application.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Printf("app stopped with error: %v", err)
		os.Exit(1)
	}

	log.Println("app stopped")
}
