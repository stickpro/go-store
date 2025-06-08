-- name: Create :one
INSERT INTO attributes (attribute_group_id, name, type, is_filterable, is_visible, sort_order, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, now())
	RETURNING *;

-- name: GetAll :many
SELECT * FROM attributes ORDER BY sort_order DESC LIMIT $1 OFFSET $2;

-- name: Update :one
UPDATE attributes
	SET name=$1, type=$2, is_filterable=$3, is_visible=$4, sort_order=$5, updated_at=now()
	WHERE id=$6
	RETURNING *;

