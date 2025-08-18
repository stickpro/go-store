-- name: Create :one
INSERT INTO attribute_groups (name, description, created_at)
	VALUES ($1, $2, now())
	RETURNING *;

-- name: Delete :exec
DELETE FROM attribute_groups WHERE id=$1;

-- name: GetAll :many
SELECT * FROM attribute_groups ORDER BY name DESC LIMIT $1 OFFSET $2;

-- name: Update :one
UPDATE attribute_groups
	SET name=$1, description=$2, updated_at=now()
	WHERE id=$3
	RETURNING *;

