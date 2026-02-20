package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hreftools/api/internal/db"
)

type UserStore interface {
	List(ctx context.Context) ([]db.User, error)
	GetById(ctx context.Context, id uuid.UUID) (db.User, error)
	GetByEmail(ctx context.Context, email string) (db.User, error)
	GetByEmailVerificationToken(ctx context.Context, id uuid.UUID) (db.User, error)
	Create(ctx context.Context, email string, emailVerified bool, emailVerificationToken uuid.NullUUID, emailVerificationTokenExpiresAt *time.Time, password string, username string, isAdmin bool, isPro bool) (db.User, error)
	Verify(ctx context.Context, id uuid.UUID) (db.User, error)
	UpdateVerificationToken(ctx context.Context, id uuid.UUID, emailVerificatinToken uuid.NullUUID, emailVerificationTokenExpiresAt *time.Time) (db.User, error)
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

func (r *userStore) GetById(ctx context.Context, id uuid.UUID) (db.User, error) {
	return r.queries.GetUserById(ctx, id)
}

func (r *userStore) GetByEmail(ctx context.Context, email string) (db.User, error) {
	return r.queries.GetUserByEmail(ctx, email)
}

func (r *userStore) GetByEmailVerificationToken(ctx context.Context, emailVerificationToken uuid.UUID) (db.User, error) {
	return r.queries.GetUserByEmailVerificationToken(ctx, uuid.NullUUID{Valid: true, UUID: emailVerificationToken})
}

func (r *userStore) Create(ctx context.Context, email string, emailVerified bool, emailVerificationToken uuid.NullUUID, emailVerificationTokenExpiresAt *time.Time, password string, username string, isAdmin bool, isPro bool) (db.User, error) {
	args := db.CreateUserParams{
		Email:                           email,
		EmailVerified:                   emailVerified,
		EmailVerificationToken:          emailVerificationToken,
		EmailVerificationTokenExpiresAt: emailVerificationTokenExpiresAt,
		Password:                        password,
		Username:                        username,
		IsAdmin:                         isAdmin,
		IsPro:                           isPro,
	}

	return r.queries.CreateUser(ctx, args)
}

func (r *userStore) Verify(ctx context.Context, id uuid.UUID) (db.User, error) {
	return r.queries.VerifyUser(ctx, id)
}

func (r *userStore) UpdateVerificationToken(ctx context.Context, id uuid.UUID, emailVerificatinToken uuid.NullUUID, emailVerificationTokenExpiresAt *time.Time) (db.User, error) {
	args := db.UpdateVerificationTokenParams{
		ID:                              id,
		EmailVerificationToken:          emailVerificatinToken,
		EmailVerificationTokenExpiresAt: emailVerificationTokenExpiresAt,
	}
	return r.queries.UpdateVerificationToken(ctx, args)
}

func (r *userStore) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteUser(ctx, id)
}
