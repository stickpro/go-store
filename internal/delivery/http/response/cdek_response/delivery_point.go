package cdek_response

import "github.com/stickpro/go-store/internal/dto"

type DeliveryPointResponse struct {
	Code        string   `json:"code"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
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
} //	@name	CDEKDeliveryPointResponse

func NewFromDTO(p dto.CDEKDeliveryPointDTO) *DeliveryPointResponse {
	return &DeliveryPointResponse{
		Code:           p.Code,
		Name:           p.Name,
		Type:           p.Type,
		CountryCode:    p.CountryCode,
		RegionCode:     p.RegionCode,
		Region:         p.Region,
		CityCode:       p.CityCode,
		City:           p.City,
		PostalCode:     p.PostalCode,
		Address:        p.Address,
		AddressFull:    p.AddressFull,
		Latitude:       p.Latitude,
		Longitude:      p.Longitude,
		WorkTime:       p.WorkTime,
		Phones:         p.Phones,
		Email:          p.Email,
		Note:           p.Note,
		TakeOnly:       p.TakeOnly,
		IsHandout:      p.IsHandout,
		IsReception:    p.IsReception,
		IsDressingRoom: p.IsDressingRoom,
		HaveCash:       p.HaveCash,
		HaveCashless:   p.HaveCashless,
		AllowedCod:     p.AllowedCod,
		WeightMin:      p.WeightMin,
		WeightMax:      p.WeightMax,
	}
}

func NewListFromDTO(points []dto.CDEKDeliveryPointDTO) []*DeliveryPointResponse {
	res := make([]*DeliveryPointResponse, 0, len(points))
	for _, p := range points {
		res = append(res, NewFromDTO(p))
	}
	return res
}
