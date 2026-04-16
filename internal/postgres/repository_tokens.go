package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/hreftools/api/internal/db"
	"github.com/hreftools/api/internal/user"
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
		Type:        t.Type,
		Description: t.Description,
		ExpiresAt:   t.ExpiresAt,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func (t *TokenRepository) Create(ctx context.Context, params user.TokenCreateParams) (user.Token, error) {
	args := db.CreateTokenParams{
		UserID:      params.UserID,
		Type:        params.Type,
		Description: params.Description,
		ExpiresAt:   params.ExpiresAt,
	}
	row, err := t.queries.CreateToken(ctx, args)
	if err != nil {
		return user.Token{}, translateTokenError(err)
	}
	return toToken(row), nil
}

func (t *TokenRepository) GetByID(ctx context.Context, id uuid.UUID) (user.Token, error) {
	row, err := t.queries.GetTokenById(ctx, id)
	if err != nil {
		return user.Token{}, translateTokenError(err)
	}
	return toToken(row), nil
}

func (t *TokenRepository) UpdateExpiresAt(ctx context.Context, params user.TokenUpdateExpiresAtParams) (user.Token, error) {
	args := db.UpdateTokenExpiresAtParams{
		ID:        params.ID,
		ExpiresAt: params.ExpiresAt,
	}
	row, err := t.queries.UpdateTokenExpiresAt(ctx, args)
	if err != nil {
		return user.Token{}, translateTokenError(err)
	}
	return toToken(row), nil
}

func (t *TokenRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return t.queries.DeleteToken(ctx, id)
}

func (t *TokenRepository) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error {
	return t.queries.DeleteTokensByUserID(ctx, userID)
}
