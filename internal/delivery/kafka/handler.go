package kafka

import (
	"context"

	"github.com/tpu-orch/avia/internal/delivery/kafka/dto"
	service "github.com/tpu-orch/avia/internal/domain/services/avia"
)

type Handler struct {
	cmdService *service.TicketCommandService
	producer   *Producer
}

func NewHandler(cmdService *service.TicketCommandService, producer *Producer) *Handler {
	return &Handler{
		cmdService: cmdService,
		producer:   producer,
	}
}

func (h *Handler) ReserveTicket(ctx context.Context, ticketID int64, compositeBookingID int64) error {
	status, err := h.cmdService.ReserveTicket(ctx, ticketID, compositeBookingID)
	if err != nil {
		return err
	}

	return h.producer.SendReservationStatusChanged(ctx, dto.ReservationStatusChangedPayload{
		CompositeBookingID: compositeBookingID,
		TicketID:           ticketID,
		Status:             string(status),
	})
}

func (h *Handler) CancelTicket(ctx context.Context, ticketID int64, compositeBookingID int64) error {
	status, err := h.cmdService.CancelTicket(ctx, ticketID, compositeBookingID)
	if err != nil {
		return err
	}

	return h.producer.SendReservationStatusChanged(ctx, dto.ReservationStatusChangedPayload{
		CompositeBookingID: compositeBookingID,
		TicketID:           ticketID,
		Status:             string(status),
	})
}
