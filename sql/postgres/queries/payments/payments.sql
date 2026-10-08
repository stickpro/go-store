-- name: GetByID :one
SELECT * FROM payments WHERE id = $1 LIMIT 1;

-- name: GetLatestByOrderID :one
-- Most recent payment attempt for the order (a customer may retry after a
-- failed/expired attempt, so an order can have more than one row).
SELECT * FROM payments WHERE order_id = $1 ORDER BY created_at DESC LIMIT 1;

-- name: ListByOrderID :many
SELECT * FROM payments WHERE order_id = $1 ORDER BY created_at DESC;

-- name: GetByProviderPaymentID :one
SELECT * FROM payments WHERE provider = $1 AND provider_payment_id = $2 LIMIT 1;

-- name: UpdateAfterInit :one
-- Fills in what the provider's Init call returned: its own payment id, the
-- URL to redirect the customer to, and the raw response for audit/debugging.
UPDATE payments
SET provider_payment_id = $2,
    payment_url         = $3,
    status               = $4,
    raw_init_response    = $5,
    updated_at           = now()
WHERE id = $1
RETURNING *;

-- name: UpdateStatus :one
-- Transitions a payment on a provider webhook/response (notification or
-- cancel/refund). raw_last_notification is overwritten with the latest
-- payload for audit/debugging.
UPDATE payments
SET status                = $2,
    raw_last_notification = $3,
    updated_at             = now()
WHERE id = $1
RETURNING *;

-- name: GetByIDForUpdate :one
SELECT * FROM payments WHERE id = $1 LIMIT 1 FOR UPDATE;

-- name: GetRefundableByOrderIDForUpdate :one
-- The order's captured payment that still has money left to refund, locked
-- so concurrent refunds see each other's pending rows.
SELECT * FROM payments
WHERE order_id = $1 AND status IN ('confirmed', 'partially_refunded')
ORDER BY created_at DESC
LIMIT 1
FOR UPDATE;

-- name: ApplyRefund :one
UPDATE payments
SET refunded_amount = $2,
    status          = $3,
    updated_at      = now()
WHERE id = $1
RETURNING *;
