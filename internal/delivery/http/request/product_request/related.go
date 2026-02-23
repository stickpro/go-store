package product_request

import "github.com/google/uuid"

type SyncRelatedProductRequest struct {
	VariantIDs []uuid.UUID `json:"variant_ids"`
} //	@name	SyncRelatedProductRequest
