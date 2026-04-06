-- name: Add :exec
INSERT INTO product_variant_categories (
    product_variant_id,
    category_id,
    created_at
) VALUES (
    $1, $2, now()
)
ON CONFLICT (product_variant_id, category_id) DO NOTHING;

-- name: AddBatch :exec
INSERT INTO product_variant_categories (product_variant_id, category_id, created_at)
SELECT sqlc.arg(product_variant_id), unnest(sqlc.arg(category_ids)::uuid[]), now()
ON CONFLICT (product_variant_id, category_id) DO NOTHING;

-- name: Remove :exec
DELETE FROM product_variant_categories
WHERE product_variant_id = $1 AND category_id = $2;

-- name: RemoveAll :exec
DELETE FROM product_variant_categories
WHERE product_variant_id = $1;

-- name: GetByVariantID :many
SELECT
    pvc.product_variant_id,
    pvc.category_id,
    pvc.created_at,
    c.name as category_name,
    c.slug as category_slug,
    c.is_enable as category_is_enable
FROM product_variant_categories pvc
INNER JOIN categories c ON pvc.category_id = c.id
WHERE pvc.product_variant_id = $1
ORDER BY c.name ASC;

-- name: GetByVariantIDs :many
SELECT
    pvc.product_variant_id,
    pvc.category_id,
    pvc.created_at,
    c.name as category_name,
    c.slug as category_slug,
    c.is_enable as category_is_enable
FROM product_variant_categories pvc
INNER JOIN categories c ON pvc.category_id = c.id
WHERE pvc.product_variant_id = ANY(sqlc.arg(variant_ids)::uuid[])
ORDER BY c.name ASC;