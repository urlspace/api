# ---- Build stage ----
FROM golang:1.26-alpine AS build

WORKDIR /src

# Download modules first so this layer is cached until go.mod/go.sum change.
COPY go.mod go.sum ./
RUN go mod download

# Install the golang-migrate CLI as a standalone tool for the pre-deploy hook.
# Separate binary — it does not enter the app's go.mod or the app binary.
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1

# Build a static binary (CGO disabled), same as `make build`.
COPY . .
RUN CGO_ENABLED=0 go build -o /out/api ./cmd/api

# ---- Runtime stage ----
FROM alpine:3.22

# CA certificates for outbound TLS (Resend API, OTLP exporter).
RUN apk add --no-cache ca-certificates

# Run as an unprivileged user.
RUN adduser -D -H -u 10001 appuser
USER appuser

WORKDIR /app

# The app binary.
COPY --from=build /out/api /usr/local/bin/api

# The migrate CLI and migration files — used only by the Railway pre-deploy hook,
# not by the app itself.
COPY --from=build /go/bin/migrate /usr/local/bin/migrate
COPY sql/migrations ./sql/migrations

ENTRYPOINT ["api"]
