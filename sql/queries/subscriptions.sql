-- name: UpsertSubscription :execrows
INSERT INTO subscriptions (
    user_id,
    polar_subscription_id,
    status,
    current_period_end,
    cancel_at_period_end,
    polar_modified_at
) VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (polar_subscription_id) DO UPDATE
SET
    status = excluded.status,
    current_period_end = excluded.current_period_end,
    cancel_at_period_end = excluded.cancel_at_period_end,
    polar_modified_at = excluded.polar_modified_at
WHERE subscriptions.polar_modified_at <= excluded.polar_modified_at;

-- name: ListSubscriptionsByUserID :many
SELECT * FROM subscriptions
WHERE user_id = $1
ORDER BY created_at DESC;
