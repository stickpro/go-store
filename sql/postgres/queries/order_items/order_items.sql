-- name: ListByOrderID :many
SELECT * FROM order_items WHERE order_id = $1 ORDER BY id;

-- name: ListByOrderIDs :many
SELECT * FROM order_items
WHERE order_id = ANY ($1::uuid[])
ORDER BY order_id, id;

-- name: GetVerifiedPurchaseOrderID :one
-- Most recent paid-or-later order by this user that contains the given variant.
-- Drives the "verified purchase" gate on product reviews; returns no row when
-- the user has not bought the variant.
SELECT o.id
FROM orders o
         JOIN order_items oi ON oi.order_id = o.id
WHERE o.user_id = sqlc.arg('user_id')::uuid
  AND oi.variant_id = sqlc.arg('variant_id')::uuid
  AND o.status IN ('paid', 'processing', 'shipped', 'delivered')
ORDER BY o.created_at DESC
LIMIT 1;
