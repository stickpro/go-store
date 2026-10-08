package order_request

// AdminRefundOrderRequest refunds an order's captured payment. There is no
// amount: by default the refund is exactly the overpayment (what the customer
// paid above the order's current total, e.g. after items were removed); with
// full it is everything not yet refunded, and the order moves to "refunded".
// Reason is kept on the refund record for the audit trail.
type AdminRefundOrderRequest struct {
	Full   bool    `json:"full"`
	Reason *string `json:"reason" validate:"omitempty,max=500"`
} //	@name	AdminRefundOrderRequest
