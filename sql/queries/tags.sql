-- name: ListTags :many
SELECT t.*, COUNT(lt.link_id) AS link_count
FROM tags t
    LEFT JOIN link_tags lt ON t.id = lt.tag_id
WHERE t.user_id = $1
GROUP BY t.id
ORDER BY t.name;

-- name: GetTag :one
SELECT * FROM tags
WHERE id = $1 AND user_id = $2
LIMIT 1;

-- name: CreateTags :execrows
INSERT INTO tags (user_id, name)
SELECT $1, unnest(sqlc.arg('names')::text [])
ON CONFLICT (user_id, name) DO NOTHING;

-- name: CountTags :one
SELECT count(*) FROM tags
WHERE user_id = $1;

-- name: GetTagsByNames :many
SELECT * FROM tags
WHERE user_id = $1 AND name = ANY(sqlc.arg('names')::text []);

-- name: UpdateTag :one
UPDATE tags
SET name = $3
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteTag :one
DELETE FROM tags
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteTagCascade :one
WITH deleted_links AS (
    DELETE FROM links
    WHERE links.user_id = $2
        AND links.id IN (SELECT link_tags.link_id FROM link_tags WHERE link_tags.tag_id = $1)
)
DELETE FROM tags
WHERE tags.id = $1 AND tags.user_id = $2
RETURNING *;

-- name: GetTagsForLink :many
SELECT t.*
FROM tags t
    JOIN link_tags lt ON t.id = lt.tag_id
WHERE lt.link_id = $1;

-- name: GetTagsForLinks :many
SELECT lt.link_id, t.*
FROM link_tags lt
    JOIN tags t ON t.id = lt.tag_id
WHERE lt.link_id = ANY(sqlc.arg('link_ids')::uuid []);

-- name: DeleteLinkTags :exec
DELETE FROM link_tags
WHERE link_id = $1;

-- name: CreateLinkTags :exec
INSERT INTO link_tags (link_id, tag_id)
SELECT $1, unnest(sqlc.arg('tag_ids')::uuid []);
