-- name: GetByEmail :one
SELECT *
FROM users
WHERE email = $1
LIMIT 1;

-- name: GetByID :one
SELECT *
FROM users
WHERE id = $1
LIMIT 1;