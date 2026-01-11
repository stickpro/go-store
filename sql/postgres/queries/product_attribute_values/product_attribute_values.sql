-- name: Add :exec
INSERT INTO product_attribute_values (
    product_id,
    attribute_value_id,
    created_at
) VALUES (
    $1, $2, now()
)
ON CONFLICT (product_id, attribute_value_id) DO NOTHING;

-- name: AddBatch :exec
INSERT INTO product_attribute_values (product_id, attribute_value_id, created_at)
SELECT sqlc.arg(product_id), unnest(sqlc.arg(attribute_value_id)::uuid[]), now()
ON CONFLICT (product_id, attribute_value_id) DO NOTHING;

-- name: Remove :exec
DELETE FROM product_attribute_values
WHERE product_id = $1 AND attribute_value_id = $2;

-- name: RemoveAll :exec
DELETE FROM product_attribute_values
WHERE product_id = $1;

-- name: GetByProductID :many
SELECT
    pav.product_id,
    pav.attribute_value_id,
    pav.created_at,
    a.id as attribute_id,
    a.name as attribute_name,
    a.slug as attribute_slug,
    a.type as attribute_type,
    a.unit as attribute_unit,
    a.is_filterable,
    a.is_visible,
    a.sort_order as attribute_sort_order,
    av.value as attribute_value,
    av.value_normalized,
    av.value_numeric,
    av.display_order as value_display_order,
    ag.id as group_id,
    ag.name as group_name,
    ag.slug as group_slug
FROM product_attribute_values pav
INNER JOIN attribute_values av ON pav.attribute_value_id = av.id
INNER JOIN attributes a ON av.attribute_id = a.id
INNER JOIN attribute_groups ag ON a.attribute_group_id = ag.id
WHERE pav.product_id = $1 AND a.is_visible = true AND av.is_active = true
ORDER BY ag.name ASC, a.sort_order ASC, av.display_order ASC;

-- name: GetFiltersForCategory :many
SELECT
    a.id as attribute_id,
    a.slug as attribute_slug,
    a.name as attribute_name,
    a.type as attribute_type,
    a.unit as attribute_unit,
    a.sort_order as attribute_sort_order,
    ag.id as group_id,
    ag.name as group_name,
    ag.slug as group_slug,
    av.id as value_id,
    av.value as value,
    av.value_normalized,
    av.value_numeric,
    av.display_order as value_display_order,
    COUNT(DISTINCT p.id) as product_count
FROM products p
INNER JOIN product_attribute_values pav ON p.id = pav.product_id
INNER JOIN attribute_values av ON pav.attribute_value_id = av.id
INNER JOIN attributes a ON av.attribute_id = a.id
INNER JOIN attribute_groups ag ON a.attribute_group_id = ag.id
WHERE
    p.category_id = $1
    AND p.is_enable = true
    AND a.is_filterable = true
    AND av.is_active = true
GROUP BY
    a.id, a.slug, a.name, a.type, a.unit, a.sort_order,
    ag.id, ag.name, ag.slug,
    av.id, av.value, av.value_normalized, av.value_numeric, av.display_order
ORDER BY
    ag.name ASC,
    a.sort_order ASC,
    a.name ASC,
    av.display_order ASC,
    av.value ASC;
