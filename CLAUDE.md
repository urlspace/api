# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go HTTP API server (`github.com/hreftools/api`) that provides a REST API. The codebase uses Go 1.26.0 and the standard library's `net/http` package, with a PostgreSQL database backend.

## Development Philosophy

**Prefer standard library**: Always use Go's standard library over third-party dependencies unless explicitly stated otherwise. This project prioritizes simplicity and minimizes external dependencies.

**Learning project - no command execution**: This is a learning project. NEVER execute commands (using Bash tool) on behalf of the user. Always provide instructions and let the user type commands themselves to build muscle memory. Only provide guidance, code suggestions, and explanations.

## Development Tools

This project uses `air` for live reloading during development. Install it globally:

```bash
go install github.com/air-verse/air@latest
```

Then initialize air configuration (creates `.air.toml`):

```bash
air init
```

**Note**: Go development tools are installed globally, not per-project. They won't appear in `go.mod` since they're not runtime dependencies.

The database layer uses `sqlc` for type-safe query generation. Generated files live in `internal/db/` and should not be edited manually.

## Development Commands

### Running the server with live reload (recommended for development)

```bash
air
```

### Running the server (manual)

```bash
go run cmd/api/main.go
```

### Using Makefile (recommended)

```bash
make build  # Build the binary
make run    # Build and run (production-like)
make dev    # Development with live reload
make test   # Run tests
make clean  # Clean build artifacts
```

The server starts on port 8080 by default (configurable via `PORT` env var) with these configured timeouts:

- ReadTimeout: 10s
- ReadHeaderTimeout: 5s
- WriteTimeout: 10s
- IdleTimeout: 120s

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `PORT` | Yes | Port the server listens on |
| `DATABASE_URL` | Yes | PostgreSQL connection string |
| `RESEND_API_KEY` | No | Resend API key for sending emails |

## Project Structure

```
api/
├── cmd/
│   └── api/
│       └── main.go                    # Entry point - wires everything together
├── internal/
│   ├── db/                            # sqlc-generated database code (do not edit)
│   │   ├── db.go
│   │   ├── models.go
│   │   ├── querier.go
│   │   ├── resources.sql.go
│   │   ├── tokens.sql.go
│   │   └── users.sql.go
│   ├── emails/                        # Email sending abstraction
│   │   ├── sender.go                  # EmailSender interface
│   │   ├── sender_resend.go           # Resend implementation
│   │   └── render.go                  # Email template rendering
│   ├── handlers/                      # HTTP handlers (one file per endpoint)
│   │   ├── authSignin.go
│   │   ├── authSignup.go
│   │   ├── authVerify.go
│   │   ├── authResendVerification.go
│   │   ├── resourcesCreate.go
│   │   ├── resourcesDelete.go
│   │   ├── resourcesGet.go
│   │   ├── resourcesList.go
│   │   ├── resourcesUpdate.go
│   │   ├── usersCreate.go
│   │   ├── usersDelete.go
│   │   ├── usersGet.go
│   │   ├── usersList.go
│   │   ├── status.go
│   │   ├── notFound.go
│   │   └── constants.go
│   ├── config/
│   │   └── config.go                  # Shared constants (session durations, context keys)
│   ├── middlewares/                   # Middleware functions
│   │   ├── auth.go                    # Auth middleware — validates session/Bearer token
│   │   ├── admin.go                   # Admin middleware — requires IsAdmin on authenticated user
│   │   ├── common.go                  # Security headers, CORS, Content-Type
│   │   ├── logging.go                 # Request logging
│   │   └── middlewares_stack.go       # Middleware chaining helper
│   ├── models/
│   │   └── models.go                  # Response structs (ResponseResource, ResponseUser)
│   ├── response/                      # JSON response helpers
│   │   ├── success.go                 # WriteJSONSuccess
│   │   └── errors.go                  # WriteJSONError, HandleDbError, etc.
│   ├── server/
│   │   └── server.go                  # HTTP server setup and route registration
│   ├── store/                         # Database access layer
│   │   ├── store.go                   # Store struct wiring db queries
│   │   ├── store_resources.go
│   │   ├── store_tokens.go
│   │   └── store_users.go
│   ├── utils/
│   │   └── utils.go
│   └── validator/                     # Input validation
│       ├── email.go
│       ├── password.go
│       ├── username.go
│       ├── url.go
│       ├── resourceTitle.go
│       ├── resourceDescription.go
│       └── token.go
├── sql/
│   ├── migrations/                    # Database migration files
│   │   ├── 000001_init.up.sql
│   │   └── 000001_init.down.sql
│   └── queries/                       # SQL queries for sqlc
│       ├── resources.sql
│       ├── tokens.sql
│       └── users.sql
├── .air.toml                          # Air live reload configuration
├── go.mod
├── Makefile
├── sqlc.yml                           # sqlc code generation config
└── CLAUDE.md
```

