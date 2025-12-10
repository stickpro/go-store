-- name: AddRelatedProducts :exec
INSERT INTO related_products (product_id, related_product_id)
SELECT sqlc.arg(product_id)::uuid, p.id
FROM products p
WHERE p.id = any(sqlc.arg(related_product_ids)::uuid[])
  AND p.id != sqlc.arg(product_id)::uuid
  AND NOT EXISTS (
    SELECT 1
    FROM related_products rp
    WHERE rp.product_id = sqlc.arg(product_id)
      AND rp.related_product_id = p.id
    );

-- name: DeleteRelatedProducts :exec
DELETE FROM related_products
WHERE product_id = $1;

-- name: GetRelatedProductsByProductID :many
SELECT p.id,
       p.name,
       p.slug,
       p.model,
       p.price,
       p.image,
       p.is_enable,
       p.stock_status
FROM related_products rp
         JOIN products p ON rp.related_product_id = p.id
WHERE rp.product_id = $1
  AND p.is_enable = true
ORDER BY p.name;

-- name: GetRelatedProductsBySlug :many
SELECT p.id,
       p.name,
       p.slug,
       p.model,
       p.price,
       p.image,
       p.is_enable,
       p.stock_status
FROM related_products rp
         JOIN products p ON rp.related_product_id = p.id
WHERE rp.product_id = (SELECT id FROM p WHERE slug = $1)
  AND p.is_enable = true
ORDER BY p.name;

-- name: DeleteSpecificRelatedProducts :exec
DELETE FROM related_products
WHERE product_id = $1
  AND related_product_id = any($2::uuid[]);
