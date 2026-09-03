package contracts

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OrderEventItem struct {
	VariantID *uuid.UUID      `json:"variant_id,omitempty"`
	Sku       *string         `json:"sku,omitempty"`
	Name      string          `json:"name"`
	UnitPrice decimal.Decimal `json:"unit_price"`
	Quantity  int64           `json:"quantity"`
	LineTotal decimal.Decimal `json:"line_total"`
}

type OrderCreatedPayload struct {
	OrderNumber int64            `json:"order_number"`
	UserID      *uuid.UUID       `json:"user_id,omitempty"`
	Email       string           `json:"email"`
	Currency    string           `json:"currency"`
	GrandTotal  decimal.Decimal  `json:"grand_total"`
	Items       []OrderEventItem `json:"items"`
	CreatedAt   time.Time        `json:"created_at"`
}

type OrderPaidPayload struct {
	OrderNumber   int64     `json:"order_number"`
	PaymentMethod *string   `json:"payment_method,omitempty"`
	PaidAt        time.Time `json:"paid_at"`
}

type OrderCancelledPayload struct {
	OrderNumber int64     `json:"order_number"`
	Reason      string    `json:"reason"`
	CancelledAt time.Time `json:"cancelled_at"`
}
