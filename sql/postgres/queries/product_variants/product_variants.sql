-- name: GetBySlug :one
SELECT * FROM product_variants WHERE slug = $1 LIMIT 1;

-- name: GetByProductID :many
SELECT * FROM product_variants WHERE product_id = $1 ORDER BY sort_order ASC;
