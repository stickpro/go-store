-- name: ListByOrderID :many
SELECT * FROM order_items WHERE order_id = $1 ORDER BY id;

-- name: ListByOrderIDs :many
SELECT * FROM order_items
WHERE order_id = ANY ($1::uuid[])
ORDER BY order_id, id;

-- name: UpdateQuantity :one
UPDATE order_items
SET quantity   = $2,
    line_total = $3
WHERE id = $1
RETURNING *;

-- name: DeleteByID :exec
DELETE FROM order_items WHERE id = $1;
