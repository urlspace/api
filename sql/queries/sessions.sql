-- name: CreateSession :one
INSERT INTO sessions (user_id, session_hash, description, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetSessionByHash :one
SELECT * FROM sessions
WHERE session_hash = $1
LIMIT 1;

-- name: UpdateSessionExpiresAt :one
UPDATE sessions
SET expires_at = $2
WHERE id = $1
RETURNING *;

-- name: ListSessionsByUserID :many
SELECT * FROM sessions
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: DeleteSessionByID :exec
DELETE FROM sessions
WHERE id = $1 AND user_id = $2;

-- name: DeleteSessionByHash :exec
DELETE FROM sessions
WHERE session_hash = $1;

-- name: DeleteSessionsByUserID :exec
DELETE FROM sessions
WHERE user_id = $1;
