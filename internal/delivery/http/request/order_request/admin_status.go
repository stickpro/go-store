package order_request

// UpdateOrderStatusRequest drives an admin-initiated order status transition.
// "pending" is never a valid target (it's the initial status only); illegal
// transitions from the order's current status are rejected by the service
// with a 409, not here.
type UpdateOrderStatusRequest struct {
	Status  string  `json:"status" validate:"required,oneof=paid processing shipped delivered cancelled refunded"`
	Comment *string `json:"comment" validate:"omitempty,max=1000"`
	// PaymentMethod is only used when Status is "paid"; ignored otherwise.
	PaymentMethod *string `json:"payment_method" validate:"omitempty,max=32"`
} //	@name	UpdateOrderStatusRequest
