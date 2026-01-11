package attribute_request

import "github.com/google/uuid"

type UpdateAttributeGroupRequest struct {
	Name        string  `json:"name" validate:"required"`
	Slug        string  `json:"slug" validate:"required"`
	Description *string `json:"description" validate:"omitempty,min=1,max=100"`
} // @name UpdateAttributeGroupRequest

type UpdateAttributeRequest struct {
	Name             string     `json:"name" validate:"required,min=1,max=255"`
	Slug             string     `json:"slug" validate:"required,min=1,max=255"`
	AttributeGroupID *uuid.UUID `json:"attribute_group_id" validate:"omitempty,uuid"`
	Type             string     `json:"type" validate:"required,oneof=select number boolean text"`
	Unit             *string    `json:"unit" validate:"omitempty,min=1,max=50"`
	IsFilterable     bool       `json:"is_filterable"`
	IsVisible        bool       `json:"is_visible"`
	IsRequired       bool       `json:"is_required"`
	SortOrder        *int32     `json:"sort_order" validate:"omitempty,min=0"`
} // @name UpdateAttributeRequest
