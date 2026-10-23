package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	kafkago "github.com/segmentio/kafka-go"
)

type CommandHandler interface {
	ReserveTicket(ctx context.Context, ticketID int64, compositeBookingID int64) error
	CancelTicket(ctx context.Context, ticketID int64, compositeBookingID int64) error
}

type Consumer struct {
	reserveReader *kafkago.Reader
	cancelReader  *kafkago.Reader
	handler       CommandHandler
}

func NewConsumer(brokers []string, reserveTopic, cancelTopic, groupID string, handler CommandHandler) *Consumer {
	return &Consumer{
		reserveReader: kafkago.NewReader(kafkago.ReaderConfig{
			Brokers: brokers,
			Topic:   reserveTopic,
			GroupID: groupID,
		}),
		cancelReader: kafkago.NewReader(kafkago.ReaderConfig{
			Brokers: brokers,
			Topic:   cancelTopic,
			GroupID: groupID,
		}),
		handler: handler,
	}
}

func (c *Consumer) Close() error {
	var err1, err2 error
	if c.reserveReader != nil {
		err1 = c.reserveReader.Close()
	}
	if c.cancelReader != nil {
		err2 = c.cancelReader.Close()
	}
	if err1 != nil {
		return err1
	}
	return err2
}

type TicketCommand struct {
	CompositeBookingID int64 `json:"compositeBookingId"`
	TicketID           int64 `json:"ticketId"`
}

func (c *Consumer) Run(ctx context.Context) error {
	errCh := make(chan error, 2)

	go func() {
		errCh <- c.consumeLoop(ctx, c.reserveReader, func(cmd TicketCommand) error {
			return c.handler.ReserveTicket(ctx, cmd.TicketID, cmd.CompositeBookingID)
		})
	}()

	go func() {
		errCh <- c.consumeLoop(ctx, c.cancelReader, func(cmd TicketCommand) error {
			return c.handler.CancelTicket(ctx, cmd.TicketID, cmd.CompositeBookingID)
		})
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

func (c *Consumer) consumeLoop(
	ctx context.Context,
	reader *kafkago.Reader,
	process func(TicketCommand) error,
) error {
	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return err
			}
			return fmt.Errorf("read kafka message: %w", err)
		}

		var cmd TicketCommand
		if err := json.Unmarshal(msg.Value, &cmd); err != nil {
			continue
		}

		if cmd.TicketID == 0 || cmd.CompositeBookingID == 0 {
			continue
		}

		if err := process(cmd); err != nil {
			_ = reader.CommitMessages(ctx, msg)
			continue
		}

		if err := reader.CommitMessages(ctx, msg); err != nil {
			return fmt.Errorf("commit kafka message: %w", err)
		}
	}
}

// TODO: Think about DLQ, Outbox, idempotency.
