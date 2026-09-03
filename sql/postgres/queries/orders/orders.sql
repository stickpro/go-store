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

-- name: ListExpiredPending :many
-- Callers must run this inside a transaction; the row locks are held until commit.
SELECT id
FROM orders
WHERE status = 'pending'
  AND created_at < $1
ORDER BY created_at
LIMIT $2
FOR UPDATE SKIP LOCKED;
