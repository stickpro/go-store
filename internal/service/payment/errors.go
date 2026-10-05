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
	// ErrNotConfirmed — Cancel was called on a payment that never reached
	// StatusConfirmed, so there is nothing to refund.
	ErrNotConfirmed = errors.New("payment: not confirmed")
)
