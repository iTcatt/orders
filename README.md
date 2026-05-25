# orders

Микросервис для управления продуктами на Go. Построен по чистой трёхслойной архитектуре: HTTP-хендлеры → use cases → хранилище.
  
## Стек

- **Go** — `net/http`, `pgx/v5`, `sqlx`, `squirrel`, `go-playground/validator`
- **PostgreSQL** — хранилище, миграции через Goose
- **Prometheus + Grafana** — метрики приложения и БД (postgres-exporter)
- **UUID v7** — идентификаторы продуктов, генерируются в use case слое
- **slog + devslog** — структурированное логирование (JSON в prod, pretty в dev)

## Особенности

- Middleware-стек: recovery, requestID, CORS, логирование запросов, Prometheus-метрики
- Цены хранятся в копейках (`BIGINT`), без float-арифметики
- Мокирование через mockery (testify-шаблон), интерфейсы изолированы по слоям
- Нагрузочное тестирование через vegeta
- Статический фронтенд — `GET /` отдаёт `frontend/index.html`
- HTTP-примеры в `http/products/` (формат `.http`, запускаются через `ijhttp`)

## Запуск

### Dev

```bash
curl https://mise.run | sh  # установить mise
mise install                # установить Go, golangci-lint, mockery, Goose, vegeta

mise run infra:up           # поднять PostgreSQL, Prometheus, Grafana
mise run start              # запустить сервис локально
```

### Prod

```bash
docker compose -f docker-compose.prod.yml up --build -d  # собрать и поднять весь стек
docker compose -f docker-compose.prod.yml logs -f        # логи
docker compose -f docker-compose.prod.yml down           # остановить
```

Миграции применяются автоматически при каждом запуске.

Grafana — `localhost:3000`, метрики сервиса — `localhost:8081/metrics`.

## Разработка

```bash
mise run test               # тесты
mise run check              # форматирование + линтер
mise run httptest           # HTTP-интеграционные тесты
```
