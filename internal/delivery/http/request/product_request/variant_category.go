package product_request

import "github.com/google/uuid"

type SyncVariantCategoriesRequest struct {
	CategoryIDs []uuid.UUID `json:"category_ids"` //nolint:tagliatelle
} //	@name	SyncVariantCategoriesRequest
