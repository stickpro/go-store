-- name: GetBySlug :one
SELECT * FROM collections WHERE slug=$1 LIMIT 1;

-- name: DeleteProductsFromCollection :exec
DELETE FROM collection_products
WHERE collection_id = $1;

-- name: AddProductsToCollection :exec
INSERT INTO collection_products (collection_id, product_id)
SELECT sqlc.arg(collection_id)::uuid, p.id
FROM products p
WHERE p.id = any(sqlc.arg(product_ids)::uuid[])
  AND NOT EXISTS (
    SELECT 1
    FROM collection_products cp
    WHERE cp.collection_id = sqlc.arg(collection_id)
  AND cp.product_id = p.id
    );

-- name: GetCollectionWithProductsBySlug :many
SELECT c.*,
       p.id        AS product_id,
       p.name      AS product_name,
       p.slug      AS product_slug,
       p.model     AS product_model,
       p.price     AS product_price,
       p.is_enable AS product_is_enable,
       p.image     AS product_image
FROM collections c
         LEFT JOIN collection_products cp ON cp.collection_id = c.id
         LEFT JOIN products p ON p.id = cp.product_id
WHERE c.slug = $1;

-- name: GetCollectionWithProductsByID :many
SELECT c.*,
       p.id        AS product_id,
       p.name      AS product_name,
       p.slug      AS product_slug,
       p.model     AS product_model,
       p.price     AS product_price,
       p.is_enable AS product_is_enable,
       p.image     AS product_image
FROM collections c
         LEFT JOIN collection_products cp ON cp.collection_id = c.id
         LEFT JOIN products p ON p.id = cp.product_id
WHERE c.id = $1;