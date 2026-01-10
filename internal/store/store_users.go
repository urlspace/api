package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jumplist/api/internal/db"
)

const (
	UserUsernameLengthMin = 3
	UserUsernameLengthMax = 32
	UserPasswordLengthMin = 12
)

type UserStore interface {
	List(ctx context.Context) ([]db.User, error)
	Get(ctx context.Context, id uuid.UUID) (db.User, error)
	Create(ctx context.Context, email string, emailVerified bool, emailVerificatinToken uuid.NullUUID, emailVerificationTokenExpiresAt *time.Time, password string, username string, isAdmin bool, isPro bool) (db.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type userStore struct {
	queries db.Querier
}

func NewUserStore(queries db.Querier) UserStore {
	return &userStore{
		queries: queries,
	}
}

func (r *userStore) List(ctx context.Context) ([]db.User, error) {
	return r.queries.ListUsers(ctx)
}

func (r *userStore) Get(ctx context.Context, id uuid.UUID) (db.User, error) {
	return r.queries.GetUser(ctx, id)
}

func (r *userStore) Create(ctx context.Context, email string, emailVerified bool, emailVerificatinToken uuid.NullUUID, emailVerificationTokenExpiresAt *time.Time, password string, username string, isAdmin bool, isPro bool) (db.User, error) {
	args := db.CreateUserParams{
		Email:                           email,
		EmailVerified:                   emailVerified,
		EmailVerificationToken:          emailVerificatinToken,
		EmailVerificationTokenExpiresAt: emailVerificationTokenExpiresAt,
		Password:                        password,
		Username:                        username,
		IsAdmin:                         isAdmin,
		IsPro:                           isPro,
	}

	return r.queries.CreateUser(ctx, args)
}

func (r *userStore) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteUser(ctx, id)
}
