-- name: Create :one
INSERT INTO attribute_groups (name, slug, description, created_at)
	VALUES ($1, $2, $3, now())
	RETURNING *;

-- name: Delete :exec
DELETE FROM attribute_groups WHERE id=$1;

-- name: GetAll :many
SELECT * FROM attribute_groups ORDER BY name DESC LIMIT $1 OFFSET $2;

-- name: Update :one
UPDATE attribute_groups
	SET name=$1, slug=$2, description=$3, updated_at=now()
	WHERE id=$4
	RETURNING *;

