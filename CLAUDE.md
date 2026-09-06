# CLAUDE.md — go-store

E-commerce REST API backend on Go (Fiber v3 + PostgreSQL + Redis + MeiliSearch + Kafka).

## Quick reference

```bash
make build              # Build binary
make run start          # Build & run (default: start command)
make lint               # golangci-lint (strict, 50+ linters)
make fmt                # gofumpt formatting
make gen-sql            # Regenerate sqlc/pgxgen models & queries
make gen-swag           # Regenerate Swagger docs
make gen-envs           # Regenerate ENVS.md and config template
go test -v ./...        # Run tests
```

## Architecture

Clean Architecture: **Delivery → Service → Repository → Storage**.

```
cmd/app/              CLI entry point (urfave/cli)
cmd/console/          Console commands (start, migrate, config, geo)
internal/
  app/                Bootstrap, DI wiring
  config/             Config structs (aconfig: YAML + env + flags)
  delivery/http/      Fiber handlers, request/response DTOs, middleware
  service/            Business logic, interfaces (IProductService, etc.)
  storage/repository/ Generated (sqlc/pgxgen) + custom query wrappers
  models/             Auto-generated DB models (models_gen.go)
  dto/                Domain DTOs and mappers
  messaging/          Kafka consumers, workers, contracts
  router/             Route registration
  server/             Fiber server init
  tools/              Helpers (apierror, validation, password, etc.)
pkg/                  Reusable packages (logger, cfg, queue, etc.)
sql/                  Migrations, seeds, pgxgen.yaml, sqlc-postgres.yaml
configs/              config.yaml template
docs/                 Swagger output, kafka-integration.md
```

## Code generation

### sqlc + pgxgen (database layer)

Config: `sql/pgxgen.yaml`, `sql/sqlc-postgres.yaml`.

**Never edit generated files directly** (`models_gen.go`, `repository_*/` generated code). Fix the source SQL in `sql/` and run `make gen-sql`.

### Swagger

Annotations live in handler functions. Regenerate with `make gen-swag`.

## Key conventions

- **Go 1.25**, module `github.com/stickpro/go-store`
- **Fiber v3** for HTTP, **pgx/v5** for Postgres, **franz-go** for Kafka
- **shopspring/decimal** for prices — never use float for money
- **google/uuid** for IDs
- JSON tags use `snake_case` (enforced by linter)
- Validation via `go-playground/validator/v10` struct tags
- Errors returned as `apierror.Errors` (standardized API responses)
- Success payloads returned via `*_response` types only — never a raw `dto`/`model`/repo row (see **HTTP responses**)
- Logging via `go.uber.org/zap` with context support
- Transactions via repository `WithTx()` method
- Config: YAML defaults → env vars → CLI flags (aconfig)

## HTTP responses

**A handler must never pass a raw `internal/dto`, `internal/models`, or `repository_*` row to `response.OkByData`.** Every endpoint returns a type from a `internal/delivery/http/response/<domain>_response` package.

- Define the response struct there with `json:"..."` tags only (no `db:` tags, no `pgtype.*` — decode nullables to `*string` / `*T` via `pkg/dbutils/pgtypeutils`).
- Add a constructor: `NewFromModel(...)`, `NewFromDTO(...)`, and for lists `NewPaginated(*base.FindResponseWithFullPagination[*X]) *base.FindResponseWithFullPagination[Y]` (see `manufacturer_response`, `category_response/listing.go`).
- `internal/dto` types are the Service↔Delivery currency only; they don't leave the process. Domain DTOs that look "clean" still get a response twin.
- Listing endpoints for a product variant return `product_response.VariantCardResponse` / `VariantListResponse` — the one shared card contract (search, category, related, collections). Don't invent a new per-endpoint variant shape.
- `@Success` annotation names the response type; for pages use `response.Result[base.FindResponseWithFullPagination[<domain>_response.XResponse]]` and blank-import `internal/storage/base` in the handler file so `swag` resolves it. Run `make gen-swag`.
- Search hits (`searchtypes.SearchResult.Hits`, `[]interface{}`) are unmarshalled via `search.UnmarshalHits[*T]` into a typed struct before mapping — never returned raw.

## Database

- PostgreSQL, migrations in `sql/postgres/migrations/` (golang-migrate)
- Create migration: `make db-create-migration <name>`
- Run migrations: `make run migrate up`

## Linting

`.golangci.yml` is strict (50+ linters). Key exclusions:
- Generated code (sqlc, migrations) is excluded
- `dupl` relaxed in handlers
- Run `make lint` before committing

## Docker

- Multi-stage Dockerfile (golang:1.23-alpine → alpine)
- `docker-compose.yml`: go-store, postgres, redis, meilisearch, maildev
