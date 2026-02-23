-- name: Create :one
INSERT INTO product_variants (product_id, category_id, name, slug, description, meta_title, meta_h1, meta_description, meta_keyword, image, sort_order, is_enable, viewed, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, now())
	RETURNING *;

-- name: Delete :exec
DELETE FROM product_variants WHERE id=$1;

-- name: GetAll :many
SELECT * FROM product_variants LIMIT $1 OFFSET $2;

-- name: Get :one
SELECT * FROM product_variants WHERE id=$1 LIMIT 1;

-- name: Update :one
UPDATE product_variants
	SET product_id=$1, category_id=$2, name=$3, slug=$4, description=$5, meta_title=$6, 
		meta_h1=$7, meta_description=$8, meta_keyword=$9, image=$10, sort_order=$11, is_enable=$12, 
		viewed=$13, updated_at=now()
	WHERE id=$14
	RETURNING *;

