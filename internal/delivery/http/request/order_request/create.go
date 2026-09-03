package order_request

import "github.com/google/uuid"

// CreateOrderRequest is the checkout payload. The cart is taken from the
// caller's session / account, not from this body. Email is required for guest
// checkout and ignored for authenticated users (their account email is used).
type CreateOrderRequest struct {
	Email string  `json:"email" validate:"omitempty,email"`
	Phone *string `json:"phone" validate:"omitempty,max=32"`

	ShipCityID     *uuid.UUID `json:"ship_city_id"`
	ShipCityName   string     `json:"ship_city_name" validate:"required,max=255"`
	ShipAddress    string     `json:"ship_address" validate:"required,max=512"`
	ShipPostcode   *string    `json:"ship_postcode" validate:"omitempty,max=16"`
	ShipRecipient  string     `json:"ship_recipient" validate:"required,max=255"`
	ShippingMethod *string    `json:"shipping_method" validate:"omitempty,max=64"`

	PaymentMethod string  `json:"payment_method" validate:"required,max=32"`
	Comment       *string `json:"comment" validate:"omitempty,max=2000"`

	// ExpectedTotal, when set, must equal the server-computed grand total or the
	// order is rejected (price/availability changed since the cart was shown).
	// Decimal string, e.g. "1990.00".
	ExpectedTotal *string `json:"expected_total" validate:"omitempty,numeric"`
} //	@name	CreateOrderRequest
