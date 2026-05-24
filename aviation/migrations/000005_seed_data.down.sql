-- Откат сидинга. Чистим только то, что вставлено в 000005_seed_data.up.sql.

BEGIN;

DELETE FROM tickets
WHERE flight_id IN (SELECT id FROM flights WHERE flight_number IN (
        'KC-101', 'KC-202', 'KC-305', 'DV-411', 'KC-520', 'DV-633', 'KC-744'))
  AND passenger_id IN (SELECT id FROM passengers WHERE email IN (
        'rafi@example.com', 'aigerim@example.com', 'daniyar@example.com',
        'madina@example.com', 'arman@example.com'));

DELETE FROM flights WHERE flight_number IN (
  'KC-101', 'KC-202', 'KC-305', 'DV-411', 'KC-520', 'DV-633', 'KC-744'
);

DELETE FROM passengers WHERE email IN (
  'rafi@example.com', 'aigerim@example.com', 'daniyar@example.com',
  'madina@example.com', 'arman@example.com'
);

DELETE FROM users WHERE username IN ('admin', 'rafi', 'aigerim');

COMMIT;
