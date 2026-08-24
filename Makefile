BINARY_NAME=bin/api

.SILENT:

# Build the application (compile only — otelc reads no OTEL_* at build time; the
# SDK setup it injects reads them at runtime).
build:
	CGO_ENABLED=0 otelc go build -o ${BINARY_NAME} ./cmd/api

# Run the already-built binary (production-like). Run `make build` first.
# Sources .env for app config (PORT, DATABASE_URL, ...) and the OTEL_* vars, so
# telemetry ships to the dev Grafana — the same .env air loads.
run:
	set -a; [ -f .env ] && . .env; set +a; ./${BINARY_NAME}

# Development mode with live reload
dev:
	air

# Clean build artifacts
clean:
	go clean
	rm -f ${BINARY_NAME}
	rm -rf bin/

# Run tests
test:
	go test ./...
# Run tests with coverage
test-coverage:
	go test -cover ./...

# Install development tools
install-tools:
	go install github.com/air-verse/air@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install go.opentelemetry.io/otelc/tool/cmd/otelc@v1.1.0

# Generate code from SQL schema and queries
gen:
	sqlc generate

# you run this thig like that
# make migration name=create_users_table
migrate-create:
	migrate create -dir ./sql/migrations -ext sql -seq $(name)

# make migrate-up db_url=postgres://postgres:postgres@localhost:5432/urlspace?sslmode=disable
migrate-up:
	migrate -path ./sql/migrations -database $(db_url) up

# make migrate-down db_url=postgres://postgres:postgres@localhost:5432/urlspace?sslmode=disable
migrate-down:
	migrate -path ./sql/migrations -database $(db_url) down

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-down-v:
	docker compose down -v


# Default target (what runs when you just type 'make')
.PHONY: build run dev clean test test-coverage install-tools gen docker-up docker-down docker-down-v migration-create migration-up migration-down
