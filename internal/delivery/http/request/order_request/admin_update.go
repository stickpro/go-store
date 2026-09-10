package order_request

import "github.com/google/uuid"

// AdminUpdateOrderRequest is the admin order-edit payload. Every field is
// optional; an omitted field leaves that value unchanged. Item lines and their
// prices cannot be changed here. Only orders in status "new" or "pending" are
// editable; editing a "new" order confirms it into "pending" and then requires a
// shipping address.
//
// Delivery: send delivery_method_code (from GET /v1/delivery/methods) or the raw
// ship_provider + ship_tariff_code pair to re-quote the carrier; leave all of
// them out to keep the stored shipping cost and window.
type AdminUpdateOrderRequest struct {
	Email         *string    `json:"email" validate:"omitempty,email,max=255"`
	Phone         *string    `json:"phone" validate:"omitempty,max=32"`
	ShipCityID    *uuid.UUID `json:"ship_city_id"`
	ShipCityName  *string    `json:"ship_city_name" validate:"omitempty,max=255"`
	ShipAddress   *string    `json:"ship_address" validate:"omitempty,max=512"`
	ShipPostcode  *string    `json:"ship_postcode" validate:"omitempty,max=16"`
	ShipRecipient *string    `json:"ship_recipient" validate:"omitempty,max=255"`

	DeliveryMethodCode *string `json:"delivery_method_code" validate:"omitempty,max=32"`
	ShipProvider       *string `json:"ship_provider" validate:"omitempty,oneof=cdek yandex_delivery pochta"`
	ShipTariffCode     *string `json:"ship_tariff_code" validate:"omitempty,max=64,required_with=ShipProvider"`
	ShipPointCode      *string `json:"ship_point_code" validate:"omitempty,max=64"`

	PaymentMethod *string `json:"payment_method" validate:"omitempty,max=32"`
	Comment       *string `json:"comment" validate:"omitempty,max=2000"`
} //	@name	AdminUpdateOrderRequest
