# avia

Репозиторий микросервиса авиабилетов.

Авиа сервис, PostgreSQL и Kafka поднимаются вместе:

1. Скопируйте `config/docker/pg.env.example` в `config/docker/pg.env`.
2. Если сервер доступен по другому публичному IP или DNS, укажите его в `KAFKA_ADVERTISED_HOST` в `pg.env` и откройте входящий TCP порт из `KAFKA_PORT` (по умолчанию `9092`).
3. Запустите `make docker-up`.

Веб-интерфейс Kafdrop после запуска будет доступен по адресу `http://localhost:9000` (на удалённом сервере — `http://<адрес-сервера>:9000`, порт задаётся через `KAFDROP_PORT`).

Внутри Compose Kafka доступна по `kafka:29092`, с других серверов — по `KAFKA_ADVERTISED_HOST:KAFKA_PORT` (по умолчанию `45.89.229.4:9092`). В настройках других сервисов указывайте этот адрес без `http://`.

Compose создаёт по одному разделу для топиков из AsyncAPI:

| Сервис | Топики |
| --- | --- |
| Авиа | `ticket.reserve`, `ticket.cancel`, `ticket.reservation-status` |
| Отели | `hotel.reserve`, `hotel.cancel`, `hotel.reservation-status` |
| Трансферы | `transfer.booking`, `transfer.booking.cancel`, `transfer.booking-status`, `transfer.cancellation-result` |

У каждого сервиса должен быть свой consumer group, чтобы он получал собственную копию событий. Подключение учебное: логин, пароль и ACL не настроены, любой доступный клиент может читать и писать во все топики. Kafdrop также работает без авторизации.
