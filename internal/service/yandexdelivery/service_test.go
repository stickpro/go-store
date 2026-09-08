package yandexdelivery

import (
	"context"
	"testing"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/dto"
)

func testPoints() []dto.YandexDeliveryPointDTO {
	return []dto.YandexDeliveryPointDTO{
		{Code: "a", Type: "pickup_point", GeoID: 213, Locality: "Москва", Latitude: 55.75, Longitude: 37.62},
		{Code: "b", Type: "terminal", GeoID: 213, Locality: "Москва", Latitude: 55.90, Longitude: 37.40},
		{Code: "c", Type: "pickup_point", GeoID: 2, Locality: "Санкт-Петербург", Latitude: 59.94, Longitude: 30.31},
	}
}

func codes(points []dto.YandexDeliveryPointDTO) []string {
	out := make([]string, len(points))
	for i, p := range points {
		out[i] = p.Code
	}
	return out
}

func TestApplyFilter(t *testing.T) {
	geo := 213
	tests := []struct {
		name   string
		filter dto.YandexDeliveryPointsFilter
		want   []string
	}{
		{"no filter", dto.YandexDeliveryPointsFilter{}, []string{"a", "b", "c"}},
		{"by type", dto.YandexDeliveryPointsFilter{Type: "terminal"}, []string{"b"}},
		{"by geo id", dto.YandexDeliveryPointsFilter{GeoID: &geo}, []string{"a", "b"}},
		{"by locality", dto.YandexDeliveryPointsFilter{Locality: "Санкт-Петербург"}, []string{"c"}},
		{
			"by bbox",
			dto.YandexDeliveryPointsFilter{BBox: &dto.GeoBBox{MinLat: 55.0, MaxLat: 56.0, MinLon: 37.0, MaxLon: 38.0}},
			[]string{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := codes(applyFilter(testPoints(), tt.filter))
			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("got %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestListDeliveryPointsReadsMemory(t *testing.T) {
	s := &Service{cfg: &config.Config{}}
	s.cfg.YandexDelivery.Enabled = true
	s.store(testPoints())

	got, err := s.ListDeliveryPoints(context.Background(), dto.YandexDeliveryPointsFilter{Type: "pickup_point"})
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"a", "c"}; len(got) != len(want) || got[0].Code != want[0] || got[1].Code != want[1] {
		t.Fatalf("got %v, want %v", codes(got), want)
	}
}

func TestListDeliveryPointsDisabled(t *testing.T) {
	s := &Service{cfg: &config.Config{}}
	s.store(testPoints())

	got, err := s.ListDeliveryPoints(context.Background(), dto.YandexDeliveryPointsFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty list when integration disabled, got %d", len(got))
	}
}
