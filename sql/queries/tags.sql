-- name: ListTags :many
SELECT t.*
FROM tags t
    LEFT JOIN link_tags lt ON t.id = lt.tag_id
WHERE t.user_id = $1
GROUP BY t.id
ORDER BY COUNT(lt.link_id) DESC, t.name;

-- name: GetTag :one
SELECT * FROM tags
WHERE id = $1 AND user_id = $2
LIMIT 1;

-- name: UpsertTag :one
INSERT INTO tags (user_id, name)
VALUES ($1, $2)
ON CONFLICT (user_id, name) DO NOTHING
RETURNING *;

-- name: GetTagByName :one
SELECT * FROM tags
WHERE user_id = $1 AND name = $2
LIMIT 1;

-- name: UpdateTag :one
UPDATE tags
SET name = $3
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteTag :one
DELETE FROM tags
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: GetTagsForLink :many
SELECT t.name
FROM tags t
    JOIN link_tags lt ON t.id = lt.tag_id
WHERE lt.link_id = $1;

-- name: GetTagsForLinks :many
SELECT lt.link_id, t.name
FROM link_tags lt
    JOIN tags t ON t.id = lt.tag_id
WHERE lt.link_id = ANY($1::uuid []);

-- name: DeleteLinkTags :exec
DELETE FROM link_tags
WHERE link_id = $1;

-- name: CreateLinkTag :exec
INSERT INTO link_tags (link_id, tag_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;
