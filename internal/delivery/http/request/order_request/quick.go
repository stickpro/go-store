package order_request

// CreateQuickOrderRequest is the one-click "quick order" payload. The cart is
// taken from the caller's session / account, not from this body. A manager calls
// the customer back to collect address, delivery and payment, so only contact
// details are required here. Email is optional (ignored for authenticated users,
// whose account email is used).
type CreateQuickOrderRequest struct {
	Name    string  `json:"name" validate:"required,max=255"`
	Phone   string  `json:"phone" validate:"required,max=32"`
	Email   *string `json:"email" validate:"omitempty,email,max=255"`
	Comment *string `json:"comment" validate:"omitempty,max=2000"`
} //	@name	CreateQuickOrderRequest
