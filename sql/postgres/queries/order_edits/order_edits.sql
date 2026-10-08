-- name: Create :one
INSERT INTO order_edits (order_id, actor, comment, lines_before, lines_after, grand_total_before, grand_total_after, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, now())
	RETURNING *;

-- name: ListByOrderID :many
SELECT * FROM order_edits WHERE order_id = $1 ORDER BY created_at DESC;
