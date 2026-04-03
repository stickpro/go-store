-- name: SyncRelatedProducts :exec
WITH deleted AS (
    DELETE FROM related_products
    WHERE variant_id = sqlc.arg(variant_id)::uuid
)
INSERT INTO related_products (variant_id, related_variant_id)
SELECT sqlc.arg(variant_id)::uuid, v.id
FROM product_variants v
WHERE v.id = ANY(sqlc.arg(related_variant_ids)::uuid[])
  AND v.id != sqlc.arg(variant_id)::uuid
ON CONFLICT DO NOTHING;

-- name: GetRelatedProductsByVariantIDs :many
SELECT rp.variant_id,
       pv.id,
       pv.name,
       pv.slug,
       p.image,
       pv.is_enable,
       pv.model,
       p.price_retail,
       p.price_business,
       p.price_wholesale,
       p.stock_status
FROM related_products rp
         JOIN product_variants pv ON rp.related_variant_id = pv.id
         JOIN products p ON pv.product_id = p.id
WHERE rp.variant_id = ANY($1::uuid[])
  AND pv.is_enable = true
ORDER BY rp.variant_id, pv.name;

-- name: GetRelatedProductsByVariantID :many
SELECT pv.id,
       pv.name,
       pv.slug,
       p.image,
       pv.is_enable,
       pv.model,
       p.price_retail,
       p.price_business,
       p.price_wholesale,
       p.stock_status
FROM related_products rp
         JOIN product_variants pv ON rp.related_variant_id = pv.id
         JOIN products p ON pv.product_id = p.id
WHERE rp.variant_id = $1
  AND pv.is_enable = true
ORDER BY pv.name;

-- name: GetRelatedProductsBySlug :many
SELECT pv.id,
       pv.name,
       pv.slug,
       p.image,
       pv.is_enable,
       pv.model,
       p.price_retail,
       p.price_business,
       p.price_wholesale,
       p.stock_status
FROM related_products rp
         JOIN product_variants pv ON rp.related_variant_id = pv.id
         JOIN products p ON pv.product_id = p.id
WHERE rp.variant_id = (SELECT id FROM product_variants pv2 WHERE pv2.slug = $1)
  AND pv.is_enable = true
ORDER BY pv.name;

-- name: DeleteSpecificRelatedProducts :exec
DELETE FROM related_products
WHERE variant_id = $1
  AND related_variant_id = ANY($2::uuid[]);
