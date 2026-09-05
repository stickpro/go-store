package order

import (
	"time"

	"github.com/shopspring/decimal"
)

// strPtr returns nil for the empty string, otherwise a pointer to a copy.
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// strOrNil is strPtr for values already known to be user input.
func strOrNil(s string) *string { return strPtr(s) }

// derefOrNow returns *t, or the current time if t is nil.
func derefOrNow(t *time.Time) time.Time {
	if t == nil {
		return time.Now()
	}
	return *t
}

func sumSubtotal(lines []checkoutLine) decimal.Decimal {
	sum := decimal.Zero
	for _, l := range lines {
		sum = sum.Add(l.lineTotal())
	}
	return sum
}
