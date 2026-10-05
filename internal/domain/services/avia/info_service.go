package service

import (
	"context"
	"time"

	"github.com/tpu-orch/avia/internal/domain/models"
	"github.com/tpu-orch/avia/internal/infra/repos"
)

type InfoService struct {
	repo repos.InfoRepository
}

func NewInfoService(repo repos.InfoRepository) *InfoService {
	return &InfoService{repo: repo}
}

func (s *InfoService) GetCities(ctx context.Context, query string) ([]models.City, error) {
	return s.repo.GetCities(ctx, query)
}

func (s *InfoService) GetTickets(ctx context.Context, from, to string) ([]models.Ticket, error) {
	return s.repo.GetTickets(ctx, from, to)
}

func (s *InfoService) GetReservationStatus(ctx context.Context, ticketID int64) (models.ReservationStatus, time.Time, error) {
	return s.repo.GetReservationStatus(ctx, ticketID)
}
