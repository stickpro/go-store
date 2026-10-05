package payment

// Status is the provider-agnostic state of one payment attempt, stored in
// payments.status. Providers map their own gateway statuses onto this set.
type Status string //	@name	PaymentProviderStatus

const (
	// StatusNew is the row created right before the provider Init call.
	StatusNew Status = "new"
	// StatusPending is an Init that succeeded: the customer has a PaymentURL
	// and hasn't finished paying yet.
	StatusPending Status = "pending"
	// StatusConfirmed is a completed, captured payment.
	StatusConfirmed Status = "confirmed"
	// StatusFailed is a terminal failure (rejected, expired, Init error).
	StatusFailed Status = "failed"
	// StatusRefunded is a confirmed payment that was fully or partially
	// cancelled/refunded after capture.
	StatusRefunded Status = "refunded"
)

func (s Status) String() string { return string(s) }

// IsTerminal reports whether a payment in this status is done changing on
// its own — further notifications for it are retries/duplicates, not new
// information, and should be acknowledged without reprocessing.
func (s Status) IsTerminal() bool {
	switch s {
	case StatusConfirmed, StatusFailed, StatusRefunded:
		return true
	default:
		return false
	}
}
