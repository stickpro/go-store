-- name: Create :one
INSERT INTO payments (order_id, provider, provider_payment_id, status, amount, currency, payment_url, raw_init_response, raw_last_notification, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now())
	RETURNING *;

-- name: Delete :exec
DELETE FROM payments WHERE id=$1;
