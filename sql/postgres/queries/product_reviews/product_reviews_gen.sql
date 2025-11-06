-- name: Create :one
INSERT INTO product_reviews (product_id, user_id, order_id, rating, title, body, status, created_at, deleted_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, now(), $8)
	RETURNING *;

-- name: Delete :exec
DELETE FROM product_reviews WHERE id=$1;

-- name: Update :one
UPDATE product_reviews
	SET product_id=$1, user_id=$2, order_id=$3, rating=$4, title=$5, body=$6, 
		status=$7, updated_at=now(), deleted_at=$8
	WHERE id=$9
	RETURNING *;

