-- name: Create :one
INSERT INTO attributes (attribute_group_id, name, value, type, is_filterable, is_visible, sort_order, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, now())
	RETURNING *;

-- name: GetAll :many
SELECT * FROM attributes ORDER BY sort_order DESC LIMIT $1 OFFSET $2;

-- name: Update :one
UPDATE attributes
	SET name=$1, value=$2, type=$3, is_filterable=$4, is_visible=$5, sort_order=$6, 
		updated_at=now()
	WHERE id=$7
	RETURNING *;

