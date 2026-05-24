-- новая таблица мест
CREATE TABLE IF NOT EXISTS seats (
    id          SERIAL PRIMARY KEY,
    flight_id   INT NOT NULL REFERENCES flights(id) ON DELETE CASCADE,
    seat_number VARCHAR(4) NOT NULL,
    class       VARCHAR(10) NOT NULL,
    price       NUMERIC(10,2) NOT NULL,
    status      VARCHAR(10) NOT NULL DEFAULT 'available',
    UNIQUE(flight_id, seat_number)
);

ALTER TABLE flights DROP COLUMN IF EXISTS price;
ALTER TABLE flights DROP COLUMN IF EXISTS available_seats;

ALTER TABLE tickets ADD COLUMN IF NOT EXISTS seat_id INT REFERENCES seats(id);
ALTER TABLE tickets DROP COLUMN IF EXISTS seat_number;
ALTER TABLE tickets DROP COLUMN IF EXISTS class;
ALTER TABLE tickets DROP COLUMN IF EXISTS price;
