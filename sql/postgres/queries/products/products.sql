-- name: GetByID :one
SELECT * FROM products WHERE id = $1 LIMIT 1;

-- name: GetBySlug :one
SELECT * FROM products WHERE slug = $1 LIMIT 1;

-- name: AddAttributesToProduct :exec
INSERT INTO attribute_products (attribute_id, product_id)
SELECT a.id, sqlc.arg(product_id)::uuid
FROM attributes a
WHERE a.id = ANY(sqlc.arg(attribute_ids)::uuid[])
ON CONFLICT DO NOTHING;

-- name: DeleteAttributesFromProduct :exec
DELETE FROM attribute_products
WHERE product_id = $1;