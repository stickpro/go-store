package attribute_request

import "github.com/google/uuid"

type CreateAttributeValueRequest struct {
	AttributeID     uuid.UUID `json:"attribute_id" validate:"required,uuid"`
	Value           string    `json:"value" validate:"required,min=1,max=255"`
	ValueNormalized *string   `json:"value_normalized" validate:"omitempty,max=255"`
	ValueNumeric    *float64  `json:"value_numeric" validate:"omitempty"`
	DisplayOrder    int32     `json:"display_order" validate:"omitempty,min=0"`
} // @name CreateAttributeValueRequest

type UpdateAttributeValueRequest struct {
	Value           string   `json:"value" validate:"required,min=1,max=255"`
	ValueNormalized *string  `json:"value_normalized" validate:"omitempty,max=255"`
	ValueNumeric    *float64 `json:"value_numeric" validate:"omitempty"`
	DisplayOrder    int32    `json:"display_order" validate:"omitempty,min=0"`
	IsActive        bool     `json:"is_active"`
} // @name UpdateAttributeValueRequest
