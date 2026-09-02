-- name: GetMediaByProductID :many
SELECT m.*
FROM product_media pm
         JOIN media m ON pm.media_id = m.id
WHERE pm.product_id = $1
ORDER BY pm.sort_order;

-- name: GetMainMediaByProductIDs :many
SELECT DISTINCT ON (pm.product_id)
    pm.product_id,
    m.id,
    m.path,
    m.width,
    m.height
FROM product_media pm
         JOIN media m ON m.id = pm.media_id
WHERE pm.product_id = ANY($1::uuid[])
ORDER BY pm.product_id, pm.sort_order;

-- name: CreateProductMedia :exec
INSERT INTO product_media (product_id, media_id, sort_order)
VALUES ($1, $2, $3) ON CONFLICT DO NOTHING;

-- name: DeleteProductMedia :exec
DELETE FROM product_media WHERE product_id = $1;

-- name: DeleteProductMediaByMediaIDs :exec
DELETE FROM product_media WHERE product_id = $1 AND media_id = ANY($2::uuid[]);