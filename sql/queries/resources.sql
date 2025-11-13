-- name: GetResource :one
SELECT * FROM resources
WHERE id = $1 LIMIT 1;

-- name: ListResources :many
SELECT * FROM resources
ORDER BY title;

-- name: CreateResource :one
INSERT INTO resources (
  title, url
) VALUES (
  $1, $2
)
RETURNING *;

-- name: UpdateResource :one
UPDATE resources
  SET title = $2,
  url = $3
WHERE id = $1
RETURNING *;

-- name: DeleteResource :exec
DELETE FROM resources
WHERE id = $1;
