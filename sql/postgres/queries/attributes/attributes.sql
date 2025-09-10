-- name: DeleteByAttributeGroupID :exec
DELETE FROM attributes WHERE attribute_group_id = $1::uuid;

-- name: GetByID :one
SELECT * FROM attributes WHERE id = $1 LIMIT 1;

-- name: GetAttributesProduct :many
select
    ag.id as group_id,
    ag.name as group_name,
    ag.description as group_description,
    json_agg(
            json_build_object(
                    'id', a.id,
                    'name', a.name,
                    'type', a.type,
                    'is_filterable', a.is_filterable,
                    'is_visible', a.is_visible,
                    'sort_order', a.sort_order,
                    'value', a.value
            ) order by a.sort_order
    ) as attributes
from attribute_products ap
         join attributes a on ap.attribute_id = a.id
         join attribute_groups ag on a.attribute_group_id = ag.id
where ap.product_id = $1::uuid
group by ag.id, ag.name, ag.description
order by ag.name;
