-- name: CreateToken :one
INSERT INTO tokens (user_id, type, description, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetTokenById :one
SELECT * FROM tokens
WHERE id = $1
LIMIT 1;

-- name: UpdateTokenExpiresAt :one
UPDATE tokens
SET expires_at = $2
WHERE id = $1
RETURNING *;

-- name: DeleteToken :exec
DELETE FROM tokens
WHERE id = $1;
