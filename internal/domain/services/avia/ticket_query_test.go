package service

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/tpu-orch/avia/internal/domain/models"
)

type mockQueryRepo struct {
	getCitiesFn       func(context.Context, string) ([]string, error)
	getTicketsRouteFn func(context.Context, string, string) ([]models.Ticket, error)
	getTicketByIDFn   func(context.Context, int64) (*models.Ticket, error)

	getCitiesCalls       int
	getTicketsRouteCalls int
	getTicketByIDCalls   int
}

func (m *mockQueryRepo) GetCities(ctx context.Context, query string) ([]string, error) {
	m.getCitiesCalls++
	if m.getCitiesFn != nil {
		return m.getCitiesFn(ctx, query)
	}
	return nil, nil
}

func (m *mockQueryRepo) GetTicketsByRoute(
	ctx context.Context,
	fromCity, toCity string,
) ([]models.Ticket, error) {
	m.getTicketsRouteCalls++
	if m.getTicketsRouteFn != nil {
		return m.getTicketsRouteFn(ctx, fromCity, toCity)
	}
	return nil, nil
}

func (m *mockQueryRepo) GetTicketByID(
	ctx context.Context,
	ticketID int64,
) (*models.Ticket, error) {
	m.getTicketByIDCalls++
	if m.getTicketByIDFn != nil {
		return m.getTicketByIDFn(ctx, ticketID)
	}
	return nil, ErrNotFound
}

func TestTicketQueryService_GetCities(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("database unavailable")

	tests := []struct {
		name      string
		query     string
		repo      *mockQueryRepo
		want      []string
		wantErr   bool
		wantErrIs error
		wantCalls int
	}{
		{
			name:  "success",
			query: "Moscow",
			repo: &mockQueryRepo{
				getCitiesFn: func(_ context.Context, query string) ([]string, error) {
					if query != "Moscow" {
						t.Errorf("query = %q, want %q", query, "Moscow")
					}
					return []string{"Moscow", "Saint Petersburg"}, nil
				},
			},
			want:      []string{"Moscow", "Saint Petersburg"},
			wantCalls: 1,
		},
		{
			name:  "trims whitespace",
			query: "  Moscow  ",
			repo: &mockQueryRepo{
				getCitiesFn: func(_ context.Context, query string) ([]string, error) {
					if query != "Moscow" {
						t.Errorf("query = %q, want %q", query, "Moscow")
					}
					return []string{"Moscow"}, nil
				},
			},
			want:      []string{"Moscow"},
			wantCalls: 1,
		},
		{
			name:      "empty query",
			query:     "",
			repo:      &mockQueryRepo{},
			want:      []string{},
			wantCalls: 1,
		},
		{
			name:  "query at maximum length",
			query: strings.Repeat("a", 100),
			repo: &mockQueryRepo{
				getCitiesFn: func(_ context.Context, query string) ([]string, error) {
					return []string{query}, nil
				},
			},
			want:      []string{strings.Repeat("a", 100)},
			wantCalls: 1,
		},
		{
			name:      "query exceeds maximum length",
			query:     strings.Repeat("a", 101),
			repo:      &mockQueryRepo{},
			wantErr:   true,
			wantErrIs: ErrInvalidParameter,
		},
		{
			name:  "repository error",
			query: "Moscow",
			repo: &mockQueryRepo{
				getCitiesFn: func(_ context.Context, _ string) ([]string, error) {
					return nil, repoErr
				},
			},
			wantErr:   true,
			wantErrIs: repoErr,
			wantCalls: 1,
		},
		{
			name:  "nil result becomes empty slice",
			query: "Unknown",
			repo: &mockQueryRepo{
				getCitiesFn: func(_ context.Context, _ string) ([]string, error) {
					return nil, nil
				},
			},
			want:      []string{},
			wantCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := NewTicketQueryService(tt.repo)
			got, err := s.GetCities(context.Background(), tt.query)

			if (err != nil) != tt.wantErr {
				t.Fatalf("GetCities() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
				t.Errorf("GetCities() error = %v, want wrapping %v", err, tt.wantErrIs)
			}
			if !tt.wantErr && !slices.Equal(got, tt.want) {
				t.Errorf("GetCities() = %v, want %v", got, tt.want)
			}
			if tt.repo.getCitiesCalls != tt.wantCalls {
				t.Errorf("GetCities() repository calls = %d, want %d",
					tt.repo.getCitiesCalls, tt.wantCalls)
			}
		})
	}
}

