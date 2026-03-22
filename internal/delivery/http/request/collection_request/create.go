package collection_request

import "github.com/google/uuid"

type CreateCollectionRequest struct {
	Name        string      `json:"name" validate:"required"`
	Description *string     `json:"description,omitempty" validate:"omitempty,max=500"`
	Slug        string      `json:"slug" validate:"required,slug"`
	VariantIDs  []uuid.UUID `json:"variant_ids,omitempty" validate:"omitempty,dive,uuid"` //nolint:tagliatelle
} //	@name	CreateCollectionRequest

type UpdateCollectionRequest struct {
	Name        string      `json:"name" validate:"required"`
	Description *string     `json:"description,omitempty" validate:"omitempty,max=500"`
	Slug        string      `json:"slug" validate:"required,slug"`
	VariantIDs  []uuid.UUID `json:"variant_ids,omitempty" validate:"omitempty,dive,uuid"` //nolint:tagliatelle
} //	@name	UpdateCollectionRequest
