package dto

// YandexDeliveryPointDTO is a Yandex Delivery (Яндекс Доставка) pickup point
// (ПВЗ or postomat) usable as an order delivery destination. It is the
// Service<->Delivery currency for the Yandex Delivery integration; it never
// touches the Yandex API's own JSON shape.
type YandexDeliveryPointDTO struct {
	Code              string `json:"code"` // Yandex pickup point id
	OperatorStationID string `json:"operator_station_id"`
	OperatorID        string `json:"operator_id"`
	Name              string `json:"name"`
	Type              string `json:"type"` // "pickup_point" or "terminal" (postomat)

	Country     string `json:"country"`
	Region      string `json:"region"`
	SubRegion   string `json:"sub_region"`
	Locality    string `json:"locality"`
	Street      string `json:"street"`
	House       string `json:"house"`
	PostalCode  string `json:"postal_code"`
	FullAddress string `json:"full_address"`
	GeoID       int    `json:"geo_id"`

	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`

	Instruction    string                      `json:"instruction"`
	Phone          string                      `json:"phone"`
	Email          string                      `json:"email"`
	PaymentMethods []string                    `json:"payment_methods"`
	TimeZone       int                         `json:"time_zone"` // UTC offset in hours
	Schedule       []YandexDeliveryScheduleDTO `json:"schedule"`

	IsYandexBranded     bool    `json:"is_yandex_branded"`
	IsMarketPartner     bool    `json:"is_market_partner"`
	IsPostOffice        bool    `json:"is_post_office"`
	AvailableForDropoff bool    `json:"available_for_dropoff"`
	DeactivationDate    *string `json:"deactivation_date"`
}

// YandexDeliveryScheduleDTO is one working-hours restriction: the point is open
// on Days (ISO weekday numbers, 1=Monday) between TimeFrom and TimeTo ("HH:MM").
type YandexDeliveryScheduleDTO struct {
	Days     []int  `json:"days"`
	TimeFrom string `json:"time_from"`
	TimeTo   string `json:"time_to"`
}
