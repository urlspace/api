-- name: GetLink :one
SELECT l.*, c.name AS collection_name
FROM links l
    LEFT JOIN collections c ON l.collection_id = c.id
WHERE l.id = $1 AND l.user_id = $2
LIMIT 1;

-- name: ListLinks :many
SELECT l.*, c.name AS collection_name
FROM links l
    LEFT JOIN collections c ON l.collection_id = c.id
WHERE l.user_id = sqlc.arg('user_id')
  AND (sqlc.narg('collection_id')::uuid IS NULL OR l.collection_id = sqlc.narg('collection_id'))
  AND (sqlc.arg('query')::text = '' OR l.title ILIKE '%' || sqlc.arg('query') || '%')
  AND (cardinality(sqlc.arg('tag_ids')::uuid[]) = 0 OR l.id IN (
      SELECT link_id FROM link_tags
      WHERE tag_id = ANY(sqlc.arg('tag_ids'))
      GROUP BY link_id
      HAVING COUNT(DISTINCT tag_id) = cardinality(sqlc.arg('tag_ids'))
  ))
  AND (sqlc.narg('favourite')::bool IS NULL OR l.favourite = sqlc.narg('favourite'))
  AND (sqlc.narg('for_later')::bool IS NULL OR l.for_later = sqlc.narg('for_later'))
ORDER BY l.created_at DESC, l.id DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountLinks :one
SELECT COUNT(*) FROM links l
WHERE l.user_id = sqlc.arg('user_id')
  AND (sqlc.narg('collection_id')::uuid IS NULL OR l.collection_id = sqlc.narg('collection_id'))
  AND (sqlc.arg('query')::text = '' OR l.title ILIKE '%' || sqlc.arg('query') || '%')
  AND (cardinality(sqlc.arg('tag_ids')::uuid[]) = 0 OR l.id IN (
      SELECT link_id FROM link_tags
      WHERE tag_id = ANY(sqlc.arg('tag_ids'))
      GROUP BY link_id
      HAVING COUNT(DISTINCT tag_id) = cardinality(sqlc.arg('tag_ids'))
  ))
  AND (sqlc.narg('favourite')::bool IS NULL OR l.favourite = sqlc.narg('favourite'))
  AND (sqlc.narg('for_later')::bool IS NULL OR l.for_later = sqlc.narg('for_later'));

-- name: CreateLink :one
INSERT INTO links (
    user_id, title, description, url, collection_id, favourite, for_later
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: UpdateLink :one
UPDATE links
SET
    title = $3,
    description = $4,
    url = $5,
    collection_id = $6,
    favourite = $7,
    for_later = $8
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteLink :one
DELETE FROM links
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: ImportLinks :exec
WITH input AS (
    SELECT
        uuidv7() AS id,
        i.title,
        i.description,
        i.url,
        NULLIF(i.collection_id, '')::uuid AS collection_id,
        i.favourite,
        i.for_later,
        i.created_at,
        i.tags
    FROM ROWS FROM (
        unnest(sqlc.arg('titles')::text []),
        unnest(sqlc.arg('descriptions')::text []),
        unnest(sqlc.arg('urls')::text []),
        unnest(sqlc.arg('collection_ids')::text []),
        unnest(sqlc.arg('favourites')::bool []),
        unnest(sqlc.arg('for_laters')::bool []),
        unnest(sqlc.arg('created_ats')::timestamptz []),
        unnest(sqlc.arg('tags')::text [])
    ) AS i (title, description, url, collection_id, favourite, for_later, created_at, tags)
), inserted_links AS (
    INSERT INTO links (
        id, user_id, title, description, url, collection_id, favourite, for_later, created_at
    )
    SELECT input.id, sqlc.arg('user_id'), input.title, input.description, input.url,
        input.collection_id, input.favourite, input.for_later, input.created_at
    FROM input
)
INSERT INTO link_tags (link_id, tag_id)
SELECT input.id, tags.id
FROM input
    CROSS JOIN LATERAL unnest(string_to_array(input.tags, ',')) AS tag_name
    JOIN tags ON tags.user_id = sqlc.arg('user_id') AND tags.name = tag_name;
