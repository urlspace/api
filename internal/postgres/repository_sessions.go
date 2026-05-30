package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/urlspace/api/internal/db"
	"github.com/urlspace/api/internal/user"
)

type SessionRepository struct {
	queries db.Querier
}

func NewSessionRepository(queries db.Querier) user.SessionRepository {
	return &SessionRepository{queries: queries}
}

func translateSessionError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return user.ErrNotFound
	}
	return err
}

func toSession(s db.Session) user.Session {
	return user.Session{
		ID:          s.ID,
		UserID:      s.UserID,
		SessionHash: s.SessionHash,
		Description: s.Description,
		ExpiresAt:   s.ExpiresAt,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}

func (r *SessionRepository) Create(ctx context.Context, params user.SessionCreateParams) (user.Session, error) {
	args := db.CreateSessionParams{
		UserID:      params.UserID,
		SessionHash: params.SessionHash,
		Description: params.Description,
		ExpiresAt:   params.ExpiresAt,
	}
	row, err := r.queries.CreateSession(ctx, args)
	if err != nil {
		return user.Session{}, translateSessionError(err)
	}
	return toSession(row), nil
}

func (r *SessionRepository) GetByHash(ctx context.Context, sessionHash string) (user.Session, error) {
	row, err := r.queries.GetSessionByHash(ctx, sessionHash)
	if err != nil {
		return user.Session{}, translateSessionError(err)
	}
	return toSession(row), nil
}

func (r *SessionRepository) UpdateExpiresAt(ctx context.Context, params user.SessionUpdateExpiresAtParams) (user.Session, error) {
	args := db.UpdateSessionExpiresAtParams{
		ID:        params.ID,
		ExpiresAt: params.ExpiresAt,
	}
	row, err := r.queries.UpdateSessionExpiresAt(ctx, args)
	if err != nil {
		return user.Session{}, translateSessionError(err)
	}
	return toSession(row), nil
}

func (r *SessionRepository) DeleteByHash(ctx context.Context, sessionHash string) error {
	return r.queries.DeleteSessionByHash(ctx, sessionHash)
}

func (r *SessionRepository) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.queries.DeleteSessionsByUserID(ctx, userID)
}
