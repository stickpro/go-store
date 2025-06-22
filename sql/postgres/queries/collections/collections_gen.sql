-- name: Create :one
INSERT INTO collections (name, description, slug, created_at)
	VALUES ($1, $2, $3, now())
	RETURNING *;

-- name: Delete :exec
DELETE FROM collections WHERE id=$1;

-- name: GetAll :many
SELECT * FROM collections LIMIT $1 OFFSET $2;

-- name: Get :one
SELECT * FROM collections WHERE id=$1 LIMIT 1;

-- name: Update :one
UPDATE collections
	SET name=$1, description=$2, slug=$3, updated_at=now()
	WHERE id=$4
	RETURNING *;

