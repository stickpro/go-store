package dto

// PochtaDeliveryPointDTO is a Russian Post (Почта России) pickup point — a post
// office (ГОПС/СОПС), a "ПВЗ" (ППВЗ) or a postomat (Почтомат) — usable as an
// order delivery destination. It is the Service<->Delivery currency for the
// Russian Post integration; it never touches the Otpravka API's own JSON shape.
type PochtaDeliveryPointDTO struct {
	Code       string `json:"code"`        // postal index, e.g. "115551"
	PostalCode string `json:"postal_code"` // same as Code, kept for symmetry with CDEK
	Name       string `json:"name"`        // brand name if any, else "Отделение <index>"
	Type       string `json:"type"`        // "ГОПС" / "СОПС" / "Почтомат" / "ППВЗ" / "ППС"
	BrandName  string `json:"brand_name"`

	Region string `json:"region"`
	Area   string `json:"area"`
	Place  string `json:"place"` // settlement / locality
	Street string `json:"street"`
	House  string `json:"house"`
	Office string `json:"office"`
	// Address is the fully formatted address string (addressFias.ads).
	Address string `json:"address"`
	// Description says where inside a building the point is, when known
	// (ecomOptions.getto), e.g. "Почтомат расположен в отделении почтовой связи".
	Description string `json:"description"`

	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`

	// Ecom is true when the point accepts e-commerce (online-shop) parcels.
	Ecom bool `json:"ecom"`

	CardPayment       bool    `json:"card_payment"`
	CashPayment       bool    `json:"cash_payment"`
	WeightLimitKg     float64 `json:"weight_limit_kg"`
	WithFitting       bool    `json:"with_fitting"`
	ContentsChecking  bool    `json:"contents_checking"`
	PartialRedemption bool    `json:"partial_redemption"`
	ReturnAvailable   bool    `json:"return_available"`

	// WorkTime is the human-readable weekly schedule, one line per weekday as
	// returned by the ОПС passport, e.g. "пн, открыто: 08:00 - 20:00".
	WorkTime []string `json:"work_time"`
}
