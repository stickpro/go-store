package shipping

import (
	"context"
	"testing"

	"github.com/stickpro/go-store/internal/dto"
)

func f64(v float64) *float64 { return &v }

func points() []dto.DeliveryPoint {
	return []dto.DeliveryPoint{
		{Code: "a", Type: "PVZ", Region: "Москва г", Locality: "Москва", PostalCode: "101000", Latitude: 55.75, Longitude: 37.62},
		{Code: "b", Type: "POSTAMAT", Region: "Москва г", Locality: "Москва", PostalCode: "115551", Latitude: 55.90, Longitude: 37.40},
		{Code: "c", Type: "PVZ", Region: "Санкт-Петербург г", Locality: "Санкт-Петербург", PostalCode: "190000", Latitude: 59.94, Longitude: 30.31},
	}
}

func codes(in []dto.DeliveryPoint) []string {
	out := make([]string, len(in))
	for i, p := range in {
		out[i] = p.Code
	}
	return out
}

func TestFilter(t *testing.T) {
	tests := []struct {
		name   string
		filter dto.DeliveryPointsFilter
		want   []string
	}{
		{"no filter", dto.DeliveryPointsFilter{}, []string{"a", "b", "c"}},
		{"by type", dto.DeliveryPointsFilter{Type: "postamat"}, []string{"b"}},
		{"by locality", dto.DeliveryPointsFilter{Locality: "санкт-петербург"}, []string{"c"}},
		{"by region substring", dto.DeliveryPointsFilter{Region: "Москва"}, []string{"a", "b"}},
		{"by index matches postal code", dto.DeliveryPointsFilter{Index: "115551"}, []string{"b"}},
		{
			"by bbox",
			dto.DeliveryPointsFilter{BBox: &dto.GeoBBox{MinLat: 55.0, MaxLat: 56.0, MinLon: 37.0, MaxLon: 38.0}},
			[]string{"a", "b"},
		},
		{
			"nearby sorts by distance",
			dto.DeliveryPointsFilter{Latitude: f64(55.76), Longitude: f64(37.61), RadiusKM: f64(60)},
			[]string{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := codes(Filter(points(), tt.filter))
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

type stubProvider struct{ code string }

func (s stubProvider) Code() string                      { return s.code }
func (s stubProvider) Enabled() bool                     { return true }
func (s stubProvider) RunCacheRefresher(context.Context) {}
func (s stubProvider) Points() []dto.DeliveryPoint       { return nil }

func TestRegistry(t *testing.T) {
	r := NewRegistry(nil, 0, ParcelDefaults{}, nil, stubProvider{"cdek"}, stubProvider{"pochta"})
	if len(r.All()) != 2 {
		t.Fatalf("All: want 2, got %d", len(r.All()))
	}
	if _, ok := r.Get("cdek"); !ok {
		t.Fatal("Get(cdek): not found")
	}
	if _, ok := r.Get("missing"); ok {
		t.Fatal("Get(missing): unexpectedly found")
	}
}
