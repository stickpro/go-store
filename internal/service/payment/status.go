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
	// StatusPartiallyRefunded is a confirmed payment with part of its amount
	// returned; the rest can still be refunded.
	StatusPartiallyRefunded Status = "partially_refunded"
	// StatusRefunded is a confirmed payment whose whole amount was
	// cancelled/refunded after capture.
	StatusRefunded Status = "refunded"
)

func (s Status) String() string { return string(s) }

// IsTerminal reports whether a payment in this status is done changing —
// further notifications for it are retries/duplicates, not new information,
// and should be acknowledged without reprocessing. A confirmed or partially
// refunded payment is not terminal: it can still be refunded.
func (s Status) IsTerminal() bool {
	switch s {
	case StatusFailed, StatusRefunded:
		return true
	default:
		return false
	}
}

// IsRefundable reports whether a payment in this status still has captured
// money that can be returned.
func (s Status) IsRefundable() bool {
	return s == StatusConfirmed || s == StatusPartiallyRefunded
}

// rank orders statuses along the payment lifecycle so a late or out-of-order
// webhook can't move a payment backwards (e.g. a delayed CONFIRMED arriving
// after a refund). Failed and refunded are both end states.
func (s Status) rank() int {
	switch s {
	case StatusNew:
		return 0
	case StatusPending:
		return 1
	case StatusConfirmed:
		return 2
	case StatusPartiallyRefunded:
		return 3
	case StatusFailed, StatusRefunded:
		return 4
	default:
		return -1
	}
}

// canAdvance reports whether a payment may move from -> to. Statuses only move
// forward; the one exception is a captured payment never "failing" — once
// money is taken, only a refund can undo it.
func canAdvance(from, to Status) bool {
	if from.IsTerminal() || to.rank() <= from.rank() {
		return false
	}
	if to == StatusFailed && from.rank() >= StatusConfirmed.rank() {
		return false
	}
	return true
}

// refundStatus is the state of one row in payment_refunds.
type refundStatus string

const (
	// refundPending is written before the provider is called; the amount
	// counts against the payment's balance until the row is resolved.
	refundPending   refundStatus = "pending"
	refundSucceeded refundStatus = "succeeded"
	refundFailed    refundStatus = "failed"
)
