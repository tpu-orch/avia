-- +goose Up

-- Аэропорты
INSERT INTO airports (code_iata, city, longitude, latitude)
VALUES
    ('LED', 'Saint Petersburg', 30.2625, 59.8003),
    ('SVO', 'Moscow', 37.4146, 55.9726),
    ('VKO', 'Moscow', 37.2615, 55.5915),
    ('AER', 'Sochi', 39.9566, 43.4499),
    ('KZN', 'Kazan', 49.2787, 55.6062),
    ('OVB', 'Novosibirsk', 82.6507, 55.0126),
    ('KGD', 'Kaliningrad', 20.5926, 54.8900),
    ('SVX', 'Yekaterinburg', 60.8027, 56.7431)
ON CONFLICT (code_iata) DO NOTHING;

-- Рейсы
INSERT INTO flights (
    departure_date,
    arrival_date,
    from_airport_id,
    to_airport_id,
    is_cancelled
)
SELECT
    v.departure_date,
    v.arrival_date,
    a_from.id,
    a_to.id,
    v.is_cancelled
FROM (
    VALUES
        ('LED', 'SVO', '2027-01-10 08:00:00+03'::timestamptz, '2027-01-10 09:30:00+03'::timestamptz, FALSE),
        ('LED', 'SVO', '2027-01-10 14:00:00+03'::timestamptz, '2027-01-10 15:30:00+03'::timestamptz, FALSE),
        ('SVO', 'LED', '2027-01-11 18:00:00+03'::timestamptz, '2027-01-11 19:30:00+03'::timestamptz, FALSE),
        ('LED', 'AER', '2027-01-12 07:00:00+03'::timestamptz, '2027-01-12 11:00:00+03'::timestamptz, FALSE),
        ('AER', 'LED', '2027-01-15 16:00:00+03'::timestamptz, '2027-01-15 20:00:00+03'::timestamptz, FALSE),
        ('SVO', 'KZN', '2027-01-13 10:00:00+03'::timestamptz, '2027-01-13 11:30:00+03'::timestamptz, FALSE),
        ('LED', 'KGD', '2027-01-14 09:00:00+03'::timestamptz, '2027-01-14 10:30:00+03'::timestamptz, TRUE),
        ('OVB', 'SVX', '2027-01-16 12:00:00+07'::timestamptz, '2027-01-16 14:00:00+05'::timestamptz, FALSE)
) AS v(from_code, to_code, departure_date, arrival_date, is_cancelled)
JOIN airports a_from ON a_from.code_iata = v.from_code
JOIN airports a_to ON a_to.code_iata = v.to_code
WHERE NOT EXISTS (
    SELECT 1
    FROM flights f
    WHERE f.departure_date = v.departure_date
      AND f.arrival_date = v.arrival_date
      AND f.from_airport_id = a_from.id
      AND f.to_airport_id = a_to.id
);

-- Билеты: по 3 места на каждый рейс
INSERT INTO tickets (price, seat, flight_id, reservation_status)
SELECT
    v.price,
    v.seat,
    f.id,
    v.reservation_status
