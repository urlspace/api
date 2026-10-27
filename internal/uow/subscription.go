package uow

import (
	"context"
	"errors"
	"log/slog"

	"github.com/urlspace/api/internal/subscription"
	"github.com/urlspace/api/internal/user"
	"go.opentelemetry.io/otel/codes"
)

func (s *Service) SyncSubscription(ctx context.Context, params subscription.UpsertParams) error {
	ctx, span := tracer.Start(ctx, "uow.SyncSubscription")
	defer span.End()

	err := s.uow.RunInTx(ctx, func(repos Repositories) error {
		// Serialize syncs per user so concurrent webhooks for different
		// subscriptions can't compute is_pro from each other's uncommitted rows.
		err := repos.Users.LockById(ctx, params.UserID)
		if errors.Is(err, user.ErrNotFound) {
			return subscription.ErrUserNotFound
		}
		if err != nil {
			return err
		}

		changed, err := repos.Subscriptions.Upsert(ctx, params)
		if err != nil {
			return err
		}
		if !changed {
			return nil
		}

		subs, err := repos.Subscriptions.ListByUserID(ctx, params.UserID)
		if err != nil {
			return err
		}

		isPro := false
		for _, sub := range subs {
			if subscription.IsProStatus(sub.Status) {
				isPro = true
				break
			}
		}

		_, err = repos.Users.UpdateIsPro(ctx, params.UserID, isPro)
		return err
	})
	if errors.Is(err, subscription.ErrUserNotFound) {
		slog.WarnContext(ctx, "polar subscription for unknown user", slog.String("user_id", params.UserID.String()), slog.String("polar_subscription_id", params.PolarSubscriptionID))
		return nil
	}
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return err
}
