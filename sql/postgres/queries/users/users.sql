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

-- name: MarkEmailVerified :exec
UPDATE users
SET email_verified_at = now(),
    updated_at        = now()
WHERE id = $1;

-- name: SetPassword :exec
UPDATE users
SET password   = $2,
    updated_at = now()
WHERE email = $1;
