package payment

import (
	"testing"

	"github.com/stickpro/go-store/internal/constant"
)

func TestRefundAllowed(t *testing.T) {
	cases := []struct {
		status constant.OrderStatus
		full   bool
		want   bool
	}{
		{constant.OrderPaid, true, true},
		{constant.OrderPaid, false, true},
		{constant.OrderCancelled, true, true},
		{constant.OrderCancelled, false, true},
		{constant.OrderProcessing, false, true},
		{constant.OrderProcessing, true, false},
		{constant.OrderShipped, false, false},
		{constant.OrderShipped, true, false},
		{constant.OrderDelivered, false, false},
		{constant.OrderDelivered, true, false},
		{constant.OrderPending, true, false},
		{constant.OrderRefunded, true, false},
	}
	for _, c := range cases {
		if got := refundAllowed(c.status, c.full); got != c.want {
			t.Errorf("refundAllowed(%s, full=%v) = %v, want %v", c.status, c.full, got, c.want)
		}
	}
}
