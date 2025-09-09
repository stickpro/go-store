package product_request

import "github.com/google/uuid"

type SyncProductAttributeRequest struct {
	AttributeIDs []uuid.UUID `json:"attribute_ids" validate:"required"` //nolint:tagliatelle
}
