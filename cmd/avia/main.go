package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	go func() {
		if err := application.Run(); err != nil {
			log.Fatalf("failed to run app: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down app...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	application.Stop(ctx)
	log.Println("app stopped")
}
