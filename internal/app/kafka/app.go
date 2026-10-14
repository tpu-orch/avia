package kafkaapp

import (
	"context"
	"errors"
	"fmt"

	deliverykafka "github.com/tpu-orch/avia/internal/delivery/kafka"
	service "github.com/tpu-orch/avia/internal/domain/services/avia"
	"github.com/tpu-orch/avia/internal/infra/config"
	"github.com/tpu-orch/avia/internal/infra/observability"
)

type App struct {
	consumer *deliverykafka.Consumer
	producer *deliverykafka.Producer
	handler  *deliverykafka.Handler
	logger   observability.Logger
}

func New(
	cfg config.KafkaConfig,
	cmdService *service.TicketCommandService,
	logger observability.Logger,
) *App {
	producer := deliverykafka.NewProducer(cfg.Brokers, cfg.ReservationStatusTopic)
	handler := deliverykafka.NewHandler(cmdService, producer)
	consumer := deliverykafka.NewConsumer(
		cfg.Brokers,
		cfg.ReserveTopic,
		cfg.CancelTopic,
		cfg.GroupID,
		handler,
	)

	return &App{
		consumer: consumer,
		producer: producer,
		handler:  handler,
		logger:   logger,
	}
}

func (a *App) Run(ctx context.Context) error {
	a.logger.Info("kafka app started")

	err := a.consumer.Run(ctx)
	if err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("run kafka consumer: %w", err)
	}

	a.logger.Info("kafka app stopped")
	return nil
}

func (a *App) Shutdown() error {
	var shutdownErr error

	if a.consumer != nil {
		if err := a.consumer.Close(); err != nil {
			shutdownErr = errors.Join(
				shutdownErr,
				fmt.Errorf("close kafka consumer: %w", err),
			)
		}
	}

	if a.producer != nil {
		if err := a.producer.Close(); err != nil {
			shutdownErr = errors.Join(
				shutdownErr,
				fmt.Errorf("close kafka producer: %w", err),
			)
		}
	}

	return shutdownErr
}
