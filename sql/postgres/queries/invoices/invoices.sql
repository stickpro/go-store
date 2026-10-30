-- name: NextNumber :one
-- Reserves the next invoice number. Taken before the insert so the number is
-- part of the immutable snapshot in data.
SELECT nextval('invoice_number_seq')::bigint AS number;

-- name: Create :one
INSERT INTO invoices (order_id, number, amount, currency, data, issued_at, due_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING *;

-- name: GetActiveByOrderID :one
SELECT * FROM invoices WHERE order_id = $1 AND status = 'active' LIMIT 1;

-- name: VoidActiveByOrderID :exec
UPDATE invoices SET status = 'void', voided_at = now()
WHERE order_id = $1 AND status = 'active';
