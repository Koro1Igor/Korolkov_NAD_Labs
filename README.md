# Orbita — Лабораторная 1 на Go + Gin

Проект сделан по структуре методички: `repository -> handler -> api`, Gin, стандартные Go HTML-шаблоны через `LoadHTMLGlob`, статика через `r.Static`.

## Запуск

```bash
cd orbit-lab1-go
go mod tidy
go run ./cmd/app
```

Открыть:

- http://localhost:8080/feed
- http://localhost:8080/add
- http://localhost:8080/services

## Три GET страницы

1. `GET /feed?id=1`
   - ID услуги передаётся query-параметром.
   - `GET /feed?id=1&next=true` открывает следующую опубликованную услугу.
   - переход по вкладке «Лента» использует `/feed` без ID и показывает первую опубликованную услугу.
2. `GET /add`
   - показывает единственную услугу в статусе `draft`.
3. `GET /services?min_payload=8300`
   - серверная фильтрация по числовому полю `MaxPayloadKg`.
   - значение фильтра возвращается в `value` input после запроса.

## Коллекция

Коллекция находится в `internal/app/repository/repository.go`.

Статусы:
- `published` — видны в ленте и плитке;
- `draft` — ровно одна услуга, видна на странице добавления;
- `deleted` — существует в коллекции, но в интерфейс не попадает.

Массив внутри услуги только один — `Likes []int`, содержащий ID пользователей. Количество лайков вычисляется в handler через `len(service.Likes)`.

## MinIO

Проект ожидает bucket `service-media` на `http://localhost:9000` с публичным чтением.

Загрузить файлы:

```text
service-media/
  images/
    falcon_9.jpeg
    falcon_heavy.jpeg
    ariane_64.jpeg
    ariane_62.jpeg
  videos/
    falcon_9.mp4
    falcon_heavy.mp4
    ariane_64.mp4
    ariane_62.mp4
```

URL изображений и видео хранятся в коллекции в двух отдельных полях `ImageURL` и `VideoURL`.

Они используются:
- `feed.html`: `poster="{{ .service.ImageURL }}"` и `<source src="{{ .service.VideoURL }}">`;
- `add.html`: изображение и два поля URL;
- `services.html`: `<img src="{{ .ImageURL }}">`.

## Что показать в Network

1. `/feed?id=1`
2. `/add`
3. `/services?min_payload=8300`

Отдельно можно нажать «Следующий» и показать `/feed?id=1&next=true`.

В Response HTML будут URL MinIO вида:

```text
http://localhost:9000/service-media/images/falcon_9.jpeg
http://localhost:9000/service-media/videos/falcon_9.mp4
```

## CSS / дизайн

Файл: `resources/css/style.css`.

Референс — SpaceX. В комментарии CSS явно перечислены:
- `#000000`
- `#FFFFFF`
- `#A3A3A3`
- hover карточек и кнопок
- прямоугольная форма карточек без скругления

## Важно

В проекте нет:
- БД;
- JavaScript;
- POST/PUT/PATCH/DELETE;
- сохранения данных на странице «Добавление».