func TestTicketQueryService_GetTickets(t *testing.T) {
	t.Parallel()

	available := models.Ticket{
		ID:                1,
		ReservationStatus: models.StatusAvailable,
		Flight:            models.Flight{IsCancelled: false},
	}
	reserved := models.Ticket{
		ID:                2,
		ReservationStatus: models.StatusReserved,
		Flight:            models.Flight{IsCancelled: false},
	}
	cancelledFlight := models.Ticket{
		ID:                3,
		ReservationStatus: models.StatusAvailable,
		Flight:            models.Flight{IsCancelled: true},
	}
	pending := models.Ticket{
		ID:                4,
		ReservationStatus: models.StatusReservationPending,
		Flight:            models.Flight{IsCancelled: false},
	}
	repoErr := errors.New("database unavailable")

	tests := []struct {
		name      string
		from      string
		to        string
		repo      *mockQueryRepo
		want      []models.Ticket
		wantErr   bool
		wantErrIs error
		wantCalls int
	}{
		{
			name: "filters reserved, pending and cancelled flights",
			from: "Moscow",
			to:   "Tomsk",
			repo: &mockQueryRepo{
				getTicketsRouteFn: func(_ context.Context, from, to string) ([]models.Ticket, error) {
					if from != "Moscow" || to != "Tomsk" {
						t.Errorf("route = %q -> %q, want Moscow -> Tomsk", from, to)
					}
					return []models.Ticket{available, reserved, cancelledFlight, pending}, nil
				},
			},
			want:      []models.Ticket{available},
			wantCalls: 1,
		},
		{
			name:      "empty departure city",
			from:      "",
			to:        "Tomsk",
			repo:      &mockQueryRepo{},
			wantErr:   true,
			wantErrIs: ErrInvalidParameter,
		},
		{
			name:      "empty destination city",
			from:      "Moscow",
			to:        "",
			repo:      &mockQueryRepo{},
			wantErr:   true,
			wantErrIs: ErrInvalidParameter,
		},
		{
			name:      "whitespace-only departure city",
			from:      "   ",
			to:        "Tomsk",
			repo:      &mockQueryRepo{},
			wantErr:   true,
			wantErrIs: ErrInvalidParameter,
		},
		{
			name:      "same cities ignoring case",
			from:      "Moscow",
			to:        "  mOsCoW ",
			repo:      &mockQueryRepo{},
			wantErr:   true,
			wantErrIs: ErrInvalidParameter,
		},
		{
			name:      "departure city exceeds maximum length",
			from:      strings.Repeat("a", 101),
			to:        "Tomsk",
			repo:      &mockQueryRepo{},
			wantErr:   true,
			wantErrIs: ErrInvalidParameter,
		},
		{
			name:      "destination city exceeds maximum length",
			from:      "Moscow",
			to:        strings.Repeat("a", 101),
			repo:      &mockQueryRepo{},
			wantErr:   true,
			wantErrIs: ErrInvalidParameter,
		},
		{
			name: "trims city names",
			from: "  Moscow ",
			to:   " Tomsk  ",
			repo: &mockQueryRepo{
				getTicketsRouteFn: func(_ context.Context, from, to string) ([]models.Ticket, error) {
					if from != "Moscow" || to != "Tomsk" {
						t.Errorf("route = %q -> %q, want Moscow -> Tomsk", from, to)
					}
					return []models.Ticket{available}, nil
				},
			},
			want:      []models.Ticket{available},
			wantCalls: 1,
		},
		{
			name: "repository error",
			from: "Moscow",
			to:   "Tomsk",
			repo: &mockQueryRepo{
				getTicketsRouteFn: func(_ context.Context, _, _ string) ([]models.Ticket, error) {
					return nil, repoErr
				},
			},
			wantErr:   true,
			wantErrIs: repoErr,
			wantCalls: 1,
		},
		{
			name: "no matching tickets returns empty slice",
			from: "Moscow",
			to:   "Tomsk",
			repo: &mockQueryRepo{
				getTicketsRouteFn: func(_ context.Context, _, _ string) ([]models.Ticket, error) {
					return []models.Ticket{reserved, cancelledFlight, pending}, nil
				},
			},
			want:      []models.Ticket{},
			wantCalls: 1,
		},
		{
			name: "nil repository result returns empty slice",
			from: "Moscow",
			to:   "Tomsk",
			repo: &mockQueryRepo{
				getTicketsRouteFn: func(_ context.Context, _, _ string) ([]models.Ticket, error) {
					return nil, nil
				},
			},
			want:      []models.Ticket{},
			wantCalls: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := NewTicketQueryService(tt.repo)
			got, err := s.GetTickets(context.Background(), tt.from, tt.to)

			if (err != nil) != tt.wantErr {
				t.Fatalf("GetTickets() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
				t.Errorf("GetTickets() error = %v, want wrapping %v", err, tt.wantErrIs)
			}
			if !tt.wantErr && !slices.Equal(got, tt.want) {
				t.Errorf("GetTickets() = %+v, want %+v", got, tt.want)
			}
			if tt.repo.getTicketsRouteCalls != tt.wantCalls {
				t.Errorf("GetTickets() repository calls = %d, want %d",
					tt.repo.getTicketsRouteCalls, tt.wantCalls)
			}
		})
	}
}

