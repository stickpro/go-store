package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/stickpro/go-store/internal/delivery/http/request/order_request"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
)

// OrderLineEditDTO is one line of the target line set of an admin item edit.
type OrderLineEditDTO struct {
	VariantID uuid.UUID
	Quantity  int64
}

// OrderItemsEditDTO replaces an order's lines with Lines (the full target set:
// a line left out is removed). Kept lines keep the price they were sold at;
// added lines are priced from the catalogue. With DryRun the edit is computed
// and validated but not saved. ExpectedVersion / ExpectedGrandTotal, when set,
// must match or the edit is rejected as stale.
type OrderItemsEditDTO struct {
	Lines              []OrderLineEditDTO
	ExpectedVersion    *int64
	ExpectedGrandTotal *decimal.Decimal
	Comment            *string
	Actor              string
	DryRun             bool
}

// OrderItemsEditResultDTO is the order after an item edit (or as it would be,
// for a dry run). RefundDue is money the customer paid above the new total —
// to be returned through a refund; it is zero for an unpaid order.
type OrderItemsEditResultDTO struct {
	Order              *OrderDTO
	PreviousGrandTotal decimal.Decimal
	RefundDue          decimal.Decimal
	Changed            bool
}

// OrderEditLineDTO is one line as recorded in the edit audit trail.
type OrderEditLineDTO struct {
	VariantID *uuid.UUID
	Sku       *string
	Name      string
	UnitPrice decimal.Decimal
	Quantity  int64
	LineTotal decimal.Decimal
}

// OrderEditDTO is one entry of an order's item-edit audit trail.
type OrderEditDTO struct {
	ID               uuid.UUID
	Actor            string
	Comment          *string
	LinesBefore      []OrderEditLineDTO
	LinesAfter       []OrderEditLineDTO
	GrandTotalBefore decimal.Decimal
	GrandTotalAfter  decimal.Decimal
	CreatedAt        time.Time
}

// OrderEditDTOFromModel maps an audit row; the line snapshots are decoded by
// the caller (they're stored as JSON).
func OrderEditDTOFromModel(e *models.OrderEdit, before, after []OrderEditLineDTO) *OrderEditDTO {
	return &OrderEditDTO{
		ID:               e.ID,
		Actor:            e.Actor,
		Comment:          pgtypeutils.DecodeText(e.Comment),
		LinesBefore:      before,
		LinesAfter:       after,
		GrandTotalBefore: e.GrandTotalBefore,
		GrandTotalAfter:  e.GrandTotalAfter,
		CreatedAt:        e.CreatedAt.Time,
	}
}

// RequestToOrderItemsEditDTO maps the admin item-edit request into the DTO.
func RequestToOrderItemsEditDTO(req *order_request.AdminUpdateOrderItemsRequest, actor string) OrderItemsEditDTO {
	return OrderItemsEditDTO{
		Lines:              orderLinesFromRequest(req.Lines),
		ExpectedVersion:    req.ExpectedVersion,
		ExpectedGrandTotal: req.ExpectedGrandTotal,
		Comment:            req.Comment,
		Actor:              actor,
	}
}

// RequestToOrderItemsPreviewDTO maps the admin item-edit preview request.
func RequestToOrderItemsPreviewDTO(req *order_request.AdminPreviewOrderItemsRequest, actor string) OrderItemsEditDTO {
	return OrderItemsEditDTO{
		Lines:  orderLinesFromRequest(req.Lines),
		Actor:  actor,
		DryRun: true,
	}
}

func orderLinesFromRequest(lines []order_request.AdminOrderLineRequest) []OrderLineEditDTO {
	out := make([]OrderLineEditDTO, 0, len(lines))
	for _, l := range lines {
		out = append(out, OrderLineEditDTO{VariantID: l.VariantID, Quantity: l.Quantity})
	}
	return out
}
