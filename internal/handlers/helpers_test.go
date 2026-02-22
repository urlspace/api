package handlers_test

import (
	"database/sql"
	"os"
	"testing"

	"github.com/hreftools/api/internal/emails"
	"github.com/hreftools/api/internal/store"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func setupTestStore(t *testing.T) *store.Store {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	pool, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	t.Cleanup(func() {
		pool.Exec("TRUNCATE users, resources RESTART IDENTITY CASCADE")
		pool.Close()
	})

	return store.New(pool)
}

type mockEmailSender struct {
	called bool
	params emails.EmailSendParams
}

func (m *mockEmailSender) Send(params emails.EmailSendParams) error {
	m.called = true
	m.params = params
	return nil
}
