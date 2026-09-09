package yandexdelivery

import (
	"testing"

	"github.com/stickpro/go-store/internal/dto"
)

func TestToDeliveryPoint(t *testing.T) {
	got := toDeliveryPoint(dto.YandexDeliveryPointDTO{
		Code:           "yd-1",
		Name:           "ПВЗ Тверская",
		Type:           "pickup_point",
		Region:         "Москва",
		Locality:       "Москва",
		FullAddress:    "Тверская 1",
		Latitude:       55.76,
		Longitude:      37.61,
		Phone:          "+7 495 000",
		PaymentMethods: []string{"card_on_receipt"},
		GeoID:          213,
	})

	if got.Provider != "yandex_delivery" || got.Code != "yd-1" || got.Locality != "Москва" {
		t.Fatalf("common fields not mapped: %+v", got)
	}
	if len(got.Phones) != 1 || got.Phones[0] != "+7 495 000" {
		t.Fatalf("phone not mapped: %+v", got.Phones)
	}
	if !got.CardPayment {
		t.Fatalf("card_on_receipt should set CardPayment")
	}
	if got.Details["geo_id"] != 213 {
		t.Fatalf("geo_id should be in details, got %v", got.Details["geo_id"])
	}
}
