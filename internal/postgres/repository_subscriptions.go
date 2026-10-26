package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/urlspace/api/internal/db"
	"github.com/urlspace/api/internal/subscription"
	"uuid"
)

type SubscriptionRepository struct {
	queries db.Querier
}

func NewSubscriptionRepository(queries db.Querier) subscription.Repository {
	return &SubscriptionRepository{queries: queries}
}

func translateSubscriptionError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return subscription.ErrUserNotFound
	}
	return err
}

func toSubscription(s db.Subscription) subscription.Subscription {
	return subscription.Subscription{
		ID:                  s.ID,
		UserID:              s.UserID,
		PolarSubscriptionID: s.PolarSubscriptionID,
		Status:              s.Status,
		CurrentPeriodEnd:    s.CurrentPeriodEnd,
		CancelAtPeriodEnd:   s.CancelAtPeriodEnd,
		PolarModifiedAt:     s.PolarModifiedAt,
		CreatedAt:           s.CreatedAt,
		UpdatedAt:           s.UpdatedAt,
	}
}

func (r *SubscriptionRepository) Upsert(ctx context.Context, params subscription.UpsertParams) (bool, error) {
	rows, err := r.queries.UpsertSubscription(ctx, db.UpsertSubscriptionParams{
		UserID:              params.UserID,
		PolarSubscriptionID: params.PolarSubscriptionID,
		Status:              params.Status,
		CurrentPeriodEnd:    params.CurrentPeriodEnd,
		CancelAtPeriodEnd:   params.CancelAtPeriodEnd,
		PolarModifiedAt:     params.PolarModifiedAt,
	})
	if err != nil {
		return false, translateSubscriptionError(err)
	}
	return rows > 0, nil
}

func (r *SubscriptionRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]subscription.Subscription, error) {
	rows, err := r.queries.ListSubscriptionsByUserID(ctx, userID)
	if err != nil {
		return nil, translateSubscriptionError(err)
	}

	subs := make([]subscription.Subscription, len(rows))
	for i, row := range rows {
		subs[i] = toSubscription(row)
	}
	return subs, nil
}
