# Binary name
BINARY_NAME=api

# Build the application
build:
	PORT=3000 go build -o ${BINARY_NAME} cmd/api/main.go

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

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-down-all:
	docker compose down -v --rmi all

lint:
	golangci-lint run ./...

format:
	golangci-lint fmt ./...

# you run this thig like that
# make migration name=create_users_table
migration:
	migrate create -dir ./sql/migrations -ext sql -seq $(name)

# Default target (what runs when you just type 'make')
.PHONY: build run dev clean test test-coverage install-tools gen
