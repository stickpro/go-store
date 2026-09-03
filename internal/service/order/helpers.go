package order

import "github.com/shopspring/decimal"

// strPtr returns nil for the empty string, otherwise a pointer to a copy.
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// strOrNil is strPtr for values already known to be user input.
func strOrNil(s string) *string { return strPtr(s) }

func sumSubtotal(lines []checkoutLine) decimal.Decimal {
	sum := decimal.Zero
	for _, l := range lines {
		sum = sum.Add(l.lineTotal())
	}
	return sum
}
