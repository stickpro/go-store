-- name: AddRelatedProducts :exec
INSERT INTO related_products (variant_id, related_variant_id)
SELECT sqlc.arg(variant_id)::uuid, pv.id
FROM product_variants pv
WHERE pv.id = any(sqlc.arg(related_variant_ids)::uuid[])
  AND pv.id != sqlc.arg(variant_id)::uuid
  AND NOT EXISTS (
    SELECT 1
    FROM related_products rp
    WHERE rp.variant_id = sqlc.arg(variant_id)
      AND rp.related_variant_id = pv.id
    );

-- name: DeleteRelatedProducts :exec
DELETE FROM related_products
WHERE variant_id = $1;

-- name: GetRelatedProductsByVariantID :many
SELECT pv.id,
       pv.name,
       pv.slug,
       pv.image,
       pv.is_enable,
       p.model,
       p.price,
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
       pv.image,
       pv.is_enable,
       p.model,
       p.price,
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
  AND related_variant_id = any($2::uuid[]);
