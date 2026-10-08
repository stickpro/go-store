package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
)

// InitPaymentDTO starts a payment attempt for an order through one provider.
type InitPaymentDTO struct {
	OrderID       uuid.UUID
	OrderNumber   int64
	Provider      string
	Amount        decimal.Decimal
	Currency      string
	Description   string
	CustomerEmail string
}

// PaymentDTO is one payment attempt row, service<->delivery currency.
type PaymentDTO struct {
	ID                uuid.UUID
	OrderID           uuid.UUID
	Provider          string
	ProviderPaymentID *string
	Status            string
	Amount            decimal.Decimal
	RefundedAmount    decimal.Decimal
	Currency          string
	PaymentURL        *string
	CreatedAt         time.Time
	UpdatedAt         *time.Time
}

// PaymentDTOFromModel maps a persisted payment row into the domain DTO.
func PaymentDTOFromModel(p *models.Payment) *PaymentDTO {
	return &PaymentDTO{
		ID:                p.ID,
		OrderID:           p.OrderID,
		Provider:          p.Provider,
		ProviderPaymentID: pgtypeutils.DecodeText(p.ProviderPaymentID),
		Status:            p.Status,
		Amount:            p.Amount,
		RefundedAmount:    p.RefundedAmount,
		Currency:          p.Currency,
		PaymentURL:        pgtypeutils.DecodeText(p.PaymentUrl),
		CreatedAt:         p.CreatedAt.Time,
		UpdatedAt:         timestampPtr(p.UpdatedAt),
	}
}

// RefundPaymentDTO asks to refund an order's captured payment. The amount is
// derived, never given: the overpayment over the order's current total, or
// with Full everything left on the payment. IdempotencyKey is chosen by the
// caller and makes a retried request return the original refund instead of a
// new one.
type RefundPaymentDTO struct {
	OrderID        uuid.UUID
	Full           bool
	IdempotencyKey string
	Reason         *string
	Actor          string
}

// PaymentRefundDTO is one refund attempt against a payment.
type PaymentRefundDTO struct {
	ID        uuid.UUID
	PaymentID uuid.UUID
	Amount    decimal.Decimal
	Status    string
	Reason    *string
	Actor     string
	CreatedAt time.Time
	UpdatedAt *time.Time
}

// PaymentRefundDTOFromModel maps a persisted refund row into the domain DTO.
func PaymentRefundDTOFromModel(r *models.PaymentRefund) *PaymentRefundDTO {
	return &PaymentRefundDTO{
		ID:        r.ID,
		PaymentID: r.PaymentID,
		Amount:    r.Amount,
		Status:    r.Status,
		Reason:    pgtypeutils.DecodeText(r.Reason),
		Actor:     r.Actor,
		CreatedAt: r.CreatedAt.Time,
		UpdatedAt: timestampPtr(r.UpdatedAt),
	}
}

// RefundResultDTO is a settled refund together with the payment it changed.
type RefundResultDTO struct {
	Payment *PaymentDTO
	Refund  *PaymentRefundDTO
}
