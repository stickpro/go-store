-- name: Create :one
INSERT INTO payment_refunds (payment_id, amount, status, idempotency_key, reason, actor, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, now())
	RETURNING *;

-- name: GetByIdempotencyKey :one
SELECT * FROM payment_refunds WHERE idempotency_key = $1 LIMIT 1;

-- name: SumPendingByPaymentID :one
-- Refunds sent to the provider whose outcome isn't known yet. They count
-- against the payment's remaining balance until resolved.
SELECT coalesce(sum(amount), 0)::numeric AS total
FROM payment_refunds
WHERE payment_id = $1 AND status = 'pending';

-- name: SumPendingByOrderID :one
-- Refunds still in flight across all of an order's payments.
SELECT coalesce(sum(r.amount), 0)::numeric AS total
FROM payment_refunds r
JOIN payments p ON p.id = r.payment_id
WHERE p.order_id = $1 AND r.status = 'pending';

-- name: ResolvePending :one
-- Settles a pending refund. Matches only a still-pending row, so a refund a
-- webhook already settled is not resolved (or counted) a second time.
UPDATE payment_refunds
SET status       = $2,
    raw_response = $3,
    updated_at   = now()
WHERE id = $1 AND status = 'pending'
RETURNING *;

-- name: SucceedAllPendingByPaymentID :exec
-- A full-refund webhook proves every refund still in flight went through.
UPDATE payment_refunds
SET status     = 'succeeded',
    updated_at = now()
WHERE payment_id = $1 AND status = 'pending';

-- name: ListByOrderID :many
SELECT r.*
FROM payment_refunds r
JOIN payments p ON p.id = r.payment_id
WHERE p.order_id = $1
ORDER BY r.created_at DESC;
