-- name: GetMediaByProductID :many
SELECT m.*
FROM product_media pm
         JOIN media m ON pm.media_id = m.id
WHERE pm.product_id = $1
ORDER BY pm.sort_order;

-- name: CreateProductMedia :exec
INSERT INTO product_media (product_id, media_id, sort_order)
VALUES ($1, $2, $3) ON CONFLICT DO NOTHING;