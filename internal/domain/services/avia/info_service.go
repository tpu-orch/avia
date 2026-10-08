package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tpu-orch/avia/internal/domain/models"
)

var (
	ErrInvalidParameter = errors.New("invalid parameter")
	ErrTicketNotFound   = errors.New("ticket not found")
	ErrNotFound         = errors.New("not found")
)

type TicketRepository interface {
	GetCities(ctx context.Context, query string) ([]string, error)
	GetTicketsByRoute(ctx context.Context, fromCity, toCity string) ([]models.Ticket, error)
	GetTicketByID(ctx context.Context, ticketID int64) (*models.Ticket, error)
	UpdateTicketStatus(ctx context.Context, ticketID int64, status models.ReservationStatus) error
}

type InfoService struct {
	repo TicketRepository
}

func NewInfoService(repo TicketRepository) *InfoService {
	return &InfoService{repo: repo}
}

func (s *InfoService) GetCities(ctx context.Context, query string) ([]string, error) {
	query = strings.TrimSpace(query)
	if len(query) > 100 {
		return nil, fmt.Errorf("%w: query parameter cannot exceed 100 characters", ErrInvalidParameter)
	}

	cities, err := s.repo.GetCities(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("repo get cities: %w", err)
	}
	if cities == nil {
		cities = []string{}
	}
	return cities, nil
}

func (s *InfoService) GetTickets(ctx context.Context, from, to string) ([]models.Ticket, error) {
	from = strings.TrimSpace(from)
	to = strings.TrimSpace(to)

	if from == "" || to == "" {
		return nil, fmt.Errorf("%w: parameters 'from' and 'to' are required", ErrInvalidParameter)
	}
	if len(from) > 100 || len(to) > 100 {
		return nil, fmt.Errorf("%w: city names cannot exceed 100 characters", ErrInvalidParameter)
	}

	if strings.EqualFold(from, to) {
		return nil, fmt.Errorf("%w: departure and arrival cities cannot be the same", ErrInvalidParameter)
	}

	rawTickets, err := s.repo.GetTicketsByRoute(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("repo get tickets by route: %w", err)
	}

	var availableTickets []models.Ticket
	for _, t := range rawTickets {
		if t.ReservationStatus == models.StatusAvailable && !t.Flight.IsCancelled {
			availableTickets = append(availableTickets, t)
		}
	}

	if availableTickets == nil {
		availableTickets = []models.Ticket{}
	}

	return availableTickets, nil
}

func (s *InfoService) GetReservationStatus(ctx context.Context, ticketID int64) (models.ReservationStatus, time.Time, error) {
	if ticketID < 1 {
		return "", time.Time{}, fmt.Errorf("%w: invalid ticket ID", ErrInvalidParameter)
	}

	ticket, err := s.repo.GetTicketByID(ctx, ticketID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", time.Time{}, ErrTicketNotFound
		}
		return "", time.Time{}, fmt.Errorf("repo get ticket by id: %w", err)
	}

	return ticket.ReservationStatus, ticket.UpdatedAt, nil
}
