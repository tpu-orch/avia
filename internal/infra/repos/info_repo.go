package repos

import (
	"context"
	"time"

	"github.com/tpu-orch/avia/internal/domain/models"
)

type InfoRepository interface {
	GetCities(ctx context.Context, query string) ([]string, error)
	GetTickets(ctx context.Context, from, to string) ([]models.Ticket, error)
	GetReservationStatus(ctx context.Context, ticketID int64) (models.ReservationStatus, time.Time, error)
}

type PostgresInfoRepo struct{}

func (r *PostgresInfoRepo) GetCities(ctx context.Context, query string) ([]string, error) {
}

func (r *PostgresInfoRepo) GetTickets(ctx context.Context, from, to string) ([]models.Ticket, error) {
}

func (r *PostgresInfoRepo) GetReservationStatus(ctx context.Context, ticketID int64) (models.ReservationStatus, time.Time, error) {
}
