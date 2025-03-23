-- name: Create :one
INSERT INTO categories (parent_id, name, slug, description, image_path, meta_title, meta_h1, meta_description, meta_keyword, is_enable, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, now())
	RETURNING *;

-- name: Update :one
UPDATE categories
	SET parent_id=$1, name=$2, slug=$3, description=$4, image_path=$5, meta_title=$6, 
		meta_h1=$7, meta_description=$8, meta_keyword=$9, is_enable=$10, updated_at=now()
	WHERE id=$11
	RETURNING *;

