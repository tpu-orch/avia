-- +goose Up
CREATE TABLE IF NOT EXISTS airports (
    id BIGSERIAL PRIMARY KEY,
    code_iata VARCHAR(3) NOT NULL UNIQUE,
    city VARCHAR(255) NOT NULL,
    longitude DOUBLE PRECISION,
    latitude DOUBLE PRECISION
);

CREATE TABLE IF NOT EXISTS flights (
    id BIGSERIAL PRIMARY KEY,
    departure_date TIMESTAMPTZ NOT NULL,
    arrival_date TIMESTAMPTZ NOT NULL,
    from_airport_id BIGINT NOT NULL REFERENCES airports(id),
    to_airport_id BIGINT NOT NULL REFERENCES airports(id),
    is_cancelled BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS tickets (
    id BIGSERIAL PRIMARY KEY,
    price NUMERIC(10, 2) NOT NULL,
    seat VARCHAR(10) NOT NULL,
    flight_id BIGINT NOT NULL REFERENCES flights(id),
    reservation_status VARCHAR(50) NOT NULL DEFAULT 'AVAILABLE',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_airports_city ON airports(city);
CREATE INDEX IF NOT EXISTS idx_tickets_status ON tickets(reservation_status);
CREATE INDEX IF NOT EXISTS idx_flights_airports ON flights(from_airport_id, to_airport_id);

-- +goose Down
DROP TABLE IF EXISTS tickets;
DROP TABLE IF EXISTS flights;
DROP TABLE IF EXISTS airports;
