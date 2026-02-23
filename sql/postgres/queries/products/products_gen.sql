-- name: Create :one
INSERT INTO products (manufacturer_id, model, sku, upc, ean, jan, isbn, mpn, location, quantity, stock_status, price, weight, length, width, height, subtract, minimum, sort_order, is_enable, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, now())
	RETURNING *;

-- name: Update :one
UPDATE products
	SET manufacturer_id=$1, model=$2, sku=$3, upc=$4, ean=$5, jan=$6, 
		isbn=$7, mpn=$8, location=$9, quantity=$10, stock_status=$11, price=$12, 
		weight=$13, length=$14, width=$15, height=$16, subtract=$17, minimum=$18, 
		sort_order=$19, is_enable=$20, updated_at=now()
	WHERE id=$21
	RETURNING *;

