package order_request

// CheckoutPreviewRequest asks for the server-computed cart total under a
// delivery choice, without creating an order. The cart is taken from the
// caller's session / account. Destination is ship_point_code (a pickup point
// from /v1/delivery/{provider}/points) or ship_postcode; with neither, and no
// carrier, the flat shipping rate applies.
type CheckoutPreviewRequest struct {
	// delivery_method_code from GET /v1/delivery/methods (preferred).
	DeliveryMethodCode *string `json:"delivery_method_code" validate:"omitempty,max=32"`
	ShipProvider       *string `json:"ship_provider" validate:"omitempty,oneof=cdek yandex_delivery pochta"`
	ShipTariffCode     *string `json:"ship_tariff_code" validate:"omitempty,max=64,required_with=ShipProvider"`
	ShipPointCode      *string `json:"ship_point_code" validate:"omitempty,max=64"`
	ShipPostcode       *string `json:"ship_postcode" validate:"omitempty,max=16"`
} //	@name	CheckoutPreviewRequest
