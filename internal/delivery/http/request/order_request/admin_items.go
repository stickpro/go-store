package order_request

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// AdminOrderLineRequest is one line of the order's target line set.
type AdminOrderLineRequest struct {
	VariantID uuid.UUID `json:"variant_id" validate:"required"`
	Quantity  int64     `json:"quantity" validate:"required,gte=1,lte=100000"`
} //	@name	AdminOrderLineRequest

// AdminPreviewOrderItemsRequest computes an item edit without saving it.
type AdminPreviewOrderItemsRequest struct {
	Lines []AdminOrderLineRequest `json:"lines" validate:"required,min=1,max=200,dive"`
} //	@name	AdminPreviewOrderItemsRequest

// AdminUpdateOrderItemsRequest replaces the order's lines with Lines — the
// full target set: a line left out is removed. expected_version and
// expected_grand_total are the values the admin saw in the preview; if the
// order changed since, the edit is rejected instead of applied blind.
type AdminUpdateOrderItemsRequest struct {
	Lines              []AdminOrderLineRequest `json:"lines" validate:"required,min=1,max=200,dive"`
	ExpectedVersion    *int64                  `json:"expected_version" validate:"required"`
	ExpectedGrandTotal *decimal.Decimal        `json:"expected_grand_total" validate:"required"`
	Comment            *string                 `json:"comment" validate:"omitempty,max=2000"`
} //	@name	AdminUpdateOrderItemsRequest
