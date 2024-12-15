-- name: Create :one
INSERT INTO products (name, model, slug, description, meta_title, meta_h1, meta_description, meta_keyword, sku, upc, ean, jan, isbn, mpn, location, quantity, stock_status, image, manufacturer_id, price, weight, length, width, height, subtract, minimum, sort_order, is_enable, viewed, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, now())
	RETURNING *;

-- name: Update :one
UPDATE products
	SET name=$1, model=$2, slug=$3, description=$4, meta_title=$5, meta_h1=$6, 
		meta_description=$7, meta_keyword=$8, sku=$9, upc=$10, ean=$11, jan=$12, 
		isbn=$13, mpn=$14, location=$15, quantity=$16, stock_status=$17, image=$18, 
		manufacturer_id=$19, price=$20, weight=$21, length=$22, width=$23, height=$24, 
		subtract=$25, minimum=$26, sort_order=$27, is_enable=$28, viewed=$29, updated_at=now()
	WHERE id=$30
	RETURNING *;

