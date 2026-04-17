-- name: CreateSession :one
INSERT INTO sessions (user_id, description, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetSessionById :one
SELECT * FROM sessions
WHERE id = $1
LIMIT 1;

-- name: UpdateSessionExpiresAt :one
UPDATE sessions
SET expires_at = $2
WHERE id = $1
RETURNING *;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE id = $1;

-- name: DeleteSessionsByUserID :exec
DELETE FROM sessions
WHERE user_id = $1;
