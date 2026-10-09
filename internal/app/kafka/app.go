package kafkaapp

import (
	"context"

	deliverykafka "github.com/tpu-orch/avia/internal/delivery/kafka"
	service "github.com/tpu-orch/avia/internal/domain/services/avia"
	"github.com/tpu-orch/avia/internal/infra/observability"
)

type App struct {
	consumer *deliverykafka.Consumer
	producer *deliverykafka.Producer
	handler  *deliverykafka.Handler
	logger   observability.Logger
}

func New(
	cfg Config,
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
	if err != nil && err != context.Canceled {
		return err
	}

	a.logger.Info("kafka app stopped")
	return nil
}

func (a *App) Shutdown() error {
	if a.consumer != nil {
		_ = a.consumer.Close()
	}
	if a.producer != nil {
		_ = a.producer.Close()
	}
	return nil
}
