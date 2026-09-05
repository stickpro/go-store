package dto

// CDEKDeliveryPointDTO is a CDEK (СДЭК) pickup point (office or postomat) usable
// as an order delivery destination. It is the Service<->Delivery currency for
// the CDEK integration; it never touches the CDEK API's own JSON shape.
type CDEKDeliveryPointDTO struct {
	Code        string   `json:"code"`
	Name        string   `json:"name"`
	Type        string   `json:"type"` // "PVZ" (office) or "POSTAMAT"
	OwnerCode   string   `json:"owner_code"`
	CountryCode string   `json:"country_code"`
	RegionCode  int      `json:"region_code"`
	Region      string   `json:"region"`
	CityCode    int      `json:"city_code"`
	City        string   `json:"city"`
	PostalCode  string   `json:"postal_code"`
	Address     string   `json:"address"`
	AddressFull string   `json:"address_full"`
	Latitude    float64  `json:"latitude"`
	Longitude   float64  `json:"longitude"`
	WorkTime    string   `json:"work_time"`
	Phones      []string `json:"phones"`
	Email       string   `json:"email"`
	Note        string   `json:"note"`

	TakeOnly       bool `json:"take_only"`
	IsHandout      bool `json:"is_handout"`
	IsReception    bool `json:"is_reception"`
	IsDressingRoom bool `json:"is_dressing_room"`
	HaveCash       bool `json:"have_cash"`
	HaveCashless   bool `json:"have_cashless"`
	AllowedCod     bool `json:"allowed_cod"`

	WeightMin float64 `json:"weight_min"`
	WeightMax float64 `json:"weight_max"`
}

// CDEKDeliveryPointsFilter narrows the cached delivery points list. All fields
// are optional; the zero value of a field means "don't filter on it".
type CDEKDeliveryPointsFilter struct {
	CityCode   *int
	PostalCode string
	Type       string // "PVZ" or "POSTAMAT"
	BBox       *CDEKDeliveryPointsBBox
}

// CDEKDeliveryPointsBBox narrows delivery points to a rectangular geographic
// area (typically a map viewport), inclusive on all bounds. It does not
// handle boxes that cross the antimeridian.
type CDEKDeliveryPointsBBox struct {
	MinLat float64
	MaxLat float64
	MinLon float64
	MaxLon float64
}
