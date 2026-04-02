-- name: GetBySlug :one
SELECT *
FROM collections
WHERE slug = $1 LIMIT 1;

-- name: DeleteProductsFromCollection :exec
DELETE
FROM collection_variants
WHERE collection_id = $1;

-- name: AddProductsToCollection :exec
INSERT INTO collection_variants (collection_id, variant_id)
SELECT sqlc.arg(collection_id)::uuid, v.id
FROM product_variants v
WHERE v.id = any (sqlc.arg(variant_ids)::uuid[])
  AND NOT EXISTS (SELECT 1
                  FROM collection_variants cv
                  WHERE cv.collection_id = sqlc.arg(collection_id)
                    AND cv.variant_id = v.id);

-- name: GetCollectionWithProductsBySlug :many
SELECT c.*,
       pv.id                 AS variant_id,
       pv.product_id         AS product_id,
       COALESCE(pv.name, '') AS product_name,
       COALESCE(pv.slug, '') AS product_slug,
       pv.model              AS product_model,
       p.price_retail        AS product_price_retail,
       p.price_business      AS product_price_business,
       p.price_wholesale     AS product_price_wholesale,
       p.is_enable           AS product_is_enable,
       pv.image              AS product_image
FROM collections c
         LEFT JOIN collection_variants cv ON cv.collection_id = c.id
         LEFT JOIN product_variants pv ON pv.id = cv.variant_id
         LEFT JOIN products p ON p.id = pv.product_id
WHERE c.slug = $1;

-- name: GetCollectionWithProductsByID :many
SELECT c.*,
       pv.id                 AS variant_id,
       pv.product_id         AS product_id,
       COALESCE(pv.name, '') AS product_name,
       COALESCE(pv.slug, '') AS product_slug,
       pv.model              AS product_model,
       p.price_retail        AS product_price_retail,
       p.price_business      AS product_price_business,
       p.price_wholesale     AS product_price_wholesale,
       p.is_enable           AS product_is_enable,
       pv.image              AS product_image
FROM collections c
         LEFT JOIN collection_variants cv ON cv.collection_id = c.id
         LEFT JOIN product_variants pv ON pv.id = cv.variant_id
         LEFT JOIN products p ON p.id = pv.product_id
WHERE c.id = $1;
