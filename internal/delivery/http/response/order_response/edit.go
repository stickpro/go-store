package order_response

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/stickpro/go-store/internal/dto"
)

// AdminOrderItemsEditResponse is the order after an item edit, or as it would
// be for a preview. refund_due is money the customer paid above the new total:
// return it with POST /v1/admin/orders/{number}/refund. It is 0 for an unpaid
// order. Send order.version and order.grand_total back as expected_version /
// expected_grand_total to apply a previewed edit.
type AdminOrderItemsEditResponse struct {
	Order              *AdminOrderResponse `json:"order"`
	PreviousGrandTotal decimal.Decimal     `json:"previous_grand_total"`
	RefundDue          decimal.Decimal     `json:"refund_due"`
	Changed            bool                `json:"changed"`
} //	@name	AdminOrderItemsEditResponse

func NewItemsEditFromDTO(d *dto.OrderItemsEditResultDTO) *AdminOrderItemsEditResponse {
	return &AdminOrderItemsEditResponse{
		Order:              NewAdminFromDTO(d.Order),
		PreviousGrandTotal: d.PreviousGrandTotal,
		RefundDue:          d.RefundDue,
		Changed:            d.Changed,
	}
}

// OrderEditLineResponse is one line as it was before or after an edit.
type OrderEditLineResponse struct {
	VariantID *uuid.UUID      `json:"variant_id"`
	Sku       *string         `json:"sku"`
	Name      string          `json:"name"`
	UnitPrice decimal.Decimal `json:"unit_price"`
	Quantity  int64           `json:"quantity"`
	LineTotal decimal.Decimal `json:"line_total"`
} //	@name	OrderEditLineResponse

// OrderEditResponse is one entry of an order's item-edit audit trail.
type OrderEditResponse struct {
	ID               uuid.UUID               `json:"id"`
	Actor            string                  `json:"actor"`
	Comment          *string                 `json:"comment"`
	LinesBefore      []OrderEditLineResponse `json:"lines_before"`
	LinesAfter       []OrderEditLineResponse `json:"lines_after"`
	GrandTotalBefore decimal.Decimal         `json:"grand_total_before"`
	GrandTotalAfter  decimal.Decimal         `json:"grand_total_after"`
	CreatedAt        time.Time               `json:"created_at"`
} //	@name	OrderEditResponse

func NewEditList(edits []*dto.OrderEditDTO) []*OrderEditResponse {
	out := make([]*OrderEditResponse, 0, len(edits))
	for _, e := range edits {
		out = append(out, &OrderEditResponse{
			ID:               e.ID,
			Actor:            e.Actor,
			Comment:          e.Comment,
			LinesBefore:      editLines(e.LinesBefore),
			LinesAfter:       editLines(e.LinesAfter),
			GrandTotalBefore: e.GrandTotalBefore,
			GrandTotalAfter:  e.GrandTotalAfter,
			CreatedAt:        e.CreatedAt,
		})
	}
	return out
}

func editLines(lines []dto.OrderEditLineDTO) []OrderEditLineResponse {
	out := make([]OrderEditLineResponse, 0, len(lines))
	for _, l := range lines {
		out = append(out, OrderEditLineResponse(l))
	}
	return out
}
