-- name: Create :one
INSERT INTO order_items (order_id, product_id, variant_id, sku, name, slug, image_path, unit_price, quantity, line_total, weight_kg, length_cm, width_cm, height_cm)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	RETURNING *;
