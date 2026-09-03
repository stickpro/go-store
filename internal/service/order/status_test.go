package order

import (
	"testing"

	"github.com/stickpro/go-store/internal/constant"
)

func TestCanTransition(t *testing.T) {
	ok := [][2]constant.OrderStatus{
		{constant.OrderPending, constant.OrderPaid},
		{constant.OrderPending, constant.OrderCancelled},
		{constant.OrderPaid, constant.OrderProcessing},
		{constant.OrderPaid, constant.OrderRefunded},
		{constant.OrderProcessing, constant.OrderShipped},
		{constant.OrderShipped, constant.OrderDelivered},
		{constant.OrderDelivered, constant.OrderRefunded},
	}
	for _, c := range ok {
		if !canTransition(c[0], c[1]) {
			t.Errorf("expected %s -> %s allowed", c[0], c[1])
		}
	}

	bad := [][2]constant.OrderStatus{
		{constant.OrderPending, constant.OrderShipped},
		{constant.OrderPending, constant.OrderDelivered},
		{constant.OrderDelivered, constant.OrderPending},
		{constant.OrderCancelled, constant.OrderPaid},
		{constant.OrderRefunded, constant.OrderProcessing},
		{constant.OrderShipped, constant.OrderCancelled},
	}
	for _, c := range bad {
		if canTransition(c[0], c[1]) {
			t.Errorf("expected %s -> %s rejected", c[0], c[1])
		}
	}
}

func TestRestocksOn(t *testing.T) {
	if !restocksOn(constant.OrderCancelled) || !restocksOn(constant.OrderRefunded) {
		t.Fatal("cancel/refund should restock")
	}
	if restocksOn(constant.OrderShipped) || restocksOn(constant.OrderPaid) {
		t.Fatal("non-terminal statuses should not restock")
	}
}
