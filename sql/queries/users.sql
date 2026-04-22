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

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE
    email
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
    display_name,
    is_admin,
    is_pro
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
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

-- name: UpdateVerificationToken :one
UPDATE users
SET
    email_verification_token = $2,
    email_verification_token_expires_at = $3
WHERE id = $1
RETURNING *;

-- name: GetUserByPasswordResetToken :one
SELECT * FROM users
WHERE password_reset_token = $1
LIMIT 1;

-- name: UpdatePasswordResetToken :one
UPDATE users
SET
    password_reset_token = $2,
    password_reset_token_expires_at = $3
WHERE id = $1
RETURNING *;

-- name: ResetUserPassword :one
UPDATE users
SET
    password = $2,
    password_reset_token = NULL,
    password_reset_token_expires_at = NULL
WHERE id = $1
RETURNING *;

-- name: DeleteUser :one
DELETE FROM users
WHERE id = $1
RETURNING *;
