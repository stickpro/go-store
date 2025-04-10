-- name: Create :one
INSERT INTO manufacturers (name, slug, description, image_path, meta_title, meta_h1, meta_description, meta_keyword, is_enable, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now())
	RETURNING *;

-- name: Update :one
UPDATE manufacturers
	SET name=$1, slug=$2, description=$3, image_path=$4, meta_title=$5, meta_h1=$6, 
		meta_description=$7, meta_keyword=$8, is_enable=$9, updated_at=now()
	WHERE id=$10
	RETURNING *;

