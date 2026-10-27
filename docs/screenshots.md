# Порядок скриншотов ЛР4

## Скрин 1 — аутентификация в Swagger в режиме инкогнито

Открыть `http://localhost:8080/swagger/index.html` в приватном окне.

Выполнить `POST /api/users/login`:

```json
{
  "email": "igor@example.com",
  "password": "igor123"
}
```

На скрине должны быть видны URL метода, `200 OK` и `access_token`.

## Скрин 2 — авторизация JWT в Swagger

Нажать `Authorize` и вставить JWT. Swagger сам добавит `Bearer` согласно security scheme.

На скрине показать активную авторизацию и BearerAuth.

## Скрин 3 — GET списка в Swagger

Выполнить `GET /api/launch-vehicles?payload_min=0&payload_max=30000`.

Показать `200 OK`, список опубликованных услуг и `is_creator`.

## Скрин 4 — GET списка в Postman без JWT

В Postman выполнить тот же `GET /api/launch-vehicles` без Authorization.

Ожидается `200 OK`.

## Скрин 5 — GET списка в Postman с JWT

Повторить GET с заголовком:

```text
Authorization: Bearer <JWT>
```

Ожидается `200 OK`. Это показывает, что чтение работает независимо от наличия JWT; при JWT корректно рассчитывается `is_creator`.

## Скрин 6 — PUT публикации от гостя

Без Authorization выполнить:

```http
PUT /api/launch-vehicles/draft
```

```json
{
  "short_description": "Proton M — ракета-носитель для вывода полезной нагрузки на орбиту.",
  "payload_kg": 23000,
  "sea_level_thrust_kn": 10000
}
```

Ожидается `401 Unauthorized`.

## Скрин 7 — PUT публикации от создателя

Повторить тот же PUT с `Authorization: Bearer <JWT>`.

Ожидается `200 OK`, `status=published`, заполненные `payload_kg`, `sea_level_thrust_kn`, `formed_at`.

Перед этим у пользователя должен существовать draft. При необходимости создать его `POST /api/launch-vehicles` с multipart `name`, `image`, `video`.

## Скрин 8 — Redis blacklist

Выполнить `POST /api/users/logout` с тем же JWT, затем:

```bash
docker exec -it orbit-lab4-redis redis-cli KEYS "orbitlab.jwt.*"
```

Для найденного ключа:

```bash
docker exec -it orbit-lab4-redis redis-cli GET <ключ>
docker exec -it orbit-lab4-redis redis-cli TTL <ключ>
```

Значение — `user_id`, TTL — оставшееся время жизни JWT. Повторный защищённый запрос с этим JWT должен вернуть `401`.

## Скрин 9 — StarUML sequence diagram

В Sequence Diagram сообщения должны использовать операции доменов из Class Diagram. В свойствах `Message` показать заполненное свойство `signature`, связанное с соответствующей операцией домена, а не просто вручную введённый текст сообщения.
