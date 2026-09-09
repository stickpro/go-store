package delivery_response

import "github.com/stickpro/go-store/internal/service/shipping"

// DeliveryMethodResponse is one checkout delivery method. The frontend renders
// its method tabs from this list: it knows the title, whether to show a
// pickup-point map (has_points), whether to expect a price (has_rates), and
// passes `code` back as delivery_method_code at checkout.
type DeliveryMethodResponse struct {
	Code       string `json:"code"`
	Title      string `json:"title"`
	Kind       string `json:"kind"` // self_pickup | pickup | courier
	Provider   string `json:"provider,omitempty"`
	TariffCode string `json:"tariff_code,omitempty"`
	Enabled    bool   `json:"enabled"`
	HasPoints  bool   `json:"has_points"`
	HasRates   bool   `json:"has_rates"`
	Free       bool   `json:"free"`
} //	@name	DeliveryMethodResponse

func NewMethodFromInfo(m shipping.MethodInfo) DeliveryMethodResponse {
	return DeliveryMethodResponse{
		Code:       m.Code,
		Title:      m.Title,
		Kind:       string(m.Kind),
		Provider:   m.Provider,
		TariffCode: m.TariffCode,
		Enabled:    m.Enabled,
		HasPoints:  m.HasPoints,
		HasRates:   m.HasRates,
		Free:       m.Free,
	}
}

func NewMethodList(methods []shipping.MethodInfo) []DeliveryMethodResponse {
	out := make([]DeliveryMethodResponse, 0, len(methods))
	for _, m := range methods {
		out = append(out, NewMethodFromInfo(m))
	}
	return out
}
