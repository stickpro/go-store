package invoice_response

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/stickpro/go-store/internal/dto"
)

// InvoiceResponse is the active invoice of an order paid by invoice. The PDF
// is served by the matching .../invoice/pdf endpoint.
type InvoiceResponse struct {
	ID       uuid.UUID       `json:"id"`
	OrderID  uuid.UUID       `json:"order_id"`
	Number   string          `json:"number"`
	Status   string          `json:"status"`
	Amount   decimal.Decimal `json:"amount"`
	Currency string          `json:"currency"`
	IssuedAt time.Time       `json:"issued_at"`
	DueAt    time.Time       `json:"due_at"`
} //	@name	InvoiceResponse

func NewFromDTO(d *dto.InvoiceDTO) *InvoiceResponse {
	return &InvoiceResponse{
		ID:       d.ID,
		OrderID:  d.OrderID,
		Number:   d.Number,
		Status:   d.Status,
		Amount:   d.Amount,
		Currency: d.Currency,
		IssuedAt: d.IssuedAt,
		DueAt:    d.DueAt,
	}
}