func TestTicketQueryService_GetReservationStatus(t *testing.T) {
	t.Parallel()

	updatedAt := time.Date(2026, time.October, 14, 12, 30, 0, 0, time.UTC)
	repoErr := errors.New("database unavailable")

	tests := []struct {
		name      string
		ticketID  int64
		repo      *mockQueryRepo
		want      models.ReservationStatus
		wantTime  time.Time
		wantErr   bool
		wantErrIs error
		wantCalls int
	}{
		{
			name:     "success",
			ticketID: 1,
			repo: &mockQueryRepo{
				getTicketByIDFn: func(_ context.Context, id int64) (*models.Ticket, error) {
					if id != 1 {
						t.Errorf("ticketID = %d, want 1", id)
					}
					return &models.Ticket{
						ID:                1,
						ReservationStatus: models.StatusReserved,
						UpdatedAt:         updatedAt,
					}, nil
				},
			},
			want:      models.StatusReserved,
			wantTime:  updatedAt,
			wantCalls: 1,
		},
		{
			name:      "zero ticket ID",
			ticketID:  0,
			repo:      &mockQueryRepo{},
			wantErr:   true,
			wantErrIs: ErrInvalidParameter,
		},
		{
			name:      "negative ticket ID",
			ticketID:  -1,
			repo:      &mockQueryRepo{},
			wantErr:   true,
			wantErrIs: ErrInvalidParameter,
		},
		{
			name:     "ticket not found",
			ticketID: 999,
			repo: &mockQueryRepo{
				getTicketByIDFn: func(_ context.Context, _ int64) (*models.Ticket, error) {
					return nil, ErrNotFound
				},
			},
			wantErr:   true,
			wantErrIs: ErrTicketNotFound,
			wantCalls: 1,
		},
		{
			name:     "wrapped not found error",
			ticketID: 999,
			repo: &mockQueryRepo{
				getTicketByIDFn: func(_ context.Context, _ int64) (*models.Ticket, error) {
					return nil, errors.Join(errors.New("repository lookup"), ErrNotFound)
				},
			},
			wantErr:   true,
			wantErrIs: ErrTicketNotFound,
			wantCalls: 1,
		},
		{
			name:     "repository error",
			ticketID: 1,
			repo: &mockQueryRepo{
				getTicketByIDFn: func(_ context.Context, _ int64) (*models.Ticket, error) {
					return nil, repoErr
				},
			},
			wantErr:   true,
			wantErrIs: repoErr,
			wantCalls: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := NewTicketQueryService(tt.repo)
			gotStatus, gotTime, err := s.GetReservationStatus(
				context.Background(),
				tt.ticketID,
			)

			if (err != nil) != tt.wantErr {
				t.Fatalf("GetReservationStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
				t.Errorf("GetReservationStatus() error = %v, want wrapping %v", err, tt.wantErrIs)
			}
			if !tt.wantErr {
				if gotStatus != tt.want {
					t.Errorf("GetReservationStatus() status = %q, want %q", gotStatus, tt.want)
				}
				if !gotTime.Equal(tt.wantTime) {
					t.Errorf("GetReservationStatus() updatedAt = %v, want %v", gotTime, tt.wantTime)
				}
			}
			if tt.repo.getTicketByIDCalls != tt.wantCalls {
				t.Errorf("GetTicketByID() calls = %d, want %d",
					tt.repo.getTicketByIDCalls, tt.wantCalls)
			}
		})
	}
}