## Architecture

**Standard Go project layout**: Following Go community conventions:

- `cmd/api/` - Application entry point, minimal logic; wires store, email sender, and server
- `internal/` - Private packages (can't be imported by external projects)
- `internal/db/` - sqlc-generated type-safe query code; backed by PostgreSQL via `pgx`
- `internal/store/` - Thin layer over `internal/db`; the only place handlers interact with the database
- `internal/handlers/` - HTTP handler functions (one file per endpoint, easy to test)
- `internal/config/` - Shared constants (`SessionExpiryDuration`, `SessionRenewalThreshold`, `UserIDContextKey`)
- `internal/middlewares/` - Reusable middleware (headers, logging, auth, admin); composed via `MiddlewareStack`
- `internal/models/` - Response structs that map db models to JSON-safe shapes
- `internal/response/` - Helpers for writing JSON success/error responses and handling DB errors
- `internal/emails/` - `EmailSender` interface + Resend implementation for transactional email
- `internal/validator/` - Pure validation functions used in handlers

**HTTP Server Setup**:

- Uses `http.Server` with explicit timeouts for security
- All routes are prefixed with `/v1/` (e.g. `GET /v1/status`)
- Route-based handlers using `http.ServeMux` with method prefixes (e.g., `GET /status`)
- Global middleware chain: `Logging` → `CommonHeaders`
- Route-level middleware: `Auth` (authenticated routes), `Auth` → `Admin` (admin-only routes)

**Response Structure**: All JSON responses follow a structured format:

- Success: `{"status": "ok", "data": "..."}`
- Error: `{"status": "error", "data": "..."}`

**Security Headers**: All responses include:

- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Content-Security-Policy: default-src 'self'`

**CORS**: Configured for open/public API with `Access-Control-Allow-Origin: *`

**Session Tokens**: Sign-in creates a session token stored in the `tokens` table. The token is set as an HTTP-only `session_id` cookie with a 30-day expiry. The `tokens` table supports two token types (defined in `handlers/constants.go`):

- `TokenTypeSession` (`"session"`) — created on sign-in, used to authenticate requests
- `TokenTypeAPI` (`"token"`) — for programmatic API access

**Authentication**: The `Auth` middleware (`middlewares/auth.go`) validates a token from the `Authorization: Bearer <uuid>` header or the `session_id` cookie. On success it stores the authenticated user's ID in the request context under `config.UserIDContextKey`. Use `utils.UserIDFromContext(ctx)` to read it in handlers. Sessions approaching expiry (< 15 days remaining) are renewed automatically via a background goroutine (sliding expiry).

**Admin routes**: The `Admin` middleware (`middlewares/admin.go`) reads the user ID set by `Auth`, fetches the user, and returns 403 if `IsAdmin` is false. Admin-only routes are protected with `MiddlewareStack(auth, admin)` (exposed as `adminOnly` in `server.go`).

## Key Dependencies

| Package | Purpose |
|---|---|
| `github.com/jackc/pgx/v5` | PostgreSQL driver |
| `github.com/google/uuid` | UUID types |
| `github.com/resend/resend-go/v3` | Transactional email via Resend |
| `golang.org/x/crypto` | Password hashing |

## API Endpoints

All endpoints are prefixed with `/v1`.

| Method | Path | Description |
|---|---|---|
| `GET` | `/v1/status` | Health check |
| `POST` | `/v1/auth/signup` | Register a new user |
| `POST` | `/v1/auth/verify` | Verify email with token |
| `POST` | `/v1/auth/resend-verification` | Resend verification email |
| `POST` | `/v1/auth/signin` | Sign in and create session |
| `GET` | `/v1/users` | List all users (admin) |
| `GET` | `/v1/users/{id}` | Get a user by ID (admin) |
| `POST` | `/v1/users` | Create a user (admin) |
| `DELETE` | `/v1/users/{id}` | Delete a user (admin) |
| `GET` | `/v1/resources` | List resources |
| `GET` | `/v1/resources/{id}` | Get a resource by ID |
| `POST` | `/v1/resources` | Create a resource |
| `PUT` | `/v1/resources/{id}` | Update a resource |
| `DELETE` | `/v1/resources/{id}` | Delete a resource |
