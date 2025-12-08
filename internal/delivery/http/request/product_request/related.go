package product_request

import "github.com/google/uuid"

type SyncRelatedProductRequest struct {
	ProductIDs []uuid.UUID `json:"product_ids"`
} // @name SyncRelatedProductRequest
