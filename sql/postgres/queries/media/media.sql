-- name: Get :one
SELECT * FROM media WHERE id = $1 LIMIT 1;