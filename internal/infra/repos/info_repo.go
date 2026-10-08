package repos

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tpu-orch/avia/internal/domain/models"
)

var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

const (
	tableAirports = "airports"
	tableFlights  = "flights"
	tableTickets  = "tickets"
)

const (
	colID                = "id"
	colCity              = "city"
	colCodeIata          = "code_iata"
	colLongitude         = "longitude"
	colLatitude          = "latitude"
	colDepartureDate     = "departure_date"
	colArrivalDate       = "arrival_date"
	colFromAirportID     = "from_airport_id"
	colToAirportID       = "to_airport_id"
	colIsCancelled       = "is_cancelled"
	colPrice             = "price"
	colSeat              = "seat"
	colFlightID          = "flight_id"
	colReservationStatus = "reservation_status"
	colUpdatedAt         = "updated_at"
)

type InfoRepository interface {
	GetCities(ctx context.Context, query string) ([]string, error)
	GetTickets(ctx context.Context, from, to string) ([]models.Ticket, error)
	GetReservationStatus(ctx context.Context, ticketID int64) (models.ReservationStatus, time.Time, error)
}

type AviaRepository interface {
	UpdateReservationStatus(ctx context.Context, ticketID int64, status models.ReservationStatus) error
}

type PostgresInfoRepo struct {
	db *pgxpool.Pool
}

func NewPostgresInfoRepo(db *pgxpool.Pool) *PostgresInfoRepo {
	return &PostgresInfoRepo{db: db}
}

func (r *PostgresInfoRepo) GetCities(ctx context.Context, query string) ([]string, error) {
	sb := psql.Select("DISTINCT " + colCity).From(tableAirports).OrderBy(colCity)
	if query != "" {
		sb = sb.Where(colCity+" ILIKE ?", "%"+query+"%")
	}

	sqlStr, args, err := sb.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build get cities query: %w", err)
	}

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("execute get cities query: %w", err)
	}
	defer rows.Close()

	var cities []string
	for rows.Next() {
		var city string
		if err := rows.Scan(&city); err != nil {
			return nil, fmt.Errorf("scan city: %w", err)
		}
		cities = append(cities, city)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return cities, nil
}

func (r *PostgresInfoRepo) GetTickets(ctx context.Context, from, to string) ([]models.Ticket, error) {
	sb := psql.Select(
		"t."+colID, "t."+colPrice, "t."+colSeat, "t."+colReservationStatus, "t."+colUpdatedAt,
		"f."+colID, "f."+colDepartureDate, "f."+colArrivalDate, "f."+colIsCancelled,
		"af."+colID, "af."+colCodeIata, "af."+colCity, "af."+colLongitude, "af."+colLatitude,
		"at."+colID, "at."+colCodeIata, "at."+colCity, "at."+colLongitude, "at."+colLatitude,
	).
		From(tableTickets+" t").
		Join(tableFlights+" f ON t."+colFlightID+" = f."+colID).
		Join(tableAirports+" af ON f."+colFromAirportID+" = af."+colID).
		Join(tableAirports+" at ON f."+colToAirportID+" = at."+colID).
		Where("t."+colReservationStatus+" = ?", models.StatusAvailable).
		Where("f."+colIsCancelled+" = ?", false).
		Where("af."+colCity+" ILIKE ?", from).
		Where("at."+colCity+" ILIKE ?", to).
		OrderBy("t." + colID)

	sqlStr, args, err := sb.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build get tickets query: %w", err)
	}

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("execute get tickets query: %w", err)
	}
	defer rows.Close()

	var tickets []models.Ticket
	for rows.Next() {
		var t models.Ticket
		var updatedAt time.Time
		var f models.Flight
		var af, at models.Airport

		if err := rows.Scan(
			&t.ID, &t.Price, &t.Seat, &t.ReservationStatus, &updatedAt,
			&f.ID, &f.DepartureDate, &f.ArrivalDate, &f.IsCancelled,
			&af.ID, &af.CodeIata, &af.City, &af.Longitude, &af.Latitude,
			&at.ID, &at.CodeIata, &at.City, &at.Longitude, &at.Latitude,
		); err != nil {
			return nil, fmt.Errorf("scan ticket: %w", err)
		}
		f.From = af
		f.To = at
		t.Flight = f
		tickets = append(tickets, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return tickets, nil
}

func (r *PostgresInfoRepo) GetReservationStatus(ctx context.Context, ticketID int64) (models.ReservationStatus, time.Time, error) {
	sqlStr, args, err := psql.Select(colReservationStatus, colUpdatedAt).
		From(tableTickets).
		Where(colID+" = ?", ticketID).
		ToSql()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("build get reservation status query: %w", err)
	}

	var status models.ReservationStatus
	var updatedAt time.Time
	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(&status, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", time.Time{}, fmt.Errorf("ticket not found: %w", err)
		}
		return "", time.Time{}, fmt.Errorf("query reservation status: %w", err)
	}

	return status, updatedAt, nil
}

func (r *PostgresInfoRepo) UpdateReservationStatus(ctx context.Context, ticketID int64, status models.ReservationStatus) error {
	sqlStr, args, err := psql.Update(tableTickets).
		Set(colReservationStatus, status).
		Set(colUpdatedAt, squirrel.Expr("NOW()")).
		Where(colID+" = ?", ticketID).
		ToSql()
	if err != nil {
		return fmt.Errorf("build update reservation status query: %w", err)
	}

	result, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		return fmt.Errorf("execute update reservation status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("ticket not found")
	}

	return nil
}
