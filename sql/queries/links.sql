-- name: GetLink :one
SELECT l.*, c.title AS collection_title
FROM links l
    LEFT JOIN collections c ON l.collection_id = c.id
WHERE l.id = $1 AND l.user_id = $2
LIMIT 1;

-- name: ListLinks :many
SELECT l.*, c.title AS collection_title
FROM links l
    LEFT JOIN collections c ON l.collection_id = c.id
WHERE l.user_id = $1
ORDER BY l.created_at;

-- name: CreateLink :one
INSERT INTO links (
    user_id, title, description, url, collection_id
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: UpdateLink :one
UPDATE links
SET
    title = $3,
    description = $4,
    url = $5,
    collection_id = $6
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteLink :one
DELETE FROM links
WHERE id = $1 AND user_id = $2
RETURNING *;