FROM (
    VALUES
        ('LED', 'SVO', '2027-01-10 08:00:00+03'::timestamptz, '10A', 5500.00::numeric, 'AVAILABLE'),
        ('LED', 'SVO', '2027-01-10 08:00:00+03'::timestamptz, '10B', 5500.00::numeric, 'AVAILABLE'),
        ('LED', 'SVO', '2027-01-10 08:00:00+03'::timestamptz, '10C', 6200.00::numeric, 'RESERVED'),

        ('LED', 'SVO', '2027-01-10 14:00:00+03'::timestamptz, '12A', 4800.00::numeric, 'AVAILABLE'),
        ('LED', 'SVO', '2027-01-10 14:00:00+03'::timestamptz, '12B', 4800.00::numeric, 'AVAILABLE'),
        ('LED', 'SVO', '2027-01-10 14:00:00+03'::timestamptz, '12C', 5100.00::numeric, 'AVAILABLE'),

        ('SVO', 'LED', '2027-01-11 18:00:00+03'::timestamptz, '14A', 6000.00::numeric, 'AVAILABLE'),
        ('SVO', 'LED', '2027-01-11 18:00:00+03'::timestamptz, '14B', 6000.00::numeric, 'RESERVATION_PENDING'),
        ('SVO', 'LED', '2027-01-11 18:00:00+03'::timestamptz, '14C', 6500.00::numeric, 'AVAILABLE'),

        ('LED', 'AER', '2027-01-12 07:00:00+03'::timestamptz, '16A', 8500.00::numeric, 'AVAILABLE'),
        ('LED', 'AER', '2027-01-12 07:00:00+03'::timestamptz, '16B', 8500.00::numeric, 'AVAILABLE'),
        ('LED', 'AER', '2027-01-12 07:00:00+03'::timestamptz, '16C', 9200.00::numeric, 'CANCELLED'),

        ('AER', 'LED', '2027-01-15 16:00:00+03'::timestamptz, '18A', 8100.00::numeric, 'AVAILABLE'),
        ('AER', 'LED', '2027-01-15 16:00:00+03'::timestamptz, '18B', 8100.00::numeric, 'AVAILABLE'),
        ('AER', 'LED', '2027-01-15 16:00:00+03'::timestamptz, '18C', 9000.00::numeric, 'AVAILABLE'),

        ('SVO', 'KZN', '2027-01-13 10:00:00+03'::timestamptz, '20A', 4200.00::numeric, 'AVAILABLE'),
        ('SVO', 'KZN', '2027-01-13 10:00:00+03'::timestamptz, '20B', 4200.00::numeric, 'AVAILABLE'),
        ('SVO', 'KZN', '2027-01-13 10:00:00+03'::timestamptz, '20C', 4700.00::numeric, 'AVAILABLE'),

        ('LED', 'KGD', '2027-01-14 09:00:00+03'::timestamptz, '22A', 3900.00::numeric, 'AVAILABLE'),
        ('LED', 'KGD', '2027-01-14 09:00:00+03'::timestamptz, '22B', 3900.00::numeric, 'AVAILABLE'),
        ('LED', 'KGD', '2027-01-14 09:00:00+03'::timestamptz, '22C', 4500.00::numeric, 'AVAILABLE'),

        ('OVB', 'SVX', '2027-01-16 12:00:00+07'::timestamptz, '24A', 7200.00::numeric, 'AVAILABLE'),
        ('OVB', 'SVX', '2027-01-16 12:00:00+07'::timestamptz, '24B', 7200.00::numeric, 'AVAILABLE'),
        ('OVB', 'SVX', '2027-01-16 12:00:00+07'::timestamptz, '24C', 7800.00::numeric, 'AVAILABLE')
) AS v(from_code, to_code, departure_date, seat, price, reservation_status)
JOIN airports a_from ON a_from.code_iata = v.from_code
JOIN airports a_to ON a_to.code_iata = v.to_code
JOIN flights f
    ON f.from_airport_id = a_from.id
   AND f.to_airport_id = a_to.id
   AND f.departure_date = v.departure_date
WHERE NOT EXISTS (
    SELECT 1
    FROM tickets t
    WHERE t.flight_id = f.id
      AND t.seat = v.seat
);

-- +goose Down
DELETE FROM tickets
WHERE flight_id IN (
    SELECT f.id
    FROM flights f
    WHERE f.departure_date >= '2027-01-10 00:00:00+03'::timestamptz
      AND f.departure_date < '2027-01-17 00:00:00+03'::timestamptz
);

DELETE FROM flights
WHERE departure_date >= '2027-01-10 00:00:00+03'::timestamptz
  AND departure_date < '2027-01-17 00:00:00+03'::timestamptz;

DELETE FROM airports
WHERE code_iata IN ('LED', 'SVO', 'VKO', 'AER', 'KZN', 'OVB', 'KGD', 'SVX');
