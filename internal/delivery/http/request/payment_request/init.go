package payment_request

// InitPaymentRequest starts a payment attempt for an order.
type InitPaymentRequest struct {
	// Provider selects the acquirer, e.g. "tbank". More values are added as
	// more providers are wired up.
	Provider string `json:"provider" validate:"required,oneof=tbank"`
} //	@name	InitPaymentRequest
