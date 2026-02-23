package contracts

import "github.com/google/uuid"

type VariantPayload struct {
	ProductExternalID string     `json:"product_external_id"`
	Name              string     `json:"name"`
	Slug              string     `json:"slug"`
	CategoryID        *uuid.UUID `json:"category_id,omitempty"`
	IsEnable          bool       `json:"is_enable"`
}
