package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/urlspace/api/internal/db"
	"github.com/urlspace/api/internal/user"
)

type UserRepository struct {
	queries db.Querier
}

func NewUserRepository(queries db.Querier) user.Repository {
	return &UserRepository{queries: queries}
}

func translateUserError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return user.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return user.ErrConflict
	}
	return err
}

func toUser(u db.User) user.User {
	return user.User{
		ID:                              u.ID,
		Email:                           u.Email,
		EmailVerified:                   u.EmailVerified,
		EmailVerificationToken:          u.EmailVerificationToken,
		EmailVerificationTokenExpiresAt: u.EmailVerificationTokenExpiresAt,
		Password:                        u.Password,
		PasswordResetToken:              u.PasswordResetToken,
		PasswordResetTokenExpiresAt:     u.PasswordResetTokenExpiresAt,
		Username:                        u.Username,
		IsAdmin:                         u.IsAdmin,
		IsPro:                           u.IsPro,
		CreatedAt:                       u.CreatedAt,
		UpdatedAt:                       u.UpdatedAt,
	}
}

func (r *UserRepository) List(ctx context.Context) ([]user.User, error) {
	rows, err := r.queries.ListUsers(ctx)
	if err != nil {
		return nil, translateUserError(err)
	}

	users := make([]user.User, len(rows))
	for i, row := range rows {
		users[i] = toUser(row)
	}
	return users, nil
}

func (r *UserRepository) GetById(ctx context.Context, id uuid.UUID) (user.User, error) {
	row, err := r.queries.GetUserById(ctx, id)
	if err != nil {
		return user.User{}, translateUserError(err)
	}
	return toUser(row), nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (user.User, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return user.User{}, translateUserError(err)
	}
	return toUser(row), nil
}

func (r *UserRepository) GetByEmailVerificationToken(ctx context.Context, emailVerificationToken uuid.UUID) (user.User, error) {
	row, err := r.queries.GetUserByEmailVerificationToken(ctx, uuid.NullUUID{Valid: true, UUID: emailVerificationToken})
	if err != nil {
		return user.User{}, translateUserError(err)
	}
	return toUser(row), nil
}

func (r *UserRepository) GetByPasswordResetToken(ctx context.Context, token uuid.UUID) (user.User, error) {
	row, err := r.queries.GetUserByPasswordResetToken(ctx, uuid.NullUUID{Valid: true, UUID: token})
	if err != nil {
		return user.User{}, translateUserError(err)
	}
	return toUser(row), nil
}

func (r *UserRepository) Create(ctx context.Context, params user.CreateParams) (user.User, error) {
	args := db.CreateUserParams{
		Email:                           params.Email,
		EmailVerified:                   params.EmailVerified,
		EmailVerificationToken:          params.EmailVerificationToken,
		EmailVerificationTokenExpiresAt: params.EmailVerificationTokenExpiresAt,
		Password:                        params.Password,
		Username:                        params.Username,
		IsAdmin:                         params.IsAdmin,
		IsPro:                           params.IsPro,
	}

	row, err := r.queries.CreateUser(ctx, args)
	if err != nil {
		return user.User{}, translateUserError(err)
	}
	return toUser(row), nil
}

func (r *UserRepository) Verify(ctx context.Context, id uuid.UUID) (user.User, error) {
	row, err := r.queries.VerifyUser(ctx, id)
	if err != nil {
		return user.User{}, translateUserError(err)
	}
	return toUser(row), nil
}

func (r *UserRepository) UpdateVerificationToken(ctx context.Context, params user.UpdateVerificationTokenParams) (user.User, error) {
	args := db.UpdateVerificationTokenParams{
		ID:                              params.ID,
		EmailVerificationToken:          params.EmailVerificationToken,
		EmailVerificationTokenExpiresAt: params.EmailVerificationTokenExpiresAt,
	}
	row, err := r.queries.UpdateVerificationToken(ctx, args)
	if err != nil {
		return user.User{}, translateUserError(err)
	}
	return toUser(row), nil
}

func (r *UserRepository) UpdatePasswordResetToken(ctx context.Context, params user.UpdatePasswordResetTokenParams) (user.User, error) {
	args := db.UpdatePasswordResetTokenParams{
		ID:                          params.ID,
		PasswordResetToken:          params.PasswordResetToken,
		PasswordResetTokenExpiresAt: params.PasswordResetTokenExpiresAt,
	}
	row, err := r.queries.UpdatePasswordResetToken(ctx, args)
	if err != nil {
		return user.User{}, translateUserError(err)
	}
	return toUser(row), nil
}

func (r *UserRepository) ResetPassword(ctx context.Context, id uuid.UUID, passwordHash string) (user.User, error) {
	args := db.ResetUserPasswordParams{
		ID:       id,
		Password: passwordHash,
	}
	row, err := r.queries.ResetUserPassword(ctx, args)
	if err != nil {
		return user.User{}, translateUserError(err)
	}
	return toUser(row), nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) (user.User, error) {
	row, err := r.queries.DeleteUser(ctx, id)
	if err != nil {
		return user.User{}, translateUserError(err)
	}
	return toUser(row), nil
}
