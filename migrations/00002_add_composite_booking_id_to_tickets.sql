-- +goose Up
ALTER TABLE tickets
ADD COLUMN composite_booking_id BIGINT;

CREATE INDEX IF NOT EXISTS idx_tickets_composite_booking_id
ON tickets(composite_booking_id);

-- +goose Down
DROP INDEX IF EXISTS idx_tickets_composite_booking_id;

ALTER TABLE tickets
DROP COLUMN IF EXISTS composite_booking_id;

