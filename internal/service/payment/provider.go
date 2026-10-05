// Package payment is the provider-agnostic acquiring layer: one Provider per
// gateway (T-Bank today; ЮKassa/CloudPayments etc. later), selected at
// runtime through Registry. It mirrors internal/service/shipping's
// Provider+Registry split — persistence (the payments table) and
// orchestration live in Service, while a Provider only talks to its gateway.
package payment

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// InitRequest starts a payment attempt at the provider. PaymentID is our own
// payments.id, echoed back by the provider in the notification/order
// reference — it is how HandleNotification finds the row to update without
// waiting on a provider payment id that may not exist yet.
type InitRequest struct {
	PaymentID     uuid.UUID
	OrderNumber   int64
	Amount        decimal.Decimal
	Currency      string
	Description   string
	CustomerEmail string
}

// InitResult is what the provider handed back for a successful Init call.
type InitResult struct {
	ProviderPaymentID string
	PaymentURL        string
	Status            Status
	RawResponse       []byte
}

// NotificationResult is a provider webhook, verified and parsed into the
// provider-agnostic shape. PaymentID is recovered from whatever field the
// provider echoes the merchant reference back in (for T-Bank, OrderId).
type NotificationResult struct {
	PaymentID         uuid.UUID
	ProviderPaymentID string
	Status            Status
	RawNotification   []byte
	// AckBody is the exact response body the provider expects for a
	// successful delivery (T-Bank wants the literal string "OK"); the HTTP
	// handler writes it back verbatim so the provider stops retrying.
	AckBody string
}

// CancelRequest asks the provider to cancel/refund a confirmed payment.
// Amount is the amount to refund; a zero value means "refund in full".
type CancelRequest struct {
	PaymentID         uuid.UUID
	ProviderPaymentID string
	Amount            decimal.Decimal
}

// CancelResult is what the provider handed back for a Cancel call.
type CancelResult struct {
	Status      Status
	RawResponse []byte
}

// Provider is one acquiring integration. Implementations must be stateless
// with respect to the payments table — all persistence is Service's job, so
// a Provider can be unit-tested against the gateway's API alone.
type Provider interface {
	// Code identifies the provider, e.g. "tbank". Stored in payments.provider
	// and orders.payment_method, and used as the :provider path segment of
	// its webhook route.
	Code() string
	// Enabled reports whether the integration is turned on in config.
	Enabled() bool
	Init(ctx context.Context, r InitRequest) (InitResult, error)
	// HandleNotification verifies and parses a raw webhook body. It must not
	// touch the database — Service reconciles the result against the
	// payments table.
	HandleNotification(ctx context.Context, raw []byte) (NotificationResult, error)
	Cancel(ctx context.Context, r CancelRequest) (CancelResult, error)
}
