-- name: DeleteByAttributeGroupID :exec
DELETE FROM attributes WHERE attribute_group_id = $1::uuid;


