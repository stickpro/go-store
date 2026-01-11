-- name: Create :one
INSERT INTO attributes (attribute_group_id, name, slug, type, unit, is_filterable, is_visible, is_required, sort_order, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now())
	RETURNING *;

-- name: Delete :exec
DELETE FROM attributes WHERE id=$1;

-- name: GetAll :many
SELECT * FROM attributes ORDER BY sort_order DESC LIMIT $1 OFFSET $2;

-- name: Update :one
UPDATE attributes
	SET attribute_group_id=$1, name=$2, slug=$3, type=$4, unit=$5, is_filterable=$6, 
		is_visible=$7, is_required=$8, sort_order=$9, updated_at=now()
	WHERE id=$10
	RETURNING *;

