-- name: ClonePublicCollection :one
WITH source AS (
    SELECT c.*
    FROM collections c
    JOIN users u ON u.id = c.user_id
    WHERE c.id = sqlc.arg(source_id)
        AND c.public = TRUE
        AND (u.is_pro = TRUE OR u.is_admin = TRUE)
), cloned AS (
    INSERT INTO collections (user_id, name, description, public)
    SELECT sqlc.arg(user_id), name, description, FALSE
    FROM source
    WHERE user_id <> sqlc.arg(user_id)
    RETURNING *
), cloned_links AS (
    INSERT INTO links (
        user_id, collection_id, title, description, url,
        favourite, for_later, created_at
    )
    SELECT c.user_id, c.id, l.title, l.description, l.url,
        FALSE, FALSE, l.created_at
    FROM source s
    JOIN links l ON l.collection_id = s.id AND l.user_id = s.user_id
    CROSS JOIN cloned c
)
SELECT id, user_id, name, description, public, created_at, updated_at, TRUE AS cloned
FROM cloned
UNION ALL
SELECT id, user_id, name, description, public, created_at, updated_at, FALSE AS cloned
FROM source
WHERE user_id = sqlc.arg(user_id);

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
