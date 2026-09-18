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

-- name: GetPublicCollection :many
SELECT c.name, c.description, c.created_at, c.updated_at,
    u.display_name, u.username,
    l.id AS link_id,
    l.title AS link_title,
    l.description AS link_description,
    l.url AS link_url,
    l.created_at AS link_created_at
FROM collections c
JOIN users u ON u.id = c.user_id
LEFT JOIN links l ON l.collection_id = c.id AND l.user_id = c.user_id
WHERE c.id = $1
    AND c.public = TRUE
    AND (u.is_pro = TRUE OR u.is_admin = TRUE)
ORDER BY l.created_at DESC, l.id DESC;
