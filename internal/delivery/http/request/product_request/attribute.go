package product_request

import "github.com/google/uuid"

type SyncProductAttributeRequest struct {
	AttributeValueIDs []uuid.UUID `json:"attribute_value_ids" validate:"required"` //nolint:tagliatelle
}
