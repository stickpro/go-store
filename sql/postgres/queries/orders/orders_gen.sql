-- name: Create :one
INSERT INTO orders (user_id, status, payment_status, payment_method, currency, email, phone, ship_city_id, ship_city_name, ship_address, ship_postcode, ship_recipient, shipping_method, subtotal, discount_total, shipping_total, tax_total, grand_total, comment, idempotency_key)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
	RETURNING *;

-- name: Get :one
SELECT * FROM orders WHERE id=$1 LIMIT 1;
