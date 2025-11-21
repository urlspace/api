-- name: GetResource :one
SELECT * FROM resources
WHERE id =
$1 LIMIT 1;

-- name: ListResources :many
SELECT * FROM resources
ORDER BY title;

-- name: CreateResource :one
INSERT INTO resources (
  title, description, url, favourite, read_later
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: UpdateResource :one
UPDATE resources
  SET title = $2,
  description = $3,
  url = $4,
  favourite = $5,
  read_later = $6
WHERE id = $1
RETURNING *;

-- name: DeleteResource :exec
DELETE FROM resources
WHERE id = $1;

