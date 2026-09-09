package dto

// DeliveryPoint is a carrier-agnostic pickup point (CDEK office, Yandex ПВЗ,
// Russian Post ОПС, …) as returned by the unified /v1/delivery/{provider}/points
// endpoint. Common fields cover what a storefront map and list need; anything
// carrier-specific lives in Details.
type DeliveryPoint struct {
	Provider   string `json:"provider"` // "cdek" | "yandex_delivery" | "pochta"
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

	// Details carries carrier-specific fields not modelled above (weight limits,
	// operator ids, schedule structs, e-commerce flags, …).
	Details map[string]any `json:"details,omitempty"`
}

// DeliveryPointsFilter narrows a provider's pickup points list. All fields are
// optional; the zero value of a field means "don't filter on it". When Latitude
// and Longitude are set (and BBox is not) the result is limited to RadiusKM
// around that point and sorted nearest-first.
type DeliveryPointsFilter struct {
	Type     string
	Locality string
	Region   string
	Index    string // matched against Code and PostalCode

	BBox *GeoBBox

	Latitude  *float64
	Longitude *float64
	RadiusKM  *float64
}
