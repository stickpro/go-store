-- name: GetByNumber :one
SELECT * FROM orders WHERE order_number = $1 LIMIT 1;

-- name: GetByNumberForUpdate :one
-- Row-locks the order for a status transition. Transaction only.
SELECT * FROM orders WHERE order_number = $1 LIMIT 1 FOR UPDATE;

-- name: GetByIdempotencyKey :one
SELECT * FROM orders WHERE idempotency_key = $1 LIMIT 1;

-- name: ListByUser :many
SELECT * FROM orders
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountByUser :one
SELECT count(*) FROM orders WHERE user_id = $1;

-- name: UpdateStatus :one
UPDATE orders
SET status = $2,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: MarkPaid :one
UPDATE orders
SET status = $2,
    payment_status = 'paid',
    payment_method = $3,
    paid_at = now(),
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: MarkCancelled :one
UPDATE orders
SET status = 'cancelled',
    cancelled_at = now(),
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: MarkRefunded :one
-- Refund also clears the payment status back to 'refunded' (unlike a plain
-- status transition, which never touches payment_status).
UPDATE orders
SET status = 'refunded',
    payment_status = 'refunded',
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: ListAdmin :many
-- Every filter is optional (NULL = don't filter on it); used by the admin order list.
SELECT * FROM orders
WHERE (sqlc.narg('status')::varchar IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('payment_status')::varchar IS NULL OR payment_status = sqlc.narg('payment_status'))
  AND (sqlc.narg('user_id')::uuid IS NULL OR user_id = sqlc.narg('user_id'))
  AND (sqlc.narg('created_from')::timestamp IS NULL OR created_at >= sqlc.narg('created_from'))
  AND (sqlc.narg('created_to')::timestamp IS NULL OR created_at <= sqlc.narg('created_to'))
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountAdmin :one
SELECT count(*) FROM orders
WHERE (sqlc.narg('status')::varchar IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('payment_status')::varchar IS NULL OR payment_status = sqlc.narg('payment_status'))
  AND (sqlc.narg('user_id')::uuid IS NULL OR user_id = sqlc.narg('user_id'))
  AND (sqlc.narg('created_from')::timestamp IS NULL OR created_at >= sqlc.narg('created_from'))
  AND (sqlc.narg('created_to')::timestamp IS NULL OR created_at <= sqlc.narg('created_to'));

-- name: ListExpiredPending :many
-- Callers must run this inside a transaction; the row locks are held until commit.
SELECT id
FROM orders
WHERE status = 'pending'
  AND created_at < $1
ORDER BY created_at
LIMIT $2
FOR UPDATE SKIP LOCKED;
