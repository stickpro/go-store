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
		Currency:          p.Currency,
		PaymentURL:        pgtypeutils.DecodeText(p.PaymentUrl),
		CreatedAt:         p.CreatedAt.Time,
		UpdatedAt:         timestampPtr(p.UpdatedAt),
	}
}
