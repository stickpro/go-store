-- name: GetByID :one
SELECT * FROM attribute_groups WHERE id = $1 LIMIT 1;


