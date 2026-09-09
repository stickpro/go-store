package delivery_request

// ListDeliveryPointsRequest is the query for GET /v1/delivery/{provider}/points.
// Every field is optional. Coordinates come in one of two mutually independent
// forms, each all-or-nothing:
//   - latitude + longitude (+ optional radius_km): a nearby search;
//   - min_lat + max_lat + min_lon + max_lon: a map-viewport box.
//
// On the coordinate fields `required_with` deliberately precedes `omitempty`:
// `omitempty` short-circuits the whole tag on a nil pointer, so the "required"
// check has to run before it.
type ListDeliveryPointsRequest struct {
	Type     string `json:"type" query:"type" validate:"omitempty,max=100"`
	Locality string `json:"locality" query:"locality" validate:"omitempty,max=200"`
	Region   string `json:"region" query:"region" validate:"omitempty,max=200"`
	Index    string `json:"index" query:"index" validate:"omitempty,max=20"`

	Latitude  *float64 `json:"latitude" query:"latitude" validate:"required_with=Longitude,omitempty,latitude"`
	Longitude *float64 `json:"longitude" query:"longitude" validate:"required_with=Latitude,omitempty,longitude"`
	RadiusKM  *float64 `json:"radius_km" query:"radius_km" validate:"omitempty,gt=0,lte=2000"`

	MinLat *float64 `json:"min_lat" query:"min_lat" validate:"required_with=MaxLat MinLon MaxLon,omitempty,latitude,ltefield=MaxLat"`
	MaxLat *float64 `json:"max_lat" query:"max_lat" validate:"required_with=MinLat MinLon MaxLon,omitempty,latitude"`
	MinLon *float64 `json:"min_lon" query:"min_lon" validate:"required_with=MinLat MaxLat MaxLon,omitempty,longitude,ltefield=MaxLon"`
	MaxLon *float64 `json:"max_lon" query:"max_lon" validate:"required_with=MinLat MaxLat MinLon,omitempty,longitude"`
} //	@name	ListDeliveryPointsRequest

// HasBBox reports whether a full map-viewport box was supplied. Validation
// guarantees the four bounds are set together, so checking one is enough.
func (r *ListDeliveryPointsRequest) HasBBox() bool {
	return r.MinLat != nil
}
