package order

import (
	"context"
	"testing"

	"github.com/stickpro/go-store/internal/config"

	"github.com/google/uuid"

	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/service/shipping"
)

// cartStub is a minimal cart.ICartService for PreviewCheckout tests.
type cartStub struct{ items []dto.CartItemsDTO }

func (c cartStub) GetCart(context.Context, dto.Owner) (*dto.CartDTO, error) {
	return &dto.CartDTO{Items: c.items}, nil
}
func (cartStub) AddItem(context.Context, dto.Owner, dto.AddCartItemDTO) (*dto.CartDTO, error) {
	return nil, nil
}
func (cartStub) UpdateQuantity(context.Context, dto.Owner, uuid.UUID, int64) (*dto.CartDTO, error) {
	return nil, nil
}
func (cartStub) RemoveItem(context.Context, dto.Owner, uuid.UUID) (*dto.CartDTO, error) {
	return nil, nil
}
func (cartStub) ClearCart(context.Context, dto.Owner) error { return nil }
func (cartStub) MergeCarts(context.Context, uuid.UUID, uuid.UUID) (*dto.CartDTO, error) {
	return nil, nil
}
func (cartStub) RawCart(context.Context, dto.Owner) (*models.Cart, error) { return nil, nil }

func TestPreviewCheckoutFlatShipping(t *testing.T) {
	s := &Service{
		cfg:          &config.Config{},
		cart:         cartStub{items: []dto.CartItemsDTO{{Available: true, Quantity: 2, Price: decOf("500"), WeightKG: decOf("1")}}},
		flatShipping: decOf("300"),
	}

	res, err := s.PreviewCheckout(context.Background(), dto.CheckoutPreviewDTO{})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Subtotal.Equal(decOf("1000")) || !res.ShippingTotal.Equal(decOf("300")) || !res.GrandTotal.Equal(decOf("1300")) {
		t.Fatalf("totals: subtotal=%s shipping=%s grand=%s", res.Subtotal, res.ShippingTotal, res.GrandTotal)
	}
	if res.ItemCount != 2 {
		t.Fatalf("item_count = %d, want 2", res.ItemCount)
	}
}

func TestPreviewCheckoutCarrierRequote(t *testing.T) {
	reg := shipping.NewRegistry(nil, 0, shipping.ParcelDefaults{}, nil,
		rateStubProvider{code: "cdek", rates: []dto.ShippingRate{
			{Provider: "cdek", TariffCode: "136", TariffName: "Склад-склад", Cost: decOf("450"), MinDays: 1, MaxDays: 3},
		}},
	)
	s := &Service{
		cfg:          &config.Config{},
		cart:         cartStub{items: []dto.CartItemsDTO{{Available: true, Quantity: 1, Price: decOf("2000"), WeightKG: decOf("2")}}},
		shipping:     reg,
		flatShipping: decOf("300"),
	}

	provider, tariff := "cdek", "136"
	res, err := s.PreviewCheckout(context.Background(), dto.CheckoutPreviewDTO{
		Shipping: dto.ShippingSelection{Provider: &provider, TariffCode: &tariff, Postcode: strp("190000")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.ShippingTotal.Equal(decOf("450")) || !res.GrandTotal.Equal(decOf("2450")) {
		t.Fatalf("shipping=%s grand=%s, want 450 / 2450", res.ShippingTotal, res.GrandTotal)
	}
	if res.Shipping.Provider == nil || *res.Shipping.Provider != "cdek" || res.Shipping.MinDays == nil || *res.Shipping.MinDays != 1 {
		t.Fatalf("shipping snapshot: %+v", res.Shipping)
	}
}

func TestPreviewCheckoutEmptyCart(t *testing.T) {
	s := &Service{cfg: &config.Config{}, cart: cartStub{}}
	if _, err := s.PreviewCheckout(context.Background(), dto.CheckoutPreviewDTO{}); err == nil {
		t.Fatal("expected ErrCartEmpty")
	}
}
