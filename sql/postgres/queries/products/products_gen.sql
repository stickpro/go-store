-- name: Create :one
INSERT INTO products (category_id, name, model, slug, description, meta_title, meta_h1, meta_description, meta_keyword, sku, upc, ean, jan, isbn, mpn, location, quantity, stock_status, image, manufacturer_id, price, weight, length, width, height, subtract, minimum, sort_order, is_enable, viewed, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, now())
	RETURNING *;

-- name: Update :one
UPDATE products
	SET category_id=$1, name=$2, model=$3, slug=$4, description=$5, meta_title=$6, 
		meta_h1=$7, meta_description=$8, meta_keyword=$9, sku=$10, upc=$11, ean=$12, 
		jan=$13, isbn=$14, mpn=$15, location=$16, quantity=$17, stock_status=$18, 
		image=$19, manufacturer_id=$20, price=$21, weight=$22, length=$23, width=$24, 
		height=$25, subtract=$26, minimum=$27, sort_order=$28, is_enable=$29, viewed=$30, 
		updated_at=now()
	WHERE id=$31
	RETURNING *;

