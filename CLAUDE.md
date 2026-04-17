# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go HTTP API server (`github.com/hreftools/api`) using Go 1.26.0, the standard library's `net/http` package, PostgreSQL via `pgx`, and OpenTelemetry for tracing.

## Development Philosophy

**Prefer standard library**: Always use Go's standard library over third-party dependencies unless explicitly stated otherwise.

**Learning project - no command execution**: NEVER execute commands (using Bash tool) on behalf of the user. Always provide instructions and let the user type commands themselves. Only provide guidance, code suggestions, and explanations.

## Development Commands

```bash
make dev              # Live reload via air
make build port=8080 db_url=... resend_api_key=...  # Build binary
make run              # Build and run
make test             # Run tests
make test-coverage    # Tests with coverage
make gen              # Regenerate sqlc code
make install-tools    # Install air, sqlc, migrate
make migrate-create name=<name>     # Create new migration
make migrate-up db_url=<url>        # Run migrations up
make migrate-down db_url=<url>      # Run migrations down
make docker-up / docker-down        # Docker Compose
```

Run a single test:
```bash
go test ./internal/user/ -run TestValidateUsername
```

## Environment Variables

| Variable         | Required | Description                       |
| ---------------- | -------- | --------------------------------- |
| `PORT`           | Yes      | Port the server listens on        |
| `DATABASE_URL`   | Yes      | PostgreSQL connection string      |
| `RESEND_API_KEY` | Yes      | Resend API key for sending emails |
| `OTEL_SERVICE_NAME` | No | OpenTelemetry service name |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | No | OTLP exporter endpoint URL |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | No | OTLP protocol (e.g., `http/protobuf`) |
| `OTEL_RESOURCE_ATTRIBUTES` | No | Additional OTEL resource attributes |
| `OTEL_EXPORTER_OTLP_HEADERS` | No | Headers for OTLP exporter (e.g., auth tokens) |

## Architecture

The codebase follows a **domain-driven layout** with two domain packages (`user`, `resource`), each defining their own models, repository interfaces, validation, services, and error types.

### Key layers

- **`cmd/api/main.go`** — Entry point. Wires config, database, repositories, services, tracing, and server. Handles graceful shutdown.
- **`internal/user/`** — User domain: `User` and `Session` models, `Repository` and `SessionRepository` interfaces, `Service` (business logic for auth flows, CRUD), validation functions, and sentinel errors.
- **`internal/resource/`** — Resource domain: `Resource` model, `Repository` interface, `Service`, validation, and sentinel errors.
- **`internal/postgres/`** — PostgreSQL implementations of repository interfaces. `Connect()` sets up the connection pool. One file per repository (`repository_users.go`, `repository_sessions.go`, `repository_resources.go`).
- **`internal/server/`** — HTTP layer: route registration (`server.go`), all handlers (`handler_*.go`), all middlewares (`middleware_*.go`), JSON response helpers and response DTOs (`helpers.go`).
- **`internal/db/`** — sqlc-generated code. Do not edit manually; regenerate with `make gen`.
- **`internal/config/`** — `LoadConfig()` reads env vars. Shared constants: session durations, context keys.
- **`internal/emails/`** — `EmailSender` interface + Resend implementation. Template rendering for transactional emails.

### Data flow

```
HTTP request → middleware stack → handler → domain Service → Repository interface → postgres implementation → sqlc/db
```

### Routing

All routes are prefixed with `/v1/` via `http.StripPrefix`. Routes are registered in `internal/server/server.go` using `http.ServeMux` with method prefixes (e.g., `GET /resources/{id}`).

Middleware composition:
- **Global**: `loggingMiddleware` → `commonHeadersMiddleware` → `maxBodySizeMiddleware`
- **Authenticated routes**: wrapped with `auth(handler)`
- **Admin routes**: wrapped with `adminOnly(handler)` (which is `middlewareStack(auth, admin)`)

### Validation pattern

Each domain package contains its own validation functions (e.g., `user/validation.go`, `resource/validation.go`). Validators are called by the service layer before any repository calls. They return sanitized values alongside errors.

### Error mapping

Each domain package has a `map_error_to_http.go` file that maps domain sentinel errors to HTTP status codes. Handlers use these to translate service errors into appropriate JSON error responses.

### Response format

All JSON responses: `{"status": "ok"|"error", "data": ...}`

### Authentication

The auth middleware validates sessions from `Authorization: Bearer <uuid>` header or `session_id` cookie. On success, stores user ID in request context via `config.UserIDContextKey`. Sessions use sliding expiry (renewed when < 15 days remaining).

## Key Dependencies

| Package                          | Purpose                        |
| -------------------------------- | ------------------------------ |
| `github.com/jackc/pgx/v5`        | PostgreSQL driver              |
| `github.com/google/uuid`         | UUID types                     |
| `github.com/resend/resend-go/v3` | Transactional email via Resend |
| `golang.org/x/crypto`            | Password hashing (argon2id)    |
| `go.opentelemetry.io/otel`       | OpenTelemetry tracing          |

## SQL & Migrations

- Migrations live in `sql/migrations/` (sequential numbered, `.up.sql`/`.down.sql`)
- Queries live in `sql/queries/` (one file per domain: `resources.sql`, `sessions.sql`, `users.sql`)
- sqlc config: `sqlc.yml` — generates to `internal/db/` with `emit_interface: true` and `emit_empty_slices: true`
