-- name: CreateCollection :one
INSERT INTO collections (user_id, name, description, public)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetCollection :one
SELECT * FROM collections
WHERE id = $1 AND user_id = $2
LIMIT 1;

-- name: ListCollections :many
SELECT c.*, COUNT(l.id) AS link_count
FROM collections c
    LEFT JOIN links l ON l.collection_id = c.id
WHERE c.user_id = $1
GROUP BY c.id
ORDER BY c.name;

-- name: UpdateCollection :one
UPDATE collections
SET name = $3, description = $4, public = $5
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteCollection :one
DELETE FROM collections
WHERE id = $1 AND user_id = $2
RETURNING *;
