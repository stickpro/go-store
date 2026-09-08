-- name: ListByOrderID :many
SELECT * FROM order_items WHERE order_id = $1 ORDER BY id;

-- name: ListByOrderIDs :many
SELECT * FROM order_items
WHERE order_id = ANY ($1::uuid[])
ORDER BY order_id, id;
