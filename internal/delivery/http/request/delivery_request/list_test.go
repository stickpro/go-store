package delivery_request

import (
	"testing"

	"github.com/stickpro/go-store/internal/tools"
)

func f(v float64) *float64 { return &v }

func TestListDeliveryPointsRequestValidation(t *testing.T) {
	v := tools.DefaultStructValidator()

	tests := []struct {
		name    string
		req     ListDeliveryPointsRequest
		wantErr bool
	}{
		{"empty is valid", ListDeliveryPointsRequest{}, false},
		{"plain filters", ListDeliveryPointsRequest{Type: "ПВЗ", Locality: "Москва", Index: "115551"}, false},
		{"nearby full", ListDeliveryPointsRequest{Latitude: f(55.75), Longitude: f(37.62), RadiusKM: f(15)}, false},
		{"nearby missing longitude", ListDeliveryPointsRequest{Latitude: f(55.75)}, true},
		{"latitude out of range", ListDeliveryPointsRequest{Latitude: f(120), Longitude: f(37.62)}, true},
		{"negative radius", ListDeliveryPointsRequest{Latitude: f(55.75), Longitude: f(37.62), RadiusKM: f(-1)}, true},
		{"bbox full", ListDeliveryPointsRequest{MinLat: f(55.0), MaxLat: f(56.0), MinLon: f(37.0), MaxLon: f(38.0)}, false},
		{"bbox partial", ListDeliveryPointsRequest{MinLat: f(55.0), MaxLat: f(56.0)}, true},
		{"bbox min greater than max", ListDeliveryPointsRequest{MinLat: f(56.0), MaxLat: f(55.0), MinLon: f(37.0), MaxLon: f(38.0)}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(&tt.req)
			if tt.wantErr && err == nil {
				t.Fatalf("expected validation error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
