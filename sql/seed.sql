INSERT INTO users (name, email, created_at)
VALUES
  ('Igor', 'igor@example.com', NOW()),
  ('Anna', 'anna@example.com', NOW()),
  ('Max', 'max@example.com', NOW()),
  ('User 4', 'user4@example.com', NOW()),
  ('User 5', 'user5@example.com', NOW()),
  ('User 6', 'user6@example.com', NOW())
ON CONFLICT (email) DO NOTHING;

DELETE FROM likes;
DELETE FROM launch_vehicles;

INSERT INTO launch_vehicles
(name, short_description, status, image_url, video_url, payload_kg, sea_level_thrust_kn, created_at, formed_at, creator_id)
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
    2
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
    3
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
    4
  ),
  (
    'Soyuz MS16',
    'Союз МС-16 — пилотируемый космический корабль серии «Союз МС», предназначенный для доставки экипажа на Международную космическую станцию и возвращения его на Землю.',
    'draft',
    '',
    '',
    4500,
    1370,
    NOW() - INTERVAL '1 day',
    NULL,
    1
  ),
  (
    'Soyuz MS12',
    'Удалённая запись launch vehicle. Она остается в БД, но не отображается в интерфейсе.',
    'deleted',
    '',
    '',
    8300,
    7607,
    NOW() - INTERVAL '10 days',
    NOW() - INTERVAL '9 days',
    5
  );

INSERT INTO likes (user_id, launch_vehicle_id, created_at)
SELECT 1, id, NOW() FROM launch_vehicles WHERE name = 'Falcon 9';
INSERT INTO likes (user_id, launch_vehicle_id, created_at)
SELECT 2, id, NOW() FROM launch_vehicles WHERE name = 'Falcon 9';
INSERT INTO likes (user_id, launch_vehicle_id, created_at)
SELECT 3, id, NOW() FROM launch_vehicles WHERE name = 'Falcon 9';

INSERT INTO likes (user_id, launch_vehicle_id, created_at)
SELECT 1, id, NOW() FROM launch_vehicles WHERE name = 'Falcon Heavy';
INSERT INTO likes (user_id, launch_vehicle_id, created_at)
SELECT 4, id, NOW() FROM launch_vehicles WHERE name = 'Falcon Heavy';
INSERT INTO likes (user_id, launch_vehicle_id, created_at)
SELECT 5, id, NOW() FROM launch_vehicles WHERE name = 'Falcon Heavy';
INSERT INTO likes (user_id, launch_vehicle_id, created_at)
SELECT 6, id, NOW() FROM launch_vehicles WHERE name = 'Falcon Heavy';

INSERT INTO likes (user_id, launch_vehicle_id, created_at)
SELECT 2, id, NOW() FROM launch_vehicles WHERE name = 'Ariane 64';
INSERT INTO likes (user_id, launch_vehicle_id, created_at)
SELECT 5, id, NOW() FROM launch_vehicles WHERE name = 'Ariane 64';
