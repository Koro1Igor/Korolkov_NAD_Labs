# Launch Vehicle — Лабораторная 1 на Go + Gin

Проект сделан по структуре методички: `repository -> handler -> api`, Gin, стандартные Go HTML-шаблоны через `LoadHTMLGlob`, статика через `r.Static`.

## Тема и поля

Тема: вывод спутников на геостационарную орбиту.

Сущность в коде: `LaunchVehicle` / `launchVehicle` / `launchVehicles`.

Два числовых поля по теме:
- `PayloadKg` — полезная нагрузка, кг;
- `SeaLevelThrustKN` — тяга на уровне моря, кН.

Фильтрация выполняется по полезной нагрузке.

## Запуск

```bash
cd orbit-lab1-go
go mod tidy
go run ./cmd/app
```

Открыть:
- http://localhost:8080/feed
- http://localhost:8080/add
- http://localhost:8080/launch-vehicles

## Три GET-страницы

1. `GET /feed?id=1`
   - ID launch vehicle передаётся query-параметром.
   - `GET /feed?id=1&next=true` открывает следующую опубликованную запись.
   - `/feed` без ID показывает первую опубликованную запись.
2. `GET /add`
   - показывает единственную запись в статусе `draft`.
   - поля можно заполнять, но сохранение в ЛР1 не реализовано.
3. `GET /launch-vehicles?payload_min=5000&payload_max=20000`
   - серверная фильтрация по диапазону `PayloadKg`;
   - значения обоих range-input сохраняются после запроса.

## Коллекция и статусы

Коллекция находится в `internal/app/repository/repository.go`.

Статусы:
- `published` — отображается в ленте и плитке;
- `draft` — ровно одна запись, отображается на странице добавления;
- `deleted` — остаётся в коллекции, но в интерфейсе не отображается.

Лайки хранятся внутри каждого `LaunchVehicle` как `Likes []int`, то есть массив ID пользователей. Количество лайков вычисляется в handler через `len(launchVehicle.Likes)`.

## MinIO

Проект ожидает bucket `service-media` на `http://localhost:9000` с публичным чтением.

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

URL изображения и видео хранятся в модели двумя отдельными полями `ImageURL` и `VideoURL`.

Использование:
- `feed.html`: видео + poster;
- `add.html`: поля выбора изображения и видео;
- `launch-vehicles.html`: изображение каждой карточки.

## Что показать в Network

1. `/feed?id=1`
2. `/add`
3. `/launch-vehicles?payload_min=5000&payload_max=20000`

Отдельно можно показать `/feed?id=1&next=true` для кнопки «Следующий».

## CSS / дизайн

Файл: `resources/css/style.css`.

Референс — SpaceX:
- фон `#000000`;
- текст `#FFFFFF`;
- дополнительный текст `#A3A3A3`;
- кнопка прозрачная/чёрная с белой рамкой, на hover инвертируется;
- карточки чёрные, прямоугольные, без скругления;
- навигация чёрная, активный пункт белый, неактивные серые.

## Важно

В проекте нет БД, JavaScript и сохранения данных. Логика фильтрации и переходов выполняется на сервере.
