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
       pv.name     AS product_name,
       pv.slug     AS product_slug,
       p.model     AS product_model,
       p.price     AS product_price,
       p.is_enable AS product_is_enable,
       pv.image    AS product_image
FROM collections c
         LEFT JOIN collection_products cp ON cp.collection_id = c.id
         LEFT JOIN products p ON p.id = cp.product_id
         LEFT JOIN LATERAL (
             SELECT name, slug, image
             FROM product_variants pv2
             WHERE pv2.product_id = p.id
             ORDER BY pv2.sort_order ASC, pv2.created_at ASC
             LIMIT 1
         ) pv ON true
WHERE c.slug = $1;

-- name: GetCollectionWithProductsByID :many
SELECT c.*,
       p.id        AS product_id,
       pv.name     AS product_name,
       pv.slug     AS product_slug,
       p.model     AS product_model,
       p.price     AS product_price,
       p.is_enable AS product_is_enable,
       pv.image    AS product_image
FROM collections c
         LEFT JOIN collection_products cp ON cp.collection_id = c.id
         LEFT JOIN products p ON p.id = cp.product_id
         LEFT JOIN LATERAL (
             SELECT name, slug, image
             FROM product_variants pv2
             WHERE pv2.product_id = p.id
             ORDER BY pv2.sort_order ASC, pv2.created_at ASC
             LIMIT 1
         ) pv ON true
WHERE c.id = $1;
