package models

import "time"

type ReservationStatus string

const (
	StatusAvailable           ReservationStatus = "AVAILABLE"
	StatusReservationPending  ReservationStatus = "RESERVATION_PENDING"
	StatusReserved            ReservationStatus = "RESERVED"
	StatusCancellationPending ReservationStatus = "CANCELLATION_PENDING"
	StatusCancelled           ReservationStatus = "CANCELLED"
)

type Ticket struct {
	ID                int64
	Price             float64
	Seat              string
	Flight            Flight
	ReservationStatus ReservationStatus
	UpdatedAt         time.Time
}
