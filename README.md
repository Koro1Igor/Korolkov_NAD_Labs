# Лабораторная работа 3

REST-сервис услуг вывода спутников на геостационарную орбиту. Проект использует Gin, PostgreSQL через GORM и MinIO для изображений и видео. Структура повторяет ЛР2: обработчики, модели, репозиторий и сериализаторы находятся в отдельных файлах без дополнительных архитектурных слоёв.

## Структура

```text
cmd/app/main.go
internal/api/server.go
internal/app/handler/handler.go
internal/app/model/model.go
internal/app/repository/repository.go
internal/app/serializer/serializer.go
resources/css/style.css
resources/defaults/default-launch-vehicle.jpg
resources/defaults/default-launch-vehicle.mp4
sql/seed.sql
templates/add.html
templates/feed.html
templates/launch-vehicles.html
docker-compose.yml
go.mod
go.sum
```

Существующие материалы сохранены в `postman/` и `docs/`.

## Запуск

```bash
docker compose up -d
go run ./cmd/app
```

Приложение: `http://localhost:8080`  
Adminer: `http://localhost:8081`  
MinIO API: `http://localhost:9000`  
MinIO Console: `http://localhost:9001`

Параметры PostgreSQL для Adminer:

| Параметр | Значение |
|---|---|
| System | PostgreSQL |
| Server | postgres |
| Username | orbit |
| Password | orbit12345 |
| Database | orbitlab |

MinIO: пользователь `minioadmin`, пароль `minioadmin123`, bucket `service-media`. Таблицы создаются GORM. После первого запуска можно выполнить `sql/seed.sql` через Adminer.

## Текущий creator

Текущий пользователь задаётся singleton-функцией `model.CurrentCreatorID()`. Константа `model.CurrentCreatorIDValue` равна `1`. Значение creator не принимается от клиента.

При создании и изменении клиент не может передавать `id`, `creator_id`, `moderator_id`, `status`, системные даты и другие неизвестные поля. JSON разбирается с запретом неизвестных полей; multipart допускает только `name`, `image` и `video`.

## REST API

В приложении ровно 10 бизнес-методов. Все они находятся под префиксом `/api`.

| № | Метод | URL | Назначение |
|---:|---|---|---|
| 1 | GET | `/api/launch-vehicles?payload_min=0&payload_max=30000` | Получить список услуг со статусом `published`; параметры нагрузки необязательны, в ответе есть `is_creator` 0/1 |
| 2 | GET | `/api/feed` | Получить элемент ленты; варианты запроса описаны ниже, в ответе есть `liked` и `is_creator` |
| 3 | GET | `/api/launch-vehicles/draft` | Получить единственный draft текущего creator, без id в URL |
| 4 | POST | `/api/launch-vehicles` | Создать draft из multipart-полей `name`, `image`, `video` |
| 5 | PUT | `/api/launch-vehicles/draft` | Опубликовать draft текущего creator |
| 6 | DELETE | `/api/launch-vehicles/:id` | Выполнить soft delete собственной услуги, например `/api/launch-vehicles/1` |
| 7 | POST | `/api/launch-vehicles/:id/like` | Установить лайк значением `0` или `1`, например `/api/launch-vehicles/1/like` |
| 8 | POST | `/api/users` | Зарегистрировать пользователя |
| 9 | POST | `/api/users/login` | Выполнить вход |
| 10 | POST | `/api/users/logout` | Выполнить выход |

### Лента

`GET /api/feed` — один бизнес-метод с двумя режимами работы.

Без параметров метод возвращает первый элемент со статусом `published`. Это начальное состояние ленты: клиенту ещё не известен идентификатор текущей карточки, поэтому передавать `id` при первом открытии нечего.

`GET /api/feed?id=<id>&next=true`, например `GET /api/feed?id=1&next=true`, возвращает следующий опубликованный элемент после услуги с указанным `id`. Переход по ленте циклический. Это два варианта вызова одного бизнес-метода `GET /api/feed`, а не два разных метода.

| Вызов | Что происходит |
|---|---|
| `GET /api/feed` | Открывается первая опубликованная карточка ленты |
| `GET /api/feed?id=1` | Возвращается опубликованная карточка с `id = 1` |
| `GET /api/feed?id=1&next=true` | Возвращается следующая опубликованная карточка; после последней выполняется переход к первой |

Если передать `next=true` без `id`, сервер вернёт `400 Bad Request`, потому что невозможно определить, после какой карточки нужно продолжать ленту.

В ответе ленты присутствуют два разных целочисленных признака:

- `liked`: `1`, если текущий пользователь поставил услуге лайк, иначе `0`;
- `is_creator`: `1`, если `creator_id` услуги совпадает с `model.CurrentCreatorID()`, иначе `0`.

Пример фрагмента ответа:

```json
{
  "id": 1,
  "creator_id": 2,
  "liked": 1,
  "is_creator": 0
}
```

Здесь пользователь поставил лайк (`liked = 1`), но не является автором услуги (`is_creator = 0`). Эти признаки независимы друг от друга.

В Postman лента представлена одним запросом `05 GET feed (id and next are optional)`. Для первого режима параметры `id` и `next` выключены, для перехода к следующей карточке их нужно включить. Поэтому коллекция содержит ровно 10 запросов — по одному на каждый бизнес-метод.

### Создание услуги

`POST /api/launch-vehicles`, тип тела `multipart/form-data`:

| Поле | Тип | Ограничение |
|---|---|---|
| `name` | text | обязательное, до 120 символов |
| `image` | file | jpg, jpeg, png или webp, до 10 МБ |
| `video` | file | mp4, mov или webm, до 60 МБ |

Файлы загружаются в MinIO. В БД сохраняются латинские сгенерированные имена вида `images/image-1791840000000000000-a1b2c3d4e5f6.jpg` и `videos/video-1791840000000000000-a1b2c3d4e5f6.mp4`.

### Публикация

`PUT /api/launch-vehicles/draft`:

```json
{
  "short_description": "Тестовая ракета-носитель",
  "payload_kg": 12000,
  "sea_level_thrust_kn": 9000
}
```

Статус меняется с `draft` на `published`, а `formed_at` формируется сервером.

### Лайк

`POST /api/launch-vehicles/:id/like`:

```json
{"value": 1}
```

Разрешены только значения `0` и `1`. Лайк можно менять только для опубликованной услуги.

### Регистрация и вход

Регистрация:

```json
{"name":"New User","email":"new@example.com","password":"123456"}
```

Вход:

```json
{"email":"igor@example.com","password":"igor123"}
```

Аутентификация не меняет фиксированного creator лабораторной работы.

## Статусы

- `draft` устанавливается только при создании.
- `published` устанавливается только при публикации собственного draft.
- `deleted` устанавливается soft delete только собственной услуги; `completed_at` формируется сервером.
- Записи `deleted` не возвращаются списком, в ленте, по id и не принимают лайки.

## Таблицы БД

### `users`

| Поле | Тип | Ограничения |
|---|---|---|
| `id` | BIGSERIAL | PK |
| `name` | VARCHAR(100) | NOT NULL |
| `email` | VARCHAR(255) | NOT NULL, UNIQUE |
| `password` | VARCHAR(255) | NOT NULL |

### `launch_vehicles`

| Поле | Тип | Ограничения |
|---|---|---|
| `id` | BIGSERIAL | PK |
| `name` | VARCHAR(120) | NOT NULL |
| `creator_id` | BIGINT | FK → users.id, NOT NULL |
| `moderator_id` | BIGINT | FK → users.id, NULL |
| `status` | VARCHAR(16) | NOT NULL |
| `short_description` | VARCHAR(500) | NULL |
| `image_name` | VARCHAR(255) | имя объекта MinIO, NULL |
| `video_name` | VARCHAR(255) | имя объекта MinIO, NULL |
| `payload_kg` | INTEGER | NULL |
| `sea_level_thrust_kn` | INTEGER | NULL |
| `created_at` | TIMESTAMP | NOT NULL |
| `formed_at` | TIMESTAMP | NULL |
| `completed_at` | TIMESTAMP | NULL |

Частичный уникальный индекс разрешает не более одного `draft` на `creator_id`.

### `likes`

| Поле | Тип | Ограничения |
|---|---|---|
| `id` | BIGSERIAL | PK |
| `user_id` | BIGINT | FK → users.id, NOT NULL |
| `launch_vehicle_id` | BIGINT | FK → launch_vehicles.id, NOT NULL |

Пара `(user_id, launch_vehicle_id)` уникальна.

## ORM и сериализация

Все обращения к PostgreSQL выполняются через GORM в `internal/app/repository/repository.go`. JSON-ответы формируются отдельно в `internal/app/serializer/serializer.go`; структуры БД напрямую клиенту не возвращаются.

## Материалы

- Postman: `postman/lab3.postman_collection.json`
- Диаграмма классов: `docs/lab3_class_diagram.drawio`
- Стандартные медиа: `resources/defaults/`
