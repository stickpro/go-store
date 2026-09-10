package order

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/service/shipping"
)

func decOf(s string) decimal.Decimal { return decimal.RequireFromString(s) }

func lineFixture() []checkoutLine {
	return []checkoutLine{{unitPrice: decOf("1000"), quantity: 2, weightKG: decOf("1")}}
}

func TestResolveShippingFallsBackToFlatWithoutCarrier(t *testing.T) {
	s := &Service{flatShipping: decOf("300"), freeShippingFrom: decOf("0")}

	got, err := s.resolveShipping(context.Background(), (dto.CreateOrderDTO{}).ShippingSelection(), lineFixture())
	if err != nil {
		t.Fatal(err)
	}
	if !got.cost.Equal(decOf("300")) || got.provider != nil {
		t.Fatalf("expected flat 300 with no carrier, got %+v", got)
	}
}

func TestResolveShippingRequoteMatchesTariff(t *testing.T) {
	reg := shipping.NewRegistry(nil, 0, shipping.ParcelDefaults{}, nil,
		rateStubProvider{code: "cdek", rates: []dto.ShippingRate{
			{Provider: "cdek", TariffCode: "136", TariffName: "Склад-склад", Cost: decOf("450"), MinDays: 1, MaxDays: 2},
			{Provider: "cdek", TariffCode: "137", TariffName: "Склад-дверь", Cost: decOf("620")},
		}},
	)
	s := &Service{shipping: reg, flatShipping: decOf("300")}

	provider, tariff := "cdek", "137"
	got, err := s.resolveShipping(context.Background(), (dto.CreateOrderDTO{
		ShipProvider: &provider, ShipTariffCode: &tariff, ShipPostcode: strp("190000"),
	}).ShippingSelection(), lineFixture())
	if err != nil {
		t.Fatal(err)
	}
	if !got.cost.Equal(decOf("620")) {
		t.Fatalf("cost = %s, want 620 (server re-quote, not client)", got.cost)
	}
	if got.provider == nil || *got.provider != "cdek" || got.tariff == nil || *got.tariff != "137" {
		t.Fatalf("carrier snapshot not set: %+v", got)
	}
	if got.method != nil {
		t.Fatalf("raw provider+tariff must not put a carrier tariff name into shipping_method, got %v", *got.method)
	}
}

func TestResolveShippingUnavailableTariff(t *testing.T) {
	reg := shipping.NewRegistry(nil, 0, shipping.ParcelDefaults{}, nil,
		rateStubProvider{code: "cdek", rates: []dto.ShippingRate{
			{Provider: "cdek", TariffCode: "136", Cost: decOf("450")},
		}},
	)
	s := &Service{shipping: reg}

	provider, tariff := "cdek", "999"
	_, err := s.resolveShipping(context.Background(), (dto.CreateOrderDTO{
		ShipProvider: &provider, ShipTariffCode: &tariff, ShipPostcode: strp("190000"),
	}).ShippingSelection(), lineFixture())
	if err == nil {
		t.Fatal("expected ErrShippingUnavailable for an unknown tariff")
	}
}

func strp(s string) *string { return &s }

// rateStubProvider satisfies shipping.RateProvider with canned rates.
type rateStubProvider struct {
	code  string
	rates []dto.ShippingRate
}

func (p rateStubProvider) Code() string                      { return p.code }
func (p rateStubProvider) Enabled() bool                     { return true }
func (p rateStubProvider) RunCacheRefresher(context.Context) {}
func (p rateStubProvider) Points() []dto.DeliveryPoint       { return nil }
func (p rateStubProvider) Quote(context.Context, shipping.RateQuery) ([]dto.ShippingRate, error) {
	return p.rates, nil
}

func TestResolveShippingMethodCodeSelfPickupFree(t *testing.T) {
	reg := shipping.NewRegistry(nil, 0, shipping.ParcelDefaults{},
		[]config.ShippingMethodConfig{{Code: "pickup", Title: "Самовывоз", Kind: "self_pickup", Free: true}},
	)
	s := &Service{shipping: reg, flatShipping: decOf("300")}

	code := "pickup"
	got, err := s.resolveShipping(context.Background(), dto.ShippingSelection{MethodCode: &code}, lineFixture())
	if err != nil {
		t.Fatal(err)
	}
	if !got.cost.IsZero() {
		t.Fatalf("self_pickup free must be 0, got %s", got.cost)
	}
}

func TestResolveShippingMethodCodeResolvesCarrier(t *testing.T) {
	reg := shipping.NewRegistry(nil, 0, shipping.ParcelDefaults{},
		[]config.ShippingMethodConfig{{Code: "cdek_pvz", Title: "СДЭК", Kind: "pickup", Provider: "cdek", Tariff: "136"}},
		rateStubProvider{code: "cdek", rates: []dto.ShippingRate{
			{Provider: "cdek", TariffCode: "136", Cost: decOf("450")},
		}},
	)
	s := &Service{shipping: reg, flatShipping: decOf("300")}

	code := "cdek_pvz"
	got, err := s.resolveShipping(context.Background(), dto.ShippingSelection{MethodCode: &code, Postcode: strp("190000")}, lineFixture())
	if err != nil {
		t.Fatal(err)
	}
	if !got.cost.Equal(decOf("450")) || got.tariff == nil || *got.tariff != "136" {
		t.Fatalf("method code should resolve to cdek/136: cost=%s %+v", got.cost, got)
	}
	if got.method == nil || *got.method != "cdek_pvz" {
		t.Fatalf("shipping_method should be the method code, got %v", got.method)
	}
}

func TestResolveShippingMethodCodeUnknown(t *testing.T) {
	reg := shipping.NewRegistry(nil, 0, shipping.ParcelDefaults{}, nil)
	s := &Service{shipping: reg}
	code := "nope"
	if _, err := s.resolveShipping(context.Background(), dto.ShippingSelection{MethodCode: &code}, lineFixture()); err == nil {
		t.Fatal("expected ErrShippingMethodUnknown")
	}
}

func TestResolveShippingMethodMarkupRoundsUp(t *testing.T) {
	reg := shipping.NewRegistry(nil, 0, shipping.ParcelDefaults{},
		[]config.ShippingMethodConfig{
			{Code: "cdek", Title: "СДЭК", Kind: "pickup", Provider: "cdek", Tariff: "136", Markup: "10"},
		},
		rateStubProvider{code: "cdek", rates: []dto.ShippingRate{
			{Provider: "cdek", TariffCode: "136", TariffName: "Посылка склад-склад", Cost: decOf("267.50")},
		}},
	)
	s := &Service{shipping: reg, flatShipping: decOf("300")}

	code := "cdek"
	got, err := s.resolveShipping(context.Background(), dto.ShippingSelection{MethodCode: &code, Postcode: strp("190000")}, lineFixture())
	if err != nil {
		t.Fatal(err)
	}
	// 267.50 * 1.10 = 294.25 -> ceil 295
	if !got.cost.Equal(decOf("295")) {
		t.Fatalf("markup+ceil: got %s, want 295", got.cost)
	}
	if got.method == nil || *got.method != "cdek" {
		t.Fatalf("shipping_method should be the method code, got %v", got.method)
	}
}
