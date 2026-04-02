package viewed_request

import "github.com/google/uuid"

type TrackViewedRequest struct {
	VariantID uuid.UUID `json:"variant_id" validate:"required"`
} //	@name	TrackViewedRequest
