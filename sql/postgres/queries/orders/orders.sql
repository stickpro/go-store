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

-- name: HasUserPurchasedVariant :one
-- Newest paid, non-cancelled/refunded order of the user that contains the
-- variant. Gates review creation to actually-purchased variants; the returned
-- order id is stored on the review.
SELECT o.id
FROM orders o
JOIN order_items oi ON oi.order_id = o.id
WHERE o.user_id = sqlc.arg('user_id')::uuid
  AND oi.variant_id = sqlc.arg('variant_id')::uuid
  AND o.payment_status = 'paid'
  AND o.status NOT IN ('cancelled', 'refunded')
ORDER BY o.created_at DESC
LIMIT 1;

-- name: DashboardOrderStats :one
-- One-shot order snapshot for the admin dashboard. Status / payment buckets and
-- `total` are all-time (current distribution); `today` and `revenue_today` use
-- the day-boundary params; `revenue_period` / `paid_orders_period` use the
-- selected range. Revenue sums grand_total of payment_status = 'paid' orders only.
SELECT
    count(*)                                                                        AS total,
    count(*) FILTER (WHERE created_at >= sqlc.arg('today_from')
                       AND created_at < sqlc.arg('today_to'))                       AS today,

    count(*) FILTER (WHERE status = 'pending')                                      AS status_pending,
    count(*) FILTER (WHERE status = 'paid')                                         AS status_paid,
    count(*) FILTER (WHERE status = 'processing')                                   AS status_processing,
    count(*) FILTER (WHERE status = 'shipped')                                      AS status_shipped,
    count(*) FILTER (WHERE status = 'delivered')                                    AS status_delivered,
    count(*) FILTER (WHERE status = 'cancelled')                                    AS status_cancelled,
    count(*) FILTER (WHERE status = 'refunded')                                     AS status_refunded,

    count(*) FILTER (WHERE payment_status = 'unpaid')                               AS payment_unpaid,
    count(*) FILTER (WHERE payment_status = 'paid')                                 AS payment_paid,
    count(*) FILTER (WHERE payment_status = 'refunded')                             AS payment_refunded,
    count(*) FILTER (WHERE payment_status = 'failed')                               AS payment_failed,

    coalesce(sum(grand_total) FILTER (WHERE payment_status = 'paid'
                                        AND created_at >= sqlc.arg('today_from')
                                        AND created_at < sqlc.arg('today_to')), 0)::numeric  AS revenue_today,
    coalesce(sum(grand_total) FILTER (WHERE payment_status = 'paid'
                                        AND created_at >= sqlc.arg('period_from')
                                        AND created_at < sqlc.arg('period_to')), 0)::numeric AS revenue_period,
    count(*) FILTER (WHERE payment_status = 'paid'
                       AND created_at >= sqlc.arg('period_from')
                       AND created_at < sqlc.arg('period_to'))                      AS paid_orders_period
FROM orders;

-- name: ListExpiredPending :many
-- Callers must run this inside a transaction; the row locks are held until commit.
SELECT id
FROM orders
WHERE status = 'pending'
  AND created_at < $1
ORDER BY created_at
LIMIT $2
FOR UPDATE SKIP LOCKED;
