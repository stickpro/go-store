-- name: GetByID :one
SELECT * FROM products WHERE id = $1 LIMIT 1;

-- name: GetBySlug :one
SELECT p.* FROM products p
INNER JOIN product_variants pv ON pv.product_id = p.id
WHERE pv.slug = $1 LIMIT 1;
