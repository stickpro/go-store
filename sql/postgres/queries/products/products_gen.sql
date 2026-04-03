-- name: Create :one
INSERT INTO products (external_id, manufacturer_id, sku, upc, ean, jan, isbn, mpn, location, quantity, stock_status, price_retail, price_business, price_wholesale, weight, length, width, height, subtract, minimum, image, sort_order, is_enable, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, now())
	RETURNING *;

-- name: Update :one
UPDATE products
	SET external_id=$1, manufacturer_id=$2, sku=$3, upc=$4, ean=$5, jan=$6, 
		isbn=$7, mpn=$8, location=$9, quantity=$10, stock_status=$11, price_retail=$12, 
		price_business=$13, price_wholesale=$14, weight=$15, length=$16, width=$17, height=$18, 
		subtract=$19, minimum=$20, image=$21, sort_order=$22, is_enable=$23, updated_at=now()
	WHERE id=$24
	RETURNING *;

