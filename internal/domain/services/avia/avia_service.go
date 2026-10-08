package service

import (
	"context"

	"github.com/tpu-orch/avia/internal/domain/models"
)

type AviaRepository interface {
	UpdateReservationStatus(ctx context.Context, ticketID int64, status models.ReservationStatus) error
}

type AviaService struct {
	repo AviaRepository
}

func NewAviaService(repo AviaRepository) *AviaService {
	return &AviaService{repo: repo}
}

func (s *AviaService) UpdateReservationStatus(ctx context.Context, ticketID int64, status models.ReservationStatus) error {
	return s.repo.UpdateReservationStatus(ctx, ticketID, status)
}
