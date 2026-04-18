-- name: CreateToken :one
INSERT INTO tokens (user_id, description, hash)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetTokenById :one
SELECT * FROM tokens
WHERE id = $1 AND user_id = $2
LIMIT 1;

-- name: GetTokenByHash :one
SELECT * FROM tokens
WHERE hash = $1
LIMIT 1;

-- name: ListTokensByUserID :many
SELECT * FROM tokens
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: UpdateTokenLastUsedAt :exec
UPDATE tokens
SET last_used_at = now()
WHERE id = $1;

-- name: DeleteToken :exec
DELETE FROM tokens
WHERE id = $1 AND user_id = $2;

-- name: DeleteTokensByUserID :exec
DELETE FROM tokens
WHERE user_id = $1;
