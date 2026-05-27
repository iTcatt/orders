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
       ↓  interfaces via internal/api/{domain}/deps.go
Use Cases (internal/usecase/)
       ↓  interfaces via internal/usecase/{domain}/deps.go
Storage (internal/storage/)
       ↓
PostgreSQL (internal/infra/postgres/)
MinIO     (internal/infra/minio/)
```

**Layer responsibilities:**
- `internal/api/product/` — Standard `net/http` handlers, request/response DTOs, input validation (go-playground/validator). Custom `Router` in `api/router.go` wraps `http.ServeMux`. Middleware stack in `api/middleware.go`: recovery, requestID, CORS, structured logging (slog + devslog), Prometheus metrics.
- `internal/api/image/` — Handles multipart file upload and image deletion. Input validation split into `validate(r)` (checks size, MIME type) and `extractInput(r, productID)` (builds `usecase.UploadImageIn`). Response DTOs in `internal/api/image/dto/`.
- `internal/usecase/product/` — Business logic. Generates product IDs as UUID v7 strings. Converts storage errors to typed use case errors (`internal/usecase/errors.go`).
- `internal/usecase/image/` — Business logic for image upload and deletion. Both operations run inside a DB transaction via `txManager`: upload does `imageRepo.Create` → `objectStore.Upload` (MinIO failure rolls back the DB insert); delete fetches the image by `imageID` alone via `imageRepo.Get` (UUID v7 is globally unique), then does `imageRepo.Delete` → `objectStore.Delete` (MinIO failure rolls back the DB delete).
- `internal/storage/products/` — PostgreSQL queries built with Masterminds/squirrel. Generic helpers live in `pkg/sqlp/sqlp.go`.
- `internal/storage/images/` — PostgreSQL queries for product images.
- `internal/infra/postgres/` — Connection pool setup (40 max open, 20 max idle, 10m idle timeout, 20m lifetime) using pgx/v5 + sqlx.
- `internal/infra/minio/` — MinIO object store client. `URL(key)` returns the public URL without uploading; `Upload` stores the object.
- `pkg/api/api.go` — Shared HTTP response helpers (e.g. `SendInternalError`).
- `pkg/sqlp/` — Generic query helpers (`Get`, `Select`, `Insert`, `Update`, `Delete`) that take `*sqlx.DB` and automatically use a transaction from context if one is present. `TxManager.RunInTx` starts a transaction, injects it into context, commits on success, rolls back on error.

**Dependency injection:** each layer defines its own interfaces (`deps.go`) that the layer above depends on. Mocks are generated with mockery (config: `.mockery.yml`, template: testify).

## DTOs

- `internal/usecase/dto.go` — shared input DTOs for use cases (`GetProductsIn`, `CreateProductIn`, `UpdateProductIn`, `UploadImageIn`).
- `internal/api/product/dto/` — HTTP-level request/response structs for the product handler.
- `internal/api/image/dto/` — HTTP-level response structs for the image handler (`UploadResponse`).

## Key Conventions

- **Price** is stored in smallest currency unit (kopecks/cents) as `BIGINT`.
- **Product IDs** are UUID v7 strings generated in the use case layer; stored as `UUID` in PostgreSQL.
- **Image object keys** follow the pattern `products/{productID}/{imageID}.{ext}`.
- **Environment** variables: in dev the service loads `.env` via `godotenv.Load()` (required, must exist). In production all vars are injected via `env_file: .env.prod` in docker-compose — no file mount needed. Both `.env` and `.env.prod` are gitignored.
- **Logging format**: JSON in production, pretty via `devslog` in dev. Service checks `env` env var; worker checks `APP_ENV`.
- **Frontend**: static `frontend/index.html` is served at `GET /` by the service.

## Mocks

Regenerate mocks after changing any interface in `./internal`:
```bash
mise run mocks
```

`mise run mocks` calls `mockery` using the mise-managed binary. **Do not use the system `mockery`** — it may be a different version and will fail to parse `.mockery.yml`. The correct binary is at:
```
~/.local/share/mise/installs/mockery/3.7.0/mockery
```

If `mise run mocks` fails because an existing mock file has compile errors (e.g. after an interface change), stub the broken file first, then regenerate:
```bash
echo "package mocks" > internal/path/to/mocks/mock_broken.go
mise run mocks
```

Mocks are placed in `mocks/` subdirectories alongside the interfaces they mock.

## Monitoring

- Prometheus scrapes `/metrics` (port 8081) every 5s.
- Grafana at `localhost:3000` — dashboards in `monitoring/dashboards/`.
- PostgreSQL metrics via postgres-exporter at port 9187.

## HTTP Test Files

Runnable API examples live in `http/products/` (`.http` format). Run via `ijhttp` with env file `http/http-client.env.json`.
