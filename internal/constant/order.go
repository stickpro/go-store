package constant

// OrderStatus is the fulfilment state of an order. Payment state is tracked
// separately in PaymentStatus.
type OrderStatus string //	@name	OrderStatus

const (
	// OrderNew is the entry status of a "quick order": the customer left only a
	// name + phone and a manager still has to collect address / delivery /
	// payment before it becomes a normal pending order.
	OrderNew        OrderStatus = "new"
	OrderPending    OrderStatus = "pending"
	OrderPaid       OrderStatus = "paid"
	OrderProcessing OrderStatus = "processing"
	OrderShipped    OrderStatus = "shipped"
	OrderDelivered  OrderStatus = "delivered"
	OrderCancelled  OrderStatus = "cancelled"
	OrderRefunded   OrderStatus = "refunded"
)

func (s OrderStatus) String() string { return string(s) }

// PaymentStatus is the payment state of an order, independent of fulfilment.
type PaymentStatus string //	@name	PaymentStatus

const (
	PaymentUnpaid   PaymentStatus = "unpaid"
	PaymentPaid     PaymentStatus = "paid"
	PaymentRefunded PaymentStatus = "refunded"
	PaymentFailed   PaymentStatus = "failed"
)

func (s PaymentStatus) String() string { return string(s) }

// Order acquisition channel, stored in orders.source.
const (
	// OrderSourceCheckout — full self-service checkout (address + delivery +
	// payment chosen by the customer).
	OrderSourceCheckout = "checkout"
	// OrderSourceQuick — one-click "quick order": name + phone only, a manager
	// collects the rest. Starts in OrderNew.
	OrderSourceQuick = "quick"
)

// Order status-history actors.
const (
	OrderActorCustomer = "customer"
	OrderActorSystem   = "system"
	// OrderActorAdmin is a prefix: "admin:<user-id>".
	OrderActorAdmin = "admin"
	// OrderActorPayment is a prefix: "payment:<provider>".
	OrderActorPayment = "payment"
)
