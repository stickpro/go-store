package delivery_response

import "github.com/stickpro/go-store/internal/dto"

// DeliveryPointResponse is one pickup point in the carrier-agnostic shape.
// Carrier-specific fields (weight limits, operator ids, schedule structs,
// e-commerce flags, …) live under Details.
type DeliveryPointResponse struct {
	Provider   string `json:"provider"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
	Region     string `json:"region"`
	Locality   string `json:"locality"`
	Address    string `json:"address"`

	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`

	Phones   []string `json:"phones"`
	Email    string   `json:"email"`
	WorkTime []string `json:"work_time"`

	CardPayment bool `json:"card_payment"`
	CashPayment bool `json:"cash_payment"`

	Details map[string]any `json:"details,omitempty"`
} //	@name	DeliveryPointResponse

// ProviderResponse describes one supported delivery carrier.
type ProviderResponse struct {
	Code    string `json:"code"`
	Enabled bool   `json:"enabled"`
} //	@name	DeliveryProviderResponse

func NewFromDTO(p dto.DeliveryPoint) *DeliveryPointResponse {
	return &DeliveryPointResponse{
		Provider:    p.Provider,
		Code:        p.Code,
		Name:        p.Name,
		Type:        p.Type,
		PostalCode:  p.PostalCode,
		Country:     p.Country,
		Region:      p.Region,
		Locality:    p.Locality,
		Address:     p.Address,
		Latitude:    p.Latitude,
		Longitude:   p.Longitude,
		Phones:      p.Phones,
		Email:       p.Email,
		WorkTime:    p.WorkTime,
		CardPayment: p.CardPayment,
		CashPayment: p.CashPayment,
		Details:     p.Details,
	}
}

func NewList(points []dto.DeliveryPoint) []*DeliveryPointResponse {
	res := make([]*DeliveryPointResponse, 0, len(points))
	for _, p := range points {
		res = append(res, NewFromDTO(p))
	}
	return res
}
