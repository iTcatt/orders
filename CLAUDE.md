# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

This project uses [mise](https://mise.jdx.dev/) to manage tools and tasks.

```bash
mise install          # Install all dev tools (Go, golangci-lint, mockery, Goose, vegeta)
mise run start        # Run the service locally (cmd/service/main.go)
mise run test         # Run all tests with -v
mise run fmt          # Format code with golangci-lint fmt
mise run lint         # Run golangci-lint
mise run check        # fmt + lint
mise run httptest     # Run HTTP integration tests via ijhttp
```

To run a single test:
```bash
go test -v ./internal/usecase/product/... -run TestName
```

### Infrastructure

```bash
mise run infra:up     # Start dev infrastructure (PostgreSQL, Prometheus, Grafana)
mise run infra:down   # Stop dev infrastructure
```

Dev workflow: `mise run infra:up` → `mise run start`

### Production

```bash
mise run prod:up      # Build and start the full production stack (app + all dependencies)
mise run prod:down    # Stop production stack
mise run prod:logs    # Follow production logs
```

Migrations run automatically via `goose-migrations` container on every `prod:up` / `infra:up`.

## Architecture

Clean three-layer architecture:

```
HTTP Handlers (internal/api/)
       ↓  interfaces via internal/api/product/deps.go
Use Cases (internal/usecase/)
       ↓  interfaces via internal/usecase/product/deps.go
Storage (internal/storage/)
       ↓
PostgreSQL (internal/infra/postgres/)
```

**Layer responsibilities:**
- `internal/api/product/` — Standard `net/http` handlers, request/response DTOs, input validation (go-playground/validator). Custom `Router` in `api/router.go` wraps `http.ServeMux`. Middleware stack in `api/middleware.go`: recovery, requestID, CORS, structured logging (slog + devslog), Prometheus metrics.
- `internal/usecase/product/` — Business logic. Generates product IDs as UUID v7 strings. Converts storage errors to typed use case errors (`internal/usecase/errors.go`).
- `internal/storage/products/` — PostgreSQL queries built with Masterminds/squirrel. Generic helpers live in `pkg/sqlp/sqlp.go`.
- `internal/infra/postgres/` — Connection pool setup (40 max open, 20 max idle, 10m idle timeout, 20m lifetime) using pgx/v5 + sqlx.
- `pkg/api/api.go` — Shared HTTP response helpers (e.g. `SendInternalError`).

**Dependency injection:** each layer defines its own interfaces (`deps.go`) that the layer above depends on. Mocks are generated with mockery (config: `.mockery.yml`, template: testify).

## Key Conventions

- **Price** is stored in smallest currency unit (kopecks/cents) as `BIGINT`.
- **Product IDs** are UUID v7 strings generated in the use case layer; stored as `UUID` in PostgreSQL.
- **Environment** variables: in dev the service loads `.env` via `godotenv.Load()` (required, must exist). In production all vars are injected via `env_file: .env.prod` in docker-compose — no file mount needed. Both `.env` and `.env.prod` are gitignored.
- **Logging format**: JSON in production, pretty via `devslog` in dev. Service checks `env` env var; worker checks `APP_ENV`.
- **Frontend**: static `frontend/index.html` is served at `GET /` by the service.

## Mocks

Regenerate mocks after changing any interface in `./internal`:
```bash
mockery
```

Mocks are placed in `mocks/` subdirectories alongside the interfaces they mock.

## Monitoring

- Prometheus scrapes `/metrics` (port 8081) every 5s.
- Grafana at `localhost:3000` — dashboards in `monitoring/dashboards/`.
- PostgreSQL metrics via postgres-exporter at port 9187.

## HTTP Test Files

Runnable API examples live in `http/products/` (`.http` format). Run via `ijhttp` with env file `http/http-client.env.json`.
