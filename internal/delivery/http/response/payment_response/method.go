package payment_response

import "github.com/stickpro/go-store/internal/service/payment"

// PaymentMethodResponse is one checkout payment method. The frontend renders
// its method tabs from this list: `code` is sent back as payment_method on
// POST /v1/orders (checkout) and, for an "online" method, as `provider` on
// POST /v1/orders/{id}/payment once the order exists.
type PaymentMethodResponse struct {
	Code     string `json:"code"`
	Title    string `json:"title"`
	Kind     string `json:"kind"` // cash | online
	Provider string `json:"provider,omitempty"`
	Enabled  bool   `json:"enabled"`
} //	@name	PaymentMethodResponse

func NewMethodFromInfo(m payment.MethodInfo) PaymentMethodResponse {
	return PaymentMethodResponse{
		Code:     m.Code,
		Title:    m.Title,
		Kind:     string(m.Kind),
		Provider: m.Provider,
		Enabled:  m.Enabled,
	}
}

func NewMethodList(methods []payment.MethodInfo) []PaymentMethodResponse {
	out := make([]PaymentMethodResponse, 0, len(methods))
	for _, m := range methods {
		out = append(out, NewMethodFromInfo(m))
	}
	return out
}
