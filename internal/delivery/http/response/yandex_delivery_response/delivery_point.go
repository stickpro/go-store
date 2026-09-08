package yandex_delivery_response

import "github.com/stickpro/go-store/internal/dto"

type ScheduleResponse struct {
	Days     []int  `json:"days"`
	TimeFrom string `json:"time_from"`
	TimeTo   string `json:"time_to"`
} //	@name	YandexDeliveryScheduleResponse

type DeliveryPointResponse struct {
	Code              string `json:"code"`
	OperatorStationID string `json:"operator_station_id"`
	OperatorID        string `json:"operator_id"`
	Name              string `json:"name"`
	Type              string `json:"type"`

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

	Instruction    string             `json:"instruction"`
	Phone          string             `json:"phone"`
	Email          string             `json:"email"`
	PaymentMethods []string           `json:"payment_methods"`
	TimeZone       int                `json:"time_zone"`
	Schedule       []ScheduleResponse `json:"schedule"`

	IsYandexBranded     bool    `json:"is_yandex_branded"`
	IsMarketPartner     bool    `json:"is_market_partner"`
	IsPostOffice        bool    `json:"is_post_office"`
	AvailableForDropoff bool    `json:"available_for_dropoff"`
	DeactivationDate    *string `json:"deactivation_date"`
} //	@name	YandexDeliveryPointResponse

func NewFromDTO(p dto.YandexDeliveryPointDTO) *DeliveryPointResponse {
	schedule := make([]ScheduleResponse, 0, len(p.Schedule))
	for _, s := range p.Schedule {
		schedule = append(schedule, ScheduleResponse{
			Days:     s.Days,
			TimeFrom: s.TimeFrom,
			TimeTo:   s.TimeTo,
		})
	}

	return &DeliveryPointResponse{
		Code:                p.Code,
		OperatorStationID:   p.OperatorStationID,
		OperatorID:          p.OperatorID,
		Name:                p.Name,
		Type:                p.Type,
		Country:             p.Country,
		Region:              p.Region,
		SubRegion:           p.SubRegion,
		Locality:            p.Locality,
		Street:              p.Street,
		House:               p.House,
		PostalCode:          p.PostalCode,
		FullAddress:         p.FullAddress,
		GeoID:               p.GeoID,
		Latitude:            p.Latitude,
		Longitude:           p.Longitude,
		Instruction:         p.Instruction,
		Phone:               p.Phone,
		Email:               p.Email,
		PaymentMethods:      p.PaymentMethods,
		TimeZone:            p.TimeZone,
		Schedule:            schedule,
		IsYandexBranded:     p.IsYandexBranded,
		IsMarketPartner:     p.IsMarketPartner,
		IsPostOffice:        p.IsPostOffice,
		AvailableForDropoff: p.AvailableForDropoff,
		DeactivationDate:    p.DeactivationDate,
	}
}

func NewListFromDTO(points []dto.YandexDeliveryPointDTO) []*DeliveryPointResponse {
	res := make([]*DeliveryPointResponse, 0, len(points))
	for _, p := range points {
		res = append(res, NewFromDTO(p))
	}
	return res
}
