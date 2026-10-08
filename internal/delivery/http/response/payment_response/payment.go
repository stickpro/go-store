package payment_response

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/stickpro/go-store/internal/dto"
)

// PaymentResponse is one payment attempt, as returned after Init or on a
// status lookup. provider_payment_id is deliberately not exposed — it's an
// acquirer-internal reference with no use on the storefront.
type PaymentResponse struct {
	ID             uuid.UUID       `json:"id"`
	OrderID        uuid.UUID       `json:"order_id"`
	Provider       string          `json:"provider"`
	Status         string          `json:"status"`
	Amount         decimal.Decimal `json:"amount"`
	RefundedAmount decimal.Decimal `json:"refunded_amount"`
	Currency       string          `json:"currency"`
	PaymentURL     *string         `json:"payment_url"`
	CreatedAt      time.Time       `json:"created_at"`
} //	@name	PaymentResponse

func NewFromDTO(p *dto.PaymentDTO) *PaymentResponse {
	return &PaymentResponse{
		ID:             p.ID,
		OrderID:        p.OrderID,
		Provider:       p.Provider,
		Status:         p.Status,
		Amount:         p.Amount,
		RefundedAmount: p.RefundedAmount,
		Currency:       p.Currency,
		PaymentURL:     p.PaymentURL,
		CreatedAt:      p.CreatedAt,
	}
}
