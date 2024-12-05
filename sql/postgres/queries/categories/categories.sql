-- name: GetByID :one
SELECT * FROM categories WHERE id = $1;

-- name: GetBySlug :one
SELECT * FROM categories WHERE slug = $1;

