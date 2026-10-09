package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/tpu-orch/avia/internal/domain/models"
	"github.com/tpu-orch/avia/internal/domain/rules"
)

type TicketCommandRepo interface {
	GetTicketByID(ctx context.Context, ticketID int64) (*models.Ticket, error)
	UpdateTicketReservation(
		ctx context.Context,
		ticketID int64,
		status models.ReservationStatus,
		compositeBookingID *int64,
	) error
}

var (
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrInvalidStatus           = errors.New("invalid reservation status")
)

type TicketCommandService struct {
	repo TicketCommandRepo
}

func NewTicketCommandService(repo TicketCommandRepo) *TicketCommandService {
	return &TicketCommandService{repo: repo}
}

func (s *TicketCommandService) ReserveTicket(
	ctx context.Context,
	ticketID int64,
	compositeBookingID int64,
) (models.ReservationStatus, error) {
	return s.updateReservationStatus(ctx, ticketID, models.StatusReservationPending, &compositeBookingID)
}

func (s *TicketCommandService) CancelTicket(
	ctx context.Context,
	ticketID int64,
	compositeBookingID int64,
) (models.ReservationStatus, error) {
	_ = compositeBookingID // потом можно валидировать совпадение с БД
	return s.updateReservationStatus(ctx, ticketID, models.StatusAvailable, nil)
}

func (s *TicketCommandService) updateReservationStatus(
	ctx context.Context,
	ticketID int64,
	newStatus models.ReservationStatus,
	compositeBookingID *int64,
) (models.ReservationStatus, error) {
	if ticketID < 1 {
		return "", fmt.Errorf("%w: invalid ticket ID", ErrInvalidParameter)
	}
	if !rules.IsValidStatus(newStatus) {
		return "", fmt.Errorf("%w: invalid target status %s", ErrInvalidStatus, newStatus)
	}

	ticket, err := s.repo.GetTicketByID(ctx, ticketID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", ErrTicketNotFound
		}
		return "", fmt.Errorf("repo get ticket by id: %w", err)
	}

	if !rules.IsValidTransition(ticket.ReservationStatus, newStatus) {
		return "", fmt.Errorf(
			"%w: cannot transition from %s to %s",
			ErrInvalidStatusTransition,
			ticket.ReservationStatus,
			newStatus,
		)
	}

	if err := s.repo.UpdateTicketReservation(ctx, ticketID, newStatus, compositeBookingID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", ErrTicketNotFound
		}
		return "", fmt.Errorf("repo update ticket reservation: %w", err)
	}

	return newStatus, nil
}
