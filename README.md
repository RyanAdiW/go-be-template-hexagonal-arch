# Go Hexagonal Architecture Template

A production-oriented Go modular-monolith template. It uses Echo, PostgreSQL, GORM, golang-migrate, cleanenv, `slog`, JWT, OpenTelemetry, Prometheus, Testify, Testcontainers, Docker, and GitHub Actions.

The goal is a pragmatic reference implementation: module ownership is explicit, boundaries are testable, and future extraction is possible without prematurely building microservices.

## Stack

| Area | Choice |
| --- | --- |
| HTTP | Echo |
| Database | PostgreSQL |
| Persistence | GORM |
| Schema migration | golang-migrate with explicit SQL migrations |
| Configuration | cleanenv from environment variables |
| Logging | Go `log/slog` |
| Validation | go-playground/validator |
| Authentication baseline | HS256 JWT access tokens |
| Observability | OpenTelemetry and Prometheus |
| Testing | Go testing, Testify, and Testcontainers |

## Architecture

Code is owned by feature modules rather than application-wide `handler`, `service`, or `repository` folders.

```text
HTTP adapter
  → application use case
    → domain
      → application port
        → PostgreSQL/GORM adapter
```

```text
internal/
├── app/                         # Application composition root
├── platform/                    # Config, database, HTTP server, security, observability
└── user/                        # Complete reference module
    ├── domain/                  # Pure entities and domain errors
    ├── application/             # Use cases and ports
    └── adapter/
        ├── inbound/http/        # Echo handlers, DTOs, and routes
        └── outbound/postgres/   # GORM models, mappers, and repository
```

### GORM boundary

`domain.User` has no GORM tags and never imports `gorm.io/*`. Persistence models live exclusively in `internal/user/adapter/outbound/postgres` and are mapped to and from domain entities.

The project also **does not use `AutoMigrate()`**. Database schema is managed only through versioned SQL files in [`migrations/`](migrations), executed by `golang-migrate` before application rollout.

## Prerequisites

- Go 1.25+
- PostgreSQL 16+, or Docker Desktop/Engine
- Docker for Testcontainers integration tests
- The `migrate` CLI is only needed for `make migrate-create`; the other migration commands use the repository's `cmd/migrate` binary.

## Quick start: local API with Docker PostgreSQL

1. Create a local environment file.

   ```bash
   cp .env.example .env
   ```

2. Replace `JWT_SECRET` with a random secret of at least 32 bytes.

3. Start PostgreSQL.

   ```bash
   make docker-up
   ```

4. Apply migrations explicitly.

   ```bash
   make migrate-up
   ```

5. Start the API.

   ```bash
   make run
   ```

The API listens on `http://localhost:8080` by default.

## Endpoints

| Method | Path | Authentication | Description |
| --- | --- | --- | --- |
| `GET` | `/health/live` | No | Process liveness; does not check PostgreSQL. |
| `GET` | `/health/ready` | No | Readiness; verifies PostgreSQL connectivity. |
| `GET` | `/metrics` | No | Prometheus metrics when `METRICS_ENABLED=true`. |
| `POST` | `/api/v1/users` | No | Creates a user. |
| `GET` | `/api/v1/users/:id` | Bearer JWT | Gets a user; the baseline policy only permits the current user. |

Create a user:

```bash
curl --request POST http://localhost:8080/api/v1/users \
  --header 'Content-Type: application/json' \
  --data '{"email":"ada@example.com","name":"Ada Lovelace"}'
```

Successful responses use a consistent envelope:

```json
{
  "data": {
    "id": "...",
    "email": "ada@example.com",
    "name": "Ada Lovelace",
    "created_at": "2026-01-01T00:00:00Z",
    "updated_at": "2026-01-01T00:00:00Z"
  }
}
```

Errors contain a stable code, a safe public message, and the request ID:

```json
{
  "error": {
    "code": "EMAIL_ALREADY_EXISTS",
    "message": "email already exists",
    "details": null
  },
  "request_id": "..."
}
```

## JWT status

The template includes an issuer utility, token verification, issuer/audience/expiry validation, auth middleware, and framework-independent authenticated identity extraction.

It intentionally does not yet expose register, login, password, or refresh-token endpoints. `POST /api/v1/users` works end-to-end today; `GET /api/v1/users/:id` requires an access token signed with `JWT_SECRET`. Add an authentication module when credential and session requirements are defined.

## Configuration

Configuration is loaded only from environment variables. The production binary does not automatically load `.env` files.

| Group | Important variables |
| --- | --- |
| Application | `APP_ENV`, `APP_NAME` |
| HTTP | `HTTP_HOST`, `HTTP_PORT`, `HTTP_READ_TIMEOUT`, `HTTP_WRITE_TIMEOUT`, `HTTP_IDLE_TIMEOUT`, `HTTP_SHUTDOWN_TIMEOUT`, `HTTP_MAX_BODY_BYTES`, `CORS_ALLOWED_ORIGINS` |
| Database | `DATABASE_HOST`, `DATABASE_PORT`, `DATABASE_USER`, `DATABASE_PASSWORD`, `DATABASE_NAME`, `DATABASE_SSL_MODE`, and connection-pool settings |
| JWT | `JWT_SECRET`, `JWT_ISSUER`, `JWT_AUDIENCE`, `JWT_ACCESS_TOKEN_TTL` |
| Observability | `OTEL_ENABLED`, `OTEL_EXPORTER_ENDPOINT`, `METRICS_ENABLED`, `LOG_LEVEL` |

Startup fails fast if database configuration is incomplete, `JWT_SECRET` is shorter than 32 bytes, connection-pool values are invalid, or tracing is enabled without an exporter endpoint.

## Database migrations

Migrations are an explicit deployment step; the API never applies them during startup.

```bash
make migrate-up
make migrate-down
make migrate-version
make migrate-force version=1
make migrate-create name=add_user_status
```

Use an expand/contract approach for risky production changes. Do not depend on down migrations to reverse destructive data changes in production.

## Testing and quality checks

```bash
make test-unit         # Unit tests with the race detector
make test-integration  # PostgreSQL Testcontainers; Docker must be running
make lint              # go vet
make fmt               # Format Go source
make build             # Build API and migration binaries
```

Unit tests use fakes at repository boundaries. Integration tests run against real PostgreSQL through Testcontainers and apply SQL migrations before repository tests execute.

## Docker

`make docker-up` starts only PostgreSQL, so the API can be run locally with `make run`.

To run the API inside Docker Compose, set `DATABASE_HOST=postgres` for the container environment, apply migrations, then start the API:

```bash
docker compose --profile migrate run --rm migrate
docker compose up -d api
```

The API image uses a multi-stage build and runs as a non-root user.

## CI

The GitHub Actions workflow checks formatting, runs `go vet`, runs unit tests with the race detector, builds binaries, and builds the Docker image. Integration tests can run in a dedicated Docker-enabled CI job as the project grows.

## Useful commands

```bash
make run
make build
make test
make fmt
make lint
make docker-up
make docker-down
```

## Security baseline

- JWT secrets come only from environment variables or a secret manager.
- Request body size is limited.
- Read, write, idle, and shutdown timeouts are configured on `http.Server`.
- CORS uses an allowlist.
- Secure response headers, panic recovery, request IDs, and structured logging are enabled.
- Passwords, JWTs, authorization headers, and secrets must never be logged.
- GORM uses parameterized queries; schema changes are explicit SQL migrations.
- The runtime container is non-root.
