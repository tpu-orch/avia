package service

import (
	"context"
	"errors"
	"testing"

	"github.com/tpu-orch/avia/internal/domain/models"
)

var ErrRepo = errors.New("database connection lost")

type mockCommandRepo struct {
	getTicketFn         func(ctx context.Context, ticketID int64) (*models.Ticket, error)
	updateReservationFn func(ctx context.Context, ticketID int64, status models.ReservationStatus, compositeBookingID *int64) error
}

func (m *mockCommandRepo) GetTicketByID(ctx context.Context, ticketID int64) (*models.Ticket, error) {
	if m.getTicketFn != nil {
		return m.getTicketFn(ctx, ticketID)
	}
	return nil, ErrNotFound
}

func (m *mockCommandRepo) UpdateTicketReservation(ctx context.Context, ticketID int64, status models.ReservationStatus, compositeBookingID *int64) error {
	if m.updateReservationFn != nil {
		return m.updateReservationFn(ctx, ticketID, status, compositeBookingID)
	}
	return nil
}

func TestTicketCommandService_ReserveTicket(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		ticketID           int64
		compositeBookingID int64
		mockRepo           *mockCommandRepo
		want               models.ReservationStatus
		wantErr            error
	}{
		{
			name:               "success",
			ticketID:           1,
			compositeBookingID: 100,
			mockRepo: &mockCommandRepo{
				getTicketFn: func(ctx context.Context, ticketID int64) (*models.Ticket, error) {
					return &models.Ticket{
						ID:                1,
						ReservationStatus: models.StatusAvailable,
					}, nil
				},
				updateReservationFn: func(ctx context.Context, ticketID int64, status models.ReservationStatus, compositeBookingID *int64) error {
					return nil
				},
			},
			want:    models.StatusReservationPending,
			wantErr: nil,
		},
		{
			name:               "invalid ticket ID",
			ticketID:           0,
			compositeBookingID: 100,
			mockRepo:           &mockCommandRepo{},
			want:               "",
			wantErr:            ErrInvalidParameter,
		},
		{
			name:               "ticket not found",
			ticketID:           999,
			compositeBookingID: 100,
			mockRepo: &mockCommandRepo{
				getTicketFn: func(ctx context.Context, ticketID int64) (*models.Ticket, error) {
					return nil, ErrNotFound
				},
			},
			want:    "",
			wantErr: ErrTicketNotFound,
		},
		{
			name:               "invalid transition",
			ticketID:           1,
			compositeBookingID: 100,
			mockRepo: &mockCommandRepo{
				getTicketFn: func(ctx context.Context, ticketID int64) (*models.Ticket, error) {
					return &models.Ticket{
						ID:                1,
						ReservationStatus: models.StatusReserved,
					}, nil
				},
			},
			want:    "",
			wantErr: ErrInvalidStatusTransition,
		},
		{
			name:               "database get ticket error",
			ticketID:           1,
			compositeBookingID: 100,
			mockRepo: &mockCommandRepo{
				getTicketFn: func(ctx context.Context, ticketID int64) (*models.Ticket, error) {
					return nil, ErrRepo
				},
			},
			want:    "",
			wantErr: ErrRepo,
		},
		{
			name:               "database update reservation status error",
			ticketID:           1,
			compositeBookingID: 100,
			mockRepo: &mockCommandRepo{
				getTicketFn: func(ctx context.Context, ticketID int64) (*models.Ticket, error) {
					return &models.Ticket{
						ID:                1,
						ReservationStatus: models.StatusAvailable,
					}, nil
				},
				updateReservationFn: func(ctx context.Context, ticketID int64, status models.ReservationStatus, compositeBookingID *int64) error {
					return ErrRepo
				},
			},
			want:    "",
			wantErr: ErrRepo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := NewTicketCommandService(tt.mockRepo)
			got, err := s.ReserveTicket(context.Background(), tt.ticketID, tt.compositeBookingID)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ReserveTicket() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("ReserveTicket() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTicketCommandService_CancelTicket(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		ticketID           int64
		compositeBookingID int64
		mockRepo           *mockCommandRepo
		want               models.ReservationStatus
		wantErr            error
	}{
		{
			name:               "success from pending",
			ticketID:           1,
			compositeBookingID: 100,
			mockRepo: &mockCommandRepo{
				getTicketFn: func(ctx context.Context, ticketID int64) (*models.Ticket, error) {
					return &models.Ticket{
						ID:                1,
						ReservationStatus: models.StatusReservationPending,
					}, nil
				},
				updateReservationFn: func(ctx context.Context, ticketID int64, status models.ReservationStatus, compositeBookingID *int64) error {
					return nil
				},
			},
			want:    models.StatusAvailable,
			wantErr: nil,
		},
		{
			name:               "invalid ticket ID",
			ticketID:           -1,
			compositeBookingID: 100,
			mockRepo:           &mockCommandRepo{},
			want:               "",
			wantErr:            ErrInvalidParameter,
		},
		{
			name:               "database get ticket error",
			ticketID:           1,
			compositeBookingID: 100,
			mockRepo: &mockCommandRepo{
				getTicketFn: func(ctx context.Context, ticketID int64) (*models.Ticket, error) {
					return nil, ErrRepo
				},
			},
			want:    "",
			wantErr: ErrRepo,
		},
		{
			name:               "ticket not found",
			ticketID:           999,
			compositeBookingID: 100,
			mockRepo: &mockCommandRepo{
				getTicketFn: func(ctx context.Context, ticketID int64) (*models.Ticket, error) {
					return nil, ErrNotFound
				},
			},
			want:    "",
			wantErr: ErrTicketNotFound,
		},
		{
			name:               "invalid transition",
			ticketID:           1,
			compositeBookingID: 100,
			mockRepo: &mockCommandRepo{
				getTicketFn: func(ctx context.Context, ticketID int64) (*models.Ticket, error) {
					return &models.Ticket{
						ID:                1,
						ReservationStatus: models.StatusReserved,
					}, nil
				},
			},
			want:    "",
			wantErr: ErrInvalidStatusTransition,
		},
		{
			name:               "database update reservation status error",
			ticketID:           1,
			compositeBookingID: 100,
			mockRepo: &mockCommandRepo{
				getTicketFn: func(ctx context.Context, ticketID int64) (*models.Ticket, error) {
					return &models.Ticket{
						ID:                1,
						ReservationStatus: models.StatusReservationPending,
					}, nil
				},
				updateReservationFn: func(ctx context.Context, ticketID int64, status models.ReservationStatus, compositeBookingID *int64) error {
					return ErrRepo
				},
			},
			want:    "",
			wantErr: ErrRepo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := NewTicketCommandService(tt.mockRepo)
			got, err := s.CancelTicket(context.Background(), tt.ticketID, tt.compositeBookingID)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CancelTicket() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("CancelTicket() = %v, want %v", got, tt.want)
			}
		})
	}
}
