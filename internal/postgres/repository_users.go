package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/urlspace/api/internal/db"
	"github.com/urlspace/api/internal/user"
	"uuid"
)

type UserRepository struct {
	queries db.Querier
}

func NewUserRepository(queries db.Querier) user.Repository {
	return &UserRepository{queries: queries}
}

func translateUserError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
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
		EmailVerificationTokenHash:      u.EmailVerificationTokenHash,
		EmailVerificationTokenExpiresAt: u.EmailVerificationTokenExpiresAt,
		Password:                        u.Password,
		PasswordResetTokenHash:          u.PasswordResetTokenHash,
		PasswordResetTokenExpiresAt:     u.PasswordResetTokenExpiresAt,
		Username:                        u.Username,
		DisplayName:                     u.DisplayName,
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

func (r *UserRepository) GetByEmailVerificationTokenHash(ctx context.Context, hash string) (user.User, error) {
	row, err := r.queries.GetUserByEmailVerificationTokenHash(ctx, &hash)
	if err != nil {
		return user.User{}, translateUserError(err)
	}
	return toUser(row), nil
}

func (r *UserRepository) GetByPasswordResetTokenHash(ctx context.Context, hash string) (user.User, error) {
	row, err := r.queries.GetUserByPasswordResetTokenHash(ctx, &hash)
	if err != nil {
		return user.User{}, translateUserError(err)
	}
	return toUser(row), nil
}

func (r *UserRepository) Create(ctx context.Context, params user.CreateParams) (user.User, error) {
	args := db.CreateUserParams{
		Email:                           params.Email,
		EmailVerified:                   params.EmailVerified,
		EmailVerificationTokenHash:      params.EmailVerificationTokenHash,
		EmailVerificationTokenExpiresAt: params.EmailVerificationTokenExpiresAt,
		Password:                        params.Password,
		Username:                        params.Username,
		DisplayName:                     params.DisplayName,
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
		EmailVerificationTokenHash:      params.EmailVerificationTokenHash,
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
		PasswordResetTokenHash:      params.PasswordResetTokenHash,
		PasswordResetTokenExpiresAt: params.PasswordResetTokenExpiresAt,
	}
	row, err := r.queries.UpdatePasswordResetToken(ctx, args)
	if err != nil {
		return user.User{}, translateUserError(err)
	}
	return toUser(row), nil
}

func (r *UserRepository) UpdateDisplayName(ctx context.Context, id uuid.UUID, displayName string) (user.User, error) {
	args := db.UpdateUserDisplayNameParams{
		ID:          id,
		DisplayName: displayName,
	}
	row, err := r.queries.UpdateUserDisplayName(ctx, args)
	if err != nil {
		return user.User{}, translateUserError(err)
	}
	return toUser(row), nil
}

func (r *UserRepository) UpdateUsername(ctx context.Context, id uuid.UUID, username string) (user.User, error) {
	args := db.UpdateUserUsernameParams{
		ID:       id,
		Username: username,
	}
	row, err := r.queries.UpdateUserUsername(ctx, args)
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
