-- name: CreateSession :one
INSERT INTO sessions (user_id, session_hash, user_agent, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetSessionByHash :one
SELECT s.*, u.is_pro, u.is_admin, u.is_blocked
FROM sessions s
JOIN users u ON u.id = s.user_id
WHERE s.session_hash = $1
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
