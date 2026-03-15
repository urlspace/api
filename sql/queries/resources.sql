-- name: GetResource :one
SELECT * FROM resources
WHERE
    id = $1 AND user_id = $2
LIMIT 1;

-- name: ListResources :many
SELECT * FROM resources
WHERE user_id = $1
ORDER BY created_at;

-- name: CreateResource :one
INSERT INTO resources (
    user_id, title, description, url, favourite, read_later
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: UpdateResource :one
UPDATE resources
SET
    title = $3,
    description = $4,
    url = $5,
    favourite = $6,
    read_later = $7
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteResource :one
DELETE FROM resources
WHERE id = $1 AND user_id = $2
RETURNING *;
