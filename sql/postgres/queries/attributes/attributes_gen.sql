-- name: Create :one
INSERT INTO attributes (attribute_group_id, name, value, type, is_filterable, is_visible, sort_order, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, now())
	RETURNING *;

-- name: Delete :exec
DELETE FROM attributes WHERE id=$1;

-- name: GetAll :many
SELECT * FROM attributes ORDER BY sort_order DESC LIMIT $1 OFFSET $2;

-- name: Update :one
UPDATE attributes
	SET attribute_group_id=$1, name=$2, value=$3, type=$4, is_filterable=$5, is_visible=$6, 
		sort_order=$7, updated_at=now()
	WHERE id=$8
	RETURNING *;

