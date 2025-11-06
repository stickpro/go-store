-- name: UpdateStatus :exec
UPDATE product_reviews
	SET  status=$1, updated_at=now()
	WHERE id=$2;

-- name: GetByID :one
SELECT * FROM product_reviews
WHERE id=$1 LIMIT 1;

-- name: SoftDelete :exec
UPDATE product_reviews
SET deleted_at= now(), updated_at=now()
WHERE id=$1;

-- name: Restore :one
UPDATE product_reviews
SET deleted_at= null, updated_at=now()
WHERE id=$1
RETURNING *;