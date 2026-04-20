package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/urlspace/api/internal/db"
	"github.com/urlspace/api/internal/user"
)

type TokenRepository struct {
	queries db.Querier
}

func NewTokenRepository(queries db.Querier) user.TokenRepository {
	return &TokenRepository{queries: queries}
}

func translateTokenError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return user.ErrNotFound
	}
	return err
}

func toToken(t db.Token) user.Token {
	return user.Token{
		ID:          t.ID,
		UserID:      t.UserID,
		Description: t.Description,
		Hash:        t.Hash,
		LastUsedAt:  t.LastUsedAt,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func (r *TokenRepository) Create(ctx context.Context, params user.TokenCreateParams) (user.Token, error) {
	args := db.CreateTokenParams{
		UserID:      params.UserID,
		Description: params.Description,
		Hash:        params.Hash,
	}
	row, err := r.queries.CreateToken(ctx, args)
	if err != nil {
		return user.Token{}, translateTokenError(err)
	}
	return toToken(row), nil
}

func (r *TokenRepository) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (user.Token, error) {
	row, err := r.queries.GetTokenById(ctx, db.GetTokenByIdParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return user.Token{}, translateTokenError(err)
	}
	return toToken(row), nil
}

func (r *TokenRepository) GetByHash(ctx context.Context, hash string) (user.Token, error) {
	row, err := r.queries.GetTokenByHash(ctx, hash)
	if err != nil {
		return user.Token{}, translateTokenError(err)
	}
	return toToken(row), nil
}

func (r *TokenRepository) List(ctx context.Context, userID uuid.UUID) ([]user.Token, error) {
	rows, err := r.queries.ListTokensByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	tokens := make([]user.Token, len(rows))
	for i, row := range rows {
		tokens[i] = toToken(row)
	}
	return tokens, nil
}

func (r *TokenRepository) UpdateLastUsedAt(ctx context.Context, id uuid.UUID) error {
	return r.queries.UpdateTokenLastUsedAt(ctx, id)
}

func (r *TokenRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return r.queries.DeleteToken(ctx, db.DeleteTokenParams{
		ID:     id,
		UserID: userID,
	})
}

func (r *TokenRepository) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.queries.DeleteTokensByUserID(ctx, userID)
}
