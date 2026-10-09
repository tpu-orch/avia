package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/tpu-orch/avia/internal/delivery/kafka/dto"
)

type Producer struct {
	writer *kafkago.Writer
	topic  string
}

func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		writer: &kafkago.Writer{
			Addr:     kafkago.TCP(brokers...),
			Balancer: &kafkago.Hash{},
		},
		topic: topic,
	}
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

func (p *Producer) SendReservationStatusChanged(ctx context.Context, payload dto.ReservationStatusChangedPayload) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	return p.writer.WriteMessages(ctx, kafkago.Message{
		Topic: p.topic,
		Value: b,
		Key:   []byte(fmt.Sprintf("%d", payload.CompositeBookingID)),
	})
}
