package shipping

import (
	"context"
	"errors"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/dto"
)

// rateStub is a Provider that also quotes, driven by canned data.
type rateStub struct {
	stubProvider
	rates []dto.ShippingRate
	err   error
}

func (s rateStub) Quote(context.Context, RateQuery) ([]dto.ShippingRate, error) {
	return s.rates, s.err
}

func rate(provider string, cost string) dto.ShippingRate {
	return dto.ShippingRate{Provider: provider, Cost: decimal.RequireFromString(cost), Currency: "RUB"}
}

func TestQuoteAllMergesSortsAndTolerates(t *testing.T) {
	reg := NewRegistry(nil, 0, ParcelDefaults{}, nil,
		rateStub{stubProvider: stubProvider{"cdek"}, rates: []dto.ShippingRate{rate("cdek", "500"), rate("cdek", "300")}},
		rateStub{stubProvider: stubProvider{"pochta"}, rates: []dto.ShippingRate{rate("pochta", "400")}},
		rateStub{stubProvider: stubProvider{"broken"}, err: errors.New("carrier down")},
		stubProvider{"yandex_delivery"}, // not a RateProvider — skipped
	)

	rates, errs := reg.QuoteAll(context.Background(), RateQuery{})

	if got := len(rates); got != 3 {
		t.Fatalf("want 3 rates, got %d (%+v)", got, rates)
	}
	if !rates[0].Cost.Equal(decimal.NewFromInt(300)) || !rates[2].Cost.Equal(decimal.NewFromInt(500)) {
		t.Fatalf("rates not sorted cheapest-first: %+v", rates)
	}
	if len(errs) != 1 {
		t.Fatalf("want 1 collected error, got %d", len(errs))
	}
}

func TestRaterSkipsDisabledAndUnsupported(t *testing.T) {
	reg := NewRegistry(nil, 0, ParcelDefaults{}, nil,
		rateStub{stubProvider: stubProvider{"cdek"}},
		stubProvider{"yandex_delivery"},
	)
	if _, ok := reg.Rater("cdek"); !ok {
		t.Fatal("cdek should be a rater")
	}
	if _, ok := reg.Rater("yandex_delivery"); ok {
		t.Fatal("yandex_delivery is not a RateProvider")
	}
	if _, ok := reg.Rater("missing"); ok {
		t.Fatal("missing carrier")
	}
}

func TestMethodsCatalogue(t *testing.T) {
	reg := NewRegistry(nil, 0, ParcelDefaults{},
		[]config.ShippingMethodConfig{
			{Code: "pickup", Title: "Самовывоз", Kind: "self_pickup", Free: true},
			{Code: "cdek", Title: "СДЭК", Kind: "pickup", Provider: "cdek", Tariff: "136"},
			{Code: "yandex", Title: "Яндекс", Kind: "pickup", Provider: "yandex_delivery"},
			{Code: "off", Title: "Отключён", Kind: "pickup", Provider: "missing"},
		},
		rateStub{stubProvider: stubProvider{"cdek"}},
		stubProvider{"yandex_delivery"}, // enabled Provider, but not a RateProvider
	)

	got := map[string]MethodInfo{}
	for _, m := range reg.Methods() {
		got[m.Code] = m
	}

	if m := got["pickup"]; !m.Enabled || m.HasPoints || m.HasRates || !m.Free {
		t.Fatalf("self_pickup: %+v", m)
	}
	if m := got["cdek"]; !m.Enabled || !m.HasPoints || !m.HasRates {
		t.Fatalf("cdek: %+v", m)
	}
	if m := got["yandex"]; !m.Enabled || !m.HasPoints || m.HasRates {
		t.Fatalf("yandex should have points but no rates: %+v", m)
	}
	if m := got["off"]; m.Enabled {
		t.Fatalf("unknown provider must be disabled: %+v", m)
	}

	if m, ok := reg.Method("cdek"); !ok || m.TariffCode != "136" {
		t.Fatalf("Method(cdek): %+v ok=%v", m, ok)
	}
}
