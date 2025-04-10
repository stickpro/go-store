-- name: GetByID :one
SELECT * FROM manufacturers WHERE id = $1 LIMIT 1;

-- name: GetBySlug :one
SELECT * FROM manufacturers WHERE slug = $1 LIMIT 1;

