package delivery_request

// CalculateRatesRequest is the body of POST /v1/delivery/rates and
// POST /v1/delivery/{provider}/rates. The destination is one of to_point_code
// (a pickup point from /points) or to_postal_code. Parcel fields are optional —
// weight in kilograms, dimensions in centimetres; a zero/omitted component
// falls back to the store's default parcel.
type CalculateRatesRequest struct {
	FromPostalCode string `json:"from_postal_code" validate:"omitempty,numeric,len=6"`
	ToPostalCode   string `json:"to_postal_code" validate:"required_without=ToPointCode,omitempty,numeric,len=6"`
	ToPointCode    string `json:"to_point_code" validate:"required_without=ToPostalCode,omitempty,max=64"`
	DeliveryType   string `json:"delivery_type" validate:"omitempty,oneof=pickup courier"`

	WeightKG      string `json:"weight_kg" validate:"omitempty,numeric"`
	LengthCM      string `json:"length_cm" validate:"omitempty,numeric"`
	WidthCM       string `json:"width_cm" validate:"omitempty,numeric"`
	HeightCM      string `json:"height_cm" validate:"omitempty,numeric"`
	DeclaredValue string `json:"declared_value" validate:"omitempty,numeric"`
} //	@name	CalculateRatesRequest

// HasParcel reports whether the request carries an explicit parcel. When it
// does not, the handler derives the parcel from the caller's cart (or the store
// default when there is no cart).
func (r *CalculateRatesRequest) HasParcel() bool {
	return r.WeightKG != "" || r.LengthCM != "" || r.WidthCM != "" || r.HeightCM != ""
}
