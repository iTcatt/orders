# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

This project uses [mise](https://mise.jdx.dev/) to manage tools and tasks.

```bash
mise install          # Install all dev tools (Go, golangci-lint, mockery, Goose)
mise run start        # Run the service
mise run test         # Run all tests with -v
mise run fmt          # Format code with gofumpt
mise run lint         # Run golangci-lint
mise run check        # fmt + lint
mise run migrate      # Apply Goose migrations (up)
mise run down         # Rollback migrations (down)
mise run httptest     # Run HTTP integration tests via httpyac
```

To run a single test:
```bash
go test -v ./internal/usecase/product/... -run TestName
```

To start the full infrastructure (PostgreSQL, Prometheus, Grafana):
```bash
docker compose up --build
```

## Architecture

Clean three-layer architecture:

```
HTTP Handlers (internal/api/)
       ↓  interfaces via internal/api/deps.go
Use Cases (internal/usecase/)
       ↓  interfaces via internal/usecase/product/deps.go
Storage (internal/storage/)
       ↓
PostgreSQL (internal/infra/postgres/)
```

**Layer responsibilities:**
- `internal/api/product/` — Chi router handlers, request/response DTOs, input validation (go-playground/validator). Middleware in `api/middleware.go` adds structured logging (slog + go-chi/httplog) and Prometheus metrics.
- `internal/usecase/product/` — Business logic. Generates product IDs (random uint32). Converts storage errors to typed use case errors (`internal/usecase/errors.go`).
- `internal/storage/products/` — PostgreSQL queries built with Masterminds/squirrel. Generic helpers live in `pkg/sqlp/sqlp.go`.
- `internal/infra/postgres/` — Connection pool setup (40 max open, 20 max idle) using pgx/v5 + sqlx.

**Dependency injection:** each layer defines its own interfaces (`deps.go`) that the layer above depends on. Mocks are generated with mockery (config: `.mockery.yml`, template: testify).

## Key Conventions

- **Price** is stored in smallest currency unit (kopecks/cents) as `BIGINT`.
- **Product IDs** are random `uint32` generated in the use case layer, not auto-increment from the DB.
- **Environment** variables are loaded from `.env`; `DB_URL` and `DB_PASSWORD` are required. See `mise.toml` for dev defaults.
- **Logging format**: JSON in production, pretty via `devslog` in dev (controlled by `APP_ENV`).

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

Runnable API examples live in `http/products/` (`.http` format, compatible with httpyac/REST Client).
