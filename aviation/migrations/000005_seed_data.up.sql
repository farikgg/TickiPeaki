-- Сидинг тестовых данных. Идемпотентно: повторный запуск ничего не сломает.
-- Все пароли пользователей — "password123" (bcrypt hash ниже).

BEGIN;

-- Пользователи (пароль для всех: "password123")
INSERT INTO users (username, password, role) VALUES
  ('admin',   '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LnkQ4dN/sDi', 'admin'),
  ('rafi',    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LnkQ4dN/sDi', 'user'),
  ('aigerim', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LnkQ4dN/sDi', 'user')
ON CONFLICT DO NOTHING;

-- Пассажиры
INSERT INTO passengers (full_name, email, phone, passport_num) VALUES
  ('Рафаэль Ахметов', 'rafi@example.com',    '+77011234567', 'N12345678'),
  ('Айгерим Бекова',  'aigerim@example.com', '+77029876543', 'N87654321'),
  ('Данияр Сейткали', 'daniyar@example.com', '+77031112233', 'N11223344'),
  ('Мадина Жакупова', 'madina@example.com',  '+77044445566', 'N44556677'),
  ('Арман Нурланов',  'arman@example.com',   '+77055557788', 'N55667788')
ON CONFLICT DO NOTHING;

-- Рейсы (без цены и количества мест — место и цена живут в seats)
INSERT INTO flights (flight_number, origin, destination, carrier, departure_time, arrival_time) VALUES
  ('KC-101', 'ALA', 'NQZ', 'Air Astana',    '2025-07-01 06:00:00+06', '2025-07-01 07:30:00+06'),
  ('KC-202', 'NQZ', 'ALA', 'Air Astana',    '2025-07-01 09:00:00+06', '2025-07-01 10:30:00+06'),
  ('KC-305', 'ALA', 'CIT', 'Air Astana',    '2025-07-05 08:00:00+06', '2025-07-05 09:10:00+06'),
  ('DV-411', 'ALA', 'NQZ', 'SCAT Airlines', '2025-07-10 14:00:00+06', '2025-07-10 15:30:00+06'),
  ('KC-520', 'NQZ', 'AKX', 'Air Astana',    '2025-07-15 07:30:00+06', '2025-07-15 09:45:00+06'),
  ('DV-633', 'ALA', 'URA', 'SCAT Airlines', '2025-08-01 11:00:00+06', '2025-08-01 13:30:00+06'),
  ('KC-744', 'CIT', 'ALA', 'Air Astana',    '2025-08-10 16:00:00+06', '2025-08-10 17:10:00+06')
ON CONFLICT DO NOTHING;

-- Места по каждому рейсу (привязываем по flight_number чтобы не зависеть от автоинкремента)
INSERT INTO seats (flight_id, seat_number, class, price, status)
SELECT f.id, v.seat_number, v.class, v.price, v.status
FROM (VALUES
  -- рейс KC-101 (ALA→NQZ)
  ('KC-101', '1A', 'first',    55000.00, 'available'),
  ('KC-101', '1B', 'first',    55000.00, 'available'),
  ('KC-101', '3A', 'business', 35000.00, 'available'),
  ('KC-101', '3B', 'business', 35000.00, 'available'),
  ('KC-101', '6A', 'economy',  25000.00, 'booked'),
  ('KC-101', '6B', 'economy',  25000.00, 'available'),
  ('KC-101', '7A', 'economy',  25000.00, 'available'),
  -- рейс KC-202 (NQZ→ALA)
  ('KC-202', '1A', 'first',    50000.00, 'available'),
  ('KC-202', '3A', 'business', 30000.00, 'available'),
  ('KC-202', '6A', 'economy',  22000.00, 'booked'),
  ('KC-202', '6B', 'economy',  22000.00, 'available'),
  -- рейс KC-305 (ALA→CIT)
  ('KC-305', '1A', 'first',    45000.00, 'available'),
  ('KC-305', '3A', 'business', 28000.00, 'available'),
  ('KC-305', '6A', 'economy',  18000.00, 'available'),
  -- рейс DV-411 (ALA→NQZ)
  ('DV-411', '1A', 'first',    48000.00, 'available'),
  ('DV-411', '3A', 'business', 32000.00, 'available'),
  ('DV-411', '6A', 'economy',  19500.00, 'available'),
  -- рейс KC-520 (NQZ→AKX)
  ('KC-520', '1A', 'first',    65000.00, 'available'),
  ('KC-520', '3A', 'business', 42000.00, 'available'),
  ('KC-520', '6A', 'economy',  31000.00, 'available'),
  -- рейс DV-633 (ALA→URA)
  ('DV-633', '1A', 'first',    60000.00, 'available'),
  ('DV-633', '3A', 'business', 38000.00, 'available'),
  ('DV-633', '6A', 'economy',  28000.00, 'available'),
  -- рейс KC-744 (CIT→ALA)
  ('KC-744', '1A', 'first',    42000.00, 'available'),
  ('KC-744', '3A', 'business', 26000.00, 'available'),
  ('KC-744', '6A', 'economy',  17500.00, 'available')
) AS v(flight_number, seat_number, class, price, status)
JOIN flights f ON f.flight_number = v.flight_number
ON CONFLICT (flight_id, seat_number) DO NOTHING;

-- Билеты — связываем пассажира с конкретным забронированным местом
INSERT INTO tickets (flight_id, passenger_id, seat_id, status, booked_at)
SELECT f.id, p.id, s.id, v.status, v.booked_at::timestamptz
FROM (VALUES
  ('KC-101', 'rafi@example.com',    '6A', 'paid',     '2025-06-01 10:00:00+06'),
  ('KC-202', 'daniyar@example.com', '6A', 'reserved', '2025-06-02 09:00:00+06')
) AS v(flight_number, email, seat_number, status, booked_at)
JOIN flights    f ON f.flight_number = v.flight_number
JOIN passengers p ON p.email = v.email
JOIN seats      s ON s.flight_id = f.id AND s.seat_number = v.seat_number
WHERE NOT EXISTS (
  SELECT 1 FROM tickets t WHERE t.seat_id = s.id
);

COMMIT;
