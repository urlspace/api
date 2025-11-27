# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go HTTP API server (`github.com/zapi-sh/api`) that provides a simple REST API. The codebase uses Go 1.25.3 and the standard library's `net/http` package.

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

The server starts on port 8080 with these configured timeouts:

- ReadTimeout: 10s
- ReadHeaderTimeout: 5s
- WriteTimeout: 10s
- IdleTimeout: 120s

## Project Structure

```
api/
├── cmd/
│   └── api/
│       └── main.go          # Entry point - wires everything together
├── internal/
│   ├── handlers/
│   │   └── status.go        # HTTP handlers (one per resource)
│   ├── middleware/
│   │   └── common.go        # Middleware (security headers, CORS, etc.)
│   └── models/
│       └── response.go      # Shared structs (requests/responses)
├── go.mod
├── Makefile
└── CLAUDE.md
```

## Architecture

**Standard Go project layout**: Following Go community conventions:

- `cmd/api/` - Application entry point, minimal logic
- `internal/` - Private packages (can't be imported by external projects)
- `internal/handlers/` - HTTP handler functions (easy to test)
- `internal/middleware/` - Reusable middleware (headers, auth, logging)
- `internal/models/` - Shared data structures

**HTTP Server Setup**:

- Uses `http.Server` with explicit timeouts for security
- Route-based handlers using `http.ServeMux` with method prefixes (e.g., `GET /status`)
- Middleware chain applies common headers (security, CORS, Content-Type)

**Response Structure**: All JSON responses follow a structured format:

- Success: `{"status": "ok", "data": "..."}`
- Error: `{"status": "error", "error": "..."}`

**Security Headers**: All responses include:

- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Content-Security-Policy: default-src 'self'`

**CORS**: Configured for open/public API with `Access-Control-Allow-Origin: *`

## API Endpoints

- `GET /status` - Health check endpoint returning service status
