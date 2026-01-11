-- name: GetByID :one
SELECT * FROM products WHERE id = $1 LIMIT 1;

-- name: GetBySlug :one
SELECT * FROM products WHERE slug = $1 LIMIT 1;
