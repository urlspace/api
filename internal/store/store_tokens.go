package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hreftools/api/internal/db"
)

type TokenCreateParams struct {
	UserID      uuid.UUID
	Type        string
	Description *string
	ExpiresAt   time.Time
}

type TokenUpdateExpiresAtParams struct {
	ID        uuid.UUID
	ExpiresAt time.Time
}

type TokenStore interface {
	Create(ctx context.Context, params TokenCreateParams) (db.Token, error)
	GetById(ctx context.Context, id uuid.UUID) (db.Token, error)
	UpdateExpiresAt(ctx context.Context, params TokenUpdateExpiresAtParams) (db.Token, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type tokenStore struct {
	queries db.Querier
}

func NewTokenStore(queries db.Querier) TokenStore {
	return &tokenStore{
		queries: queries,
	}
}

func (t *tokenStore) Create(ctx context.Context, params TokenCreateParams) (db.Token, error) {
	args := db.CreateTokenParams{
		UserID:      params.UserID,
		Type:        params.Type,
		Description: params.Description,
		ExpiresAt:   params.ExpiresAt,
	}
	return t.queries.CreateToken(ctx, args)
}

func (t *tokenStore) GetById(ctx context.Context, id uuid.UUID) (db.Token, error) {
	return t.queries.GetTokenById(ctx, id)
}

func (t *tokenStore) UpdateExpiresAt(ctx context.Context, params TokenUpdateExpiresAtParams) (db.Token, error) {
	args := db.UpdateTokenExpiresAtParams{
		ID:        params.ID,
		ExpiresAt: params.ExpiresAt,
	}
	return t.queries.UpdateTokenExpiresAt(ctx, args)
}

func (t *tokenStore) Delete(ctx context.Context, id uuid.UUID) error {
	return t.queries.DeleteToken(ctx, id)
}
