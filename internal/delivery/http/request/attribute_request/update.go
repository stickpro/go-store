package attribute_request

import "github.com/google/uuid"

type UpdateAttributeGroupRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description *string `json:"description" validate:"omitempty,min=1,max=100"`
}

type UpdateAttributeRequest struct {
	Name             string     `json:"name" validate:"required"`
	Value            string     `json:"value" validate:"required,min=1"`
	AttributeGroupID *uuid.UUID `json:"attribute_group_id" validate:"omitempty,uuid"`
	Type             string     `json:"type" validate:"required,oneof=string"`
	IsFilterable     bool       `json:"is_filterable" validate:"required"`
	IsVisible        bool       `json:"is_visible" validate:"required"`
	SortOrder        *int32     `json:"sort_order" validate:"omitempty,min=0"`
}
