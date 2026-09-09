package pochta

import (
	"testing"

	"github.com/stickpro/go-store/internal/dto"
)

func TestToDeliveryPoint(t *testing.T) {
	got := toDeliveryPoint(dto.PochtaDeliveryPointDTO{
		Code:          "115551",
		PostalCode:    "115551",
		Name:          "Отделение 115551",
		Type:          "ГОПС",
		Region:        "Москва г",
		Place:         "Москва",
		Address:       "Домодедовская ул, 20",
		Latitude:      55.61,
		Longitude:     37.70,
		Ecom:          true,
		CardPayment:   true,
		WeightLimitKg: 20,
		WorkTime:      []string{"пн, открыто: 08:00 - 20:00"},
	})

	if got.Provider != "pochta" || got.Code != "115551" || got.Locality != "Москва" || got.Country != "RU" {
		t.Fatalf("common fields not mapped: %+v", got)
	}
	if !got.CardPayment || len(got.WorkTime) != 1 {
		t.Fatalf("payment/worktime not mapped: %+v", got)
	}
	if got.Details["ecom"] != true || got.Details["weight_limit_kg"].(float64) != 20 {
		t.Fatalf("ecom details not mapped: %v", got.Details)
	}
}
