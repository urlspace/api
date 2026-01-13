-- name: GetUserById :one
SELECT * FROM users
WHERE
    id
    =
    $1
LIMIT 1;

-- name: GetUserByEmailVerificationToken :one
SELECT * FROM users
WHERE
    email_verification_token
    =
    $1
LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at;

-- name: CreateUser :one
INSERT INTO users (
    email,
    email_verified,
    email_verification_token,
    email_verification_token_expires_at,
    password,
    username,
    is_admin,
    is_pro
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: VerifyUser :one
UPDATE users
SET
    email_verified = TRUE,
    email_verification_token = NULL,
    email_verification_token_expires_at = NULL
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
