package cart_request

import "github.com/google/uuid"

type AddCartItemRequest struct {
	ProductID uuid.UUID `json:"product_id" validate:"required"`
	VariantID uuid.UUID `json:"variant_id" validate:"required"`
	Quantity  int64     `json:"quantity" validate:"required,min=1"`
} //	@name	AddCartItemRequest
