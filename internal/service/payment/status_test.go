package payment

import (
	"errors"
	"testing"

	"github.com/shopspring/decimal"
)

func TestCanAdvance(t *testing.T) {
	cases := []struct {
		from, to Status
		want     bool
	}{
		{StatusNew, StatusPending, true},
		{StatusPending, StatusConfirmed, true},
		{StatusPending, StatusFailed, true},
		{StatusConfirmed, StatusPartiallyRefunded, true},
		{StatusConfirmed, StatusRefunded, true},
		{StatusPartiallyRefunded, StatusRefunded, true},

		// Duplicates and late/out-of-order deliveries.
		{StatusPending, StatusPending, false},
		{StatusConfirmed, StatusConfirmed, false},
		{StatusConfirmed, StatusPending, false},
		{StatusPartiallyRefunded, StatusConfirmed, false},
		{StatusPartiallyRefunded, StatusPartiallyRefunded, false},
		{StatusRefunded, StatusPartiallyRefunded, false},

		// Captured money can only go back through a refund.
		{StatusConfirmed, StatusFailed, false},
		{StatusPartiallyRefunded, StatusFailed, false},

		// End states.
		{StatusFailed, StatusConfirmed, false},
		{StatusRefunded, StatusConfirmed, false},
	}

	for _, c := range cases {
		if got := canAdvance(c.from, c.to); got != c.want {
			t.Errorf("canAdvance(%s, %s) = %v, want %v", c.from, c.to, got, c.want)
		}
	}
}

func TestRefundAmount(t *testing.T) {
	d := decimal.RequireFromString

	cases := []struct {
		name                              string
		captured, refunded, pending, keep string
		want                              string
		wantErr                           error
	}{
		{name: "full refund of untouched payment", captured: "1000", refunded: "0", pending: "0", keep: "0", want: "1000"},
		{name: "overpayment after items were removed", captured: "1000", refunded: "0", pending: "0", keep: "850", want: "150"},
		{name: "overpayment already partly returned", captured: "1000", refunded: "100", pending: "0", keep: "850", want: "50"},
		{name: "full refund after a partial one", captured: "1000", refunded: "150", pending: "0", keep: "0", want: "850"},
		{name: "pending refund counts as returned", captured: "1000", refunded: "0", pending: "100", keep: "850", want: "50"},
		{name: "nothing owed", captured: "1000", refunded: "0", pending: "0", keep: "1000", wantErr: ErrNothingToRefund},
		{name: "overpayment fully returned", captured: "1000", refunded: "150", pending: "0", keep: "850", wantErr: ErrNothingToRefund},
		{name: "everything already refunded", captured: "1000", refunded: "1000", pending: "0", keep: "0", wantErr: ErrNothingToRefund},
		{name: "rest in flight", captured: "1000", refunded: "0", pending: "150", keep: "850", wantErr: ErrRefundPending},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := refundAmount(d(c.captured), d(c.refunded), d(c.pending), d(c.keep))
			if c.wantErr != nil {
				if !errors.Is(err, c.wantErr) {
					t.Fatalf("err = %v, want %v", err, c.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if !got.Equal(d(c.want)) {
				t.Errorf("refundAmount() = %s, want %s", got, c.want)
			}
		})
	}
}
