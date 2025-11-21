package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/zapi-sh/api/internal/server"
	"github.com/zapi-sh/api/internal/store"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	// TODO: this this should come from env
	databaseUrl    = "postgres://postgres:postgres@localhost:5432/zapishdb?sslmode=disable"
	pathMigrations = "file://sql/migrations"
)

func run(ctx context.Context) error {
	pool, err := sql.Open("pgx", databaseUrl)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := pool.Close(); err != nil {
			log.Printf("Failed to close database connection: %v", err)
		}
	}()

	// configure db connection pool
	pool.SetMaxOpenConns(25)
	pool.SetMaxIdleConns(5)
	pool.SetConnMaxLifetime(5 * time.Minute)
	pool.SetConnMaxIdleTime(5 * time.Minute)

	// verify the db connectoin
	if err := pool.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	m, err := migrate.New(pathMigrations, databaseUrl)
	if err != nil {
		log.Fatal(err)
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal(err)
	}

	store := store.NewStore(pool)

	srv := server.New(store)

	chServer := make(chan error, 1)

	log.Println("Starting server on :8080")
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			chServer <- err
		}
		close(chServer)
	}()

	ctxSignal, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case <-ctxSignal.Done():
		log.Printf("Server shutting down due to signal: %v", context.Cause(ctxSignal))
	case err := <-chServer:
		log.Printf("Server error: %v", err)
		return err
	}

	ctxTimeout, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()

	if err := srv.Shutdown(ctxTimeout); err != nil {
		log.Printf("Server shutdown failed: %v", err)

		if closeErr := srv.Close(); closeErr != nil {
			log.Printf("Server close failed: %v", closeErr)
			return errors.Join(err, closeErr)
		}

		return err
	}

	log.Println("Server exited gracefully")
	return nil
}

func main() {
	ctx := context.Background()

	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}
