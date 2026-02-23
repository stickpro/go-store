-- name: Create :one
INSERT INTO products (external_id, manufacturer_id, model, sku, upc, ean, jan, isbn, mpn, location, quantity, stock_status, price, weight, length, width, height, subtract, minimum, sort_order, is_enable, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, now())
	RETURNING *;

-- name: Update :one
UPDATE products
	SET external_id=$1, manufacturer_id=$2, model=$3, sku=$4, upc=$5, ean=$6, 
		jan=$7, isbn=$8, mpn=$9, location=$10, quantity=$11, stock_status=$12, 
		price=$13, weight=$14, length=$15, width=$16, height=$17, subtract=$18, 
		minimum=$19, sort_order=$20, is_enable=$21, updated_at=now()
	WHERE id=$22
	RETURNING *;

