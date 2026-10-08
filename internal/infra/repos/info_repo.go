package repos

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tpu-orch/avia/internal/domain/models"
	service "github.com/tpu-orch/avia/internal/domain/services/avia"
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

type PostgresTicketRepo struct {
	db *pgxpool.Pool
}

func NewPostgresTicketRepo(db *pgxpool.Pool) *PostgresTicketRepo {
	return &PostgresTicketRepo{db: db}
}

func (r *PostgresTicketRepo) GetCities(ctx context.Context, query string) ([]string, error) {
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

func (r *PostgresTicketRepo) GetTicketsByRoute(ctx context.Context, fromCity, toCity string) ([]models.Ticket, error) {
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
		Where("LOWER(af."+colCity+") = LOWER(?)", fromCity).
		Where("LOWER(at."+colCity+") = LOWER(?)", toCity).
		OrderBy("t." + colID)

	sqlStr, args, err := sb.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build get tickets by route query: %w", err)
	}

	rows, err := r.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("execute get tickets by route query: %w", err)
	}
	defer rows.Close()

	var tickets []models.Ticket
	for rows.Next() {
		var t models.Ticket
		var f models.Flight
		var af, at models.Airport

		if err := rows.Scan(
			&t.ID, &t.Price, &t.Seat, &t.ReservationStatus, &t.UpdatedAt,
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

func (r *PostgresTicketRepo) GetTicketByID(ctx context.Context, ticketID int64) (*models.Ticket, error) {
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
		Where("t."+colID+" = ?", ticketID)

	sqlStr, args, err := sb.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build get ticket by id query: %w", err)
	}

	var t models.Ticket
	var f models.Flight
	var af, at models.Airport

	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(
		&t.ID, &t.Price, &t.Seat, &t.ReservationStatus, &t.UpdatedAt,
		&f.ID, &f.DepartureDate, &f.ArrivalDate, &f.IsCancelled,
		&af.ID, &af.CodeIata, &af.City, &af.Longitude, &af.Latitude,
		&at.ID, &at.CodeIata, &at.City, &at.Longitude, &at.Latitude,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, service.ErrNotFound
		}
		return nil, fmt.Errorf("query ticket by id: %w", err)
	}
	f.From = af
	f.To = at
	t.Flight = f

	return &t, nil
}

func (r *PostgresTicketRepo) UpdateTicketStatus(ctx context.Context, ticketID int64, status models.ReservationStatus) error {
	sqlStr, args, err := psql.Update(tableTickets).
		Set(colReservationStatus, status).
		Set(colUpdatedAt, squirrel.Expr("NOW()")).
		Where(colID+" = ?", ticketID).
		ToSql()
	if err != nil {
		return fmt.Errorf("build update ticket status query: %w", err)
	}

	result, err := r.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		return fmt.Errorf("execute update ticket status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return service.ErrNotFound
	}

	return nil
}
