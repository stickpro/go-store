package order

import "github.com/stickpro/go-store/internal/constant"

// allowedTransitions is the order fulfilment state machine. A missing key (e.g.
// "cancelled", "refunded") is terminal.
var allowedTransitions = map[constant.OrderStatus][]constant.OrderStatus{
	constant.OrderNew:        {constant.OrderPending, constant.OrderCancelled},
	constant.OrderPending:    {constant.OrderPaid, constant.OrderCancelled},
	constant.OrderPaid:       {constant.OrderProcessing, constant.OrderCancelled, constant.OrderRefunded},
	constant.OrderProcessing: {constant.OrderShipped, constant.OrderCancelled, constant.OrderRefunded},
	constant.OrderShipped:    {constant.OrderDelivered, constant.OrderRefunded},
	constant.OrderDelivered:  {constant.OrderRefunded},
}

// canTransition reports whether an order may move from -> to.
func canTransition(from, to constant.OrderStatus) bool {
	for _, s := range allowedTransitions[from] {
		if s == to {
			return true
		}
	}
	return false
}

// restocksOn reports whether moving into this status should return reserved
// stock to the catalogue (cancel / refund of an order that had already
// decremented stock).
func restocksOn(to constant.OrderStatus) bool {
	return to == constant.OrderCancelled || to == constant.OrderRefunded
}
