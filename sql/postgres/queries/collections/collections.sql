-- name: GetBySlug :one
SELECT * FROM collections WHERE slug=$1 LIMIT 1;

