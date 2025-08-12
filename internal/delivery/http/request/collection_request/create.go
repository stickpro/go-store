package collection_request

import "github.com/google/uuid"

type CreateCollectionRequest struct {
	Name        string      `json:"name" validate:"required"`
	Description *string     `json:"description,omitempty" validate:"omitempty,max=500"`
	Slug        string      `json:"slug" validate:"required,slug"`
	ProductIDs  []uuid.UUID `json:"product_ids,omitempty" validate:"omitempty,dive,uuid"`
} // @name CreateCollectionRequest

type UpdateCollectionRequest struct {
	Name        string      `json:"name" validate:"required"`
	Description *string     `json:"description,omitempty" validate:"omitempty,max=500"`
	Slug        string      `json:"slug" validate:"required,slug"`
	ProductIDs  []uuid.UUID `json:"product_ids,omitempty" validate:"omitempty,dive,uuid"`
} // @name UpdateCollectionRequest
