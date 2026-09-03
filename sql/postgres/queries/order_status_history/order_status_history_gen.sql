-- name: Create :one
INSERT INTO order_status_history (order_id, from_status, to_status, actor, comment)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING *;
