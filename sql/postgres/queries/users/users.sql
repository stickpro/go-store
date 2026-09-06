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

-- name: DashboardCustomerStats :one
-- Customer counters for the admin dashboard. Admin accounts and soft-deleted
-- users are excluded; new_today uses the store-timezone day boundary params.
SELECT
    count(*) FILTER (WHERE created_at >= sqlc.arg('today_from')
                       AND created_at < sqlc.arg('today_to')) AS new_today,
    count(*)                                                  AS total
FROM users
WHERE deleted_at IS NULL
  AND is_admin IS NOT TRUE;
