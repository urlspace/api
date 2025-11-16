package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
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

func main() {
	pool, err := sql.Open("pgx", "postgres://postgres:postgres@localhost:5432/zapishdb?sslmode=disable")
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

	m, err := migrate.New(
		"file://sql/migrations",
		"postgres://postgres:postgres@localhost:5432/zapishdb?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	if err := m.Up(); err != nil {
		log.Fatal(err)
	}

	store := store.NewStore(pool)

	s := server.New(store)

	go func() {
		log.Printf("Starting server on %s", s.Addr)

		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	ctxSignal, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctxSignal.Done()

	log.Println("Server shutting down.")

	ctxTimeout, stop := context.WithTimeout(context.Background(), 10*time.Second)
	defer stop()

	if err := s.Shutdown(ctxTimeout); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server closed.")
}
