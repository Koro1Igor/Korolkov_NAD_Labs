INSERT INTO users (name, email, password)
VALUES
    ('Igor', 'igor@example.com', 'igor123'),
    ('Anna', 'anna@example.com', 'anna123'),
    ('Max', 'max@example.com', 'max123'),
    ('User 4', 'user4@example.com', 'user4123'),
    ('User 5', 'user5@example.com', 'user5123'),
    ('User 6', 'user6@example.com', 'user6123')
ON CONFLICT (email)
DO UPDATE SET
    name = EXCLUDED.name,
    password = EXCLUDED.password;

DELETE FROM likes;
DELETE FROM launch_vehicles;

INSERT INTO launch_vehicles (
    name,
    short_description,
    status,
    image_url,
    video_url,
    payload_kg,
    sea_level_thrust_kn,
    created_at,
    formed_at,
    creator_id
)
VALUES
(
    'Falcon 9',
    'Falcon 9 — ракета-носитель SpaceX. Максимальная полезная нагрузка на геопереходную орбиту — 8 300 кг.',
    'published',
    'http://localhost:9000/service-media/images/falcon9.jpeg',
    'http://localhost:9000/service-media/videos/falcon9_launch.mp4',
    8300,
    7607,
    NOW() - INTERVAL '5 days',
    NOW() - INTERVAL '4 days',
    (SELECT id FROM users WHERE email = 'anna@example.com')
),
(
    'Falcon Heavy',
    'Falcon Heavy — тяжёлая ракета-носитель SpaceX. Максимальная полезная нагрузка на геопереходную орбиту — 26 700 кг.',
    'published',
    'http://localhost:9000/service-media/images/falcon_heavy.jpeg',
    'http://localhost:9000/service-media/videos/falcon_heavy_launch.mp4',
    26700,
    22819,
    NOW() - INTERVAL '4 days',
    NOW() - INTERVAL '3 days',
    (SELECT id FROM users WHERE email = 'max@example.com')
),
(
    'Ariane 64',
    'Ariane 64 — четырёхбустерная конфигурация Ariane 6. Максимальная полезная нагрузка на стандартную геопереходную орбиту — 11 500 кг.',
    'published',
    'http://localhost:9000/service-media/images/ariane64.jpeg',
    'http://localhost:9000/service-media/videos/ariane64_launch.mp4',
    11500,
    1370,
    NOW() - INTERVAL '3 days',
    NOW() - INTERVAL '2 days',
    (SELECT id FROM users WHERE email = 'user4@example.com')
),
(
    'Soyuz MS16',
    NULL,
    'draft',
    NULL,
    NULL,
    NULL,
    NULL,
    NOW() - INTERVAL '1 day',
    NULL,
    (SELECT id FROM users WHERE email = 'igor@example.com')
),
(
    'Soyuz MS12',
    'Удалённая запись launch vehicle. Она остаётся в БД, но не отображается в интерфейсе.',
    'deleted',
    NULL,
    NULL,
    8300,
    7607,
    NOW() - INTERVAL '10 days',
    NOW() - INTERVAL '9 days',
    (SELECT id FROM users WHERE email = 'user5@example.com')
);

INSERT INTO likes (user_id, launch_vehicle_id)
SELECT
    (SELECT id FROM users WHERE email = 'igor@example.com'),
    id
FROM launch_vehicles
WHERE name = 'Falcon 9';

INSERT INTO likes (user_id, launch_vehicle_id)
SELECT
    (SELECT id FROM users WHERE email = 'anna@example.com'),
    id
FROM launch_vehicles
WHERE name = 'Falcon 9';

INSERT INTO likes (user_id, launch_vehicle_id)
SELECT
    (SELECT id FROM users WHERE email = 'max@example.com'),
    id
FROM launch_vehicles
WHERE name = 'Falcon 9';

INSERT INTO likes (user_id, launch_vehicle_id)
SELECT
    (SELECT id FROM users WHERE email = 'igor@example.com'),
    id
FROM launch_vehicles
WHERE name = 'Falcon Heavy';

INSERT INTO likes (user_id, launch_vehicle_id)
SELECT
    (SELECT id FROM users WHERE email = 'user4@example.com'),
    id
FROM launch_vehicles
WHERE name = 'Falcon Heavy';

INSERT INTO likes (user_id, launch_vehicle_id)
SELECT
    (SELECT id FROM users WHERE email = 'user5@example.com'),
    id
FROM launch_vehicles
WHERE name = 'Falcon Heavy';

INSERT INTO likes (user_id, launch_vehicle_id)
SELECT
    (SELECT id FROM users WHERE email = 'user6@example.com'),
    id
FROM launch_vehicles
WHERE name = 'Falcon Heavy';

INSERT INTO likes (user_id, launch_vehicle_id)
SELECT
    (SELECT id FROM users WHERE email = 'anna@example.com'),
    id
FROM launch_vehicles
WHERE name = 'Ariane 64';

INSERT INTO likes (user_id, launch_vehicle_id)
SELECT
    (SELECT id FROM users WHERE email = 'user5@example.com'),
    id
FROM launch_vehicles
WHERE name = 'Ariane 64';