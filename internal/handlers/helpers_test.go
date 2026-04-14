package handlers_test

import (
	"database/sql"
	"os"
	"testing"

	"github.com/hreftools/api/internal/db"
	"github.com/hreftools/api/internal/emails"
	"github.com/hreftools/api/internal/postgres"
	"github.com/hreftools/api/internal/resource"
	"github.com/hreftools/api/internal/user"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func setupTestDB(t *testing.T) *sql.DB {
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
		pool.Exec("TRUNCATE users, resources, tokens RESTART IDENTITY CASCADE")
		pool.Close()
	})

	return pool
}

func setupUserService(t *testing.T) *user.Service {
	t.Helper()

	pool := setupTestDB(t)
	queries := db.New(pool)
	userRepo := postgres.NewUserRepository(queries)
	tokenRepo := postgres.NewTokenRepository(queries)
	emailSender := &mockEmailSender{}

	return user.NewService(userRepo, tokenRepo, emailSender)
}

func setupResourceService(t *testing.T) *resource.Service {
	t.Helper()

	pool := setupTestDB(t)
	queries := db.New(pool)
	resourceRepo := postgres.NewResourceRepository(queries)

	return resource.NewService(resourceRepo)
}

func setupServices(t *testing.T) (*user.Service, *resource.Service, *mockEmailSender) {
	t.Helper()

	pool := setupTestDB(t)
	queries := db.New(pool)
	userRepo := postgres.NewUserRepository(queries)
	tokenRepo := postgres.NewTokenRepository(queries)
	resourceRepo := postgres.NewResourceRepository(queries)
	emailSender := &mockEmailSender{}

	userSvc := user.NewService(userRepo, tokenRepo, emailSender)
	resourceSvc := resource.NewService(resourceRepo)

	return userSvc, resourceSvc, emailSender
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
