# OrbitLab — Лабораторная работа №4

Тема: веб-сервис для публикации ракет-носителей, используемых для вывода спутников на геостационарную орбиту.

ЛР4 завершает backend для SPA: добавлены JWT-аутентификация, авторизация методов создателя, Redis blacklist для logout и Swagger UI.

## Запуск

```bash
cp .env.example .env
set -a
source .env
set +a
docker compose up -d postgres adminer minio redis
go mod tidy
go run ./cmd/app
```

Swagger UI: `http://localhost:8080/swagger/index.html`

Adminer: `http://localhost:8081`

MinIO Console: `http://localhost:9001`

## Авторизация

После `POST /api/users/login` сервер возвращает JWT. Для защищённых методов передаётся заголовок:

```text
Authorization: Bearer <access_token>
```

Пользователь для операций создателя извлекается из JWT. Поле `creator_id` клиентом не передаётся.

`POST /api/users/logout` добавляет идентификатор JWT в Redis blacklist на оставшееся время жизни токена. После logout тот же JWT больше не проходит middleware.

Проверка blacklist:

```bash
docker exec -it orbit-lab4-redis redis-cli KEYS "orbitlab.jwt.*"
docker exec -it orbit-lab4-redis redis-cli GET <ключ>
docker exec -it orbit-lab4-redis redis-cli TTL <ключ>
```

Значение blacklist-ключа — `user_id`, TTL совпадает с оставшимся временем жизни JWT.

## HTTP методы

| № | Метод | URL | Доступ | Назначение |
|---|---|---|---|---|
| 1 | GET | `/api/launch-vehicles?payload_min=0&payload_max=30000` | гость/создатель | опубликованные услуги с фильтрацией; `is_creator` 0/1 |
| 2 | GET | `/api/feed` | гость/создатель | первый опубликованный элемент ленты; `liked` и `is_creator` 0/1 |
| 3 | GET | `/api/launch-vehicles/draft` | создатель | черновик текущего пользователя |
| 4 | POST | `/api/launch-vehicles` | создатель | создать draft из `name`, `image`, `video`; файлы сохраняются в MinIO |
| 5 | PUT | `/api/launch-vehicles/draft` | создатель | публикация своего draft |
| 6 | DELETE | `/api/launch-vehicles/:id` | создатель | soft delete только своей услуги |
| 7 | POST | `/api/launch-vehicles/:id/like` | создатель | `value=1` поставить лайк, `value=0` снять |
| 8 | POST | `/api/users` | гость | регистрация |
| 9 | POST | `/api/users/login` | гость | аутентификация и выдача JWT |
| 10 | POST | `/api/users/logout` | создатель | JWT помещается в Redis blacklist |

Дополнительный режим ленты: `GET /api/feed?id=11&next=true` возвращает следующую опубликованную запись после `id=11`; при достижении конца выполняется циклический переход к первой опубликованной записи.

## Примеры JSON

### Регистрация

```json
{
  "name": "Igor",
  "email": "igor@example.com",
  "password": "igor123"
}
```

### Login

```json
{
  "email": "igor@example.com",
  "password": "igor123"
}
```

Ответ:

```json
{
  "access_token": "<jwt>",
  "token_type": "Bearer",
  "expires_at": "2026-10-27T08:00:00+03:00",
  "user": {
    "id": 1,
    "name": "Igor",
    "email": "igor@example.com"
  }
}
```

### Публикация

```json
{
  "short_description": "Ракета-носитель для вывода полезной нагрузки на орбиту.",
  "payload_kg": 23000,
  "sea_level_thrust_kn": 10000
}
```

### Лайк

```json
{
  "value": 1
}
```

## Доступ без JWT

Без JWT успешно выполняются:

- `GET /api/launch-vehicles`
- `GET /api/feed`
- `POST /api/users`
- `POST /api/users/login`

Методы создателя без JWT возвращают `401 Unauthorized`.

## Таблицы БД

### `users`

| Поле | Тип | Ограничение |
|---|---|---|
| `id` | BIGSERIAL | PK |
| `name` | VARCHAR(100) | NOT NULL |
| `email` | VARCHAR(255) | NOT NULL, UNIQUE |
| `password` | VARCHAR(255) | NOT NULL, bcrypt hash после регистрации/первого login |

### `launch_vehicles`

| Поле | Тип | Ограничение |
|---|---|---|
| `id` | BIGSERIAL | PK |
| `name` | VARCHAR(120) | NOT NULL |
| `creator_id` | BIGINT | FK → `users.id`, NOT NULL |
| `moderator_id` | BIGINT | FK → `users.id`, NULL |
| `status` | VARCHAR(16) | NOT NULL |
| `short_description` | VARCHAR(500) | NULL до публикации |
| `image_name` | VARCHAR(255) | имя объекта MinIO |
| `video_name` | VARCHAR(255) | имя объекта MinIO |
| `payload_kg` | INTEGER | NULL до публикации |
| `sea_level_thrust_kn` | INTEGER | NULL до публикации |
| `created_at` | TIMESTAMP | NOT NULL |
| `formed_at` | TIMESTAMP | NULL до публикации |
| `completed_at` | TIMESTAMP | NULL до soft delete |

Частичный уникальный индекс допускает не более одного `draft` на `creator_id`.

### `likes`

| Поле | Тип | Ограничение |
|---|---|---|
| `id` | BIGSERIAL | PK |
| `user_id` | BIGINT | FK → `users.id`, NOT NULL |
| `launch_vehicle_id` | BIGINT | FK → `launch_vehicles.id`, NOT NULL |

Пара `(user_id, launch_vehicle_id)` уникальна.

## Сервисы

- Go + Gin — HTTP API
- GORM + PostgreSQL — модели и бизнес-данные
- MinIO — изображения и короткие видео
- Redis — blacklist JWT
- Swagger UI/OpenAPI — описание и ручное тестирование API
- Adminer — просмотр PostgreSQL
