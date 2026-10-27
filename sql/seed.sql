TRUNCATE TABLE likes, launch_vehicles, users RESTART IDENTITY;

INSERT INTO users (name, email, password)
VALUES
    ('Igor', 'igor@example.com', 'igor123'),
    ('Anna', 'anna@example.com', 'anna123'),
    ('Max', 'max@example.com', 'max123'),
    ('User 4', 'user4@example.com', 'user4123'),
    ('User 5', 'user5@example.com', 'user5123'),
    ('User 6', 'user6@example.com', 'user6123');

INSERT INTO launch_vehicles (
    name,
    creator_id,
    moderator_id,
    status,
    short_description,
    image_name,
    video_name,
    payload_kg,
    sea_level_thrust_kn,
    created_at,
    formed_at,
    completed_at
)
VALUES
(
    'Falcon 9',
    2,
    NULL,
    'published',
    'Falcon 9 — ракета-носитель SpaceX. Максимальная полезная нагрузка на геопереходную орбиту — 8 300 кг.',
    'images/falcon9.jpeg',
    'videos/falcon9_launch.mp4',
    8300,
    7607,
    NOW() - INTERVAL '5 days',
    NOW() - INTERVAL '4 days',
    NULL
),
(
    'Falcon Heavy',
    3,
    NULL,
    'published',
    'Falcon Heavy — тяжёлая ракета-носитель SpaceX. Максимальная полезная нагрузка на геопереходную орбиту — 26 700 кг.',
    'images/falcon_heavy.jpeg',
    'videos/falcon_heavy_launch.mp4',
    26700,
    22819,
    NOW() - INTERVAL '4 days',
    NOW() - INTERVAL '3 days',
    NULL
),
(
    'Ariane 64',
    4,
    NULL,
    'published',
    'Ariane 64 — четырёхбустерная конфигурация Ariane 6. Максимальная полезная нагрузка на стандартную геопереходную орбиту — 11 500 кг.',
    'images/ariane64.jpeg',
    'videos/ariane64_launch.mp4',
    11500,
    1370,
    NOW() - INTERVAL '3 days',
    NOW() - INTERVAL '2 days',
    NULL
),
(
    'Soyuz MS16',
    6,
    NULL,
    'draft',
    NULL,
    NULL,
    NULL,
    NULL,
    NULL,
    NOW() - INTERVAL '1 day',
    NULL,
    NULL
),
(
    'Soyuz MS12',
    5,
    NULL,
    'deleted',
    'Удалённая запись launch vehicle.',
    NULL,
    NULL,
    8300,
    7607,
    NOW() - INTERVAL '10 days',
    NOW() - INTERVAL '9 days',
    NOW() - INTERVAL '8 days'
);

INSERT INTO likes (user_id, launch_vehicle_id)
VALUES
    (1, 1),
    (2, 1),
    (3, 1),
    (1, 2),
    (4, 2),
    (5, 2),
    (6, 2),
    (2, 3),
    (5, 3);
