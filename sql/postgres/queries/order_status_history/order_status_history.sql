-- name: ListByOrderID :many
SELECT * FROM order_status_history
WHERE order_id = $1
ORDER BY created_at;
