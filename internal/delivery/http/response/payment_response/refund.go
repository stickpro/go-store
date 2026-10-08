package payment_response

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/stickpro/go-store/internal/dto"
)

// RefundResponse is one refund attempt against an order's payment. status is
// pending (sent to the acquirer, outcome not known yet), succeeded or failed.
type RefundResponse struct {
	ID        uuid.UUID       `json:"id"`
	PaymentID uuid.UUID       `json:"payment_id"`
	Amount    decimal.Decimal `json:"amount"`
	Status    string          `json:"status"`
	Reason    *string         `json:"reason"`
	Actor     string          `json:"actor"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt *time.Time      `json:"updated_at"`
} //	@name	RefundResponse

func NewRefundFromDTO(r *dto.PaymentRefundDTO) *RefundResponse {
	return &RefundResponse{
		ID:        r.ID,
		PaymentID: r.PaymentID,
		Amount:    r.Amount,
		Status:    r.Status,
		Reason:    r.Reason,
		Actor:     r.Actor,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

func NewRefundList(refunds []*dto.PaymentRefundDTO) []*RefundResponse {
	out := make([]*RefundResponse, 0, len(refunds))
	for _, r := range refunds {
		out = append(out, NewRefundFromDTO(r))
	}
	return out
}
