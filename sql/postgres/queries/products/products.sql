-- name: GetByID :one
SELECT * FROM products WHERE id = $1 LIMIT 1;

-- name: GetByExternalID :one
SELECT * FROM products WHERE external_id = $1 LIMIT 1;

-- name: GetBySlug :one
SELECT p.* FROM products p
INNER JOIN product_variants pv ON pv.product_id = p.id
WHERE pv.slug = $1 LIMIT 1;

-- name: GetCartItemsByVariantIDs :many
SELECT p.id        AS product_id,
       p.price_retail,
       p.price_business,
       p.price_wholesale,
       p.quantity   AS max_quantity,
       p.is_enable  AS product_enabled,
       pv.id        AS variant_id,
       pv.name,
       pv.slug,
       pv.is_enable AS variant_enabled,
       img.id       AS image_id,
       img.path     AS image_path,
       img.width    AS image_width,
       img.height   AS image_height
FROM products p
         JOIN product_variants pv ON pv.product_id = p.id
         LEFT JOIN LATERAL (
             SELECT pm.media_id FROM product_media pm
             WHERE pm.product_id = p.id ORDER BY pm.sort_order LIMIT 1
         ) mm ON true
         LEFT JOIN media img ON img.id = mm.media_id
WHERE pv.id = ANY ($1::uuid[]);
