package payment

import "errors"

var (
	// ErrProviderNotFound — no provider is registered under the requested
	// code, or it exists but is disabled in config.
	ErrProviderNotFound = errors.New("payment: provider not found")
	// ErrNotFound — no payment row matches the lookup.
	ErrNotFound = errors.New("payment: not found")
	// ErrInvalidSignature — a webhook's signature/token did not verify; the
	// payload must be discarded, never acted on.
	ErrInvalidSignature = errors.New("payment: invalid notification signature")
	// ErrNotConfirmed — a refund was requested for an order with no captured
	// payment, or one already refunded in full: there is nothing to refund.
	ErrNotConfirmed = errors.New("payment: not confirmed")
	// ErrAmountMismatch — a provider confirmed a payment for a different
	// amount than was requested. The payment is left unconfirmed for manual
	// review instead of marking the order paid.
	ErrAmountMismatch = errors.New("payment: confirmed amount does not match")
	// ErrProviderRejected — the provider answered and explicitly declined the
	// operation (as opposed to a timeout/network error with unknown outcome).
	// Providers wrap their gateway errors with it.
	ErrProviderRejected = errors.New("payment: rejected by provider")

	// ErrNothingToRefund — the customer has paid no more than the order's
	// current total (or, for a full refund, everything is already returned).
	ErrNothingToRefund = errors.New("payment: nothing owed to the customer")
	// ErrRefundPending — the refund was sent to the provider but its outcome
	// is not known yet (timeout/network error), or another refund is still
	// in flight for the whole remaining balance. It must not be retried under
	// a new idempotency key until resolved.
	ErrRefundPending = errors.New("payment: refund outcome pending")
	// ErrRefundRejected — the provider declined the refund; nothing was
	// returned to the customer.
	ErrRefundRejected = errors.New("payment: refund rejected")
	// ErrIdempotencyConflict — the idempotency key was already used for a
	// refund of another order.
	ErrIdempotencyConflict = errors.New("payment: idempotency key reused for a different refund")
)
