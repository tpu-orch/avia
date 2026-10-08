package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/tpu-orch/avia/internal/domain/models"
	"github.com/tpu-orch/avia/internal/domain/rules"
)

var (
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrInvalidStatus           = errors.New("invalid reservation status")
)

type AviaService struct {
	repo TicketRepository
}

func NewAviaService(repo TicketRepository) *AviaService {
	return &AviaService{repo: repo}
}

func (s *AviaService) UpdateReservationStatus(ctx context.Context, ticketID int64, newStatus models.ReservationStatus) error {
	if ticketID < 1 {
		return fmt.Errorf("%w: invalid ticket ID", ErrInvalidParameter)
	}
	if !rules.IsValidStatus(newStatus) {
		return fmt.Errorf("%w: invalid target status %s", ErrInvalidStatus, newStatus)
	}

	// Read ticket first to check current status and validate transition rules (read-before-update)
	ticket, err := s.repo.GetTicketByID(ctx, ticketID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrTicketNotFound
		}
		return fmt.Errorf("repo get ticket by id: %w", err)
	}

	// Validate status transition business rule using domain/rules
	if !rules.IsValidTransition(ticket.ReservationStatus, newStatus) {
		return fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidStatusTransition, ticket.ReservationStatus, newStatus)
	}

	return s.repo.UpdateTicketStatus(ctx, ticketID, newStatus)
}
