BINARY_NAME=api

.SILENT:

# Build the application
# make build port=port db_url=db_url resend_api_key=resend_api_key
build:
	PORT=$(port) DATABASE_URL=$(db_url) RESEND_API_KEY=$(resend_api_key) go build -o ${BINARY_NAME} cmd/api/main.go

# Run the built binary (production-like)
run: build
	./${BINARY_NAME}

# Development mode with live reload
dev:
	air

# Clean build artifacts
clean:
	go clean
	rm -f ${BINARY_NAME}

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

# Generate code from SQL schema and queries
gen:
	sqlc generate

# you run this thig like that
# make migration name=create_users_table
migrate-create:
	migrate create -dir ./sql/migrations -ext sql -seq $(name)

# make migrate-up db_url=postgres://postgres:postgres@localhost:5432/jumplist?sslmode=disable
migrate-up:
	migrate -path ./sql/migrations -database $(db_url) up

# make migrate-down db_url=postgres://postgres:postgres@localhost:5432/jumplist?sslmode=disable
migrate-down:
	migrate -path ./sql/migrations -database $(db_url) down

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-down-v:
	docker compose down -v

lint:
	golangci-lint run ./...

format:
	golangci-lint fmt ./...

# Default target (what runs when you just type 'make')
.PHONY: build run dev clean test test-coverage install-tools gen docker-up docker-down docker-down-v lint format migration-create migration-up migration-down
