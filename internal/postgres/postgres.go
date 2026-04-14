package postgres

import (
	"database/sql"
	"fmt"
	"time"
)

func Connect(databaseURL string) (*sql.DB, error) {
	pool, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// configure db connection pool
	pool.SetMaxOpenConns(25)
	pool.SetMaxIdleConns(5)
	pool.SetConnMaxLifetime(5 * time.Minute)
	pool.SetConnMaxIdleTime(5 * time.Minute)

	// verify the db connection
	if err := pool.Ping(); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}
